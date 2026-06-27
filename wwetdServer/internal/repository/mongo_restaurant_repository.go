package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"wwetd-server/internal/domain"
)

type MongoRestaurantRepository struct {
	collection *mongo.Collection
}

func NewMongoRestaurantRepository(collection *mongo.Collection) *MongoRestaurantRepository {
	return &MongoRestaurantRepository{collection: collection}
}

func (r *MongoRestaurantRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "data_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *MongoRestaurantRepository) FindDetailByDataID(ctx context.Context, dataID string) (*domain.RestaurantDetail, error) {
	var detail domain.RestaurantDetail
	err := r.collection.FindOne(ctx, bson.M{"data_id": dataID}).Decode(&detail)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrRestaurantNotFound
	}
	if err != nil {
		return nil, err
	}

	return &detail, nil
}

func (r *MongoRestaurantRepository) UpsertDetail(ctx context.Context, detail *domain.RestaurantDetail) error {
	now := time.Now().UTC()
	if detail.CreatedAt.IsZero() {
		detail.CreatedAt = now
	}
	detail.UpdatedAt = now

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"data_id": detail.DataID},
		bson.M{
			"$set": bson.M{
				"detail":     detail.Detail,
				"updated_at": detail.UpdatedAt,
			},
			"$setOnInsert": bson.M{
				"data_id":    detail.DataID,
				"created_at": detail.CreatedAt,
			},
		},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}
