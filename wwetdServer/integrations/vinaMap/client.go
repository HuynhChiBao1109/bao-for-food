package vinamap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"wwetd-server/internal/config"
	"wwetd-server/internal/dto"
	"wwetd-server/internal/interfaces"
)

const (
	searchPath = "/api/v2/place/nearbysearch/json"
	placePath  = "/api/v2/place/details/json"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	radius     int
	placeType  string
}

var _ interfaces.NamedMapClient = (*Client)(nil)

func NewClient(cfg config.VinamapConfig) *Client {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, searchPath)
	baseURL = strings.TrimSuffix(baseURL, placePath)

	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		radius:     cfg.Radius,
		placeType:  cfg.PlaceType,
	}
}

func (c *Client) ProviderName() string {
	return "vinamap"
}

func (c *Client) Search(ctx context.Context, params dto.MapSearchParams) (json.RawMessage, error) {
	endpoint, err := c.buildURL(searchPath)
	if err != nil {
		return nil, fmt.Errorf("parse vinamap search url: %w", err)
	}

	query := endpoint.Query()
	query.Set("location", fmt.Sprintf("%.6f,%.6f", params.Lat, params.Lng))
	query.Set("radius", strconv.Itoa(c.radius))
	query.Set("key", c.apiKey)
	query.Set("new_admin", "true")
	query.Set("include_old_admin", "true")
	if c.placeType != "" {
		query.Set("type", c.placeType)
	}
	endpoint.RawQuery = query.Encode()

	payload, err := c.getJSON(ctx, endpoint.String(), "search")
	if err != nil {
		return nil, err
	}

	return normalizeSearchResponse(payload, params.Limit)
}

func (c *Client) Place(ctx context.Context, params dto.MapPlaceParams) (json.RawMessage, error) {
	endpoint, err := c.buildURL(placePath)
	if err != nil {
		return nil, fmt.Errorf("parse vinamap place url: %w", err)
	}

	query := endpoint.Query()
	query.Set("place_id", params.DataID)
	query.Set("key", c.apiKey)
	query.Set("new_admin", "true")
	query.Set("include_old_admin", "true")
	endpoint.RawQuery = query.Encode()

	payload, err := c.getJSON(ctx, endpoint.String(), "place")
	if err != nil {
		return nil, err
	}

	return normalizePlaceResponse(payload)
}

func (c *Client) buildURL(path string) (*url.URL, error) {
	return url.Parse(c.baseURL + path)
}

func (c *Client) getJSON(ctx context.Context, endpoint string, operation string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create vinamap %s request: %w", operation, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call vinamap %s: %w", operation, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("read vinamap %s response: %w", operation, err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("vinamap %s returned status %d: %s", operation, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if !json.Valid(body) {
		return nil, fmt.Errorf("vinamap %s returned invalid json", operation)
	}

	var status struct {
		Status       string `json:"status"`
		ErrorMessage string `json:"error_message"`
	}
	if err := json.Unmarshal(body, &status); err == nil && status.Status != "" && status.Status != "OK" && status.Status != "ZERO_RESULTS" {
		return nil, fmt.Errorf("vinamap %s returned %s: %s", operation, status.Status, status.ErrorMessage)
	}

	return json.RawMessage(body), nil
}

func normalizeSearchResponse(payload json.RawMessage, limit int) (json.RawMessage, error) {
	var response struct {
		Results []map[string]interface{} `json:"results"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, fmt.Errorf("decode vinamap search response: %w", err)
	}

	if limit > 0 && len(response.Results) > limit {
		response.Results = response.Results[:limit]
	}
	for _, place := range response.Results {
		normalizePlace(place)
	}

	normalized, err := json.Marshal(map[string]interface{}{"results": response.Results})
	if err != nil {
		return nil, fmt.Errorf("encode vinamap search response: %w", err)
	}
	return normalized, nil
}

func normalizePlaceResponse(payload json.RawMessage) (json.RawMessage, error) {
	var response struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, fmt.Errorf("decode vinamap place response: %w", err)
	}
	if response.Result == nil {
		return nil, fmt.Errorf("vinamap place response is missing result")
	}

	normalizePlace(response.Result)
	normalized, err := json.Marshal(response.Result)
	if err != nil {
		return nil, fmt.Errorf("encode vinamap place response: %w", err)
	}
	return normalized, nil
}

func normalizePlace(place map[string]interface{}) {
	placeID, _ := place["place_id"].(string)
	name, _ := place["name"].(string)
	formattedAddress, _ := place["formatted_address"].(string)

	if place["data_id"] == nil {
		place["data_id"] = placeID
	}
	if place["title"] == nil {
		place["title"] = name
	}

	location := map[string]interface{}{
		"address": map[string]interface{}{"full": formattedAddress},
	}
	if geometry, ok := place["geometry"].(map[string]interface{}); ok {
		if coordinates, ok := geometry["location"].(map[string]interface{}); ok {
			location["lat"] = coordinates["lat"]
			location["lng"] = coordinates["lng"]
		}
	}
	place["location"] = location

	if place["type"] == nil {
		if types, ok := place["types"].([]interface{}); ok && len(types) > 0 {
			place["type"] = types[0]
		}
	}
	if place["reviews"] == nil && place["user_ratings_total"] != nil {
		place["reviews"] = place["user_ratings_total"]
	}
	if place["link_google_maps"] == nil && place["url"] != nil {
		place["link_google_maps"] = place["url"]
	}

	normalizePhotos(place)
	normalizeOpeningState(place)
}

func normalizePhotos(place map[string]interface{}) {
	photos, ok := place["photos"].([]interface{})
	if !ok || len(photos) == 0 {
		return
	}

	media := make([]interface{}, 0, len(photos))
	for _, photo := range photos {
		if item, ok := photo.(map[string]interface{}); ok && item["url"] != nil {
			media = append(media, item)
		}
	}
	if len(media) > 0 {
		place["photos"] = []interface{}{map[string]interface{}{"media": media}}
	}
}

func normalizeOpeningState(place map[string]interface{}) {
	openingHours, ok := place["opening_hours"].(map[string]interface{})
	if !ok {
		return
	}

	openNow, ok := openingHours["open_now"].(bool)
	if !ok {
		return
	}

	text := "Closed"
	if openNow {
		text = "Open"
	}
	place["open_state"] = map[string]interface{}{
		"is_open_now": openNow,
		"text":        text,
	}
}

func DefaultTimeout() time.Duration {
	return 8 * time.Second
}
