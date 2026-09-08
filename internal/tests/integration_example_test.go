//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrationClientIsReachable(t *testing.T) {
	client := Client()
	if client == nil {
		t.Skip("integration tests not initialized")
	}
	// Simple health check via Recipes list – should return empty list, not error
	ctx := context.Background()
	recipes, err := client.Recipes().List(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, recipes)
}
