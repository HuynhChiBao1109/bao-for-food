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
)

type restaurantService struct {
	cache       cache.Store
	piso        interfaces.PisoSearcher
	restaurants repository.RestaurantRepository
	locationTTL time.Duration
}

func NewRestaurantService(cacheStore cache.Store, pisoSearcher interfaces.PisoSearcher, restaurantRepo repository.RestaurantRepository, locationTTL time.Duration) interfaces.RestaurantService {
	return &restaurantService{
		cache:       cacheStore,
		piso:        pisoSearcher,
		restaurants: restaurantRepo,
		locationTTL: locationTTL,
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
	}, nil
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

func randomIndex(length int) (int, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
	if err != nil {
		return 0, err
	}

	return int(value.Int64()), nil
}
