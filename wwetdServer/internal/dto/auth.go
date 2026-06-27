package dto

import (
	"time"

	"wwetd-server/internal/domain"
)

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,min=8,max=20"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required,min=8,max=20"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type RequestOTPRequest struct {
	Phone string `json:"phone" binding:"required,min=8,max=20"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone" binding:"required,min=8,max=20"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type AuthUserResponse struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthResponse struct {
	Token string           `json:"token"`
	User  AuthUserResponse `json:"user"`
}

type OTPResponse struct {
	Phone    string `json:"phone"`
	TTL      int    `json:"ttl"`
	DebugOTP string `json:"debug_otp,omitempty"`
}

func NewAuthUserResponse(user *domain.AuthUser) AuthUserResponse {
	return AuthUserResponse{
		ID:        user.ID.Hex(),
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
	}
}
