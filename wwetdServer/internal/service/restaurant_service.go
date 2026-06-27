package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"math/big"
	"net"
	"strings"
	"time"

	"wwetd-server/internal/cache"
	"wwetd-server/internal/domain"
	"wwetd-server/internal/dto"
	"wwetd-server/internal/interfaces"
	"wwetd-server/internal/repository"
)

const (
	defaultRestaurantQuery = "quán ăn gần đây"
	viewedRestaurantsTTL   = 24 * time.Hour
)

type restaurantService struct {
	cache           cache.Store
	piso            interfaces.PisoSearcher
	restaurants     repository.RestaurantRepository
	userRestaurants repository.UserRestaurantRepository
	locationTTL     time.Duration
}

func NewRestaurantService(cacheStore cache.Store, pisoSearcher interfaces.PisoSearcher, restaurantRepo repository.RestaurantRepository, userRestaurantRepo repository.UserRestaurantRepository, locationTTL time.Duration) interfaces.RestaurantService {
	return &restaurantService{
		cache:           cacheStore,
		piso:            pisoSearcher,
		restaurants:     restaurantRepo,
		userRestaurants: userRestaurantRepo,
		locationTTL:     locationTTL,
	}
}

func (s *restaurantService) SearchNearby(ctx context.Context, query dto.NearbyRestaurantsQuery) (dto.NearbyRestaurantsResponse, error) {
	location, source := s.resolveLocation(ctx, query)

	searchQuery := strings.TrimSpace(query.Query)
	if searchQuery == "" {
		searchQuery = defaultRestaurantQuery
	}

	results, err := s.piso.Search(ctx, dto.PisoSearchParams{
		Query: searchQuery,
		Lat:   location.Lat,
		Lng:   location.Lng,
		Limit: query.Limit,
	})
	if err != nil {
		return dto.NearbyRestaurantsResponse{}, err
	}

	return dto.NearbyRestaurantsResponse{
		Location: location,
		Source:   source,
		Results:  results,
	}, nil
}

func (s *restaurantService) PickNearby(ctx context.Context, query dto.NearbyRestaurantsQuery) (dto.PickRestaurantResponse, error) {
	if query.Limit <= 0 {
		query.Limit = 20
	}

	nearby, err := s.SearchNearby(ctx, query)
	if err != nil {
		return dto.PickRestaurantResponse{}, err
	}

	restaurants, err := extractRestaurants(nearby.Results)
	if err != nil {
		return dto.PickRestaurantResponse{}, err
	}
	if len(restaurants) == 0 {
		return dto.PickRestaurantResponse{}, domain.ErrRestaurantNotFound
	}

	restaurants = s.filterViewedRestaurants(ctx, query.UserID, restaurants)
	if len(restaurants) == 0 {
		return dto.PickRestaurantResponse{}, domain.ErrRestaurantNotFound
	}

	index, err := randomIndex(len(restaurants))
	if err != nil {
		return dto.PickRestaurantResponse{}, err
	}

	selected := restaurants[index]
	dataID, err := extractDataID(selected)
	if err != nil {
		return dto.PickRestaurantResponse{}, err
	}

	detail, detailSource, err := s.getPlaceDetail(ctx, dataID, nearby.Location)
	if err != nil {
		return dto.PickRestaurantResponse{}, err
	}

	return dto.PickRestaurantResponse{
		Location:     nearby.Location,
		Source:       nearby.Source,
		DetailSource: detailSource,
		Restaurant:   detail,
		IsSaved:      s.isSaved(ctx, query.UserID, dataID),
	}, nil
}

func (s *restaurantService) GetDetail(ctx context.Context, userID string, dataID string, location dto.ClientLocation) (dto.RestaurantDetailResponse, error) {
	detail, detailSource, err := s.getPlaceDetail(ctx, dataID, location)
	if err != nil {
		return dto.RestaurantDetailResponse{}, err
	}

	return dto.RestaurantDetailResponse{
		DetailSource: detailSource,
		Restaurant:   detail,
		IsSaved:      s.isSaved(ctx, userID, dataID),
	}, nil
}

func (s *restaurantService) RecordViewed(ctx context.Context, userID string, dataID string, location dto.ClientLocation) error {
	detail, _, err := s.getPlaceDetail(ctx, dataID, location)
	if err != nil {
		return err
	}

	items, _ := s.getViewedItems(ctx, userID)
	now := time.Now().UTC()
	next := make([]domain.UserRestaurant, 0, len(items)+1)
	next = append(next, domain.UserRestaurant{
		UserID:     userID,
		DataID:     dataID,
		Detail:     detail,
		RecordedAt: now,
	})
	for _, item := range items {
		if item.DataID != dataID {
			next = append(next, item)
		}
	}

	payload, err := json.Marshal(next)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, viewedRestaurantsKey(userID), string(payload), viewedRestaurantsTTL)
}

func (s *restaurantService) ListViewed(ctx context.Context, userID string) ([]dto.UserRestaurantItem, error) {
	items, err := s.getViewedItems(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toUserRestaurantItems(items), nil
}

func (s *restaurantService) SaveRestaurant(ctx context.Context, userID string, dataID string, location dto.ClientLocation) error {
	detail, _, err := s.getPlaceDetail(ctx, dataID, location)
	if err != nil {
		return err
	}

	return s.userRestaurants.UpsertSaved(ctx, &domain.UserRestaurant{
		UserID: userID,
		DataID: dataID,
		Detail: detail,
	})
}

func (s *restaurantService) ListSaved(ctx context.Context, userID string) ([]dto.UserRestaurantItem, error) {
	items, err := s.userRestaurants.ListSaved(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toUserRestaurantItems(items), nil
}

func (s *restaurantService) resolveLocation(ctx context.Context, query dto.NearbyRestaurantsQuery) (dto.ClientLocation, string) {
	ip := normalizeIP(query.IP)

	if query.HasLat && query.HasLng {
		location := dto.ClientLocation{IP: ip, Lat: query.Lat, Lng: query.Lng}
		s.cacheLocation(ctx, location)
		return location, "request"
	}

	if location, ok := s.getCachedLocation(ctx, ip); ok {
		return location, "cache"
	}

	location := generateLocationFromIP(ip)
	s.cacheLocation(ctx, location)
	return location, "generated"
}

func (s *restaurantService) getCachedLocation(ctx context.Context, ip string) (dto.ClientLocation, bool) {
	if s.cache == nil {
		return dto.ClientLocation{}, false
	}

	payload, err := s.cache.Get(ctx, locationCacheKey(ip))
	if err != nil {
		return dto.ClientLocation{}, false
	}

	var location dto.ClientLocation
	if err := json.Unmarshal([]byte(payload), &location); err != nil {
		return dto.ClientLocation{}, false
	}

	return location, true
}

func (s *restaurantService) cacheLocation(ctx context.Context, location dto.ClientLocation) {
	if s.cache == nil || s.locationTTL <= 0 {
		return
	}

	payload, err := json.Marshal(location)
	if err != nil {
		return
	}

	_ = s.cache.Set(ctx, locationCacheKey(location.IP), string(payload), s.locationTTL)
}

func locationCacheKey(ip string) string {
	return fmt.Sprintf("locations:ip:%s", ip)
}

func viewedRestaurantsKey(userID string) string {
	return fmt.Sprintf("users:%s:restaurants:viewed", userID)
}

func normalizeIP(value string) string {
	host := strings.TrimSpace(value)
	if host == "" {
		return "unknown"
	}

	if parsed := net.ParseIP(host); parsed != nil {
		return parsed.String()
	}

	if splitHost, _, err := net.SplitHostPort(host); err == nil {
		if parsed := net.ParseIP(splitHost); parsed != nil {
			return parsed.String()
		}
	}

	parts := strings.Split(host, ",")
	if len(parts) > 0 {
		if parsed := net.ParseIP(strings.TrimSpace(parts[0])); parsed != nil {
			return parsed.String()
		}
	}

	return host
}

func generateLocationFromIP(ip string) dto.ClientLocation {
	if parsed := net.ParseIP(ip); parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified() {
		return dto.ClientLocation{IP: ip, Lat: 10.776889, Lng: 106.700806}
	}

	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(ip))
	sum := hasher.Sum64()

	latSeed := float64(uint32(sum>>32)) / float64(math.MaxUint32)
	lngSeed := float64(uint32(sum)) / float64(math.MaxUint32)

	return dto.ClientLocation{
		IP:  ip,
		Lat: 10.776889 + ((latSeed - 0.5) * 0.08),
		Lng: 106.700806 + ((lngSeed - 0.5) * 0.08),
	}
}

func extractRestaurants(payload json.RawMessage) ([]json.RawMessage, error) {
	var wrapped struct {
		LocalResult []json.RawMessage `json:"local_result"`
		Results     []json.RawMessage `json:"results"`
		Data        []json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(payload, &wrapped); err != nil {
		return nil, err
	}

	switch {
	case len(wrapped.LocalResult) > 0:
		return wrapped.LocalResult, nil
	case len(wrapped.Results) > 0:
		return wrapped.Results, nil
	case len(wrapped.Data) > 0:
		return wrapped.Data, nil
	default:
		var direct []json.RawMessage
		if err := json.Unmarshal(payload, &direct); err == nil {
			return direct, nil
		}
		return nil, domain.ErrRestaurantNotFound
	}
}

func extractDataID(payload json.RawMessage) (string, error) {
	var restaurant struct {
		DataID string `json:"data_id"`
	}

	if err := json.Unmarshal(payload, &restaurant); err != nil {
		return "", err
	}
	if strings.TrimSpace(restaurant.DataID) == "" {
		return "", domain.ErrRestaurantNotFound
	}

	return restaurant.DataID, nil
}

func (s *restaurantService) getPlaceDetail(ctx context.Context, dataID string, location dto.ClientLocation) (json.RawMessage, string, error) {
	if s.restaurants != nil {
		detail, err := s.restaurants.FindDetailByDataID(ctx, dataID)
		if err == nil && json.Valid(detail.Detail) {
			return normalizePlaceDetail(detail.Detail), "database", nil
		}
		if err != nil && !errors.Is(err, domain.ErrRestaurantNotFound) {
			return nil, "", err
		}
	}

	detail, err := s.piso.Place(ctx, dto.PisoPlaceParams{
		DataID: dataID,
		Lat:    location.Lat,
		Lng:    location.Lng,
	})
	if err != nil {
		return nil, "", err
	}

	detail = normalizePlaceDetail(detail)

	if s.restaurants != nil {
		if err := s.restaurants.UpsertDetail(ctx, &domain.RestaurantDetail{
			DataID: dataID,
			Detail: detail,
		}); err != nil {
			return nil, "", err
		}
	}

	return detail, "piso", nil
}

func normalizePlaceDetail(payload json.RawMessage) json.RawMessage {
	var wrapped struct {
		PlaceResult json.RawMessage `json:"place_result"`
	}

	if err := json.Unmarshal(payload, &wrapped); err == nil && len(wrapped.PlaceResult) > 0 {
		return wrapped.PlaceResult
	}

	return payload
}

func (s *restaurantService) filterViewedRestaurants(ctx context.Context, userID string, restaurants []json.RawMessage) []json.RawMessage {
	if strings.TrimSpace(userID) == "" {
		return restaurants
	}

	viewedItems, err := s.getViewedItems(ctx, userID)
	if err != nil || len(viewedItems) == 0 {
		return restaurants
	}

	viewed := make(map[string]struct{}, len(viewedItems))
	for _, item := range viewedItems {
		viewed[item.DataID] = struct{}{}
	}

	filtered := make([]json.RawMessage, 0, len(restaurants))
	for _, restaurant := range restaurants {
		dataID, err := extractDataID(restaurant)
		if err != nil {
			continue
		}
		if _, ok := viewed[dataID]; !ok {
			filtered = append(filtered, restaurant)
		}
	}

	return filtered
}

func (s *restaurantService) getViewedItems(ctx context.Context, userID string) ([]domain.UserRestaurant, error) {
	if s.cache == nil || strings.TrimSpace(userID) == "" {
		return []domain.UserRestaurant{}, nil
	}

	payload, err := s.cache.Get(ctx, viewedRestaurantsKey(userID))
	if err != nil {
		return []domain.UserRestaurant{}, nil
	}

	var items []domain.UserRestaurant
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *restaurantService) isSaved(ctx context.Context, userID string, dataID string) bool {
	if s.userRestaurants == nil || strings.TrimSpace(userID) == "" || strings.TrimSpace(dataID) == "" {
		return false
	}

	ok, err := s.userRestaurants.IsSaved(ctx, userID, dataID)
	return err == nil && ok
}

func toUserRestaurantItems(items []domain.UserRestaurant) []dto.UserRestaurantItem {
	response := make([]dto.UserRestaurantItem, 0, len(items))
	for _, item := range items {
		response = append(response, dto.UserRestaurantItem{
			DataID:     item.DataID,
			Detail:     item.Detail,
			RecordedAt: item.RecordedAt.Format(time.RFC3339),
		})
	}
	return response
}

func randomIndex(length int) (int, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
	if err != nil {
		return 0, err
	}

	return int(value.Int64()), nil
}
