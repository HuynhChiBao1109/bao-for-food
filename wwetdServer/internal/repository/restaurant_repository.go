package repository

import (
	"context"

	"wwetd-server/internal/domain"
)

type RestaurantRepository interface {
	FindDetailByDataID(ctx context.Context, dataID string) (*domain.RestaurantDetail, error)
	UpsertDetail(ctx context.Context, detail *domain.RestaurantDetail) error
}
