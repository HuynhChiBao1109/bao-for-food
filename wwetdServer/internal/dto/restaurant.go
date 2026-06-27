package dto

import "encoding/json"

type NearbyRestaurantsQuery struct {
	IP     string
	UserID string
	Query  string
	Limit  int
	Lat    float64
	Lng    float64
	HasLat bool
	HasLng bool
}

type ClientLocation struct {
	IP  string  `json:"ip"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type PisoSearchParams struct {
	Query string
	Lat   float64
	Lng   float64
	Limit int
}

type PisoPlaceParams struct {
	DataID string
	Lat    float64
	Lng    float64
}

type NearbyRestaurantsResponse struct {
	Location ClientLocation  `json:"location"`
	Source   string          `json:"source"`
	Results  json.RawMessage `json:"results"`
}

type PickRestaurantResponse struct {
	Location     ClientLocation  `json:"location"`
	Source       string          `json:"source"`
	DetailSource string          `json:"detail_source"`
	Restaurant   json.RawMessage `json:"restaurant"`
	IsSaved      bool            `json:"is_saved,omitempty"`
}

type RestaurantDetailResponse struct {
	DetailSource string          `json:"detail_source"`
	Restaurant   json.RawMessage `json:"restaurant"`
	IsSaved      bool            `json:"is_saved,omitempty"`
}

type UserRestaurantItem struct {
	DataID     string          `json:"data_id"`
	Detail     json.RawMessage `json:"detail"`
	RecordedAt string          `json:"recorded_at"`
}
