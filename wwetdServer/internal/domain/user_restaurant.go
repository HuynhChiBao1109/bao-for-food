package domain

import (
	"encoding/json"
	"time"
)

type UserRestaurant struct {
	UserID     string          `bson:"user_id" json:"user_id"`
	DataID     string          `bson:"data_id" json:"data_id"`
	Detail     json.RawMessage `bson:"detail" json:"detail"`
	RecordedAt time.Time       `bson:"recorded_at" json:"recorded_at"`
}
