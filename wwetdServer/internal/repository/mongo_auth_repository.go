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

type MongoAuthRepository struct {
	collection *mongo.Collection
}

func NewMongoAuthRepository(collection *mongo.Collection) *MongoAuthRepository {
	return &MongoAuthRepository{collection: collection}
}

func (r *MongoAuthRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "phone", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *MongoAuthRepository) Create(ctx context.Context, user *domain.AuthUser) error {
	now := time.Now().UTC()
	if user.ID.IsZero() {
		user.ID = bson.NewObjectID()
	}
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		return domain.ErrPhoneAlreadyExists
	}
	return err
}

func (r *MongoAuthRepository) FindByPhone(ctx context.Context, phone string) (*domain.AuthUser, error) {
	var user domain.AuthUser
	err := r.collection.FindOne(ctx, bson.M{"phone": phone}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrAuthUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MongoAuthRepository) FindByID(ctx context.Context, id string) (*domain.AuthUser, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrAuthUserNotFound
	}

	var user domain.AuthUser
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrAuthUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MongoAuthRepository) UpdateName(ctx context.Context, id string, name string) error {
	return r.updateProfileField(ctx, id, "name", name)
}

func (r *MongoAuthRepository) UpdateAvatar(ctx context.Context, id string, avatar string) error {
	return r.updateProfileField(ctx, id, "avatar", avatar)
}

func (r *MongoAuthRepository) updateProfileField(ctx context.Context, id string, field string, value string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrAuthUserNotFound
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{field: value, "updated_at": time.Now().UTC()}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrAuthUserNotFound
	}
	return nil
}
