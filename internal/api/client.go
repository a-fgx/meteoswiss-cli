// Package api provides a client for the unofficial MeteoSwiss app backend API.
//
// The API is reverse-engineered from the official MeteoSwiss iOS/Android app.
// It is not an officially documented public API; use at your own risk.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	baseURL   = "https://app-prod-ws.meteoswiss-app.ch/v1"
	userAgent = "meteocli/0.1 (github.com/user/meteocli)"
)

// Client is an HTTP client for the MeteoSwiss app API.
type Client struct {
	http    *http.Client
	baseURL string
}

// New creates a new Client with a sensible default timeout.
func New() *Client {
	return &Client{
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: baseURL,
	}
}

// PLZDetail fetches current weather and the 10-day forecast for a Swiss
// postal code. The API expects a 6-digit PLZ (e.g. 8000 → 800000).
func (c *Client) PLZDetail(plz int) (*PLZDetail, error) {
	plz6 := plz6(plz)
	url := fmt.Sprintf("%s/plzDetail?plz=%d", c.baseURL, plz6)
	var result PLZDetail
	if err := c.get(url, &result); err != nil {
		return nil, fmt.Errorf("fetching PLZ detail for %d: %w", plz, err)
	}
	return &result, nil
}

// get performs a GET request and JSON-decodes the response body into dst.
func (c *Client) get(url string, dst any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP %d from %s", resp.StatusCode, url)
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

// plz6 converts a 4-digit Swiss postal code to the 6-digit format the API
// expects (e.g. 8000 → 800000, 3012 → 301200).
func plz6(plz int) int {
	return plz * 100
}
