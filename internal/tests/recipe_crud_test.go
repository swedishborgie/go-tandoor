//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/recipe"
)

func TestIntegrationRecipeCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)

	ctx := context.Background()

	// Create a minimal recipe
	r := &recipe.Recipe{
		Name:     "Integration Test Recipe",
		Servings: 2,
		Steps:    []recipe.Step{},
	}
	created, err := client.Recipes().Create(ctx, r)
	if err != nil {
		// try to get more details
		t.Logf("Create error: %v", err)
	}
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, "Integration Test Recipe", created.Name)
	require.Greater(t, created.ID, 0)

	// List recipes and ensure it's present
	list, err := client.Recipes().List(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, list)
	found := false
	for _, rec := range list.Results {
		if rec.ID == created.ID {
			found = true
			break
		}
	}
	require.True(t, found, "created recipe should appear in list")

	// Get by ID
	got, err := client.Recipes().Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, "Integration Test Recipe", got.Name)

	// Update name
	got.Name = "Integration Test Recipe Updated"
	updated, err := client.Recipes().Update(ctx, got)
	require.NoError(t, err)
	require.Equal(t, "Integration Test Recipe Updated", updated.Name)

	// Delete
	err = client.Recipes().Delete(ctx, updated.ID)
	require.NoError(t, err)

	// Verify deletion – Get should error
	_, err = client.Recipes().Get(ctx, updated.ID)
	require.Error(t, err)
}
