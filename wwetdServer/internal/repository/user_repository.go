package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

	"wwetd-server/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, limit int64, skip int64) ([]domain.User, error)
}

type IndexEnsurer interface {
	EnsureIndexes(ctx context.Context) error
}
