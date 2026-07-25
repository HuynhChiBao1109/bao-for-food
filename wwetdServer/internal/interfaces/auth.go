package interfaces

import (
	"context"

	"wwetd-server/internal/dto"
)

type AuthService interface {
	Register(ctx context.Context, request dto.RegisterRequest) (dto.AuthResponse, error)
	Login(ctx context.Context, request dto.LoginRequest) (dto.AuthResponse, error)
	RequestOTP(ctx context.Context, request dto.RequestOTPRequest) (dto.OTPResponse, error)
	VerifyOTP(ctx context.Context, request dto.VerifyOTPRequest) (dto.AuthResponse, error)
	Refresh(ctx context.Context, request dto.RefreshTokenRequest) (dto.AuthResponse, error)
	GetMe(ctx context.Context, userID string) (dto.AuthUserResponse, error)
	UpdateName(ctx context.Context, userID string, request dto.UpdateNameRequest) (dto.AuthUserResponse, error)
	UpdateAvatar(ctx context.Context, userID string, avatar string) (dto.AuthUserResponse, error)
	UserIDFromToken(ctx context.Context, token string) (string, error)
}
