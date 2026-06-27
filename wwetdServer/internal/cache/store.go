package cache

import (
	"context"
	"errors"
	"time"
)

var ErrMiss = errors.New("cache miss")

type Message struct {
	Channel string
	Payload string
}

type Store interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Publish(ctx context.Context, channel string, payload string) error
}

type Subscriber interface {
	Subscribe(ctx context.Context, channels ...string) (<-chan Message, error)
}
