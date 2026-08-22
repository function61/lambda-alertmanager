// Structure of data for all state changes
package amdomain

import (
	"time"

	"github.com/function61/eventhorizon/pkg/ehevent"
)

var Types = ehevent.Allocators{
	"AlertRaised":               func() ehevent.Event { return &AlertRaised{} },
	"AlertAcknowledged":         func() ehevent.Event { return &AlertAcknowledged{} },
	"UnnoticedAlertsNotified":   func() ehevent.Event { return &UnnoticedAlertsNotified{} },
	"HttpMonitorCreated":        func() ehevent.Event { return &HTTPMonitorCreated{} },
	"HttpMonitorEnabledUpdated": func() ehevent.Event { return &HTTPMonitorEnabledUpdated{} },
	"HttpMonitorDeleted":        func() ehevent.Event { return &HTTPMonitorDeleted{} },
	"DeadMansSwitchCreated":     func() ehevent.Event { return &DeadMansSwitchCreated{} },
	"DeadMansSwitchCheckin":     func() ehevent.Event { return &DeadMansSwitchCheckin{} },
	"DeadMansSwitchDeleted":     func() ehevent.Event { return &DeadMansSwitchDeleted{} },
}

// ------

type AlertRaised struct {
	meta    ehevent.EventMeta
	ID      string `json:"Id"`
	Subject string
	Details string
}

func (e *AlertRaised) MetaType() string         { return "AlertRaised" }
func (e *AlertRaised) Meta() *ehevent.EventMeta { return &e.meta }

func NewAlertRaised(
	id string,
	subject string,
	details string,
	meta ehevent.EventMeta,
) *AlertRaised {
	return &AlertRaised{
		meta:    meta,
		ID:      id,
		Subject: subject,
		Details: details,
	}
}

// ------

type AlertAcknowledged struct {
	meta ehevent.EventMeta
	ID   string `json:"Id"`
}

func (e *AlertAcknowledged) MetaType() string         { return "AlertAcknowledged" }
func (e *AlertAcknowledged) Meta() *ehevent.EventMeta { return &e.meta }

func NewAlertAcknowledged(
	id string,
	meta ehevent.EventMeta,
) *AlertAcknowledged {
	return &AlertAcknowledged{
		meta: meta,
		ID:   id,
	}
}

// ------

type UnnoticedAlertsNotified struct {
	meta     ehevent.EventMeta
	AlertIDs []string `json:"AlertIds"`
}

func (e *UnnoticedAlertsNotified) MetaType() string         { return "UnnoticedAlertsNotified" }
func (e *UnnoticedAlertsNotified) Meta() *ehevent.EventMeta { return &e.meta }

func NewUnnoticedAlertsNotified(
	alertIds []string,
	meta ehevent.EventMeta,
) *UnnoticedAlertsNotified {
	return &UnnoticedAlertsNotified{
		meta:     meta,
		AlertIDs: alertIds,
	}
}

// ------

type HTTPMonitorCreated struct {
	meta    ehevent.EventMeta
	ID      string `json:"Id"`
	Enabled bool
	URL     string `json:"Url"`
	Find    string
}

func (e *HTTPMonitorCreated) MetaType() string         { return "HttpMonitorCreated" }
func (e *HTTPMonitorCreated) Meta() *ehevent.EventMeta { return &e.meta }

func NewHTTPMonitorCreated(
	id string,
	enabled bool,
	url string,
	find string,
	meta ehevent.EventMeta,
) *HTTPMonitorCreated {
	return &HTTPMonitorCreated{
		meta:    meta,
		ID:      id,
		Enabled: enabled,
		URL:     url,
		Find:    find,
	}
}

// ------

type HTTPMonitorEnabledUpdated struct {
	meta    ehevent.EventMeta
	ID      string `json:"Id"`
	Enabled bool
}

func (e *HTTPMonitorEnabledUpdated) MetaType() string         { return "HttpMonitorEnabledUpdated" }
func (e *HTTPMonitorEnabledUpdated) Meta() *ehevent.EventMeta { return &e.meta }

func NewHTTPMonitorEnabledUpdated(
	id string,
	enabled bool,
	meta ehevent.EventMeta,
) *HTTPMonitorEnabledUpdated {
	return &HTTPMonitorEnabledUpdated{
		meta:    meta,
		ID:      id,
		Enabled: enabled,
	}
}

// ------

type HTTPMonitorDeleted struct {
	meta ehevent.EventMeta
	ID   string `json:"Id"`
}

func (e *HTTPMonitorDeleted) MetaType() string         { return "HttpMonitorDeleted" }
func (e *HTTPMonitorDeleted) Meta() *ehevent.EventMeta { return &e.meta }

func NewHTTPMonitorDeleted(
	id string,
	meta ehevent.EventMeta,
) *HTTPMonitorDeleted {
	return &HTTPMonitorDeleted{
		meta: meta,
		ID:   id,
	}
}

// ------

type DeadMansSwitchCreated struct {
	meta    ehevent.EventMeta
	Subject string
	TTL     time.Time `json:"Ttl"`
}

func (e *DeadMansSwitchCreated) MetaType() string         { return "DeadMansSwitchCreated" }
func (e *DeadMansSwitchCreated) Meta() *ehevent.EventMeta { return &e.meta }

func NewDeadMansSwitchCreated(
	subject string,
	ttl time.Time,
	meta ehevent.EventMeta,
) *DeadMansSwitchCreated {
	return &DeadMansSwitchCreated{
		meta:    meta,
		Subject: subject,
		TTL:     ttl,
	}
}

// ------

type DeadMansSwitchCheckin struct {
	meta    ehevent.EventMeta
	Subject string
	TTL     time.Time `json:"Ttl"`
}

func (e *DeadMansSwitchCheckin) MetaType() string         { return "DeadMansSwitchCheckin" }
func (e *DeadMansSwitchCheckin) Meta() *ehevent.EventMeta { return &e.meta }

func NewDeadMansSwitchCheckin(
	subject string,
	ttl time.Time,
	meta ehevent.EventMeta,
) *DeadMansSwitchCheckin {
	return &DeadMansSwitchCheckin{
		meta:    meta,
		Subject: subject,
		TTL:     ttl,
	}
}

// ------

type DeadMansSwitchDeleted struct {
	meta    ehevent.EventMeta
	Subject string
}

func (e *DeadMansSwitchDeleted) MetaType() string         { return "DeadMansSwitchDeleted" }
func (e *DeadMansSwitchDeleted) Meta() *ehevent.EventMeta { return &e.meta }

func NewDeadMansSwitchDeleted(
	subject string,
	meta ehevent.EventMeta,
) *DeadMansSwitchDeleted {
	return &DeadMansSwitchDeleted{
		meta:    meta,
		Subject: subject,
	}
}
