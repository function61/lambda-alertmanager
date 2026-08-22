package amstate

import (
	"time"

	"github.com/function61/gokit/crypto/cryptoutil"
)

func FindAlertWithSubject(subject string, alerts []Alert) *Alert {
	for _, alert := range alerts {
		if alert.Subject == subject {
			return &alert
		}
	}

	return nil
}

func HasAlertWithID(id string, alerts []Alert) bool {
	for _, alert := range alerts {
		if alert.ID == id {
			return true
		}
	}

	return false
}

// unnoticed = not acked within 4 hours
func GetUnnoticedAlerts(alerts []Alert, now time.Time) []Alert {
	unnoticed := []Alert{}
	for _, alert := range alerts {
		if now.Sub(alert.Timestamp) >= 4*time.Hour {
			unnoticed = append(unnoticed, alert)
		}
	}

	return unnoticed
}

func FindHTTPMonitorWithID(id string, monitors []HTTPMonitor) *HTTPMonitor {
	for _, monitor := range monitors {
		if monitor.ID == id {
			return &monitor
		}
	}

	return nil
}

func EnabledHTTPMonitors(monitors []HTTPMonitor) []HTTPMonitor {
	enabled := []HTTPMonitor{}

	for _, monitor := range monitors {
		if monitor.Enabled {
			enabled = append(enabled, monitor)
		}
	}

	return enabled
}

func FindDeadMansSwitchWithSubject(subject string, dmss []DeadMansSwitch) *DeadMansSwitch {
	for _, dms := range dmss {
		if dms.Subject == subject {
			return &dms
		}
	}

	return nil
}

func GetExpiredDeadMansSwitches(switches []DeadMansSwitch, now time.Time) []DeadMansSwitch {
	expired := []DeadMansSwitch{}
	for _, sw := range switches {
		if !now.Before(sw.TTL) {
			expired = append(expired, sw)
		}
	}

	return expired
}

func NewAlertID() string {
	return cryptoutil.RandBase64UrlWithoutLeadingDash(6)
}

func NewHTTPMonitorID() string {
	return cryptoutil.RandBase64UrlWithoutLeadingDash(6)
}
