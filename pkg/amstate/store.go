package amstate

import (
	"context"
	"encoding/json"
	"github.com/function61/eventhorizon/pkg/ehclient"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/eventhorizon/pkg/ehreader"
	"github.com/function61/gokit/logex"
	"github.com/function61/lambda-alertmanager/pkg/amdomain"
	"log"
	"sort"
	"sync"
	"time"
)

const (
	Stream = "/alertmanager"
)

func newStateFormat() stateFormat {
	return stateFormat{
		ActiveAlerts:     map[string]Alert{},
		HttpMonitors:     map[string]HttpMonitor{},
		DeadMansSwitches: map[string]DeadMansSwitch{},
	}
}

type Store struct {
	version ehclient.Cursor
	mu      sync.Mutex
	state   stateFormat // for easy snapshotting
	logl    *logex.Leveled
}

func New(tenant ehreader.Tenant, logger *log.Logger) *Store {
	return &Store{
		version: ehclient.Beginning(tenant.Stream(Stream)),
		state:   newStateFormat(),
		logl:    logex.Levels(logger),
	}
}

func (s *Store) Version() ehclient.Cursor {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.version
}

func (s *Store) InstallSnapshot(snap *ehreader.Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.version = snap.Cursor
	s.state = stateFormat{}

	return json.Unmarshal(snap.Data, &s.state)
}

func (s *Store) Snapshot() (*ehreader.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.state, "", "\t")
	if err != nil {
		return nil, err
	}

	return ehreader.NewSnapshot(s.version, data), nil
}

func (s *Store) ActiveAlerts() []Alert {
	s.mu.Lock()
	defer s.mu.Unlock()

	alerts := []Alert{}
	for _, alert := range s.state.ActiveAlerts {
		alerts = append(alerts, alert)
	}

	sort.Slice(alerts, func(i, j int) bool { return alerts[i].Timestamp.Before(alerts[j].Timestamp) })

	return alerts
}

func (s *Store) HttpMonitors() []HttpMonitor {
	s.mu.Lock()
	defer s.mu.Unlock()

	monitors := []HttpMonitor{}
	for _, alert := range s.state.HttpMonitors {
		monitors = append(monitors, alert)
	}

	sort.Slice(monitors, func(i, j int) bool { return monitors[i].Created.Before(monitors[j].Created) })

	return monitors
}

func (s *Store) DeadMansSwitches() []DeadMansSwitch {
	s.mu.Lock()
	defer s.mu.Unlock()

	deadMansSwitches := []DeadMansSwitch{}
	for _, dms := range s.state.DeadMansSwitches {
		deadMansSwitches = append(deadMansSwitches, dms)
	}

	sort.Slice(deadMansSwitches, func(i, j int) bool {
		return deadMansSwitches[i].Subject < deadMansSwitches[j].Subject
	})

	return deadMansSwitches
}

func (s *Store) LastUnnoticedAlertsNotified() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.state.LastUnnoticedAlertsNotified
}

func (s *Store) GetEventTypes() ehevent.Allocators {
	return amdomain.Types
}

func (s *Store) ProcessEvents(_ context.Context, processAndCommit ehreader.EventProcessorHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return processAndCommit(
		s.version,
		func(ev ehevent.Event) error { return s.processEvent(ev) },
		func(version ehclient.Cursor) error {
			s.version = version
			return nil
		})
}

func (s *Store) processEvent(ev ehevent.Event) error {
	s.logl.Debug.Println(ev.MetaType())

	switch e := ev.(type) {
	case *amdomain.AlertRaised:
		s.state.ActiveAlerts[e.Id] = Alert{
			Id:        e.Id,
			Subject:   e.Subject,
			Details:   e.Details,
			Timestamp: e.Meta().Timestamp,
		}
	case *amdomain.AlertAcknowledged:
		delete(s.state.ActiveAlerts, e.Id)
	case *amdomain.HttpMonitorCreated:
		s.state.HttpMonitors[e.Id] = HttpMonitor{
			Id:      e.Id,
			Created: e.Meta().Timestamp,
			Enabled: e.Enabled,
			Url:     e.Url,
			Find:    e.Find,
		}
	case *amdomain.HttpMonitorEnabledUpdated:
		mon := s.state.HttpMonitors[e.Id]
		mon.Enabled = e.Enabled
		s.state.HttpMonitors[e.Id] = mon
	case *amdomain.HttpMonitorDeleted:
		delete(s.state.HttpMonitors, e.Id)
	case *amdomain.DeadMansSwitchCreated:
		s.state.DeadMansSwitches[e.Subject] = DeadMansSwitch{
			Subject: e.Subject,
			Ttl:     e.Ttl,
		}
	case *amdomain.DeadMansSwitchCheckin:
		dms := s.state.DeadMansSwitches[e.Subject]
		dms.Ttl = e.Ttl
		s.state.DeadMansSwitches[e.Subject] = dms
	case *amdomain.DeadMansSwitchDeleted:
		delete(s.state.DeadMansSwitches, e.Subject)
	case *amdomain.UnnoticedAlertsNotified:
		s.state.LastUnnoticedAlertsNotified = e.Meta().Timestamp
	default:
		return ehreader.UnsupportedEventTypeErr(ev)
	}

	return nil
}

type App struct {
	State  *Store
	Reader *ehreader.Reader
	Writer ehclient.Writer
	Logger *log.Logger
}

// helper
func (a *App) AppendAfter(ctx context.Context, cur ehclient.Cursor, events ...ehevent.Event) error {
	serialized := []string{}
	for _, event := range events {
		serialized = append(serialized, ehevent.Serialize(event))
	}

	// helper mainly written b/c we don't care for returned cursor
	_, err := a.Writer.AppendAfter(ctx, cur, serialized)
	return err
}

func LoadUntilRealtime(
	ctx context.Context,
	tenantCtx *ehreader.TenantCtxWithSnapshots,
	logger *log.Logger,
) (*App, error) {
	store := New(tenantCtx.Tenant, logger)

	a := &App{
		store,
		ehreader.NewWithSnapshots(
			store,
			tenantCtx.Client,
			tenantCtx.SnapshotStore,
			logger),
		tenantCtx.Client,
		logger}

	if err := a.Reader.LoadUntilRealtime(ctx); err != nil {
		return nil, err
	}

	return a, nil
}
