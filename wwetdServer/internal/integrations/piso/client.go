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
)

const searchPath = "/api/maps/search"

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

func NewClient(cfg config.PisoConfig) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:     cfg.APIKey,
	}
}

func (c *Client) Search(ctx context.Context, params dto.PisoSearchParams) (json.RawMessage, error) {
	endpoint, err := url.Parse(c.baseURL + searchPath)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create piso search request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call piso search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("read piso search response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("piso search returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if !json.Valid(body) {
		return nil, fmt.Errorf("piso search returned invalid json")
	}

	return json.RawMessage(body), nil
}

func DefaultTimeout() time.Duration {
	return 8 * time.Second
}
