//go:build integration
// +build integration

package e2e

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/stretchr/testify/require"
)

// firstListID returns the id of the first result of a paginated list tool.
func firstListID(t *testing.T, c *client.Client, tool string, args map[string]any) int {
	t.Helper()
	out := callTool(t, c, tool, args)
	results, ok := out["results"].([]any)
	require.True(t, ok, "expected paginated results, got %v", out)
	require.NotEmpty(t, results, "list should not be empty")
	return int(results[0].(map[string]any)["id"].(float64))
}

// TestE2EPropertyAttachIdempotent verifies the property_attach composite:
// create, then re-run updates the same property.
func TestE2EPropertyAttachIdempotent(t *testing.T) {
	c := newMCPClient(t)
	suffix := runSuffix()

	// Fresh spaces have no units or property types; create what we need.
	gram := callTool(t, c, "unit_create", map[string]any{"name": "gram-" + suffix})
	gramID := int(gram["id"].(float64))
	pt := callTool(t, c, "property_type_create", map[string]any{"name": "e2e-ptype-" + suffix, "unit": "g"})
	ptID := int(pt["id"].(float64))
	food := callTool(t, c, "food_create", map[string]any{"name": "e2e-pattach-food-" + suffix})
	foodID := int(food["id"].(float64))

	// create (per_100_unit_id pins the basis unit; the fresh instance has no
	// unit named "gram")
	created := callTool(t, c, "property_attach", map[string]any{
		"food_id": foodID, "property_type_id": ptID, "property_amount": 42,
		"per_100_unit_id": gramID,
	})
	require.Equal(t, "created", created["action"])

	// idempotent re-run updates the existing property
	updated := callTool(t, c, "property_attach", map[string]any{
		"food_id": foodID, "property_type_id": ptID, "property_amount": 99,
	})
	require.Equal(t, "updated", updated["action"])

	// the food's properties reflect the final amount
	fetched := callTool(t, c, "food_get", map[string]any{"id": foodID})
	props, _ := fetched["properties"].([]any)
	require.Len(t, props, 1)
	require.Equal(t, float64(99), props[0].(map[string]any)["property_amount"])
}

// TestE2EFoodEnsureCycle verifies the food_ensure composite: report-only mode
// (force_create=false) finds nothing, the real run creates the missing food,
// and a second run finds it (idempotent).
func TestE2EFoodEnsureCycle(t *testing.T) {
	c := newMCPClient(t)
	name := "e2e-ensure-" + runSuffix()

	// report-only: the name does not exist yet, and nothing is created
	report := callTool(t, c, "food_ensure", map[string]any{"names": []any{name}, "force_create": false})
	summary := report["summary"].(map[string]any)
	require.Equal(t, float64(0), summary["created"])
	require.Equal(t, float64(1), summary["not_found"])
	entry := report["results"].(map[string]any)[name].(map[string]any)
	require.Equal(t, "not_found", entry["status"])

	// apply: creates the food (force_create must be explicit — the default
	// is report-only, so a wrong name cannot create a near-duplicate)
	out := callTool(t, c, "food_ensure", map[string]any{"names": []any{name}, "force_create": true})
	report = out
	summary = report["summary"].(map[string]any)
	require.Equal(t, float64(1), summary["created"])

	// second run: finds it, creates nothing
	out2 := callTool(t, c, "food_ensure", map[string]any{"names": []any{name}})
	report2 := out2
	summary2 := report2["summary"].(map[string]any)
	require.Equal(t, float64(1), summary2["found"])
	require.Equal(t, float64(0), summary2["created"])
}

// TestE2EFoodAuditInspectAndFix verifies food_audit_inspect surfaces a
// suggested canonical name and food_audit_fix renames the food.
func TestE2EFoodAuditInspectAndFix(t *testing.T) {
	c := newMCPClient(t)
	suffix := runSuffix()

	rawName := "E2E  audit_fix FOOD!! " + suffix
	food := callTool(t, c, "food_create", map[string]any{"name": rawName})
	foodID := int(food["id"].(float64))

	inspect := callTool(t, c, "food_audit_inspect", map[string]any{"food_id": foodID})
	require.Equal(t, foodID, int(inspect["food"].(map[string]any)["id"].(float64)))
	suggested, ok := inspect["suggested_name"].(string)
	require.True(t, ok, "inspect should suggest a name: %v", inspect)
	require.NotEqual(t, rawName, suggested)

	// preview: would rename (read-only, no write)
	preview := callTool(t, c, "food_audit_fix_preview", map[string]any{"food_id": foodID})
	require.Equal(t, suggested, preview["new_name"])

	// apply: renames
	fixed := callTool(t, c, "food_audit_fix", map[string]any{"food_id": foodID})
	require.Equal(t, suggested, fixed["new_name"])
	require.Equal(t, suggested, fixed["food"].(map[string]any)["name"])
}

// TestE2EFoodFindDuplicates verifies duplicate detection groups similar names.
func TestE2EFoodFindDuplicates(t *testing.T) {
	c := newMCPClient(t)
	suffix := runSuffix()

	f1 := callTool(t, c, "food_create", map[string]any{"name": "e2e dupbase " + suffix})
	f2 := callTool(t, c, "food_create", map[string]any{"name": "e2e dupbase " + suffix + " extra"})
	id1 := int(f1["id"].(float64))
	id2 := int(f2["id"].(float64))

	out := callTool(t, c, "food_find_duplicates", map[string]any{"threshold": 0.5})
	groups, ok := out["groups"].([]any)
	require.True(t, ok, "expected groups, got %v", out)

	found := false
	for _, g := range groups {
		members := g.(map[string]any)["foods"].([]any)
		var ids []int
		for _, m := range members {
			ids = append(ids, int(m.(map[string]any)["id"].(float64)))
		}
		if containsInt(ids, id1) && containsInt(ids, id2) {
			found = true
		}
	}
	require.True(t, found, "the two similar foods should be grouped, groups: %v", groups)
}

func containsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// TestE2EShoppingRecipeCreateEntries verifies the shopping composite derives
// scaled entries from recipe ingredients and creates them on a list.
func TestE2EShoppingRecipeCreateEntries(t *testing.T) {
	c := newMCPClient(t)
	suffix := runSuffix()

	food := callTool(t, c, "food_create", map[string]any{"name": "e2e-shop-food-" + suffix})
	foodID := int(food["id"].(float64))

	recipe := callTool(t, c, "recipe_create", map[string]any{
		"data": map[string]any{
			"name":     "e2e-shop-recipe-" + suffix,
			"servings": 2,
			"steps": []any{map[string]any{
				"instruction": "mix",
				"ingredients": []any{map[string]any{"food": foodID, "unit": nil, "amount": 100}},
			}},
		},
	})
	recipeID := int(recipe["id"].(float64))

	list := callTool(t, c, "shopping_list_create", map[string]any{"name": "e2e-shop-list-" + suffix})
	listID := int(list["id"].(float64))

	// apply at double servings: 100 -> 200 (creates a shopping list recipe
	// link and the scaled entries)
	out := callTool(t, c, "shopping_recipe_create_entries", map[string]any{
		"recipe_id": recipeID, "shopping_list_ids": []any{listID}, "servings": 4,
	})
	entries, ok := out["entries"].([]any)
	require.True(t, ok, "expected entries in result: %v", out)
	require.NotEmpty(t, entries, "bulk create should return the derived entries")

	// the entry landed on the list, scaled (all=true returns a flat array)
	entryText, isErr := callToolRaw(t, c, "shopping_entry_list", map[string]any{"all": true})
	require.False(t, isErr)
	var results []map[string]any
	require.NoError(t, json.Unmarshal([]byte(entryText), &results))
	var scaled []float64
	for _, m := range results {
		if m["food"] != nil {
			if fid, ok := m["food"].(map[string]any); ok && int(fid["id"].(float64)) == foodID {
				scaled = append(scaled, m["amount"].(float64))
			}
		}
	}
	require.Contains(t, scaled, float64(200), "expected a scaled entry for amount 200, got %v", scaled)
}

// TestE2EAutoPlan verifies meal_plan_auto_plan produces a meal plan entry.
func TestE2EAutoPlan(t *testing.T) {
	c := newMCPClient(t)
	suffix := runSuffix()

	mealTypeID := firstListID(t, c, "meal_type_list", nil)
	kw := callTool(t, c, "keyword_create", map[string]any{"name": "e2e-autoplan-" + suffix})
	kwID := int(kw["id"].(float64))

	recipe := callTool(t, c, "recipe_create", map[string]any{
		"data": map[string]any{
			"name":     "e2e-autoplan-recipe-" + suffix,
			"servings": 2,
			// auto-plan only considers internal recipes.
			"internal": true,
			// Tandoor requires the ingredients key on every step, even when empty.
			"steps":    []any{map[string]any{"instruction": "do it", "ingredients": []any{}}},
			"keywords": []any{kwID},
		},
	})
	recipeID := int(recipe["id"].(float64))

	// Meal plans older than 90 days are invisible to meal_plan_list, so plan
	// for tomorrow.
	planDate := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	start := planDate
	end := planDate

	out := callTool(t, c, "meal_plan_auto_plan", map[string]any{
		"start_date": start, "end_date": end, "meal_type_id": mealTypeID,
		"keyword_ids": []any{kwID},
	})
	// the endpoint echoes back the processed request
	require.Equal(t, float64(mealTypeID), out["meal_type_id"])

	// the plan exists and references the recipe (all=true returns a flat array)
	planText, isErr := callToolRaw(t, c, "meal_plan_list", map[string]any{"all": true})
	require.False(t, isErr)
	var results []map[string]any
	require.NoError(t, json.Unmarshal([]byte(planText), &results))
	found := false
	for _, m := range results {
		if m["recipe"] != nil {
			if rid, ok := m["recipe"].(map[string]any); ok && int(rid["id"].(float64)) == recipeID {
				found = true
			}
		}
	}
	require.True(t, found, "expected a meal plan for the recipe, plans: %v", results)
}

// TestE2EShareLinkCycle verifies share link creation and deletion.
func TestE2EShareLinkCycle(t *testing.T) {
	c := newMCPClient(t)

	recipe := callTool(t, c, "recipe_create", map[string]any{
		"data": map[string]any{
			"name":  "e2e-share-" + runSuffix(),
			"steps": []any{map[string]any{"instruction": "share me", "ingredients": []any{}}},
		},
	})
	recipeID := int(recipe["id"].(float64))

	link := callTool(t, c, "share_link_create", map[string]any{"recipe_id": recipeID})
	linkPK := int(link["pk"].(float64))
	require.Greater(t, linkPK, 0)
	// Tandoor builds "<base>/recipe/<pk>/?share=<uuid>".
	require.Contains(t, link["link"].(string), "?share=")
	require.NotEmpty(t, link["share"].(string))
	// Tandoor exposes no share-link delete endpoint (GET-only view), so
	// creation is the full cycle.
}

// TestE2EAuthTokenCycle verifies access token create/list/delete.
func TestE2EAuthTokenCycle(t *testing.T) {
	c := newMCPClient(t)
	expires := "2099-01-01T00:00:00Z"

	created := callTool(t, c, "auth_create_token", map[string]any{
		"expires": expires, "scope": "read",
	})
	id := int(created["id"].(float64))
	require.Greater(t, id, 0)
	require.NotEmpty(t, created["token"])

	// visible in the list (auth_list_tokens returns a plain array)
	text, isErr := callToolRaw(t, c, "auth_list_tokens", nil)
	require.False(t, isErr)
	var tokens []map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &tokens))
	found := false
	for _, r := range tokens {
		if int(r["id"].(float64)) == id {
			found = true
		}
	}
	require.True(t, found, "created token should appear in auth_list_tokens")

	deleted := callTool(t, c, "auth_delete_token", map[string]any{"id": id})
	require.Equal(t, true, deleted["deleted"])

	_, isGone := callToolRaw(t, c, "auth_get_token", map[string]any{"id": id})
	require.True(t, isGone, "deleted token should be gone")
}

// TestE2EHouseholdCycle verifies household create/update/delete.
func TestE2EHouseholdCycle(t *testing.T) {
	c := newMCPClient(t)
	name := "e2e-hh-" + runSuffix()

	created := callTool(t, c, "household_create", map[string]any{"name": name})
	id := int(created["id"].(float64))
	require.Greater(t, id, 0)

	updated := callTool(t, c, "household_update", map[string]any{"id": id, "name": name + "-up"})
	require.Equal(t, name+"-up", updated["name"])

	deleted := callTool(t, c, "household_delete", map[string]any{"id": id})
	require.Equal(t, true, deleted["deleted"])
	_, isErr := callToolRaw(t, c, "household_get", map[string]any{"id": id})
	require.True(t, isErr, "deleted household should be gone")
}

// TestE2ESearchPreferencePatch verifies the search preference PATCH passthrough.
func TestE2ESearchPreferencePatch(t *testing.T) {
	c := newMCPClient(t)

	// The e2e user's preference row is created on first access; the API
	// PATCHs against the user id (the API token user's id is 1 here).
	patch := callTool(t, c, "search_preference_patch", map[string]any{
		"user_id": 1,
		"data":    map[string]any{"trigram_threshold": 0.6},
	})
	require.Equal(t, float64(0.6), patch["trigram_threshold"])
}

// TestE2EFdcNotConfigured verifies fdc tools report a clean configuration
// error when no FDC key is configured (the e2e stack has none).
func TestE2EFdcNotConfigured(t *testing.T) {
	c := newMCPClient(t)
	text, isErr := callToolRaw(t, c, "fdc_search", map[string]any{"query": "apple"})
	require.True(t, isErr)
	require.Contains(t, text, "FDC")

	text, isErr = callToolRaw(t, c, "fdc_get_food", map[string]any{"fdc_id": 1})
	require.True(t, isErr)
	require.Contains(t, text, "FDC")
}

// TestE2ESyncCycle verifies sync config create/patch/delete.
func TestE2ESyncCycle(t *testing.T) {
	c := newMCPClient(t)
	// Create a local storage for the sync config to point at.
	st := callTool(t, c, "storage_create", map[string]any{"name": "e2e-sync-storage-" + runSuffix(), "method": "LOCAL", "path": "/tmp"})
	stID := int(st["id"].(float64))
	require.Greater(t, stID, 0)

	created := callTool(t, c, "sync_create", map[string]any{
		"data": map[string]any{"storage": stID, "path": "/e2e-sync/" + runSuffix()},
	})
	id := int(created["id"].(float64))
	require.Greater(t, id, 0)

	patched := callTool(t, c, "sync_patch", map[string]any{
		"id": id, "data": map[string]any{"active": false},
	})
	require.Equal(t, false, patched["active"])

	deleted := callTool(t, c, "sync_delete", map[string]any{"id": id})
	require.Equal(t, true, deleted["deleted"])

	_ = callTool(t, c, "storage_delete", map[string]any{"id": stID})
}
