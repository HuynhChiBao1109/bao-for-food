package repository

import (
	"context"

	"wwetd-server/internal/domain"
)

type UserRestaurantRepository interface {
	UpsertSaved(ctx context.Context, item *domain.UserRestaurant) error
	ListSaved(ctx context.Context, userID string) ([]domain.UserRestaurant, error)
	IsSaved(ctx context.Context, userID string, dataID string) (bool, error)
}
