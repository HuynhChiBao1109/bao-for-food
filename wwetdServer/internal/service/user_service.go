package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"wwetd-server/internal/cache"
	"wwetd-server/internal/domain"
	"wwetd-server/internal/dto"
	"wwetd-server/internal/interfaces"
	"wwetd-server/internal/repository"
)

const usersListCacheKey = "users:list"

type userService struct {
	repo          repository.UserRepository
	cache         cache.Store
	userCacheTTL  time.Duration
	pubSubChannel string
}

func NewUserService(repo repository.UserRepository, cacheStore cache.Store, userCacheTTL time.Duration, pubSubChannel string) interfaces.UserService {
	return &userService{
		repo:          repo,
		cache:         cacheStore,
		userCacheTTL:  userCacheTTL,
		pubSubChannel: pubSubChannel,
	}
}

func (s *userService) Create(ctx context.Context, input dto.CreateUserInput) (*domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	name := strings.TrimSpace(input.Name)

	if existing, err := s.repo.FindByEmail(ctx, email); err == nil && existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	} else if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	user := &domain.User{
		Name:  name,
		Email: email,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	s.cacheUser(ctx, user)
	s.deleteCache(ctx, usersListCacheKey)
	s.publishEvent(ctx, "user.created", user)

	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	if user, ok := s.getCachedUser(ctx, id); ok {
		return user, nil
	}

	user, err := s.repo.FindByID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	s.cacheUser(ctx, user)
	return user, nil
}

func (s *userService) List(ctx context.Context, query dto.ListUsersQuery) ([]domain.User, error) {
	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}

	skip := (page - 1) * limit
	return s.repo.List(ctx, limit, skip)
}

func (s *userService) getCachedUser(ctx context.Context, id string) (*domain.User, bool) {
	if s.cache == nil {
		return nil, false
	}

	payload, err := s.cache.Get(ctx, userCacheKey(id))
	if err != nil {
		return nil, false
	}

	var user domain.User
	if err := json.Unmarshal([]byte(payload), &user); err != nil {
		return nil, false
	}

	return &user, true
}

func (s *userService) cacheUser(ctx context.Context, user *domain.User) {
	if s.cache == nil || s.userCacheTTL <= 0 {
		return
	}

	payload, err := json.Marshal(user)
	if err != nil {
		return
	}

	_ = s.cache.Set(ctx, userCacheKey(user.ID.Hex()), string(payload), s.userCacheTTL)
}

func (s *userService) deleteCache(ctx context.Context, keys ...string) {
	if s.cache == nil {
		return
	}

	_ = s.cache.Delete(ctx, keys...)
}

func (s *userService) publishEvent(ctx context.Context, eventType string, data interface{}) {
	if s.cache == nil || s.pubSubChannel == "" {
		return
	}

	event := dto.RealtimeEvent{
		Type:   eventType,
		Data:   data,
		SentAt: time.Now().UTC(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	_ = s.cache.Publish(ctx, s.pubSubChannel, string(payload))
}

func userCacheKey(id string) string {
	return fmt.Sprintf("users:%s", id)
}
