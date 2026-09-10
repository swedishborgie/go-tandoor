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

func TestAuth401_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api-token-auth/", r.URL.Path)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	svc := NewService(testutil.NewMockExecutor(server.URL))
	require.NoError(t, svc.Auth401(context.Background(), "user", "wrongpass"))
}

func TestAuth401_Non401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"auth_token":"x"}`))
	}))
	t.Cleanup(server.Close)

	svc := NewService(testutil.NewMockExecutor(server.URL))
	err := svc.Auth401(context.Background(), "user", "pass")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected 401")
}
