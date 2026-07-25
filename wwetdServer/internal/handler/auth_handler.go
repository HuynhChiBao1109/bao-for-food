package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"wwetd-server/internal/domain"
	"wwetd-server/internal/dto"
	"wwetd-server/internal/interfaces"
)

type AuthHandler struct {
	service   interfaces.AuthService
	uploadDir string
}

const maxAvatarSize = 5 << 20

func NewAuthHandler(service interfaces.AuthService, uploadDir string) *AuthHandler {
	return &AuthHandler{service: service, uploadDir: uploadDir}
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

func (h *AuthHandler) UpdateName(c *gin.Context) {
	userID, ok := h.requiredUserID(c)
	if !ok {
		return
	}

	var request dto.UpdateNameRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.UpdateName(c.Request.Context(), userID, request)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	respondOK(c, response)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, ok := h.requiredUserID(c)
	if !ok {
		return
	}

	response, err := h.service.GetMe(c.Request.Context(), userID)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	respondOK(c, response)
}

func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	userID, ok := h.requiredUserID(c)
	if !ok {
		return
	}

	currentUser, err := h.service.GetMe(c.Request.Context(), userID)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAvatarSize+(1<<20))
	header, err := c.FormFile("avatar")
	if err != nil {
		respondError(c, http.StatusBadRequest, "avatar file is required")
		return
	}
	if header.Size <= 0 || header.Size > maxAvatarSize {
		respondError(c, http.StatusBadRequest, "avatar must be 5 MB or smaller")
		return
	}

	source, err := header.Open()
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid avatar file")
		return
	}
	defer source.Close()

	head := make([]byte, 512)
	read, err := io.ReadFull(source, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		respondError(c, http.StatusBadRequest, "invalid avatar file")
		return
	}

	extension, ok := avatarExtension(http.DetectContentType(head[:read]))
	if !ok {
		respondError(c, http.StatusBadRequest, "avatar must be JPEG, PNG, or WebP")
		return
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		respondError(c, http.StatusBadRequest, "invalid avatar file")
		return
	}

	fileName, err := avatarFileName(userID, extension)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create avatar")
		return
	}

	avatarDir := filepath.Join(h.uploadDir, "avatars")
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create upload directory")
		return
	}

	destinationPath := filepath.Join(avatarDir, fileName)
	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to save avatar")
		return
	}

	copied, copyErr := io.Copy(destination, io.LimitReader(source, maxAvatarSize+1))
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil || copied > maxAvatarSize {
		_ = os.Remove(destinationPath)
		respondError(c, http.StatusInternalServerError, "failed to save avatar")
		return
	}

	avatarPath := "/uploads/avatars/" + fileName
	response, err := h.service.UpdateAvatar(c.Request.Context(), userID, avatarPath)
	if err != nil {
		_ = os.Remove(destinationPath)
		handleAuthError(c, err)
		return
	}

	h.removePreviousAvatar(currentUser.Avatar)
	respondOK(c, response)
}

func (h *AuthHandler) requiredUserID(c *gin.Context) (string, bool) {
	token := bearerToken(c)
	if token == "" {
		respondError(c, http.StatusUnauthorized, "login required")
		return "", false
	}

	userID, err := h.service.UserIDFromToken(c.Request.Context(), token)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "login required")
		return "", false
	}
	return userID, true
}

func (h *AuthHandler) removePreviousAvatar(avatar string) {
	const avatarPrefix = "/uploads/avatars/"
	if !strings.HasPrefix(avatar, avatarPrefix) {
		return
	}

	fileName := filepath.Base(avatar)
	if fileName == "." || fileName == string(filepath.Separator) {
		return
	}
	_ = os.Remove(filepath.Join(h.uploadDir, "avatars", fileName))
}

func avatarExtension(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func avatarFileName(userID string, extension string) (string, error) {
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate avatar filename: %w", err)
	}
	return userID + "-" + hex.EncodeToString(random) + extension, nil
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
	case errors.Is(err, domain.ErrInvalidName):
		respondError(c, http.StatusBadRequest, "name is required")
	case errors.Is(err, domain.ErrInvalidPhone):
		respondError(c, http.StatusBadRequest, "invalid phone")
	case errors.Is(err, domain.ErrInvalidPassword):
		respondError(c, http.StatusBadRequest, "password must contain 6 to 72 characters")
	default:
		respondError(c, http.StatusInternalServerError, "internal server error")
	}
}
