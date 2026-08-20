package vinamap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wwetd-server/internal/config"
	"wwetd-server/internal/dto"
)

func TestClientUsesMapVinaEndpointsAndNormalizesResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")

		switch request.URL.Path {
		case searchPath:
			if request.URL.Query().Get("location") != "10.776889,106.700806" {
				t.Fatalf("unexpected location: %q", request.URL.Query().Get("location"))
			}
			if request.URL.Query().Get("radius") != "3000" || request.URL.Query().Get("key") != "secret" {
				t.Fatalf("missing radius or API key")
			}
			_, _ = response.Write([]byte(`{
				"status":"OK",
				"results":[
					{"place_id":"place-a","name":"Quán A","formatted_address":"Quận 1","geometry":{"location":{"lat":10.7,"lng":106.7}}},
					{"place_id":"place-b","name":"Quán B"}
				]
			}`))
		case placePath:
			if request.URL.Query().Get("place_id") != "place-a" {
				t.Fatalf("unexpected place_id: %q", request.URL.Query().Get("place_id"))
			}
			_, _ = response.Write([]byte(`{
				"status":"OK",
				"result":{
					"place_id":"place-a",
					"name":"Quán A",
					"formatted_address":"Quận 1",
					"photos":[{"url":"https://example.com/a.jpg"}],
					"opening_hours":{"open_now":true}
				}
			}`))
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	client := NewClient(config.VinamapConfig{
		APIKey:    "secret",
		BaseURL:   server.URL,
		Timeout:   time.Second,
		Radius:    3000,
		PlaceType: "restaurant",
	})

	searchPayload, err := client.Search(context.Background(), dto.MapSearchParams{
		Lat: 10.776889, Lng: 106.700806, Limit: 1,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	var search struct {
		Results []map[string]interface{} `json:"results"`
	}
	if err := json.Unmarshal(searchPayload, &search); err != nil {
		t.Fatalf("decode normalized search: %v", err)
	}
	if len(search.Results) != 1 || search.Results[0]["data_id"] != "place-a" || search.Results[0]["title"] != "Quán A" {
		t.Fatalf("unexpected normalized search response: %s", searchPayload)
	}

	placePayload, err := client.Place(context.Background(), dto.MapPlaceParams{DataID: "place-a"})
	if err != nil {
		t.Fatalf("Place returned error: %v", err)
	}
	var place map[string]interface{}
	if err := json.Unmarshal(placePayload, &place); err != nil {
		t.Fatalf("decode normalized place: %v", err)
	}
	if place["data_id"] != "place-a" || place["title"] != "Quán A" {
		t.Fatalf("unexpected normalized place response: %s", placePayload)
	}
	if place["open_state"] == nil || place["photos"] == nil {
		t.Fatalf("expected normalized opening state and photos: %s", placePayload)
	}
}
