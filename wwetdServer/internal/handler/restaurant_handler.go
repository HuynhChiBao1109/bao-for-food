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
	auth    interfaces.AuthService
}

func NewRestaurantHandler(service interfaces.RestaurantService, auth interfaces.AuthService) *RestaurantHandler {
	return &RestaurantHandler{service: service, auth: auth}
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
	query.UserID = h.optionalUserID(c)

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

func (h *RestaurantHandler) RecordViewed(c *gin.Context) {
	userID, ok := h.requiredUserID(c)
	if !ok {
		return
	}

	location := dto.ClientLocation{Lat: parseFloatQuery(c, "lat"), Lng: parseFloatQuery(c, "lng")}
	if err := h.service.RecordViewed(c.Request.Context(), userID, c.Param("data_id"), location); err != nil {
		respondError(c, http.StatusBadGateway, "failed to record viewed restaurant")
		return
	}

	respondOK(c, gin.H{"saved": true})
}

func (h *RestaurantHandler) ListViewed(c *gin.Context) {
	userID, ok := h.requiredUserID(c)
	if !ok {
		return
	}

	items, err := h.service.ListViewed(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list viewed restaurants")
		return
	}

	respondOK(c, items)
}

func (h *RestaurantHandler) Save(c *gin.Context) {
	userID, ok := h.requiredUserID(c)
	if !ok {
		return
	}

	location := dto.ClientLocation{Lat: parseFloatQuery(c, "lat"), Lng: parseFloatQuery(c, "lng")}
	if err := h.service.SaveRestaurant(c.Request.Context(), userID, c.Param("data_id"), location); err != nil {
		respondError(c, http.StatusBadGateway, "failed to save restaurant")
		return
	}

	respondOK(c, gin.H{"saved": true})
}

func (h *RestaurantHandler) ListSaved(c *gin.Context) {
	userID, ok := h.requiredUserID(c)
	if !ok {
		return
	}

	items, err := h.service.ListSaved(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list saved restaurants")
		return
	}

	respondOK(c, items)
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

func (h *RestaurantHandler) optionalUserID(c *gin.Context) string {
	token := bearerToken(c)
	if token == "" {
		return ""
	}

	userID, err := h.auth.UserIDFromToken(c.Request.Context(), token)
	if err != nil {
		return ""
	}
	return userID
}

func (h *RestaurantHandler) requiredUserID(c *gin.Context) (string, bool) {
	userID := h.optionalUserID(c)
	if userID == "" {
		respondError(c, http.StatusUnauthorized, "login required")
		return "", false
	}
	return userID, true
}

func bearerToken(c *gin.Context) string {
	value := c.GetHeader("Authorization")
	if !strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return ""
	}
	return strings.TrimSpace(value[7:])
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

func parseFloatQuery(c *gin.Context, key string) float64 {
	value, err := strconv.ParseFloat(c.Query(key), 64)
	if err != nil {
		return 0
	}
	return value
}
