//go:build integration
// +build integration

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// runSuffix returns a short unique-ish string so reruns never collide with
// leftover objects from a previously interrupted suite.
func runSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%1e9)
}

func TestE2EKeywordWriteCycle(t *testing.T) {
	c := newMCPClient(t)
	name := "e2e-keyword-" + runSuffix()

	created := callTool(t, c, "keyword_create", map[string]any{"name": name})
	id := int(created["id"].(float64))
	require.Greater(t, id, 0)
	require.Equal(t, name, created["name"])

	// update + patch (label is a read-only method field; description is writable)
	updated := callTool(t, c, "keyword_update", map[string]any{"id": id, "name": name + "-up"})
	require.Equal(t, name+"-up", updated["name"])
	patched := callTool(t, c, "keyword_patch", map[string]any{"id": id, "description": "e2e keyword desc"})
	require.Equal(t, "e2e keyword desc", patched["description"])

	// delete
	del := callTool(t, c, "keyword_delete", map[string]any{"id": id})
	require.Equal(t, true, del["deleted"])

	// gone
	_, isErr := callToolRaw(t, c, "keyword_get", map[string]any{"id": id})
	require.True(t, isErr, "deleted keyword should be gone")
}

func TestE2EFoodWriteCycle(t *testing.T) {
	c := newMCPClient(t)
	name := "e2e-food-" + runSuffix()

	created := callTool(t, c, "food_create", map[string]any{"name": name, "description": "e2e food"})
	id := int(created["id"].(float64))
	require.Greater(t, id, 0)

	patched := callTool(t, c, "food_patch", map[string]any{"id": id, "ignore_shopping": true})
	require.Equal(t, true, patched["ignore_shopping"])

	del := callTool(t, c, "food_delete", map[string]any{"id": id})
	require.Equal(t, true, del["deleted"])
}

func TestE2EUnitWriteCycle(t *testing.T) {
	c := newMCPClient(t)
	name := "e2e-unit-" + runSuffix()

	created := callTool(t, c, "unit_create", map[string]any{"name": name})
	id := int(created["id"].(float64))
	require.Greater(t, id, 0)

	// merge: create a second unit and merge into the first.
	// The merge endpoint returns {msg: ...}; success means the source is gone.
	other := callTool(t, c, "unit_create", map[string]any{"name": name + "-alt"})
	otherID := int(other["id"].(float64))
	callTool(t, c, "unit_merge", map[string]any{"source_id": otherID, "target_id": id})
	_, isErr := callToolRaw(t, c, "unit_get", map[string]any{"id": otherID})
	require.True(t, isErr, "merged source unit should be deleted")

	del := callTool(t, c, "unit_delete", map[string]any{"id": id})
	require.Equal(t, true, del["deleted"])
}

// The API only lists/updates/deletes steps and ingredients that belong to a
// visible recipe (both viewsets filter by recipe__space), so these write
// cycles run through a recipe: the step and ingredient are created as part of
// the recipe payload, then managed through the standalone endpoints.
func TestE2EStepWriteCycle(t *testing.T) {
	c := newMCPClient(t)

	recipe := callTool(t, c, "recipe_create", map[string]any{"data": map[string]any{
		"name":     "e2e-step-recipe-" + runSuffix(),
		"servings": 1,
		"steps": []any{map[string]any{
			"instruction": "e2e stir the pot",
			"name":        "e2e-stir",
			"ingredients": []any{},
		}},
	}})
	recipeID := int(recipe["id"].(float64))
	stepID := int(recipe["steps"].([]any)[0].(map[string]any)["id"].(float64))
	require.Greater(t, stepID, 0)

	got := callTool(t, c, "step_get", map[string]any{"id": stepID})
	require.Equal(t, "e2e-stir", got["name"])

	patched := callTool(t, c, "step_patch", map[string]any{"id": stepID, "time": 12})
	require.Equal(t, float64(12), patched["time"])

	del := callTool(t, c, "step_delete", map[string]any{"id": stepID})
	require.Equal(t, true, del["deleted"])

	require.Equal(t, true, callTool(t, c, "recipe_delete", map[string]any{"id": recipeID})["deleted"])
}

func TestE2EIngredientWriteCycle(t *testing.T) {
	c := newMCPClient(t)

	food := callTool(t, c, "food_create", map[string]any{"name": "e2e-ing-food-" + runSuffix()})
	foodID := int(food["id"].(float64))

	recipe := callTool(t, c, "recipe_create", map[string]any{"data": map[string]any{
		"name":     "e2e-ingredient-recipe-" + runSuffix(),
		"servings": 1,
		"steps": []any{map[string]any{
			"instruction": "mix",
			"ingredients": []any{map[string]any{"food": foodID, "unit": nil, "amount": 2, "note": "e2e seasoning"}},
		}},
	}})
	recipeID := int(recipe["id"].(float64))
	ingID := int(recipe["steps"].([]any)[0].(map[string]any)["ingredients"].([]any)[0].(map[string]any)["id"].(float64))
	require.Greater(t, ingID, 0)

	patched := callTool(t, c, "ingredient_patch", map[string]any{"id": ingID, "note": "a pinch of e2e salt"})
	require.Equal(t, "a pinch of e2e salt", patched["note"])

	del := callTool(t, c, "ingredient_delete", map[string]any{"id": ingID})
	require.Equal(t, true, del["deleted"])

	require.Equal(t, true, callTool(t, c, "recipe_delete", map[string]any{"id": recipeID})["deleted"])
	require.Equal(t, true, callTool(t, c, "food_delete", map[string]any{"id": foodID})["deleted"])
}
