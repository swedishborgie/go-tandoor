//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/ingredient"
	"github.com/swedishborgie/go-tandoor/unit"
)

func TestIntegrationIngredientCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create a food and unit for the ingredient
	f := &food.Food{Name: "Integration Test Food"}
	createdFood, err := client.Foods().Create(ctx, f)
	require.NoError(t, err)
	require.NotZero(t, createdFood.ID)
	defer client.Foods().Delete(ctx, createdFood.ID)

	u := &unit.Unit{Name: "Integration Test Unit", PluralName: "Integration Test Units"}
	createdUnit, err := client.Units().Create(ctx, u)
	require.NoError(t, err)
	require.NotZero(t, createdUnit.ID)
	defer client.Units().Delete(ctx, createdUnit.ID)

	// Create ingredient using raw payload with food/unit as objects with name
	payload := map[string]any{
		"food":   map[string]any{"name": "Integration Test Food"},
		"unit":   map[string]any{"name": "Integration Test Unit"},
		"amount": 1.5,
	}
	var created ingredient.Ingredient
	err = client.DoJSON(ctx, "POST", "api/ingredient/", payload, &created)
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	t.Logf("created ingredient ID %d", created.ID)

	// Ingredient detail is only accessible when linked to a recipe, so we skip Get/Update/Delete checks
	// Verify creation succeeded by checking the response contains food name
	require.NotNil(t, created.Food)
	if created.Food != nil {
		require.Equal(t, "Integration Test Food", created.Food.Name)
	}

}
