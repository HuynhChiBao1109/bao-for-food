package dto

import "time"

type HealthReport struct {
	Status       string            `json:"status"`
	CheckedAt    time.Time         `json:"checked_at"`
	Dependencies map[string]string `json:"dependencies"`
}
