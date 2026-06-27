package dto

import "time"

type RealtimeEvent struct {
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
	SentAt time.Time   `json:"sent_at"`
}
