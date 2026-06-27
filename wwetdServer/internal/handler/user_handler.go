package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"wwetd-server/internal/domain"
	"wwetd-server/internal/dto"
	"wwetd-server/internal/interfaces"
)

type UserHandler struct {
	service interfaces.UserService
}

func NewUserHandler(service interfaces.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(c *gin.Context) {
	var request dto.CreateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.Create(c.Request.Context(), dto.CreateUserInput{
		Name:  request.Name,
		Email: request.Email,
	})
	if err != nil {
		handleUserError(c, err)
		return
	}

	respondCreated(c, dto.NewUserResponse(user))
}

func (h *UserHandler) GetByID(c *gin.Context) {
	user, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleUserError(c, err)
		return
	}

	respondOK(c, dto.NewUserResponse(user))
}

func (h *UserHandler) List(c *gin.Context) {
	page := parseInt64Query(c, "page", 1)
	limit := parseInt64Query(c, "limit", 20)

	users, err := h.service.List(c.Request.Context(), dto.ListUsersQuery{
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list users")
		return
	}

	respondOK(c, dto.NewUserResponses(users))
}

func handleUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidID):
		respondError(c, http.StatusBadRequest, "invalid user id")
	case errors.Is(err, domain.ErrUserNotFound):
		respondError(c, http.StatusNotFound, "user not found")
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		respondError(c, http.StatusConflict, "email already exists")
	default:
		respondError(c, http.StatusInternalServerError, "internal server error")
	}
}

func parseInt64Query(c *gin.Context, key string, fallback int64) int64 {
	value := c.Query(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
