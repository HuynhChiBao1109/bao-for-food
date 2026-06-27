package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"wwetd-server/internal/config"
)

type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewClient(ctx context.Context, cfg config.MongoConfig) (*Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	dbClient := &Client{
		client:   client,
		database: client.Database(cfg.Database),
	}

	if err := dbClient.Ping(ctx); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	return dbClient, nil
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, readpref.Primary())
}

func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

func (c *Client) Database() *mongo.Database {
	return c.database
}

func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
