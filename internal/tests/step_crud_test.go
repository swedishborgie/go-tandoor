//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrationStepCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create a recipe with a step in one go
	recipePayload := map[string]any{
		"name": "Step Test Recipe",
		"steps": []any{
			map[string]any{
				"order":       1,
				"instruction": "Test step text",
				"ingredients": []any{},
			},
		},
	}
	var recipe struct {
		ID int `json:"id"`
	}
	err := client.DoJSON(ctx, "POST", "api/recipe/", recipePayload, &recipe)
	require.NoError(t, err)
	require.NotZero(t, recipe.ID)
	defer func() { _ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/recipe/%d/", recipe.ID), nil, nil) }()

	// Get recipe and verify step exists
	var gotRecipe struct {
		ID    int `json:"id"`
		Steps []struct {
			ID          int    `json:"id"`
			Instruction string `json:"instruction"`
		} `json:"steps"`
	}
	err = client.DoJSON(ctx, "GET", fmt.Sprintf("api/recipe/%d/", recipe.ID), nil, &gotRecipe)
	require.NoError(t, err)
	require.NotEmpty(t, gotRecipe.Steps)
	stepID := gotRecipe.Steps[0].ID
	require.Equal(t, "Test step text", gotRecipe.Steps[0].Instruction)

	// Update step via recipe patch
	updateRecipePayload := map[string]any{
		"steps": []any{
			map[string]any{
				"id":          stepID,
				"order":       1,
				"instruction": "Updated step text",
				"ingredients": []any{},
			},
		},
	}
	var updatedRecipe struct {
		ID int `json:"id"`
	}
	err = client.DoJSON(ctx, "PATCH", fmt.Sprintf("api/recipe/%d/", recipe.ID), updateRecipePayload, &updatedRecipe)
	require.NoError(t, err)

	// Verify update
	var verify struct {
		Steps []struct {
			Instruction string `json:"instruction"`
		} `json:"steps"`
	}
	err = client.DoJSON(ctx, "GET", fmt.Sprintf("api/recipe/%d/", recipe.ID), nil, &verify)
	require.NoError(t, err)
	require.Equal(t, "Updated step text", verify.Steps[0].Instruction)
}
