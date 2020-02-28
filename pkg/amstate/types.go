package amstate

import (
	"time"
)

// for snapshots
type stateFormat struct {
	LastUnnoticedAlertsNotified time.Time                 `json:"last_unnoticed_alerts_notified"`
	ActiveAlerts                map[string]Alert          `json:"active_alerts"`
	HttpMonitors                map[string]HttpMonitor    `json:"http_monitors"`
	DeadMansSwitches            map[string]DeadMansSwitch `json:"dead_mans_switches"`
}

type Alert struct {
	Id        string    `json:"alert_key"` // name in JSON for backwards compat
	Subject   string    `json:"subject"`   // same type of error should always have same subject
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

type HttpMonitor struct {
	Id      string    `json:"id"`
	Created time.Time `json:"created"`
	Enabled bool      `json:"enabled"`
	Url     string    `json:"url"`
	Find    string    `json:"find"`
}

type DeadMansSwitch struct {
	Subject string    `json:"subject"`
	Ttl     time.Time `json:"ttl"`
}
