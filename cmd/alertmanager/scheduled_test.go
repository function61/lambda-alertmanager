package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/eventhorizon/pkg/ehreader"
	"github.com/function61/eventhorizon/pkg/ehreader/ehreadertest"
	"github.com/function61/gokit/assert"
	"github.com/function61/lambda-alertmanager/pkg/amdomain"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"os"
	"testing"
	"time"
)

var t0 = time.Date(2019, 9, 7, 12, 0, 0, 0, time.UTC)

func TestCheckAndAlertForUnnoticedAlerts(t *testing.T) {
	ctx := context.Background()

	testStreamName := "/t-42/alertmanager"

	// have an alert raised at T+0
	eventLog := ehreadertest.NewEventLog()
	eventLog.AppendE(
		testStreamName,
		amdomain.NewAlertRaised(
			"a14308bba82f",
			"The building is on fire",
			"Fire sensor in room 456 went off",
			ehevent.MetaSystemUser(t0)))

	app, err := amstate.LoadUntilRealtime(
		ctx,
		ehreader.NewTenantCtxWithSnapshots(
			ehreader.TenantId("42"),
			eventLog,
			ehreader.NewInMemSnapshotStore()),
		nil)
	assert.Ok(t, err)

	// lame hack to get ackLink() to produce URLs while in testing
	os.Setenv("API_ENDPOINT", "https://alertmanager.com/api")

	unnoticedAlertAtT0Plus := func(plus time.Duration) string { // helper
		var publishedAlert *amstate.Alert

		// capture maybe-published alert
		assert.Ok(t, checkAndAlertForUnnoticedAlerts(ctx, app, func(alert amstate.Alert) error {
			publishedAlert = &alert

			return nil
		}, t0.Add(plus)))

		publishedAlertAsJson, err := json.MarshalIndent(publishedAlert, "", "  ")
		assert.Ok(t, err)

		return string(publishedAlertAsJson)
	}

	/*	we want alert at 4 hour mark when alert is un-acked, then *every* hour again:

		0:00 (an alert is raised)
		1:00
		2:00
		3:00
		4:00 => first "unnoticed" alert
		4:30
		5:00 => second "unnoticed" alert
		...
	*/
	assert.EqualString(t, unnoticedAlertAtT0Plus(0*time.Hour), "null")
	assert.EqualString(t, unnoticedAlertAtT0Plus(1*time.Hour), "null")
	assert.EqualString(t, unnoticedAlertAtT0Plus(2*time.Hour), "null")
	assert.EqualString(t, unnoticedAlertAtT0Plus(3*time.Hour), "null")
	assert.EqualString(t, unnoticedAlertAtT0Plus(4*time.Hour), `{
  "alert_key": "",
  "subject": "Un-acked alerts",
  "details": "There are 1 un-acked alert(s):\n\nThe building is on fire https://alertmanager.com/api/alerts/acknowledge?id=a14308bba82f\n\nGo take care of them!",
  "timestamp": "2019-09-07T16:00:00Z"
}`)
	assert.EqualString(t, unnoticedAlertAtT0Plus(4*time.Hour+30*time.Minute), "null")
	assert.EqualString(t, unnoticedAlertAtT0Plus(5*time.Hour), `{
  "alert_key": "",
  "subject": "Un-acked alerts",
  "details": "There are 1 un-acked alert(s):\n\nThe building is on fire https://alertmanager.com/api/alerts/acknowledge?id=a14308bba82f\n\nGo take care of them!",
  "timestamp": "2019-09-07T17:00:00Z"
}`)
	assert.EqualString(t, unnoticedAlertAtT0Plus(5*time.Hour+30*time.Minute), "null")
	assert.EqualString(t, unnoticedAlertAtT0Plus(6*time.Hour), `{
  "alert_key": "",
  "subject": "Un-acked alerts",
  "details": "There are 1 un-acked alert(s):\n\nThe building is on fire https://alertmanager.com/api/alerts/acknowledge?id=a14308bba82f\n\nGo take care of them!",
  "timestamp": "2019-09-07T18:00:00Z"
}`)
}

func TestParseTtlSpec(t *testing.T) {
	tcs := []struct {
		input  string
		output string
	}{
		{
			"+24h",
			"2019-09-08T12:00:00Z",
		},
		{
			"+1h",
			"2019-09-07T13:00:00Z",
		},
		{
			"+1d@18:00",
			"2019-09-08T18:00:00Z",
		},
		{
			"+14d@10:00",
			"2019-09-21T10:00:00Z",
		},
		{
			"2019-09-10T01:13:00Z",
			"2019-09-10T01:13:00Z",
		},
		{
			"foobar",
			"error: not in RFC3339: foobar",
		},
		{
			"+12x",
			"error: duration in bad format",
		},
		{
			"+1d@18:0x",
			"error: duration in bad format",
		},
	}

	for _, tc := range tcs {
		tc := tc // pin
		t.Run(tc.input, func(t *testing.T) {
			ttl, err := parseTtlSpec(tc.input, t0)
			var actual string
			if err != nil {
				actual = fmt.Sprintf("error: %v", err)
			} else {
				actual = ttl.Format(time.RFC3339Nano)
			}

			assert.EqualString(t, actual, tc.output)
		})
	}
}
