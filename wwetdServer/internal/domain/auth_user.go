package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthUser struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Avatar       string        `bson:"avatar" json:"avatar"`
	Phone        string        `bson:"phone" json:"phone"`
	PasswordHash string        `bson:"password_hash" json:"-"`
	Salt         string        `bson:"salt" json:"-"`
	CreatedAt    time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at" json:"updated_at"`
}
