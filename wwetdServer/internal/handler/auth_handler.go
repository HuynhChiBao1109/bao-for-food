package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"wwetd-server/internal/domain"
	"wwetd-server/internal/dto"
	"wwetd-server/internal/interfaces"
)

type AuthHandler struct {
	service interfaces.AuthService
}

func NewAuthHandler(service interfaces.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var request dto.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.Register(c.Request.Context(), request)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	respondCreated(c, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request dto.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.Login(c.Request.Context(), request)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	respondOK(c, response)
}

func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var request dto.RequestOTPRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.RequestOTP(c.Request.Context(), request)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	respondOK(c, response)
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var request dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.VerifyOTP(c.Request.Context(), request)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	respondOK(c, response)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var request dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.Refresh(c.Request.Context(), request)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	respondOK(c, response)
}

func handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrPhoneAlreadyExists):
		respondError(c, http.StatusConflict, "phone already exists")
	case errors.Is(err, domain.ErrInvalidCredentials):
		respondError(c, http.StatusUnauthorized, "invalid phone or password")
	case errors.Is(err, domain.ErrAuthUserNotFound):
		respondError(c, http.StatusNotFound, "user not found")
	case errors.Is(err, domain.ErrInvalidOTP):
		respondError(c, http.StatusUnauthorized, "invalid otp")
	default:
		respondError(c, http.StatusInternalServerError, "internal server error")
	}
}
