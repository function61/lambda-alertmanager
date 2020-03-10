package amstate

import (
	"github.com/function61/gokit/cryptorandombytes"
	"time"
)

func FindAlertWithSubject(subject string, alerts []Alert) *Alert {
	for _, alert := range alerts {
		if alert.Subject == subject {
			return &alert
		}
	}

	return nil
}

func HasAlertWithId(id string, alerts []Alert) bool {
	for _, alert := range alerts {
		if alert.Id == id {
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

func FindHttpMonitorWithId(id string, monitors []HttpMonitor) *HttpMonitor {
	for _, monitor := range monitors {
		if monitor.Id == id {
			return &monitor
		}
	}

	return nil
}

func EnabledHttpMonitors(monitors []HttpMonitor) []HttpMonitor {
	enabled := []HttpMonitor{}

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
		if !now.Before(sw.Ttl) {
			expired = append(expired, sw)
		}
	}

	return expired
}

func NewAlertId() string {
	return cryptorandombytes.Base64UrlWithoutLeadingDash(6)
}

func NewHttpMonitorId() string {
	return cryptorandombytes.Base64UrlWithoutLeadingDash(6)
}
