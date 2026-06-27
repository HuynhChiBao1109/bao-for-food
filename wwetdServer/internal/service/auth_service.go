package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"

	"wwetd-server/internal/cache"
	"wwetd-server/internal/domain"
	"wwetd-server/internal/dto"
	"wwetd-server/internal/interfaces"
	"wwetd-server/internal/repository"
)

const (
	passwordIterations = 120_000
	passwordKeyLength  = 32
	otpTTL             = 5 * time.Minute
	authSessionTTL     = 24 * time.Hour
)

var nonDigitPattern = regexp.MustCompile(`\D+`)

type authService struct {
	repo  repository.AuthRepository
	cache cache.Store
}

func NewAuthService(repo repository.AuthRepository, cacheStore cache.Store) interfaces.AuthService {
	return &authService{repo: repo, cache: cacheStore}
}

func (s *authService) Register(ctx context.Context, request dto.RegisterRequest) (dto.AuthResponse, error) {
	phone := normalizePhone(request.Phone)
	salt, err := randomBase64(16)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	user := &domain.AuthUser{
		Phone:        phone,
		Salt:         salt,
		PasswordHash: hashPassword(request.Password, salt),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return dto.AuthResponse{}, err
	}

	return s.issueAuth(ctx, user)
}

func (s *authService) Login(ctx context.Context, request dto.LoginRequest) (dto.AuthResponse, error) {
	phone := normalizePhone(request.Phone)
	user, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return dto.AuthResponse{}, domain.ErrInvalidCredentials
	}

	expectedHash := hashPassword(request.Password, user.Salt)
	if subtle.ConstantTimeCompare([]byte(expectedHash), []byte(user.PasswordHash)) != 1 {
		return dto.AuthResponse{}, domain.ErrInvalidCredentials
	}

	return s.issueAuth(ctx, user)
}

func (s *authService) RequestOTP(ctx context.Context, request dto.RequestOTPRequest) (dto.OTPResponse, error) {
	phone := normalizePhone(request.Phone)
	if _, err := s.repo.FindByPhone(ctx, phone); err != nil {
		return dto.OTPResponse{}, domain.ErrAuthUserNotFound
	}

	otp, err := randomOTP()
	if err != nil {
		return dto.OTPResponse{}, err
	}

	if err := s.cache.Set(ctx, otpCacheKey(phone), otp, otpTTL); err != nil {
		return dto.OTPResponse{}, err
	}

	return dto.OTPResponse{
		Phone:    phone,
		TTL:      int(otpTTL.Seconds()),
		DebugOTP: otp,
	}, nil
}

func (s *authService) VerifyOTP(ctx context.Context, request dto.VerifyOTPRequest) (dto.AuthResponse, error) {
	phone := normalizePhone(request.Phone)
	cachedOTP, err := s.cache.Get(ctx, otpCacheKey(phone))
	if err != nil {
		return dto.AuthResponse{}, domain.ErrInvalidOTP
	}
	if subtle.ConstantTimeCompare([]byte(cachedOTP), []byte(strings.TrimSpace(request.OTP))) != 1 {
		return dto.AuthResponse{}, domain.ErrInvalidOTP
	}

	user, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	_ = s.cache.Delete(ctx, otpCacheKey(phone))
	return s.issueAuth(ctx, user)
}

func (s *authService) issueAuth(ctx context.Context, user *domain.AuthUser) (dto.AuthResponse, error) {
	token, err := randomBase64(32)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	if s.cache != nil {
		if err := s.cache.Set(ctx, authSessionKey(token), user.ID.Hex(), authSessionTTL); err != nil {
			return dto.AuthResponse{}, err
		}
	}

	return dto.AuthResponse{
		Token: token,
		User:  dto.NewAuthUserResponse(user),
	}, nil
}

func hashPassword(password string, salt string) string {
	key := pbkdf2.Key([]byte(password), []byte(salt), passwordIterations, passwordKeyLength, sha256.New)
	return base64.RawStdEncoding.EncodeToString(key)
}

func normalizePhone(phone string) string {
	return nonDigitPattern.ReplaceAllString(phone, "")
}

func randomBase64(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func randomOTP() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func otpCacheKey(phone string) string {
	return fmt.Sprintf("auth:otp:%s", phone)
}

func authSessionKey(token string) string {
	return fmt.Sprintf("auth:sessions:%s", token)
}
