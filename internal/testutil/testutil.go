// Package testutil provides helpers for testing service packages.
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// MockExecutor implements executor.Executor using an *httptest.Server.
type MockExecutor struct {
	serverURL string
}

// DoJSON executes a JSON request against the mock server.
func (m *MockExecutor) DoJSON(ctx context.Context, method, path string, body any, v any) error {
	subPath, query := splitQuery(path)
	p := "/"
	if subPath != "" {
		p = "/" + subPath
	}
	if !hasTrailingSlash(p) {
		p += "/"
	}
	u := m.serverURL + p
	if query != "" {
		u += "?" + query
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("mock: marshal body: %w", err)
		}
		bodyReader = io.NopCloser(strings.NewReader(string(data)))
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("mock: create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("mock: request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("mock: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		mockErr := &MockError{
			StatusCode: resp.StatusCode,
			Body:       raw,
		}
		// Attempt to parse DRF error for Message.
		var rawMap map[string]any
		if json.Unmarshal(raw, &rawMap) == nil {
			if d, ok := rawMap["detail"]; ok {
				if ds, ok := d.(string); ok {
					mockErr.Message = ds
				}
			} else if nfe, ok := rawMap["non_field_errors"]; ok {
				if arr, ok := nfe.([]any); ok && len(arr) > 0 {
					if s, ok := arr[0].(string); ok {
						mockErr.Message = s
					}
				}
			}
		}
		return mockErr
	}

	if v != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, v); err != nil {
			return fmt.Errorf("mock: unmarshal response: %w", err)
		}
	}

	return nil
}

// DoRaw executes a raw request against the mock server.
func (m *MockExecutor) DoRaw(ctx context.Context, method, path string, body any) (*http.Response, error) {
	subPath, query := splitQuery(path)
	p := "/"
	if subPath != "" {
		p = "/" + subPath
	}
	if !hasTrailingSlash(p) {
		p += "/"
	}
	u := m.serverURL + p
	if query != "" {
		u += "?" + query
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = body.(io.Reader)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("mock: create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return http.DefaultClient.Do(req)
}

// DoRawNoAuth executes a raw request without auth.
func (m *MockExecutor) DoRawNoAuth(ctx context.Context, method, path string, body any) (*http.Response, error) {
	return m.DoRaw(ctx, method, path, body)
}

// BaseURLOrigin returns the mock server origin.
func (m *MockExecutor) BaseURLOrigin() string {
	u, _ := url.Parse(m.serverURL)
	return u.Scheme + "://" + u.Host
}

// MockError mimics tandoor.TandoorError for testing.
type MockError struct {
	StatusCode int
	Body       []byte
	Message    string
}

func (e *MockError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = string(e.Body)
	}
	return fmt.Sprintf("mock: API error %d: %s", e.StatusCode, msg)
}

// IsNotFound reports if the error is not found.
func (e *MockError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsForbidden reports if the error is forbidden.
func (e *MockError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsUnauthorized reports if the error is unauthorized.
func (e *MockError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// Details returns parsed field errors (mock implementation).
func (e *MockError) Details() map[string][]string {
	return nil
}

// NonFieldErrors returns non-field errors (mock implementation).
func (e *MockError) NonFieldErrors() []string {
	return nil
}

// NewMockExecutor creates a MockExecutor for the given test server URL.
func NewMockExecutor(serverURL string) *MockExecutor {
	return &MockExecutor{serverURL: serverURL}
}

func splitQuery(raw string) (path, query string) {
	idx := strings.Index(raw, "?")
	if idx == -1 {
		return raw, ""
	}
	return raw[:idx], raw[idx+1:]
}

func hasTrailingSlash(p string) bool {
	return len(p) > 0 && p[len(p)-1] == '/'
}
