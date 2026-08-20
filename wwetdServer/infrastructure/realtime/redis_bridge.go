package realtime

import (
	"context"

	"wwetd-server/internal/cache"
)

type RedisBridge struct {
	subscriber cache.Subscriber
	hub        *Hub
	channels   []string
}

func NewRedisBridge(subscriber cache.Subscriber, hub *Hub, channels []string) *RedisBridge {
	return &RedisBridge{
		subscriber: subscriber,
		hub:        hub,
		channels:   channels,
	}
}

func (b *RedisBridge) Start(ctx context.Context) error {
	messages, err := b.subscriber.Subscribe(ctx, b.channels...)
	if err != nil {
		return err
	}

	go func() {
		for message := range messages {
			b.hub.Broadcast([]byte(message.Payload))
		}
	}()

	return nil
}
