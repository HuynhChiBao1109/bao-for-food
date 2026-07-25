package domain

import "errors"

var (
	ErrInvalidID          = errors.New("invalid id")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrRestaurantNotFound = errors.New("restaurant not found")
	ErrPhoneAlreadyExists = errors.New("phone already exists")
	ErrAuthUserNotFound   = errors.New("auth user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidOTP         = errors.New("invalid otp")
	ErrInvalidName        = errors.New("invalid name")
	ErrInvalidPhone       = errors.New("invalid phone")
	ErrInvalidPassword    = errors.New("invalid password")
)
