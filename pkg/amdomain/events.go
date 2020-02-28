// Structure of data for all state changes
package amdomain

import (
	"github.com/function61/eventhorizon/pkg/ehevent"
	"time"
)

var Types = ehevent.Allocators{
	"AlertRaised":               func() ehevent.Event { return &AlertRaised{} },
	"AlertAcknowledged":         func() ehevent.Event { return &AlertAcknowledged{} },
	"UnnoticedAlertsNotified":   func() ehevent.Event { return &UnnoticedAlertsNotified{} },
	"HttpMonitorCreated":        func() ehevent.Event { return &HttpMonitorCreated{} },
	"HttpMonitorEnabledUpdated": func() ehevent.Event { return &HttpMonitorEnabledUpdated{} },
	"HttpMonitorDeleted":        func() ehevent.Event { return &HttpMonitorDeleted{} },
	"DeadMansSwitchCreated":     func() ehevent.Event { return &DeadMansSwitchCreated{} },
	"DeadMansSwitchCheckin":     func() ehevent.Event { return &DeadMansSwitchCheckin{} },
	"DeadMansSwitchDeleted":     func() ehevent.Event { return &DeadMansSwitchDeleted{} },
}

// ------

type AlertRaised struct {
	meta    ehevent.EventMeta
	Id      string
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
		Id:      id,
		Subject: subject,
		Details: details,
	}
}

// ------

type AlertAcknowledged struct {
	meta ehevent.EventMeta
	Id   string
}

func (e *AlertAcknowledged) MetaType() string         { return "AlertAcknowledged" }
func (e *AlertAcknowledged) Meta() *ehevent.EventMeta { return &e.meta }

func NewAlertAcknowledged(
	id string,
	meta ehevent.EventMeta,
) *AlertAcknowledged {
	return &AlertAcknowledged{
		meta: meta,
		Id:   id,
	}
}

// ------

type UnnoticedAlertsNotified struct {
	meta     ehevent.EventMeta
	AlertIds []string
}

func (e *UnnoticedAlertsNotified) MetaType() string         { return "UnnoticedAlertsNotified" }
func (e *UnnoticedAlertsNotified) Meta() *ehevent.EventMeta { return &e.meta }

func NewUnnoticedAlertsNotified(
	alertIds []string,
	meta ehevent.EventMeta,
) *UnnoticedAlertsNotified {
	return &UnnoticedAlertsNotified{
		meta:     meta,
		AlertIds: alertIds,
	}
}

// ------

type HttpMonitorCreated struct {
	meta    ehevent.EventMeta
	Id      string
	Enabled bool
	Url     string
	Find    string
}

func (e *HttpMonitorCreated) MetaType() string         { return "HttpMonitorCreated" }
func (e *HttpMonitorCreated) Meta() *ehevent.EventMeta { return &e.meta }

func NewHttpMonitorCreated(
	id string,
	enabled bool,
	url string,
	find string,
	meta ehevent.EventMeta,
) *HttpMonitorCreated {
	return &HttpMonitorCreated{
		meta:    meta,
		Id:      id,
		Enabled: enabled,
		Url:     url,
		Find:    find,
	}
}

// ------

type HttpMonitorEnabledUpdated struct {
	meta    ehevent.EventMeta
	Id      string
	Enabled bool
}

func (e *HttpMonitorEnabledUpdated) MetaType() string         { return "HttpMonitorEnabledUpdated" }
func (e *HttpMonitorEnabledUpdated) Meta() *ehevent.EventMeta { return &e.meta }

func NewHttpMonitorEnabledUpdated(
	id string,
	enabled bool,
	meta ehevent.EventMeta,
) *HttpMonitorEnabledUpdated {
	return &HttpMonitorEnabledUpdated{
		meta:    meta,
		Id:      id,
		Enabled: enabled,
	}
}

// ------

type HttpMonitorDeleted struct {
	meta ehevent.EventMeta
	Id   string
}

func (e *HttpMonitorDeleted) MetaType() string         { return "HttpMonitorDeleted" }
func (e *HttpMonitorDeleted) Meta() *ehevent.EventMeta { return &e.meta }

func NewHttpMonitorDeleted(
	id string,
	meta ehevent.EventMeta,
) *HttpMonitorDeleted {
	return &HttpMonitorDeleted{
		meta: meta,
		Id:   id,
	}
}

// ------

type DeadMansSwitchCreated struct {
	meta    ehevent.EventMeta
	Subject string
	Ttl     time.Time
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
		Ttl:     ttl,
	}
}

// ------

type DeadMansSwitchCheckin struct {
	meta    ehevent.EventMeta
	Subject string
	Ttl     time.Time
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
		Ttl:     ttl,
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
