package repository

import (
	"context"

	"wwetd-server/internal/domain"
)

type AuthRepository interface {
	Create(ctx context.Context, user *domain.AuthUser) error
	FindByID(ctx context.Context, id string) (*domain.AuthUser, error)
	FindByPhone(ctx context.Context, phone string) (*domain.AuthUser, error)
	UpdateName(ctx context.Context, id string, name string) error
	UpdateAvatar(ctx context.Context, id string, avatar string) error
}
