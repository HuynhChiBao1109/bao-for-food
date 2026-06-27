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
	UserIDFromToken(ctx context.Context, token string) (string, error)
}
