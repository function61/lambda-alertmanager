package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/alertmanagertypes"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// every hour or so
func handleCloudwatchScheduledEvent(ctx context.Context, event events.CloudWatchEvent) error {
	now := event.Time // not sure if this is worth it

	if err := checkForUnAckedAlerts(now); err != nil {
		return err
	}

	if err := checkForDeadMansSwitches(now); err != nil {
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

	oldAlertSubjects := []string{}

	for _, alert := range alerts {
		if now.Sub(alert.Timestamp) > 4*time.Hour {
			oldAlertSubjects = append(oldAlertSubjects, alert.Subject+" "+ackLink(alert))
		}
	}

	if len(oldAlertSubjects) > 0 {
		details := fmt.Sprintf(
			"There are %d un-acked alert(s):\n\n%s\n\nGo take care of them!",
			len(oldAlertSubjects),
			strings.Join(oldAlertSubjects, "\n"))

		// skip ingestion to bypass rate limiting (this scheduled function is not invoked
		// too often) and deduplication. besides, we want to keep reminding the operator
		// to take care of this situation
		return publishAlert(alertmanagertypes.Alert{
			Key:       "", // = not stored in the stateful store, comment above explains why
			Subject:   "Un-acked alerts",
			Details:   details,
			Timestamp: now,
		})
	}

	return nil
}

func checkForDeadMansSwitches(now time.Time) error {
	switches, err := getDeadMansSwitches()
	if err != nil {
		return err
	}

	overdue := getOverdueSwitches(switches, now)

	if len(overdue) == 0 {
		return nil
	}

	for _, overdueSwitch := range overdue {
		alert := alertmanagertypes.Alert{
			Subject:   overdueSwitch.Subject,
			Timestamp: now,
			Details:   fmt.Sprintf("Check-in late by %s (%s)", now.Sub(overdueSwitch.TTL), overdueSwitch.TTL.Format(time.RFC3339Nano)),
		}

		if _, err := ingestAlert(alert); err != nil {
			return err
		}
	}

	return nil
}

func getOverdueSwitches(switches []alertmanagertypes.DeadMansSwitch, now time.Time) []alertmanagertypes.DeadMansSwitch {
	ret := []alertmanagertypes.DeadMansSwitch{}

	for _, sw := range switches {
		if now.Before(sw.TTL) {
			continue
		}

		ret = append(ret, sw)
	}

	return ret
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
