package fdc

import (
	"fmt"
	"net/http"
	"net/url"
)

// ClientOption configures a Client via the functional options pattern.
type ClientOption func(*Client) error

// WithAPIKey sets the API key used for authentication.
//
// The key is obtained from https://api.nal.usda.gov/user/register
// and is sent as the "api_key" query parameter on every request.
func WithAPIKey(key string) ClientOption {
	return func(c *Client) error {
		if key == "" {
			return fmt.Errorf("fdc: API key is required")
		}
		c.apiKey = key
		return nil
	}
}

// WithHTTPClient sets a custom http.Client for all HTTP requests.
//
// This allows callers to configure timeouts, TLS settings, transport
// options, or middleware (e.g., retry logic).
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("fdc: http.Client must not be nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithBaseURL overrides the API base URL.
//
// Only needed for testing or non-standard API mirrors.
// The default is "https://api.nal.usda.gov/fdc".
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		if baseURL == "" {
			return fmt.Errorf("fdc: base URL is required")
		}
		u, err := url.Parse(baseURL)
		if err != nil {
			return fmt.Errorf("fdc: invalid base URL %q: %w", baseURL, err)
		}
		c.baseURL = u
		return nil
	}
}
