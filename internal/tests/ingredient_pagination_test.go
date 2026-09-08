//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/ingredient"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestIntegrationIngredientPagination(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create recipes with steps containing ingredients, to make ingredients visible via the ingredient list endpoint
	createdRecipeIDs := make([]int, 0, 13)
	for i := 0; i < 13; i++ {
		recipePayload := map[string]any{
			"name": fmt.Sprintf("Pag Recipe %d", i),
			"steps": []any{
				map[string]any{
					"instruction": "Step 1",
					"ingredients": []any{
						map[string]any{
							"amount": float64(i + 1),
							"food":   map[string]any{"name": fmt.Sprintf("Pag Food %d", i)},
							"unit":   map[string]any{"name": fmt.Sprintf("Pag Unit %d", i)},
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
		createdRecipeIDs = append(createdRecipeIDs, createdRec.ID)
	}
	// Cleanup recipes
	defer func() {
		for _, id := range createdRecipeIDs {
			_ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/recipe/%d/", id), nil, nil)
		}
	}()

	// List ingredients page 1 with page size 5
	opts := &ingredient.ListOptions{ListOptions: pagination.ListOptions{Page: 1, PageSize: 5}}
	page1, err := client.Ingredients().List(ctx, opts)
	require.NoError(t, err)
	require.Len(t, page1.Results, 5)
	require.True(t, page1.HasNext())

	// Page 2
	opts.Page = 2
	page2, err := client.Ingredients().List(ctx, opts)
	require.NoError(t, err)
	require.Len(t, page2.Results, 5)

	// Page 3
	opts.Page = 3
	page3, err := client.Ingredients().List(ctx, opts)
	require.NoError(t, err)
	require.Len(t, page3.Results, 3)
}
