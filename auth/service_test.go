package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

func TestAuthService_Auth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api-token-auth/", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token":"tda_abc123","scope":"read write","expires":"2027-01-01T00:00:00Z","user_id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	token, err := NewService(mock).Auth(context.Background(), "user", "pass")
	require.NoError(t, err)
	assert.Equal(t, "tda_abc123", token.Token)
	assert.Equal(t, 1, token.UserID)
}

func TestAuthService_Auth_401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"detail":"Invalid credentials."}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	_, err := NewService(mock).Auth(context.Background(), "user", "wrong")
	require.Error(t, err)
	assert.ErrorContains(t, err, "auth error 401")
}

func TestAuthService_ListAccessTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/access-token/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"token":"tda_x","expires":"2027-01-01T00:00:00Z","scope":"read","created":"2025-01-01T00:00:00Z","updated":"2025-01-01T00:00:00Z"}]`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	tokens, err := NewService(mock).ListAccessTokens(context.Background())
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, "tda_x", tokens[0].Token)
}

func TestAuthService_CreateAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":2,"token":"tda_new","expires":"2027-01-01T00:00:00Z","scope":"read write","created":"2025-01-01T00:00:00Z","updated":"2025-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	token, err := NewService(mock).CreateAccessToken(context.Background(), &AccessToken{Scope: "read write"})
	require.NoError(t, err)
	assert.Equal(t, "tda_new", token.Token)
}

func TestAuthService_GetAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/access-token/5/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":5,"token":"tda_five","expires":"2027-01-01T00:00:00Z","scope":"read","created":"2025-01-01T00:00:00Z","updated":"2025-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	token, err := NewService(mock).GetAccessToken(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, "tda_five", token.Token)
}

func TestAuthService_DeleteAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/access-token/5/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).DeleteAccessToken(context.Background(), 5)
	require.NoError(t, err)
}
