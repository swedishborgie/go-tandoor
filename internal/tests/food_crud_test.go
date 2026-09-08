//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/food"
)

func TestIntegrationFoodCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)

	ctx := context.Background()

	// Create food
	f := &food.Food{
		Name: "Integration Test Food",
	}
	created, err := client.Foods().Create(ctx, f)
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, "Integration Test Food", created.Name)
	require.Greater(t, created.ID, 0)

	// Get
	got, err := client.Foods().Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)

	// Update
	got.Name = "Integration Test Food Updated"
	updated, err := client.Foods().Update(ctx, got)
	require.NoError(t, err)
	require.Equal(t, "Integration Test Food Updated", updated.Name)

	// Delete
	err = client.Foods().Delete(ctx, updated.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = client.Foods().Get(ctx, updated.ID)
	require.Error(t, err)
}
