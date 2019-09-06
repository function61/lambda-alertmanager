package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/alertmanagertypes"
	"time"
)

// every hour or so
func handleCloudwatchScheduledEvent(ctx context.Context, event events.CloudWatchEvent) error {
	now := event.Time // not sure if this is worth it

	if err := checkForUnAckedAlerts(now); err != nil {
		return err
	}

	return nil
}

// check for old un-acked alerts and if there is, send an alarm to notify the operator
func checkForUnAckedAlerts(now time.Time) error {
	alerts, err := getAlerts()
	if err != nil {
		return err
	}

	oldAlertCount := 0

	for _, alert := range alerts {
		if now.Sub(alert.Timestamp) > 4*time.Hour {
			oldAlertCount++
		}
	}

	if oldAlertCount > 0 {
		// skip ingestion to bypass rate limiting (this scheduled function is not invoked
		// too often) and deduplication. besides, we want to keep reminding the operator
		// to take care of this situation
		return publishAlert(alertmanagertypes.Alert{
			Key:       "", // = not stored in the stateful store, comment above explains why
			Subject:   "Un-acked alerts",
			Details:   fmt.Sprintf("There are %d un-acked alert(s). Go take care of them!", oldAlertCount),
			Timestamp: now,
		})
	}

	return nil
}
