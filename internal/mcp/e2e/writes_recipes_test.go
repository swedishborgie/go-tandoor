//go:build integration
// +build integration

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2EShoppingWriteCycle(t *testing.T) {
	c := newMCPClient(t)

	list := callTool(t, c, "shopping_list_create", map[string]any{"name": "e2e-list-" + runSuffix(), "color": "#123456"})
	listID := int(list["id"].(float64))
	require.Greater(t, listID, 0)

	// entry with a food
	food := callTool(t, c, "food_create", map[string]any{"name": "e2e-shop-food-" + runSuffix()})
	foodID := int(food["id"].(float64))

	entry := callTool(t, c, "shopping_entry_create", map[string]any{
		"food_id":  foodID,
		"amount":   1.5,
		"list_ids": []any{listID},
	})
	entryID := int(entry["id"].(float64))
	require.Greater(t, entryID, 0)

	// bulk update: check it off
	bulk := callTool(t, c, "shopping_entry_bulk_update", map[string]any{
		"ids":     []any{entryID},
		"checked": true,
	})
	require.Contains(t, bulk, "ids")

	require.Equal(t, true, callTool(t, c, "shopping_entry_delete", map[string]any{"id": entryID})["deleted"])
	require.Equal(t, true, callTool(t, c, "shopping_list_delete", map[string]any{"id": listID})["deleted"])
	require.Equal(t, true, callTool(t, c, "food_delete", map[string]any{"id": foodID})["deleted"])
}

func TestE2ERecipeWriteCycle(t *testing.T) {
	c := newMCPClient(t)
	name := "e2e-recipe-" + runSuffix()

	created := callTool(t, c, "recipe_create", map[string]any{
		"data": map[string]any{"name": name, "servings": 3, "working_time": 15, "steps": []any{}},
	})
	id := int(created["id"].(float64))
	require.Greater(t, id, 0)
	require.Equal(t, name, created["name"])

	// patch
	patched := callTool(t, c, "recipe_patch", map[string]any{
		"id":   id,
		"data": map[string]any{"servings": 5, "description": "e2e patched"},
	})
	require.Equal(t, float64(5), patched["servings"])

	// batch update
	batch := callTool(t, c, "recipe_batch_update", map[string]any{
		"recipes":      []any{id},
		"working_time": 20,
	})
	require.Contains(t, batch, "results")

	// add to shopping (creates a shopping list recipe)
	shop := callTool(t, c, "recipe_add_to_shopping", map[string]any{"id": id, "servings": 5})
	require.Equal(t, true, shop["added"])

	// delete
	require.Equal(t, true, callTool(t, c, "recipe_delete", map[string]any{"id": id})["deleted"])
	_, isErr := callToolRaw(t, c, "recipe_get", map[string]any{"id": id})
	require.True(t, isErr, "deleted recipe should be gone")
}

func TestE2EShoppingListAddRecipe(t *testing.T) {
	c := newMCPClient(t)

	list := callTool(t, c, "shopping_list_create", map[string]any{"name": "e2e-list2-" + runSuffix()})
	listID := int(list["id"].(float64))

	food := callTool(t, c, "food_create", map[string]any{"name": "e2e-ar-food-" + runSuffix()})
	foodID := int(food["id"].(float64))

	recipe := callTool(t, c, "recipe_create", map[string]any{
		"data": map[string]any{
			"name":     "e2e-ar-recipe-" + runSuffix(),
			"servings": 2,
			"steps": []any{map[string]any{
				"instruction": "mix",
				"ingredients": []any{map[string]any{"food": foodID, "unit": nil, "amount": 4}},
			}},
		},
	})
	recipeID := int(recipe["id"].(float64))

	added := callTool(t, c, "shopping_list_add_recipe", map[string]any{
		"list_id": listID, "recipe_id": recipeID, "servings": 4,
	}) // response is the validated serializer data: {entries, shopping_lists_ids}
	created := added["entries"].([]any)
	require.Len(t, created, 1)
	// recipe servings 2, ingredient amount 4, target 4 servings → factor 2 → amount 8
	require.Equal(t, float64(8), created[0].(map[string]any)["amount"], "amount should scale with servings")

	// the entry should exist on the list
	entries := callTool(t, c, "shopping_entry_list", map[string]any{})
	require.GreaterOrEqual(t, len(entries["results"].([]any)), 1)

	// cleanup: remove created shopping entries (find ones for our food)
	for _, e := range entries["results"].([]any) {
		em := e.(map[string]any)
		if f, ok := em["food"].(map[string]any); ok && int(f["id"].(float64)) == foodID {
			_ = callTool(t, c, "shopping_entry_delete", map[string]any{"id": int(em["id"].(float64))})
		}
	}
	require.Equal(t, true, callTool(t, c, "shopping_list_delete", map[string]any{"id": listID})["deleted"])
	require.Equal(t, true, callTool(t, c, "recipe_delete", map[string]any{"id": recipeID})["deleted"])
	require.Equal(t, true, callTool(t, c, "food_delete", map[string]any{"id": foodID})["deleted"])
}
