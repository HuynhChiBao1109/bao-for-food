package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"wwetd-server/internal/domain"
	"wwetd-server/internal/dto"
)

func TestAuthServiceRegisterCreatesProfileWithoutName(t *testing.T) {
	repo := newFakeAuthRepository()
	svc := NewAuthService(repo, newFakeCache())

	response, err := svc.Register(context.Background(), dto.RegisterRequest{
		Phone:    "0901234567",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if response.User.Name != "" {
		t.Fatalf("expected name to be completed after registration, got %q", response.User.Name)
	}
}

func TestAuthServiceRegisterRejectsShortPassword(t *testing.T) {
	svc := NewAuthService(newFakeAuthRepository(), newFakeCache())

	_, err := svc.Register(context.Background(), dto.RegisterRequest{
		Phone:    "0901234567",
		Password: "12345",
	})
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestAuthServiceLoginRejectsShortPassword(t *testing.T) {
	svc := NewAuthService(newFakeAuthRepository(), newFakeCache())

	_, err := svc.Login(context.Background(), dto.LoginRequest{
		Phone:    "0901234567",
		Password: "12345",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceRegisterRejectsInjectionLikePhone(t *testing.T) {
	svc := NewAuthService(newFakeAuthRepository(), newFakeCache())

	_, err := svc.Register(context.Background(), dto.RegisterRequest{
		Phone:    `0901234567' || '1'='1`,
		Password: "secret123",
	})
	if !errors.Is(err, domain.ErrInvalidPhone) {
		t.Fatalf("expected ErrInvalidPhone, got %v", err)
	}
}

func TestAuthServiceUpdateNameCompletesExistingProfile(t *testing.T) {
	repo := newFakeAuthRepository()
	user := &domain.AuthUser{
		ID:    bson.NewObjectID(),
		Phone: "0901234567",
	}
	repo.usersByPhone[user.Phone] = user
	repo.usersByID[user.ID.Hex()] = user
	svc := NewAuthService(repo, newFakeCache())

	response, err := svc.UpdateName(
		context.Background(),
		user.ID.Hex(),
		dto.UpdateNameRequest{Name: "  Minh Anh  "},
	)
	if err != nil {
		t.Fatalf("UpdateName returned error: %v", err)
	}

	if response.Name != "Minh Anh" {
		t.Fatalf("expected updated name, got %q", response.Name)
	}
	if user.Name != "Minh Anh" {
		t.Fatalf("expected stored user name to be updated, got %q", user.Name)
	}
}

func TestAuthServiceGetMeDoesNotExposeCredentials(t *testing.T) {
	repo := newFakeAuthRepository()
	user := &domain.AuthUser{
		ID:           bson.NewObjectID(),
		Name:         "Minh Anh",
		Phone:        "0901234567",
		PasswordHash: "private-hash",
		Salt:         "private-salt",
	}
	repo.usersByPhone[user.Phone] = user
	repo.usersByID[user.ID.Hex()] = user
	svc := NewAuthService(repo, newFakeCache())

	response, err := svc.GetMe(context.Background(), user.ID.Hex())
	if err != nil {
		t.Fatalf("GetMe returned error: %v", err)
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal GetMe response: %v", err)
	}
	body := string(payload)
	if strings.Contains(body, "password") || strings.Contains(body, "salt") ||
		strings.Contains(body, "private-hash") || strings.Contains(body, "private-salt") {
		t.Fatalf("GetMe exposed credential data: %s", body)
	}
}

func TestAuthServiceUpdateAvatarStoresPublicPath(t *testing.T) {
	repo := newFakeAuthRepository()
	user := &domain.AuthUser{
		ID:    bson.NewObjectID(),
		Phone: "0901234567",
	}
	repo.usersByPhone[user.Phone] = user
	repo.usersByID[user.ID.Hex()] = user
	svc := NewAuthService(repo, newFakeCache())

	response, err := svc.UpdateAvatar(
		context.Background(),
		user.ID.Hex(),
		"/uploads/avatars/user-avatar.jpg",
	)
	if err != nil {
		t.Fatalf("UpdateAvatar returned error: %v", err)
	}
	if response.Avatar != "/uploads/avatars/user-avatar.jpg" {
		t.Fatalf("expected updated avatar path, got %q", response.Avatar)
	}
}

type fakeAuthRepository struct {
	usersByID    map[string]*domain.AuthUser
	usersByPhone map[string]*domain.AuthUser
}

func newFakeAuthRepository() *fakeAuthRepository {
	return &fakeAuthRepository{
		usersByID:    make(map[string]*domain.AuthUser),
		usersByPhone: make(map[string]*domain.AuthUser),
	}
}

func (r *fakeAuthRepository) Create(_ context.Context, user *domain.AuthUser) error {
	if _, exists := r.usersByPhone[user.Phone]; exists {
		return domain.ErrPhoneAlreadyExists
	}
	if user.ID.IsZero() {
		user.ID = bson.NewObjectID()
	}
	r.usersByID[user.ID.Hex()] = user
	r.usersByPhone[user.Phone] = user
	return nil
}

func (r *fakeAuthRepository) FindByID(_ context.Context, id string) (*domain.AuthUser, error) {
	user, exists := r.usersByID[id]
	if !exists {
		return nil, domain.ErrAuthUserNotFound
	}
	return user, nil
}

func (r *fakeAuthRepository) FindByPhone(_ context.Context, phone string) (*domain.AuthUser, error) {
	user, exists := r.usersByPhone[phone]
	if !exists {
		return nil, domain.ErrAuthUserNotFound
	}
	return user, nil
}

func (r *fakeAuthRepository) UpdateName(_ context.Context, id string, name string) error {
	user, exists := r.usersByID[id]
	if !exists {
		return domain.ErrAuthUserNotFound
	}
	user.Name = name
	return nil
}

func (r *fakeAuthRepository) UpdateAvatar(_ context.Context, id string, avatar string) error {
	user, exists := r.usersByID[id]
	if !exists {
		return domain.ErrAuthUserNotFound
	}
	user.Avatar = avatar
	return nil
}
