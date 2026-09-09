//go:build integration
// +build integration

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2EPropertyWriteCycle(t *testing.T) {
	c := newMCPClient(t)

	// property type first
	createdType := callTool(t, c, "property_type_create", map[string]any{"name": "e2e-calorie-type-" + runSuffix()})
	typeID := int(createdType["id"].(float64))
	require.Greater(t, typeID, 0)

	// property (standalone, no food — 2.6.13 property serializer has no food FK)
	dry := callTool(t, c, "property_create", map[string]any{"property_type_id": typeID, "property_amount": 250.5, "dry_run": true})
	require.Equal(t, "/api/property/", dry["path"])

	created := callTool(t, c, "property_create", map[string]any{"property_type_id": typeID, "property_amount": 250.5})
	id := int(created["id"].(float64))
	require.Greater(t, id, 0)

	// cleanup in dependency order
	require.Equal(t, true, callTool(t, c, "property_delete", map[string]any{"id": id})["deleted"])
	require.Equal(t, true, callTool(t, c, "property_type_delete", map[string]any{"id": typeID})["deleted"])
}

func TestE2EBookWriteCycle(t *testing.T) {
	c := newMCPClient(t)
	name := "e2e-book-" + runSuffix()

	created := callTool(t, c, "book_create", map[string]any{"name": name, "description": "e2e book"})
	id := int(created["id"].(float64))
	require.Greater(t, id, 0)

	// book entry needs a recipe; create a minimal one
	recipe := callTool(t, c, "recipe_create", map[string]any{"data": map[string]any{"name": "e2e-book-entry-recipe-" + runSuffix(), "servings": 1, "steps": []any{}}})
	recipeID := int(recipe["id"].(float64))

	entry := callTool(t, c, "book_entry_create", map[string]any{"book_id": id, "recipe_id": recipeID})
	entryID := int(entry["id"].(float64))
	require.Greater(t, entryID, 0)

	require.Equal(t, true, callTool(t, c, "book_entry_delete", map[string]any{"id": entryID})["deleted"])
	require.Equal(t, true, callTool(t, c, "recipe_delete", map[string]any{"id": recipeID})["deleted"])
	require.Equal(t, true, callTool(t, c, "book_delete", map[string]any{"id": id})["deleted"])
}

func TestE2EStorageWriteCycle(t *testing.T) {
	c := newMCPClient(t)
	name := "e2e-storage-" + runSuffix()

	created := callTool(t, c, "storage_create", map[string]any{"name": name, "method": "LOCAL", "path": "/tmp"})
	id := int(created["id"].(float64))
	require.Greater(t, id, 0)

	patched := callTool(t, c, "storage_patch", map[string]any{"id": id, "path": "/tmp/e2e"})
	require.Equal(t, "/tmp/e2e", patched["path"])

	require.Equal(t, true, callTool(t, c, "storage_delete", map[string]any{"id": id})["deleted"])
}

func TestE2EInventoryWriteCycle(t *testing.T) {
	c := newMCPClient(t)

	// find the seeded household
	hhList := callTool(t, c, "household_list", map[string]any{})
	results := hhList["results"].([]any)
	require.NotEmpty(t, results)
	householdID := int(results[0].(map[string]any)["id"].(float64))

	// location
	loc := callTool(t, c, "inventory_location_create", map[string]any{"name": "e2e-loc-" + runSuffix(), "household_id": householdID})
	locID := int(loc["id"].(float64))
	require.Greater(t, locID, 0)

	// entry: needs a food and a unit (the API requires the unit field)
	food := callTool(t, c, "food_create", map[string]any{"name": "e2e-inv-food-" + runSuffix()})
	foodID := int(food["id"].(float64))
	unit := callTool(t, c, "unit_create", map[string]any{"name": "e2e-inv-unit-" + runSuffix()})
	unitID := int(unit["id"].(float64))

	entry := callTool(t, c, "inventory_entry_create", map[string]any{"food_id": foodID, "location_id": locID, "unit_id": unitID, "amount": 3.5})
	entryID := int(entry["id"].(float64))
	require.Greater(t, entryID, 0)

	patched := callTool(t, c, "inventory_entry_patch", map[string]any{"id": entryID, "amount": 7.25})
	require.Equal(t, 7.25, patched["amount"])

	require.Equal(t, true, callTool(t, c, "inventory_entry_delete", map[string]any{"id": entryID})["deleted"])
	require.Equal(t, true, callTool(t, c, "inventory_location_delete", map[string]any{"id": locID})["deleted"])
	require.Equal(t, true, callTool(t, c, "food_delete", map[string]any{"id": foodID})["deleted"])
}

func TestE2EMealplanWriteCycle(t *testing.T) {
	c := newMCPClient(t)

	// find the seeded Breakfast meal type
	mtList := callTool(t, c, "meal_type_list", map[string]any{})
	results := mtList["results"].([]any)
	mealTypeID := 0
	for _, r := range results {
		if r.(map[string]any)["name"] == "Breakfast" {
			mealTypeID = int(r.(map[string]any)["id"].(float64))
		}
	}
	require.NotZero(t, mealTypeID, "seeded Breakfast meal type not found")

	plan := callTool(t, c, "meal_plan_create", map[string]any{
		"from_date":    "2026-12-01",
		"meal_type_id": mealTypeID,
		"title":        "e2e breakfast",
		"servings":     2,
	})
	id := int(plan["id"].(float64))
	require.Greater(t, id, 0)

	// meal type lifecycle
	mt := callTool(t, c, "meal_type_create", map[string]any{"name": "e2e-snack-" + runSuffix(), "time": "15:00", "color": "#00ff00"})
	mtID := int(mt["id"].(float64))
	require.Greater(t, mtID, 0)
	require.Equal(t, true, callTool(t, c, "meal_type_delete", map[string]any{"id": mtID})["deleted"])

	require.Equal(t, true, callTool(t, c, "meal_plan_delete", map[string]any{"id": id})["deleted"])
}

func TestE2ECookLogCreate(t *testing.T) {
	c := newMCPClient(t)

	recipe := callTool(t, c, "recipe_create", map[string]any{"data": map[string]any{"name": "e2e-cooklog-recipe-" + runSuffix(), "servings": 2, "steps": []any{}}})
	recipeID := int(recipe["id"].(float64))

	log := callTool(t, c, "cook_log_create", map[string]any{"recipe_id": recipeID, "servings": 2, "rating": 4, "comment": "e2e delicious"})
	id := int(log["id"].(float64))
	require.Greater(t, id, 0)

	// verify via list
	logs := callTool(t, c, "cook_log_list", map[string]any{})
	require.Contains(t, logs, "results")

	require.Equal(t, true, callTool(t, c, "recipe_delete", map[string]any{"id": recipeID})["deleted"])
}
