package service

import (
	"context"
	"encoding/json"
	"math"
	"testing"
	"time"

	"wwetd-server/internal/domain"
	"wwetd-server/internal/dto"
)

func TestRestaurantServiceSearchNearbyGeneratesCachesAndCallsPiso(t *testing.T) {
	ctx := context.Background()
	cacheStore := newFakeCache()
	restaurantRepo := newFakeRestaurantRepo()
	userRestaurantRepo := newFakeUserRestaurantRepo()
	piso := &fakePisoSearcher{searchResponse: json.RawMessage(`{"local_result":[]}`)}
	svc := NewRestaurantService(cacheStore, piso, restaurantRepo, userRestaurantRepo, time.Hour)

	result, err := svc.SearchNearby(ctx, dto.NearbyRestaurantsQuery{
		IP:    "8.8.8.8",
		Query: "phở",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("SearchNearby returned error: %v", err)
	}

	if result.Source != "generated" {
		t.Fatalf("expected generated location source, got %q", result.Source)
	}
	if len(piso.searchCalls) != 1 {
		t.Fatalf("expected one piso search call, got %d", len(piso.searchCalls))
	}
	if piso.searchCalls[0].Query != "phở" {
		t.Fatalf("expected query phở, got %q", piso.searchCalls[0].Query)
	}
	if piso.searchCalls[0].Limit != 10 {
		t.Fatalf("expected limit 10, got %d", piso.searchCalls[0].Limit)
	}
	if _, ok := cacheStore.values[locationCacheKey("8.8.8.8")]; !ok {
		t.Fatalf("expected generated location to be cached")
	}
}

func TestRestaurantServiceSearchNearbyUsesCachedLocation(t *testing.T) {
	ctx := context.Background()
	cacheStore := newFakeCache()
	restaurantRepo := newFakeRestaurantRepo()
	userRestaurantRepo := newFakeUserRestaurantRepo()
	cachedLocation := dto.ClientLocation{IP: "8.8.4.4", Lat: 10.8, Lng: 106.7}
	payload, err := json.Marshal(cachedLocation)
	if err != nil {
		t.Fatalf("marshal cached location: %v", err)
	}
	cacheStore.values[locationCacheKey("8.8.4.4")] = string(payload)

	piso := &fakePisoSearcher{searchResponse: json.RawMessage(`{"local_result":[]}`)}
	svc := NewRestaurantService(cacheStore, piso, restaurantRepo, userRestaurantRepo, time.Hour)

	result, err := svc.SearchNearby(ctx, dto.NearbyRestaurantsQuery{IP: "8.8.4.4"})
	if err != nil {
		t.Fatalf("SearchNearby returned error: %v", err)
	}

	if result.Source != "cache" {
		t.Fatalf("expected cache location source, got %q", result.Source)
	}
	if len(piso.searchCalls) != 1 {
		t.Fatalf("expected one piso search call, got %d", len(piso.searchCalls))
	}
	if piso.searchCalls[0].Lat != cachedLocation.Lat || piso.searchCalls[0].Lng != cachedLocation.Lng {
		t.Fatalf("expected cached coordinates, got lat=%f lng=%f", piso.searchCalls[0].Lat, piso.searchCalls[0].Lng)
	}
}

func TestRestaurantServicePickNearbyReturnsOneRestaurant(t *testing.T) {
	ctx := context.Background()
	cacheStore := newFakeCache()
	restaurantRepo := newFakeRestaurantRepo()
	userRestaurantRepo := newFakeUserRestaurantRepo()
	piso := &fakePisoSearcher{
		searchResponse: json.RawMessage(`{
		"local_result": [
			{"data_id": "place-a", "title": "Quán A"},
			{"data_id": "place-b", "title": "Quán B"}
		]
	}`),
		placeResponse: json.RawMessage(`{"data_id": "place-a", "title": "Quán A detail"}`),
	}
	svc := NewRestaurantService(cacheStore, piso, restaurantRepo, userRestaurantRepo, time.Hour)

	result, err := svc.PickNearby(ctx, dto.NearbyRestaurantsQuery{IP: "8.8.8.8"})
	if err != nil {
		t.Fatalf("PickNearby returned error: %v", err)
	}

	var restaurant map[string]interface{}
	if err := json.Unmarshal(result.Restaurant, &restaurant); err != nil {
		t.Fatalf("unmarshal restaurant: %v", err)
	}
	if restaurant["title"] != "Quán A detail" {
		t.Fatalf("expected restaurant detail title, got %v", restaurant["title"])
	}
	if result.DetailSource != "piso" {
		t.Fatalf("expected detail source piso, got %q", result.DetailSource)
	}
	if len(piso.placeCalls) != 1 {
		t.Fatalf("expected one piso place call, got %d", len(piso.placeCalls))
	}
	if _, ok := restaurantRepo.details[piso.placeCalls[0].DataID]; !ok {
		t.Fatalf("expected place detail to be saved in repository")
	}
}

func TestRestaurantServicePickNearbyUsesSavedRestaurantDetail(t *testing.T) {
	ctx := context.Background()
	cacheStore := newFakeCache()
	restaurantRepo := newFakeRestaurantRepo()
	userRestaurantRepo := newFakeUserRestaurantRepo()
	restaurantRepo.details["place-a"] = json.RawMessage(`{"data_id": "place-a", "title": "Saved detail"}`)

	piso := &fakePisoSearcher{
		searchResponse: json.RawMessage(`{
		"local_result": [
			{"data_id": "place-a", "title": "Quán A"}
		]
	}`),
		placeResponse: json.RawMessage(`{"data_id": "place-a", "title": "Piso detail"}`),
	}
	svc := NewRestaurantService(cacheStore, piso, restaurantRepo, userRestaurantRepo, time.Hour)

	result, err := svc.PickNearby(ctx, dto.NearbyRestaurantsQuery{IP: "8.8.8.8"})
	if err != nil {
		t.Fatalf("PickNearby returned error: %v", err)
	}

	var restaurant map[string]interface{}
	if err := json.Unmarshal(result.Restaurant, &restaurant); err != nil {
		t.Fatalf("unmarshal restaurant: %v", err)
	}
	if restaurant["title"] != "Saved detail" {
		t.Fatalf("expected saved restaurant detail, got %v", restaurant["title"])
	}
	if result.DetailSource != "database" {
		t.Fatalf("expected detail source database, got %q", result.DetailSource)
	}
	if len(piso.placeCalls) != 0 {
		t.Fatalf("expected no piso place call, got %d", len(piso.placeCalls))
	}
}

func TestRestaurantServicePickNearbyExpandsSearchWhenNearestRestaurantsWereViewed(t *testing.T) {
	ctx := context.Background()
	cacheStore := newFakeCache()
	restaurantRepo := newFakeRestaurantRepo()
	userRestaurantRepo := newFakeUserRestaurantRepo()
	userID := "user-a"

	viewedItems := make([]domain.UserRestaurant, 0, todayRestaurantLimit)
	for index := 1; index <= todayRestaurantLimit; index++ {
		viewedItems = append(viewedItems, domain.UserRestaurant{
			UserID: userID,
			DataID: "viewed-" + string(rune('a'+index-1)),
		})
	}
	viewedPayload, err := json.Marshal(viewedItems)
	if err != nil {
		t.Fatalf("marshal viewed restaurants: %v", err)
	}
	cacheStore.values[viewedRestaurantsKey(userID)] = string(viewedPayload)

	piso := &fakePisoSearcher{
		searchResponses: []json.RawMessage{
			json.RawMessage(`{
				"local_result": [
					{"data_id": "viewed-a"}, {"data_id": "viewed-b"},
					{"data_id": "viewed-c"}, {"data_id": "viewed-d"},
					{"data_id": "viewed-e"}, {"data_id": "viewed-f"},
					{"data_id": "viewed-g"}, {"data_id": "viewed-h"},
					{"data_id": "viewed-i"}, {"data_id": "viewed-j"}
				]
			}`),
			json.RawMessage(`{"local_result":[{"data_id":"place-new","title":"Quán mới"}]}`),
		},
		placeResponse: json.RawMessage(`{"data_id":"place-new","title":"Quán mới detail"}`),
	}
	svc := NewRestaurantService(cacheStore, piso, restaurantRepo, userRestaurantRepo, time.Hour)

	result, err := svc.PickNearby(ctx, dto.NearbyRestaurantsQuery{
		UserID: userID,
		Lat:    10.7,
		Lng:    106.7,
		HasLat: true,
		HasLng: true,
	})
	if err != nil {
		t.Fatalf("PickNearby returned error: %v", err)
	}

	if len(piso.searchCalls) != 2 {
		t.Fatalf("expected two piso search calls, got %d", len(piso.searchCalls))
	}
	if piso.searchCalls[0].Limit != todayRestaurantLimit {
		t.Fatalf("expected today limit %d, got %d", todayRestaurantLimit, piso.searchCalls[0].Limit)
	}
	if piso.searchCalls[0].Query != defaultRestaurantQuery {
		t.Fatalf("expected default restaurant query, got %q", piso.searchCalls[0].Query)
	}
	if math.Abs(piso.searchCalls[1].Lat-10.71) > 0.000001 ||
		math.Abs(piso.searchCalls[1].Lng-106.71) > 0.000001 {
		t.Fatalf(
			"expected expanded coordinates lat=10.71 lng=106.71, got lat=%f lng=%f",
			piso.searchCalls[1].Lat,
			piso.searchCalls[1].Lng,
		)
	}
	if result.Location.Lat != 10.7 || result.Location.Lng != 106.7 {
		t.Fatalf("expected original response location, got lat=%f lng=%f", result.Location.Lat, result.Location.Lng)
	}

	var restaurant map[string]interface{}
	if err := json.Unmarshal(result.Restaurant, &restaurant); err != nil {
		t.Fatalf("unmarshal selected restaurant: %v", err)
	}
	if restaurant["data_id"] != "place-new" {
		t.Fatalf("expected unviewed restaurant place-new, got %v", restaurant["data_id"])
	}
}

type fakePisoSearcher struct {
	searchResponse  json.RawMessage
	searchResponses []json.RawMessage
	placeResponse   json.RawMessage
	searchCalls     []dto.PisoSearchParams
	placeCalls      []dto.PisoPlaceParams
}

func (s *fakePisoSearcher) Search(_ context.Context, params dto.PisoSearchParams) (json.RawMessage, error) {
	s.searchCalls = append(s.searchCalls, params)
	index := len(s.searchCalls) - 1
	if index < len(s.searchResponses) {
		return s.searchResponses[index], nil
	}
	return s.searchResponse, nil
}

func (s *fakePisoSearcher) Place(_ context.Context, params dto.PisoPlaceParams) (json.RawMessage, error) {
	s.placeCalls = append(s.placeCalls, params)
	return s.placeResponse, nil
}

type fakeRestaurantRepo struct {
	details map[string]json.RawMessage
}

func newFakeRestaurantRepo() *fakeRestaurantRepo {
	return &fakeRestaurantRepo{details: make(map[string]json.RawMessage)}
}

func (r *fakeRestaurantRepo) FindDetailByDataID(_ context.Context, dataID string) (*domain.RestaurantDetail, error) {
	detail, ok := r.details[dataID]
	if !ok {
		return nil, domain.ErrRestaurantNotFound
	}

	return &domain.RestaurantDetail{
		DataID: dataID,
		Detail: detail,
	}, nil
}

func (r *fakeRestaurantRepo) UpsertDetail(_ context.Context, detail *domain.RestaurantDetail) error {
	r.details[detail.DataID] = detail.Detail
	return nil
}

type fakeUserRestaurantRepo struct {
	saved []domain.UserRestaurant
}

func newFakeUserRestaurantRepo() *fakeUserRestaurantRepo {
	return &fakeUserRestaurantRepo{}
}

func (r *fakeUserRestaurantRepo) UpsertSaved(_ context.Context, item *domain.UserRestaurant) error {
	r.saved = append(r.saved, *item)
	return nil
}

func (r *fakeUserRestaurantRepo) ListSaved(_ context.Context, userID string) ([]domain.UserRestaurant, error) {
	var items []domain.UserRestaurant
	for _, item := range r.saved {
		if item.UserID == userID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeUserRestaurantRepo) IsSaved(_ context.Context, userID string, dataID string) (bool, error) {
	for _, item := range r.saved {
		if item.UserID == userID && item.DataID == dataID {
			return true, nil
		}
	}
	return false, nil
}
