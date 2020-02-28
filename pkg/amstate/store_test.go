package amstate

import (
	"context"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/eventhorizon/pkg/ehreader"
	"github.com/function61/eventhorizon/pkg/ehreader/ehreadertest"
	"github.com/function61/gokit/assert"
	"github.com/function61/lambda-alertmanager/pkg/amdomain"
	"testing"
	"time"
)

const (
	testStreamName = "/t-42/alertmanager"
)

var (
	t0 = time.Date(2020, 2, 20, 14, 2, 0, 0, time.UTC)
)

func TestAlerts(t *testing.T) {
	ctx := context.Background()

	eventLog := ehreadertest.NewEventLog()
	eventLog.AppendE(
		testStreamName,
		amdomain.NewAlertRaised(
			"a14308bba82f",
			"The building is on fire",
			"Fire sensor in room 456 went off",
			ehevent.MetaSystemUser(t0)))
	eventLog.AppendE(
		testStreamName,
		amdomain.NewAlertRaised(
			"1a33032c9081",
			"Water damage detected",
			"Water leak sensor in room 456 went off",
			ehevent.MetaSystemUser(t0.Add(2*time.Minute))))

	app, err := LoadUntilRealtime(ctx, ehreader.NewTenantCtxWithSnapshots(ehreader.TenantId("42"), eventLog, ehreader.NewInMemSnapshotStore()), nil)
	assert.Ok(t, err)

	assert.Assert(t, FindAlertWithSubject("The building is on fire", app.State.ActiveAlerts()) != nil)
	assert.Assert(t, FindAlertWithSubject("Everything is calm", app.State.ActiveAlerts()) == nil)

	alerts := app.State.ActiveAlerts()

	assert.EqualJson(t, alerts, `[
  {
    "alert_key": "a14308bba82f",
    "subject": "The building is on fire",
    "details": "Fire sensor in room 456 went off",
    "timestamp": "2020-02-20T14:02:00Z"
  },
  {
    "alert_key": "1a33032c9081",
    "subject": "Water damage detected",
    "details": "Water leak sensor in room 456 went off",
    "timestamp": "2020-02-20T14:04:00Z"
  }
]`)

	eventLog.AppendE(
		testStreamName,
		amdomain.NewAlertAcknowledged(
			"a14308bba82f",
			ehevent.MetaSystemUser(t0.Add(1*time.Hour))))

	assert.Ok(t, app.Reader.LoadUntilRealtime(ctx))

	assert.Assert(t, len(app.State.ActiveAlerts()) == 1)
}

func TestHttpMonitors(t *testing.T) {
	ctx := context.Background()

	eventLog := ehreadertest.NewEventLog()
	eventLog.AppendE(
		testStreamName,
		amdomain.NewHttpMonitorCreated(
			"49365a17244e",
			true,
			"https://function61.com/",
			"Welcome to the best page in the universe",
			ehevent.MetaSystemUser(t0)))

	app, err := LoadUntilRealtime(ctx, ehreader.NewTenantCtxWithSnapshots(ehreader.TenantId("42"), eventLog, ehreader.NewInMemSnapshotStore()), nil)
	assert.Ok(t, err)

	assert.EqualJson(t, app.State.HttpMonitors()[0], `{
  "id": "49365a17244e",
  "created": "2020-02-20T14:02:00Z",
  "enabled": true,
  "url": "https://function61.com/",
  "find": "Welcome to the best page in the universe"
}`)

	eventLog.AppendE(
		testStreamName,
		amdomain.NewHttpMonitorEnabledUpdated(
			"49365a17244e",
			false,
			ehevent.MetaSystemUser(t0)))

	assert.Ok(t, app.Reader.LoadUntilRealtime(ctx))

	assert.Assert(t, !app.State.HttpMonitors()[0].Enabled)

	eventLog.AppendE(
		testStreamName,
		amdomain.NewHttpMonitorDeleted(
			"49365a17244e",
			ehevent.MetaSystemUser(t0)))

	assert.Ok(t, app.Reader.LoadUntilRealtime(ctx))

	assert.Assert(t, len(app.State.HttpMonitors()) == 0)
}

func TestDeadMansSwitches(t *testing.T) {
	ctx := context.Background()

	eventLog := ehreadertest.NewEventLog()
	eventLog.AppendE(
		testStreamName,
		amdomain.NewDeadMansSwitchCreated(
			"Joonas checkins",
			t0.Add(2*time.Hour),
			ehevent.MetaSystemUser(t0)))

	app, err := LoadUntilRealtime(ctx, ehreader.NewTenantCtxWithSnapshots(ehreader.TenantId("42"), eventLog, ehreader.NewInMemSnapshotStore()), nil)
	assert.Ok(t, err)

	assert.EqualJson(t, app.State.DeadMansSwitches(), `[
  {
    "subject": "Joonas checkins",
    "ttl": "2020-02-20T16:02:00Z"
  }
]`)

	eventLog.AppendE(
		testStreamName,
		amdomain.NewDeadMansSwitchCheckin(
			"Joonas checkins",
			t0.Add(3*time.Hour),
			ehevent.MetaSystemUser(t0)))

	assert.Ok(t, app.Reader.LoadUntilRealtime(ctx))

	assert.EqualJson(t, app.State.DeadMansSwitches(), `[
  {
    "subject": "Joonas checkins",
    "ttl": "2020-02-20T17:02:00Z"
  }
]`)

	switches := app.State.DeadMansSwitches()

	assert.Assert(t, len(GetExpiredDeadMansSwitches(switches, t0.Add(1*time.Hour))) == 0)
	assert.Assert(t, len(GetExpiredDeadMansSwitches(switches, t0.Add(2*time.Hour))) == 0)
	assert.Assert(t, len(GetExpiredDeadMansSwitches(switches, t0.Add(3*time.Hour))) == 1)

	eventLog.AppendE(
		testStreamName,
		amdomain.NewDeadMansSwitchDeleted(
			"Joonas checkins",
			ehevent.MetaSystemUser(t0)))

	assert.Ok(t, app.Reader.LoadUntilRealtime(ctx))

	assert.EqualJson(t, app.State.DeadMansSwitches(), `[]`)
}

func TestGetUnnoticedAlerts(t *testing.T) {
	ctx := context.Background()

	eventLog := ehreadertest.NewEventLog()
	eventLog.AppendE(
		testStreamName,
		amdomain.NewAlertRaised(
			"a14308bba82f",
			"The building is on fire",
			"Fire sensor in room 456 went off",
			ehevent.MetaSystemUser(t0)))

	app, err := LoadUntilRealtime(
		ctx,
		ehreader.NewTenantCtxWithSnapshots(
			ehreader.TenantId("42"),
			eventLog,
			ehreader.NewInMemSnapshotStore()),
		nil)
	assert.Ok(t, err)

	unnoticedAlertCountT0Plus := func(plus time.Duration) int {
		return len(GetUnnoticedAlerts(app.State.ActiveAlerts(), t0.Add(plus)))
	}

	assert.Assert(t, unnoticedAlertCountT0Plus(0*time.Hour) == 0)
	assert.Assert(t, unnoticedAlertCountT0Plus(3*time.Hour) == 0)
	assert.Assert(t, unnoticedAlertCountT0Plus(4*time.Hour) == 1)

	assert.EqualJson(t, app.State.LastUnnoticedAlertsNotified(), `"0001-01-01T00:00:00Z"`)

	eventLog.AppendE(testStreamName, amdomain.NewUnnoticedAlertsNotified(
		[]string{"a14308bba82f"},
		ehevent.MetaSystemUser(t0)))

	assert.Ok(t, app.Reader.LoadUntilRealtime(ctx))

	assert.EqualJson(t, app.State.LastUnnoticedAlertsNotified(), `"2020-02-20T14:02:00Z"`)
}
