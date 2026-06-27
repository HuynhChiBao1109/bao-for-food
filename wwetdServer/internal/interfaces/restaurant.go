package interfaces

import (
	"context"
	"encoding/json"

	"wwetd-server/internal/dto"
)

type PisoSearcher interface {
	Search(ctx context.Context, params dto.PisoSearchParams) (json.RawMessage, error)
	Place(ctx context.Context, params dto.PisoPlaceParams) (json.RawMessage, error)
}

type RestaurantService interface {
	SearchNearby(ctx context.Context, query dto.NearbyRestaurantsQuery) (dto.NearbyRestaurantsResponse, error)
	PickNearby(ctx context.Context, query dto.NearbyRestaurantsQuery) (dto.PickRestaurantResponse, error)
	GetDetail(ctx context.Context, userID string, dataID string, location dto.ClientLocation) (dto.RestaurantDetailResponse, error)
	RecordViewed(ctx context.Context, userID string, dataID string, location dto.ClientLocation) error
	ListViewed(ctx context.Context, userID string) ([]dto.UserRestaurantItem, error)
	SaveRestaurant(ctx context.Context, userID string, dataID string, location dto.ClientLocation) error
	ListSaved(ctx context.Context, userID string) ([]dto.UserRestaurantItem, error)
}
