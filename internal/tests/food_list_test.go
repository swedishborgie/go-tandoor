//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrationFoodList(t *testing.T) {
	client := Client()
	require.NotNil(t, client)

	ctx := context.Background()
	foods, err := client.Foods().List(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, foods)
	// Should be able to list without error, even if empty
}
