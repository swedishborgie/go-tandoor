//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/recipe"
)

func TestIntegrationIngredientViaRecipe(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create a recipe with a step containing an ingredient
	recipePayload := map[string]any{
		"name": "Ingredient Via Recipe Test",
		"steps": []any{
			map[string]any{
				"instruction": "Mix ingredients",
				"ingredients": []any{
					map[string]any{
						"amount": 2.0,
						"food":   map[string]any{"name": "Via Recipe Food"},
						"unit":   map[string]any{"name": "Via Recipe Unit"},
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

	// Retrieve recipe
	var rec recipe.Recipe
	err = client.DoJSON(ctx, "GET", fmt.Sprintf("api/recipe/%d/", createdRec.ID), nil, &rec)
	require.NoError(t, err)
	require.Len(t, rec.Steps, 1)
	require.Len(t, rec.Steps[0].Ingredients, 1)
	ing := rec.Steps[0].Ingredients[0]
	require.InDelta(t, 2.0, ing.Amount, 0.001)

	// Update ingredient amount via recipe update
	ing.Amount = 3.5
	rec.Steps[0].Ingredients[0] = ing
	updated, err := client.Recipes().Update(ctx, &rec)
	require.NoError(t, err)
	require.Len(t, updated.Steps[0].Ingredients, 1)
	require.InDelta(t, 3.5, updated.Steps[0].Ingredients[0].Amount, 0.001)
}
