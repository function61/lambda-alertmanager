package main

// Ingesting is the act of taking alert from some system, doing deduplication and alert
// limiting (only N amount of active alerts are allowed). We either accept or drop the alert.

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/gokit/logex"
	"github.com/function61/lambda-alertmanager/pkg/amdomain"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"os"
	"strconv"
)

// invoked for "AlertManager-ingest" SNS topic
func handleSnsIngest(ctx context.Context, event events.SNSEvent) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	candidateAlerts := []amstate.Alert{}

	for _, msg := range event.Records {
		candidateAlerts = append(candidateAlerts, amstate.Alert{
			Id:        amstate.NewAlertId(),
			Subject:   msg.SNS.Subject,
			Details:   msg.SNS.Message,
			Timestamp: msg.SNS.Timestamp,
		})
	}

	_, err = ingestAlerts(ctx, candidateAlerts, app)
	return err
}

// this is somewhat of a hack to pass candidate-phase alerts as the same struct as we get
// from the actual persisted State
func ingestAlerts(ctx context.Context, candidateAlerts []amstate.Alert, app *amstate.App) (bool, error) {
	ingestedAny := false

	maxActiveAlerts, err := getMaxFiringAlerts()
	if err != nil {
		return false, err
	}

	// this call is free (unless we actually call Append()), so no reason to optimize by
	// checking for alert length
	if err := app.Reader.TransactWrite(ctx, func() error {
		alertEvents := []ehevent.Event{}

		alerts := deduplicateAndRatelimit(candidateAlerts, app.State, maxActiveAlerts)

		// raise alerts for failures
		for _, alert := range alerts {
			alertEvents = append(alertEvents, amdomain.NewAlertRaised(
				alert.Id,
				alert.Subject,
				alert.Details,
				ehevent.MetaSystemUser(alert.Timestamp)))
		}

		if len(alertEvents) == 0 {
			return nil // nothing to do
		}

		if err := app.AppendAfter(ctx, app.State.Version(), alertEvents...); err != nil {
			return err
		}

		ingestedAny = true

		for _, alert := range alerts {
			if err := publishAlert(alert); err != nil {
				logex.Levels(app.Logger).Error.Printf("publishAlert: %v", err)
			}
		}

		return nil
	}); err != nil {
		return ingestedAny, err
	}

	return ingestedAny, nil
}

func deduplicateAndRatelimit(
	alerts []amstate.Alert,
	state *amstate.Store,
	maxActiveAlerts int,
) []amstate.Alert {
	processed := []amstate.Alert{}

	activeAlerts := state.ActiveAlerts()

	for _, alert := range alerts {
		// no more "room"?
		if len(activeAlerts) >= maxActiveAlerts {
			continue
		}

		// deduplication
		if amstate.FindAlertWithSubject(alert.Subject, activeAlerts) != nil {
			continue
		}

		processed = append(processed, alert)
	}

	return processed
}

func getMaxFiringAlerts() (int, error) {
	fromEnvStr := os.Getenv("MAX_FIRING_ALERTS")
	if fromEnvStr == "" {
		return 5, nil // default
	}

	return strconv.Atoi(fromEnvStr)
}
