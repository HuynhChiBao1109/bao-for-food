package handler

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"wwetd-server/internal/domain"
	"wwetd-server/internal/dto"
	"wwetd-server/internal/interfaces"
)

type RestaurantHandler struct {
	service interfaces.RestaurantService
}

func NewRestaurantHandler(service interfaces.RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{service: service}
}

func (h *RestaurantHandler) SearchNearby(c *gin.Context) {
	query, ok := nearbyQueryFromRequest(c)
	if !ok {
		return
	}

	result, err := h.service.SearchNearby(c.Request.Context(), query)
	if err != nil {
		respondError(c, http.StatusBadGateway, "failed to search nearby restaurants")
		return
	}

	respondOK(c, result)
}

func (h *RestaurantHandler) PickNearby(c *gin.Context) {
	query, ok := nearbyQueryFromRequest(c)
	if !ok {
		return
	}

	result, err := h.service.PickNearby(c.Request.Context(), query)
	if err != nil {
		if errors.Is(err, domain.ErrRestaurantNotFound) {
			respondError(c, http.StatusNotFound, "restaurant not found")
			return
		}

		respondError(c, http.StatusBadGateway, "failed to pick nearby restaurant")
		return
	}

	respondOK(c, result)
}

func nearbyQueryFromRequest(c *gin.Context) (dto.NearbyRestaurantsQuery, bool) {
	lat, hasLat, err := parseOptionalFloat(c.Query("lat"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid lat")
		return dto.NearbyRestaurantsQuery{}, false
	}

	lng, hasLng, err := parseOptionalFloat(c.Query("lng"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid lng")
		return dto.NearbyRestaurantsQuery{}, false
	}

	if hasLat != hasLng {
		respondError(c, http.StatusBadRequest, "lat and lng must be provided together")
		return dto.NearbyRestaurantsQuery{}, false
	}

	return dto.NearbyRestaurantsQuery{
		IP:     clientIP(c),
		Query:  c.DefaultQuery("query", "quán ăn"),
		Limit:  parseIntQuery(c, "limit", 20),
		Lat:    lat,
		Lng:    lng,
		HasLat: hasLat,
		HasLng: hasLng,
	}, true
}

func clientIP(c *gin.Context) string {
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			if ip := strings.TrimSpace(parts[0]); net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	if realIP := strings.TrimSpace(c.GetHeader("X-Real-IP")); realIP != "" {
		if net.ParseIP(realIP) != nil {
			return realIP
		}
	}

	return c.ClientIP()
}

func parseOptionalFloat(value string) (float64, bool, error) {
	if strings.TrimSpace(value) == "" {
		return 0, false, nil
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false, err
	}
	return parsed, true, nil
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
