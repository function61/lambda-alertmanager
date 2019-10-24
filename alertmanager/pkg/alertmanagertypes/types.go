package alertmanagertypes

import (
	"time"
)

type Alert struct {
	// will be filled at ingestion time with sequential number: 1, 2, 3, ...
	// this is to implement race condition -free rate limiting by utilizing unique column constraint on save
	Key       string    `json:"alert_key"`
	Subject   string    `json:"subject"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

type DeadMansSwitch struct {
	Subject string    `json:"subject"`
	TTL     time.Time `json:"ttl"`
}

type DeadMansSwitchCheckinRequest struct {
	Subject string `json:"subject"`
	TTL     string `json:"ttl"`
}

func NewDeadMansSwitchCheckinRequest(subject string, ttl string) DeadMansSwitchCheckinRequest {
	return DeadMansSwitchCheckinRequest{
		Subject: subject,
		TTL:     ttl,
	}
}

func (a *Alert) Equal(other Alert) bool {
	return a.Subject == other.Subject
}
