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
	accessTokenTTL     = 24 * time.Hour
	refreshTokenTTL    = 7 * 24 * time.Hour
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

func (s *authService) Refresh(ctx context.Context, request dto.RefreshTokenRequest) (dto.AuthResponse, error) {
	refreshToken := strings.TrimSpace(request.RefreshToken)
	if s.cache == nil || refreshToken == "" {
		return dto.AuthResponse{}, domain.ErrInvalidCredentials
	}

	userID, err := s.cache.Get(ctx, refreshTokenKey(refreshToken))
	if err != nil {
		return dto.AuthResponse{}, domain.ErrInvalidCredentials
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return dto.AuthResponse{}, domain.ErrInvalidCredentials
	}

	_ = s.cache.Delete(ctx, refreshTokenKey(refreshToken))
	return s.issueAuth(ctx, user)
}

func (s *authService) UserIDFromToken(ctx context.Context, token string) (string, error) {
	if s.cache == nil || strings.TrimSpace(token) == "" {
		return "", domain.ErrInvalidCredentials
	}

	userID, err := s.cache.Get(ctx, accessTokenKey(strings.TrimSpace(token)))
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}
	return userID, nil
}

func (s *authService) issueAuth(ctx context.Context, user *domain.AuthUser) (dto.AuthResponse, error) {
	accessToken, err := randomBase64(32)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	refreshToken, err := randomBase64(32)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	if s.cache != nil {
		if err := s.cache.Set(ctx, accessTokenKey(accessToken), user.ID.Hex(), accessTokenTTL); err != nil {
			return dto.AuthResponse{}, err
		}
		if err := s.cache.Set(ctx, refreshTokenKey(refreshToken), user.ID.Hex(), refreshTokenTTL); err != nil {
			return dto.AuthResponse{}, err
		}
	}

	return dto.AuthResponse{
		Token:            accessToken,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int(accessTokenTTL.Seconds()),
		RefreshExpiresIn: int(refreshTokenTTL.Seconds()),
		User:             dto.NewAuthUserResponse(user),
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

func accessTokenKey(token string) string {
	return fmt.Sprintf("auth:access:%s", token)
}

func refreshTokenKey(token string) string {
	return fmt.Sprintf("auth:refresh:%s", token)
}
