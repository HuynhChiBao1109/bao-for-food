package piso

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
	searchPath = "/api/maps/search"
	placePath  = "/api/maps/place"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

var _ interfaces.NamedMapClient = (*Client)(nil)

func NewClient(cfg config.PisoConfig) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:     cfg.APIKey,
	}
}

func (c *Client) ProviderName() string {
	return "piso"
}

func (c *Client) Search(ctx context.Context, params dto.PisoSearchParams) (json.RawMessage, error) {
	endpoint, err := c.buildURL(searchPath)
	if err != nil {
		return nil, fmt.Errorf("parse piso search url: %w", err)
	}

	query := endpoint.Query()
	if params.Query != "" {
		query.Set("q", params.Query)
	}
	query.Set("lat", strconv.FormatFloat(params.Lat, 'f', 6, 64))
	query.Set("lng", strconv.FormatFloat(params.Lng, 'f', 6, 64))
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	endpoint.RawQuery = query.Encode()

	return c.getJSON(ctx, endpoint.String(), "search")
}

func (c *Client) Place(ctx context.Context, params dto.PisoPlaceParams) (json.RawMessage, error) {
	endpoint, err := c.buildURL(placePath)
	if err != nil {
		return nil, fmt.Errorf("parse piso place url: %w", err)
	}

	query := endpoint.Query()
	query.Set("data_id", params.DataID)
	if params.Lat != 0 {
		query.Set("lat", strconv.FormatFloat(params.Lat, 'f', 6, 64))
	}
	if params.Lng != 0 {
		query.Set("lng", strconv.FormatFloat(params.Lng, 'f', 6, 64))
	}
	endpoint.RawQuery = query.Encode()

	return c.getJSON(ctx, endpoint.String(), "place")
}

func (c *Client) buildURL(path string) (*url.URL, error) {
	return url.Parse(c.baseURL + path)
}

func (c *Client) getJSON(ctx context.Context, endpoint string, operation string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create piso %s request: %w", operation, err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call piso %s: %w", operation, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("read piso %s response: %w", operation, err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("piso %s returned status %d: %s", operation, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if !json.Valid(body) {
		return nil, fmt.Errorf("piso %s returned invalid json", operation)
	}

	return json.RawMessage(body), nil
}

func DefaultTimeout() time.Duration {
	return 8 * time.Second
}
