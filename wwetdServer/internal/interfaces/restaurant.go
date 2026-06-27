package interfaces

import (
	"context"
	"encoding/json"

	"wwetd-server/internal/dto"
)

type PisoSearcher interface {
	Search(ctx context.Context, params dto.PisoSearchParams) (json.RawMessage, error)
}

type RestaurantService interface {
	SearchNearby(ctx context.Context, query dto.NearbyRestaurantsQuery) (dto.NearbyRestaurantsResponse, error)
	PickNearby(ctx context.Context, query dto.NearbyRestaurantsQuery) (dto.PickRestaurantResponse, error)
}
