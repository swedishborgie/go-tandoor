package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/swedishborgie/go-tandoor/internal/executor"
)

// Service provides access to Auth API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new AuthService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// Error represents an authentication error returned by the API.
type Error struct {
	StatusCode int
	Body       []byte
}

func (e *Error) Error() string {
	if len(e.Body) > 0 {
		return fmt.Sprintf("auth error %d: %s", e.StatusCode, string(e.Body))
	}
	return fmt.Sprintf("auth error %d", e.StatusCode)
}

// Auth authenticates with the /api-token-auth/ endpoint.
//
// This method bypasses the standard request flow because it uses
// form-encoded data without a Bearer token.
func (s *Service) Auth(ctx context.Context, username, password string) (*Token, error) {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	resp, err := s.exec.DoRawNoAuth(ctx, "POST", "api-token-auth/", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("tandoor: auth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tandoor: read auth response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &Error{
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("tandoor: unmarshal auth response: %w", err)
	}
	return &token, nil
}

// Auth401 authenticates and returns a clean 401 error on failure.
func (s *Service) Auth401(ctx context.Context, username, password string) error {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	resp, err := s.exec.DoRawNoAuth(ctx, "POST", "api-token-auth/", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("tandoor: auth401 request failed: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("tandoor: expected 401, got %d", resp.StatusCode)
	}
	return nil
}

// ListAccessTokens lists all OAuth2 access tokens.
//
// Note: this endpoint returns a plain array, not a paginated response.
func (s *Service) ListAccessTokens(ctx context.Context) ([]AccessToken, error) {
	var tokens []AccessToken
	if err := s.exec.DoJSON(ctx, "GET", "api/access-token/", nil, &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

// CreateAccessToken creates a new OAuth2 access token.
func (s *Service) CreateAccessToken(ctx context.Context, token *AccessToken) (*AccessToken, error) {
	var result AccessToken
	if err := s.exec.DoJSON(ctx, "POST", "api/access-token/", token, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAccessToken retrieves an access token by ID.
func (s *Service) GetAccessToken(ctx context.Context, id int) (*AccessToken, error) {
	var token AccessToken
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/access-token/%d/", id), nil, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteAccessToken removes an access token by ID.
func (s *Service) DeleteAccessToken(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/access-token/%d/", id), nil, nil)
}
