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
	piso := &fakePisoSearcher{response: json.RawMessage(`{"local_result":[]}`)}
	svc := NewRestaurantService(cacheStore, piso, time.Hour)

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
	if len(piso.calls) != 1 {
		t.Fatalf("expected one piso call, got %d", len(piso.calls))
	}
	if piso.calls[0].Query != "phở" {
		t.Fatalf("expected query phở, got %q", piso.calls[0].Query)
	}
	if piso.calls[0].Limit != 10 {
		t.Fatalf("expected limit 10, got %d", piso.calls[0].Limit)
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

	piso := &fakePisoSearcher{response: json.RawMessage(`{"local_result":[]}`)}
	svc := NewRestaurantService(cacheStore, piso, time.Hour)

	result, err := svc.SearchNearby(ctx, dto.NearbyRestaurantsQuery{IP: "8.8.4.4"})
	if err != nil {
		t.Fatalf("SearchNearby returned error: %v", err)
	}

	if result.Source != "cache" {
		t.Fatalf("expected cache location source, got %q", result.Source)
	}
	if len(piso.calls) != 1 {
		t.Fatalf("expected one piso call, got %d", len(piso.calls))
	}
	if piso.calls[0].Lat != cachedLocation.Lat || piso.calls[0].Lng != cachedLocation.Lng {
		t.Fatalf("expected cached coordinates, got lat=%f lng=%f", piso.calls[0].Lat, piso.calls[0].Lng)
	}
}

func TestRestaurantServicePickNearbyReturnsOneRestaurant(t *testing.T) {
	ctx := context.Background()
	cacheStore := newFakeCache()
	piso := &fakePisoSearcher{response: json.RawMessage(`{
		"local_result": [
			{"title": "Quán A"},
			{"title": "Quán B"}
		]
	}`)}
	svc := NewRestaurantService(cacheStore, piso, time.Hour)

	result, err := svc.PickNearby(ctx, dto.NearbyRestaurantsQuery{IP: "8.8.8.8"})
	if err != nil {
		t.Fatalf("PickNearby returned error: %v", err)
	}

	var restaurant map[string]interface{}
	if err := json.Unmarshal(result.Restaurant, &restaurant); err != nil {
		t.Fatalf("unmarshal restaurant: %v", err)
	}
	if restaurant["title"] == "" {
		t.Fatalf("expected restaurant title")
	}
}

type fakePisoSearcher struct {
	response json.RawMessage
	calls    []dto.PisoSearchParams
}

func (s *fakePisoSearcher) Search(_ context.Context, params dto.PisoSearchParams) (json.RawMessage, error) {
	s.calls = append(s.calls, params)
	return s.response, nil
}
