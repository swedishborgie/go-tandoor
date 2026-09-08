//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrationRecipeBookCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create RecipeBook
	payload := map[string]any{
		"name":   "Test RecipeBook",
		"shared": []any{},
	}
	var created struct {
		ID int `json:"id"`
	}
	err := client.DoJSON(ctx, "POST", "api/recipe-book/", payload, &created)
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	defer func() { _ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/recipe-book/%d/", created.ID), nil, nil) }()

	// Get RecipeBook
	var got struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err = client.DoJSON(ctx, "GET", fmt.Sprintf("api/recipe-book/%d/", created.ID), nil, &got)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, "Test RecipeBook", got.Name)

	// Update RecipeBook
	updatePayload := map[string]any{
		"name": "Updated RecipeBook",
	}
	var updated struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err = client.DoJSON(ctx, "PATCH", fmt.Sprintf("api/recipe-book/%d/", created.ID), updatePayload, &updated)
	require.NoError(t, err)
	require.Equal(t, created.ID, updated.ID)
	require.Equal(t, "Updated RecipeBook", updated.Name)

	// List RecipeBooks
	var list struct {
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
	}
	err = client.DoJSON(ctx, "GET", "api/recipe-book/", nil, &list)
	require.NoError(t, err)
	found := false
	for _, r := range list.Results {
		if r.ID == created.ID {
			found = true
			break
		}
	}
	require.True(t, found)
}
