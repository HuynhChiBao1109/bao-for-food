package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"wwetd-server/internal/dto"
)

func TestRestaurantServiceSearchNearbyGeneratesCachesAndCallsPiso(t *testing.T) {
	ctx := context.Background()
	cacheStore := newFakeCache()
	piso := &fakePisoSearcher{searchResponse: json.RawMessage(`{"local_result":[]}`)}
	svc := NewRestaurantService(cacheStore, piso, time.Hour, time.Hour)

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
	cachedLocation := dto.ClientLocation{IP: "8.8.4.4", Lat: 10.8, Lng: 106.7}
	payload, err := json.Marshal(cachedLocation)
	if err != nil {
		t.Fatalf("marshal cached location: %v", err)
	}
	cacheStore.values[locationCacheKey("8.8.4.4")] = string(payload)

	piso := &fakePisoSearcher{searchResponse: json.RawMessage(`{"local_result":[]}`)}
	svc := NewRestaurantService(cacheStore, piso, time.Hour, time.Hour)

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
	piso := &fakePisoSearcher{
		searchResponse: json.RawMessage(`{
		"local_result": [
			{"data_id": "place-a", "title": "Quán A"},
			{"data_id": "place-b", "title": "Quán B"}
		]
	}`),
		placeResponse: json.RawMessage(`{"data_id": "place-a", "title": "Quán A detail"}`),
	}
	svc := NewRestaurantService(cacheStore, piso, time.Hour, time.Hour)

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
	if _, ok := cacheStore.values[placeDetailCacheKey(piso.placeCalls[0].DataID)]; !ok {
		t.Fatalf("expected place detail to be cached")
	}
}

type fakePisoSearcher struct {
	searchResponse json.RawMessage
	placeResponse  json.RawMessage
	searchCalls    []dto.PisoSearchParams
	placeCalls     []dto.PisoPlaceParams
}

func (s *fakePisoSearcher) Search(_ context.Context, params dto.PisoSearchParams) (json.RawMessage, error) {
	s.searchCalls = append(s.searchCalls, params)
	return s.searchResponse, nil
}

func (s *fakePisoSearcher) Place(_ context.Context, params dto.PisoPlaceParams) (json.RawMessage, error) {
	s.placeCalls = append(s.placeCalls, params)
	return s.placeResponse, nil
}
