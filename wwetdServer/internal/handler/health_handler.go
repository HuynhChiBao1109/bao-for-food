package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"wwetd-server/internal/interfaces"
)

type HealthHandler struct {
	service interfaces.HealthChecker
}

func NewHealthHandler(service interfaces.HealthChecker) *HealthHandler {
	return &HealthHandler{service: service}
}

func (h *HealthHandler) Check(c *gin.Context) {
	report := h.service.Check(c.Request.Context())
	status := http.StatusOK
	if report.Status != "ok" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, report)
}
