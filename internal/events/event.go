package events

import "time"

type EventType string

const (
	EventCheckComplete EventType = "check_complete"
	EventStateChange   EventType = "state_change"
)

type Event struct {
	Type     EventType `json:"type"`
	TargetID int       `json:"target_id"`
	At       time.Time `json:"at"`

	Success        bool    `json:"success"`
	StatusCode     int     `json:"status_code,omitempty"`
	ResponseTimeMS int     `json:"response_time_ms,omitempty"`
	ErrorMessage   *string `json:"error_message,omitempty"`

	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}
