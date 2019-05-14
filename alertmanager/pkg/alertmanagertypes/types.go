package alertmanagertypes

import (
	"time"
)

type Alert struct {
	Key       string    `json:"alert_key"`
	Subject   string    `json:"subject"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}
