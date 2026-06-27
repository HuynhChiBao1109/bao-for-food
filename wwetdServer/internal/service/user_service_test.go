package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"wwetd-server/internal/cache"
	"wwetd-server/internal/domain"
	"wwetd-server/internal/dto"
)

func TestUserServiceCreateNormalizesEmailCachesAndPublishes(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	cacheStore := newFakeCache()
	svc := NewUserService(repo, cacheStore, time.Minute, "realtime.test")

	user, err := svc.Create(ctx, dto.CreateUserInput{
		Name:  " Ada Lovelace ",
		Email: " ADA@EXAMPLE.COM ",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if user.Name != "Ada Lovelace" {
		t.Fatalf("expected trimmed name, got %q", user.Name)
	}
	if user.Email != "ada@example.com" {
		t.Fatalf("expected normalized email, got %q", user.Email)
	}
	if _, ok := cacheStore.values[userCacheKey(user.ID.Hex())]; !ok {
		t.Fatalf("expected user to be cached")
	}
	if len(cacheStore.published) != 1 {
		t.Fatalf("expected one published event, got %d", len(cacheStore.published))
	}
	if cacheStore.published[0].channel != "realtime.test" {
		t.Fatalf("expected event channel realtime.test, got %q", cacheStore.published[0].channel)
	}
}

func TestUserServiceCreateRejectsDuplicateEmail(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	existing := &domain.User{
		ID:    bson.NewObjectID(),
		Name:  "Ada Lovelace",
		Email: "ada@example.com",
	}
	repo.usersByEmail[existing.Email] = existing

	svc := NewUserService(repo, newFakeCache(), time.Minute, "realtime.test")
	_, err := svc.Create(ctx, dto.CreateUserInput{
		Name:  "Ada Lovelace",
		Email: "ADA@example.com",
	})
	if !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Fatalf("expected duplicate email error, got %v", err)
	}
}

func TestUserServiceGetByIDReturnsCachedUser(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	cacheStore := newFakeCache()
	svc := NewUserService(repo, cacheStore, time.Minute, "realtime.test")

	user := &domain.User{
		ID:        bson.NewObjectID(),
		Name:      "Ada Lovelace",
		Email:     "ada@example.com",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	cacheStore.values[userCacheKey(user.ID.Hex())] = mustMarshalUser(t, user)

	found, err := svc.GetByID(ctx, user.ID.Hex())
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if found.ID != user.ID {
		t.Fatalf("expected cached user id %s, got %s", user.ID.Hex(), found.ID.Hex())
	}
	if repo.findByIDCalls != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.findByIDCalls)
	}
}

type fakeUserRepo struct {
	usersByID     map[bson.ObjectID]*domain.User
	usersByEmail  map[string]*domain.User
	findByIDCalls int
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		usersByID:    make(map[bson.ObjectID]*domain.User),
		usersByEmail: make(map[string]*domain.User),
	}
}

func (r *fakeUserRepo) Create(_ context.Context, user *domain.User) error {
	now := time.Now().UTC()
	if user.ID.IsZero() {
		user.ID = bson.NewObjectID()
	}
	user.CreatedAt = now
	user.UpdatedAt = now

	clone := *user
	r.usersByID[user.ID] = &clone
	r.usersByEmail[user.Email] = &clone
	return nil
}

func (r *fakeUserRepo) FindByID(_ context.Context, id bson.ObjectID) (*domain.User, error) {
	r.findByIDCalls++
	user, ok := r.usersByID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}

	clone := *user
	return &clone, nil
}

func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	user, ok := r.usersByEmail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}

	clone := *user
	return &clone, nil
}

func (r *fakeUserRepo) List(_ context.Context, _ int64, _ int64) ([]domain.User, error) {
	users := make([]domain.User, 0, len(r.usersByID))
	for _, user := range r.usersByID {
		users = append(users, *user)
	}
	return users, nil
}

type publishedMessage struct {
	channel string
	payload string
}

type fakeCache struct {
	values    map[string]string
	published []publishedMessage
}

func newFakeCache() *fakeCache {
	return &fakeCache{values: make(map[string]string)}
}

func (c *fakeCache) Get(_ context.Context, key string) (string, error) {
	value, ok := c.values[key]
	if !ok {
		return "", cache.ErrMiss
	}
	return value, nil
}

func (c *fakeCache) Set(_ context.Context, key string, value string, _ time.Duration) error {
	c.values[key] = value
	return nil
}

func (c *fakeCache) Delete(_ context.Context, keys ...string) error {
	for _, key := range keys {
		delete(c.values, key)
	}
	return nil
}

func (c *fakeCache) Publish(_ context.Context, channel string, payload string) error {
	c.published = append(c.published, publishedMessage{channel: channel, payload: payload})
	return nil
}

func mustMarshalUser(t *testing.T, user *domain.User) string {
	t.Helper()

	payload, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal user: %v", err)
	}
	return string(payload)
}
