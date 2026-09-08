// Package fdc provides a Go client for the USDA FoodData Central API.
//
// It covers all public endpoints: food lookup by ID, batch food retrieval,
// paged food listing, and keyword-based food search.
//
// # Quick Start
//
//	client, err := fdc.NewClient(
//		fdc.WithAPIKey("your-api-key"),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	food, err := client.GetFood(context.Background(), 534358, fdc.FormatFull, nil)
package fdc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	// DefaultBaseURL is the base URL for the FDC API.
	DefaultBaseURL = "https://api.nal.usda.gov/fdc"

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
)

// Client is the FDC API client.
type Client struct {
	apiKey     string
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new FDC API client.
//
// apiKey is required and should be obtained from https://fdc.nal.usda.gov/.
// Additional configuration can be passed via functional options.
func NewClient(opts ...ClientOption) (*Client, error) {
	u, _ := url.Parse(DefaultBaseURL)
	c := &Client{
		baseURL: u,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.apiKey == "" {
		return nil, fmt.Errorf("fdc: API key is required")
	}

	return c, nil
}

// do executes an HTTP request and decodes the JSON response into v.
//
// The apiKey is appended as a query parameter on every request.
// If the response status is not in the 2xx range, an *Error is returned.
func (c *Client) do(ctx context.Context, method string, subPath string, body any, v any) error {
	// Build URL.
	subPath, query := splitQuery(subPath)
	p := path.Join(c.baseURL.Path, subPath)

	q := url.Values{}
	if query != "" {
		parsed, err := url.ParseQuery(query)
		if err == nil {
			for k, vals := range parsed {
				for _, v := range vals {
					q.Add(k, v)
				}
			}
		}
	}
	q.Set("api_key", c.apiKey)

	u := &url.URL{
		Scheme:   c.baseURL.Scheme,
		Host:     c.baseURL.Host,
		Path:     p,
		RawQuery: q.Encode(),
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("fdc: marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("fdc: create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fdc: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("fdc: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &Error{
			StatusCode: resp.StatusCode,
			Body:       respBody,
		}
	}

	if v != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, v); err != nil {
			return fmt.Errorf("fdc: unmarshal response: %w", err)
		}
	}

	return nil
}

// splitQuery splits a path that may contain a query string into
// the path prefix and the raw query string.
func splitQuery(raw string) (pathStr, query string) {
	idx := len(raw)
	for i, r := range raw {
		if r == '?' {
			idx = i
			break
		}
	}
	if idx == len(raw) {
		return raw, ""
	}
	return raw[:idx], raw[idx+1:]
}
