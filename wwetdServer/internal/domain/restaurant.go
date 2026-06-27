package domain

import (
	"encoding/json"
	"time"
)

type RestaurantDetail struct {
	DataID    string          `bson:"data_id" json:"data_id"`
	Detail    json.RawMessage `bson:"detail" json:"detail"`
	CreatedAt time.Time       `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time       `bson:"updated_at" json:"updated_at"`
}
