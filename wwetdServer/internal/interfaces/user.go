package interfaces

import (
	"context"

	"wwetd-server/internal/domain"
	"wwetd-server/internal/dto"
)

type UserService interface {
	Create(ctx context.Context, input dto.CreateUserInput) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	List(ctx context.Context, query dto.ListUsersQuery) ([]domain.User, error)
}
