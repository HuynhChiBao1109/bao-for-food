package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"wwetd-server/internal/dto"
	"wwetd-server/internal/interfaces"
)

type mapClientStrategy struct {
	client interfaces.NamedMapClient
}

var _ interfaces.NamedMapClient = (*mapClientStrategy)(nil)

func NewMapClient(provider string, clients ...interfaces.NamedMapClient) (interfaces.NamedMapClient, error) {
	requestedProvider := normalizeProviderName(provider)

	for _, client := range clients {
		if client != nil && normalizeProviderName(client.ProviderName()) == requestedProvider {
			return &mapClientStrategy{client: client}, nil
		}
	}

	return nil, fmt.Errorf("unsupported map provider %q; expected piso or vinamap", provider)
}

func (s *mapClientStrategy) ProviderName() string {
	return normalizeProviderName(s.client.ProviderName())
}

func (s *mapClientStrategy) Search(ctx context.Context, params dto.MapSearchParams) (json.RawMessage, error) {
	return s.client.Search(ctx, params)
}

func (s *mapClientStrategy) Place(ctx context.Context, params dto.MapPlaceParams) (json.RawMessage, error) {
	return s.client.Place(ctx, params)
}

func normalizeProviderName(value string) string {
	name := strings.ToLower(strings.TrimSpace(value))
	switch name {
	case "vina", "vina-map", "vina_map", "mapvina":
		return "vinamap"
	default:
		return name
	}
}
