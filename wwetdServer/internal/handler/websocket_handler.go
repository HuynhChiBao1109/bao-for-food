package handler

import (
	"net/http"
	"wwetd-server/infrastructure/realtime"

	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	hub *realtime.Hub
}

func NewWebSocketHandler(hub *realtime.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

func (h *WebSocketHandler) Handle(c *gin.Context) {
	if err := h.hub.ServeHTTP(c.Writer, c.Request); err != nil {
		respondError(c, http.StatusBadRequest, "websocket upgrade failed")
		return
	}
}
