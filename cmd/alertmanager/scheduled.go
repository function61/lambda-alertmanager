package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/lambda-alertmanager/pkg/amdomain"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// runs every minute
func handleCloudwatchScheduledEvent(ctx context.Context, now time.Time) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	if err := checkAndAlertForUnnoticedAlerts(ctx, app, publishAlert, now); err != nil {
		return err
	}

	if err := alertForDeadExpiredMansSwitches(ctx, app, now); err != nil {
		return err
	}

	if err := httpMonitorScanAndAlertFailures(ctx, app); err != nil {
		return err
	}

	return nil
}

type alertDirectPublisherFn func(amstate.Alert) error

// check for old unnoticed alerts (not acked within 4 hours) and send an alarm to notify the operator,
// keep sending every hour
func checkAndAlertForUnnoticedAlerts(
	ctx context.Context,
	app *amstate.App,
	alertDirectPublisher alertDirectPublisherFn,
	now time.Time,
) error {
	unnoticedAlerts := amstate.GetUnnoticedAlerts(app.State.ActiveAlerts(), now)

	unnoticedAlertSubjects := []string{}
	unnoticedAlertIds := []string{}

	for _, alert := range unnoticedAlerts {
		unnoticedAlertSubjects = append(unnoticedAlertSubjects, alert.Subject+" "+ackLink(alert))
		unnoticedAlertIds = append(unnoticedAlertIds, alert.Id)
	}

	if len(unnoticedAlertIds) == 0 {
		return nil
	}

	errAlreadyNotified := errors.New("already notified")

	if err := app.Reader.TransactWrite(ctx, func() error {
		// only notify about unnoticed alerts once an hour
		if now.Sub(app.State.LastUnnoticedAlertsNotified()) < 1*time.Hour {
			return errAlreadyNotified
		}

		return app.AppendAfter(ctx, app.State.Version(), amdomain.NewUnnoticedAlertsNotified(
			unnoticedAlertIds,
			ehevent.MetaSystemUser(now)))
	}); err != nil {
		if err == errAlreadyNotified {
			return nil // not actually an error
		} else {
			return err
		}
	}

	details := fmt.Sprintf(
		"There are %d un-acked alert(s):\n\n%s\n\nGo take care of them!",
		len(unnoticedAlertSubjects),
		strings.Join(unnoticedAlertSubjects, "\n"))

	// skip ingestion to bypass rate limiting (this scheduled function is not invoked
	// too often) and deduplication. besides, we want to keep reminding the operator
	// to take care of this situation
	return alertDirectPublisher(amstate.Alert{
		Subject:   "Un-acked alerts",
		Details:   details,
		Timestamp: now,
	})
}

func alertForDeadExpiredMansSwitches(ctx context.Context, app *amstate.App, now time.Time) error {
	candidateAlerts := []amstate.Alert{}

	for _, dms := range amstate.GetExpiredDeadMansSwitches(app.State.DeadMansSwitches(), now) {
		candidateAlerts = append(candidateAlerts, deadMansSwitchToAlert(dms, now))
	}

	// ok with len(alerts) == 0
	return ingestAlerts(ctx, candidateAlerts, app)
}

var plusDayAtStaticTimeRe = regexp.MustCompile(`^\+([0-9]+)d@([0-9]{2}):([0-9]{2})$`)

func parseTtlSpec(spec string, now time.Time) (time.Time, error) {
	// +1d@12:00
	if match := plusDayAtStaticTimeRe.FindStringSubmatch(spec); match != nil {
		// below errors should never trigger because regexp guarantees they're in good format
		day, err := strconv.Atoi(match[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("bad day component: %v", err)
		}
		hour, err := strconv.Atoi(match[2])
		if err != nil {
			return time.Time{}, fmt.Errorf("bad hour component: %v", err)
		}
		minute, err := strconv.Atoi(match[3])
		if err != nil {
			return time.Time{}, fmt.Errorf("bad minute component: %v", err)
		}

		return time.Date(now.Year(), now.Month(), now.Day()+day, hour, minute, 0, 0, time.UTC), nil
	} else if strings.HasPrefix(spec, "+") { // +24h
		duration, err := time.ParseDuration(spec[1:])
		if err != nil {
			return time.Time{}, errors.New("duration in bad format")
		}

		return now.Add(duration), nil
	} else {
		ttl, err := time.Parse(time.RFC3339Nano, spec)
		if err != nil {
			err = fmt.Errorf("not in RFC3339: %s", spec)
		}
		return ttl, err
	}
}

func deadMansSwitchToAlert(dms amstate.DeadMansSwitch, now time.Time) amstate.Alert {
	return amstate.Alert{
		Id:        amstate.NewAlertId(),
		Subject:   dms.Subject,
		Details:   fmt.Sprintf("Check-in late by %s (%s)", now.Sub(dms.Ttl), dms.Ttl.Format(time.RFC3339Nano)),
		Timestamp: now,
	}
}
