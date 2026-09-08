//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/shopping"
)

func TestIntegrationShoppingRecipeCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create a recipe with minimal steps + ingredients via raw payload
	recipePayload := map[string]any{
		"name": "Shopping Recipe Test",
		"steps": []any{
			map[string]any{
				"instruction": "Step 1",
				"ingredients": []any{
					map[string]any{
						"amount": 1.0,
						"food":   map[string]any{"name": "Test Food"},
						"unit":   map[string]any{"name": "Test Unit"},
					},
				},
			},
		},
	}
	var createdRec struct {
		ID int `json:"id"`
	}
	err := client.DoJSON(ctx, "POST", "api/recipe/", recipePayload, &createdRec)
	require.NoError(t, err)
	require.NotZero(t, createdRec.ID)
	defer func() { _ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/recipe/%d/", createdRec.ID), nil, nil) }()

	// Create shopping list recipe via raw payload
	payload := map[string]any{
		"recipe":   createdRec.ID,
		"servings": 2,
	}
	var createdSR shopping.ListRecipe
	err = client.DoJSON(ctx, "POST", "api/shopping-list-recipe/", payload, &createdSR)
	require.NoError(t, err)
	require.NotZero(t, createdSR.ID)

	url := fmt.Sprintf("api/shopping-list-recipe/%d/", createdSR.ID)

	// Get
	var got shopping.ListRecipe
	err = client.DoJSON(ctx, "GET", url, nil, &got)
	require.NoError(t, err)
	require.Equal(t, createdSR.ID, got.ID)

	// Update servings
	updatePayload := map[string]any{"servings": 3}
	var updated shopping.ListRecipe
	err = client.DoJSON(ctx, "PATCH", url, updatePayload, &updated)
	require.NoError(t, err)
	require.InDelta(t, 3.0, updated.Servings, 0.001)

	// Cleanup
	_ = client.DoJSON(ctx, "DELETE", url, nil, nil)
}
