package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"wwetd-server/internal/domain"
)

type MongoUserRestaurantRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRestaurantRepository(collection *mongo.Collection) *MongoUserRestaurantRepository {
	return &MongoUserRestaurantRepository{collection: collection}
}

func (r *MongoUserRestaurantRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "data_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *MongoUserRestaurantRepository) UpsertSaved(ctx context.Context, item *domain.UserRestaurant) error {
	now := time.Now().UTC()
	item.RecordedAt = now

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"user_id": item.UserID, "data_id": item.DataID},
		bson.M{"$set": bson.M{
			"user_id":     item.UserID,
			"data_id":     item.DataID,
			"detail":      item.Detail,
			"recorded_at": item.RecordedAt,
		}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (r *MongoUserRestaurantRepository) ListSaved(ctx context.Context, userID string) ([]domain.UserRestaurant, error) {
	cursor, err := r.collection.Find(
		ctx,
		bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{Key: "recorded_at", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []domain.UserRestaurant
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]domain.UserRestaurant, 0)
	}
	return items, nil
}

func (r *MongoUserRestaurantRepository) IsSaved(ctx context.Context, userID string, dataID string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"user_id": userID, "data_id": dataID})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
