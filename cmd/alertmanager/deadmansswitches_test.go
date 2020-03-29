package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/function61/eventhorizon/pkg/ehclient"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/eventhorizon/pkg/ehreader"
	"github.com/function61/eventhorizon/pkg/ehreader/ehreadertest"
	"github.com/function61/gokit/assert"
	"github.com/function61/lambda-alertmanager/pkg/amdomain"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
)

func TestDeadmansswitchCheckin(t *testing.T) {
	ctx := context.Background()

	testStreamName := "/t-42/alertmanager"

	eventLog := ehreadertest.NewEventLog()
	// not significant (we just need an event in the log for initial read to work)
	// TODO: add CreateStream() on the testing log?
	eventLog.AppendE(
		testStreamName,
		amdomain.NewUnnoticedAlertsNotified(
			[]string{"dummyid"},
			ehevent.MetaSystemUser(t0)))

	app, err := amstate.LoadUntilRealtime(
		ctx,
		ehreader.NewTenantCtxWithSnapshots(
			ehreader.TenantId("42"),
			eventLog,
			ehreader.NewInMemSnapshotStore()),
		nil)
	assert.Ok(t, err)

	alertAcked, err := deadmansswitchCheckin(
		ctx,
		"My test switch",
		t0.Add(1*time.Hour),
		app,
		t0)
	assert.Ok(t, err)
	assert.Assert(t, !alertAcked)

	dumper := newEventDumper(testStreamName, eventLog, amdomain.Types)

	assert.EqualString(t, dumper.Dump(), `
2019-09-07T12:00:00.000Z UnnoticedAlertsNotified    {"AlertIds":["dummyid"]}
2019-09-07T12:00:00.000Z DeadMansSwitchCreated    {"Subject":"My test switch","Ttl":"2019-09-07T13:00:00Z"}
2019-09-07T12:00:00.000Z DeadMansSwitchCheckin    {"Subject":"My test switch","Ttl":"2019-09-07T13:00:00Z"}`)

	// 30 minutes lapses, and we send another checkin

	alertAcked, err = deadmansswitchCheckin(
		ctx,
		"My test switch",
		t0.Add(90*time.Minute),
		app,
		t0.Add(30*time.Minute))
	assert.Ok(t, err)
	assert.Assert(t, !alertAcked)

	assert.EqualString(t, dumper.Dump(), `
2019-09-07T12:00:00.000Z UnnoticedAlertsNotified    {"AlertIds":["dummyid"]}
2019-09-07T12:00:00.000Z DeadMansSwitchCreated    {"Subject":"My test switch","Ttl":"2019-09-07T13:00:00Z"}
2019-09-07T12:00:00.000Z DeadMansSwitchCheckin    {"Subject":"My test switch","Ttl":"2019-09-07T13:00:00Z"}
2019-09-07T12:30:00.000Z DeadMansSwitchCheckin    {"Subject":"My test switch","Ttl":"2019-09-07T13:30:00Z"}`)
}

func newEventDumper(stream string, eventLog ehclient.Reader, types ehevent.Allocators) *eventDumper {
	d := &eventDumper{
		cur:   ehclient.Beginning(stream),
		dump:  []string{""}, // to get newline at beginning
		types: types,
	}
	d.reader = ehreader.New(d, eventLog, nil)
	return d
}

// For testing, allows you to print your event log
// TODO: move this to eventhorizon testing helper package?
type eventDumper struct {
	cur    ehclient.Cursor
	reader *ehreader.Reader
	dump   []string
	types  ehevent.Allocators
}

func (e *eventDumper) Dump() string {
	if err := e.reader.LoadUntilRealtime(context.Background()); err != nil {
		panic(err)
	}

	return strings.Join(e.dump, "\n")
}

func (e *eventDumper) GetEventTypes() ehevent.Allocators {
	return e.types
}

func (e *eventDumper) ProcessEvents(ctx context.Context, handle ehreader.EventProcessorHandler) error {
	return handle(
		e.cur,
		func(ev ehevent.Event) error {
			e.dump = append(e.dump, ehevent.Serialize(ev))
			return nil
		},
		func(commitCursor ehclient.Cursor) error {
			e.cur = commitCursor
			return nil
		})
}
