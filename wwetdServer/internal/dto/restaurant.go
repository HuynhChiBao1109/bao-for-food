package dto

import "encoding/json"

type NearbyRestaurantsQuery struct {
	IP     string
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

type NearbyRestaurantsResponse struct {
	Location ClientLocation  `json:"location"`
	Source   string          `json:"source"`
	Results  json.RawMessage `json:"results"`
}

type PickRestaurantResponse struct {
	Location   ClientLocation  `json:"location"`
	Source     string          `json:"source"`
	Restaurant json.RawMessage `json:"restaurant"`
}
