package repository

import (
	"context"

	"wwetd-server/internal/domain"
)

type AuthRepository interface {
	Create(ctx context.Context, user *domain.AuthUser) error
	FindByPhone(ctx context.Context, phone string) (*domain.AuthUser, error)
}
