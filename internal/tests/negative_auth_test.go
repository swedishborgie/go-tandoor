//go:build integration
// +build integration

package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrationNegativeAuth(t *testing.T) {
	ctx := context.Background()
	// Unauthenticated POST should be rejected
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:8080/api/recipe/", nil)
	require.NoError(t, err)
	// No Authorization header
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	// Expect 401 Unauthorized or 403 Forbidden
	require.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden, "expected 401/403, got %d", resp.StatusCode)
}
