package strategy

import (
	"context"
	"encoding/json"
	"testing"

	"wwetd-server/internal/dto"
)

func TestNewMapClientSelectsProviderAndAliases(t *testing.T) {
	piso := &fakeMapClient{name: "piso"}
	vina := &fakeMapClient{name: "vinamap"}

	client, err := NewMapClient("mapvina", piso, vina)
	if err != nil {
		t.Fatalf("NewMapClient returned error: %v", err)
	}
	if client.ProviderName() != "vinamap" {
		t.Fatalf("expected vinamap provider, got %q", client.ProviderName())
	}

	if _, err := client.Search(context.Background(), dto.MapSearchParams{}); err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if vina.searchCalls != 1 || piso.searchCalls != 0 {
		t.Fatalf("expected only vinamap to receive the search call")
	}
}

func TestNewMapClientRejectsUnknownProvider(t *testing.T) {
	_, err := NewMapClient("unknown", &fakeMapClient{name: "piso"})
	if err == nil {
		t.Fatal("expected an unsupported provider error")
	}
}

type fakeMapClient struct {
	name        string
	searchCalls int
}

func (c *fakeMapClient) ProviderName() string {
	return c.name
}

func (c *fakeMapClient) Search(context.Context, dto.MapSearchParams) (json.RawMessage, error) {
	c.searchCalls++
	return json.RawMessage(`{"results":[]}`), nil
}

func (c *fakeMapClient) Place(context.Context, dto.MapPlaceParams) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
