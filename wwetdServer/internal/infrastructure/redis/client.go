package redisinfra

import (
	"context"
	"errors"
	"time"

	redis "github.com/redis/go-redis/v9"

	"wwetd-server/internal/cache"
	"wwetd-server/internal/config"
)

type Client struct {
	client *redis.Client
}

func NewClient(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	redisClient := &Client{client: client}
	if err := redisClient.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}

	return redisClient, nil
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	value, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", cache.ErrMiss
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *Client) Publish(ctx context.Context, channel string, payload string) error {
	return c.client.Publish(ctx, channel, payload).Err()
}

func (c *Client) Subscribe(ctx context.Context, channels ...string) (<-chan cache.Message, error) {
	pubsub := c.client.Subscribe(ctx, channels...)
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, err
	}

	out := make(chan cache.Message)
	in := pubsub.Channel()

	go func() {
		defer close(out)
		defer pubsub.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case message, ok := <-in:
				if !ok {
					return
				}

				select {
				case out <- cache.Message{Channel: message.Channel, Payload: message.Payload}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
