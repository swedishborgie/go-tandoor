// cmd/tandoor/commands_test.go
//
// Smoke + targeted tests for every CLI command group, run against the
// generic fake Tandoor API (fakeAPI in cli_test.go).
package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runTable runs each CLI invocation and asserts success/failure.
func runTable(t *testing.T, f *fakeAPI, cases []struct {
	name    string
	args    []string
	wantErr string // substring expected in the error; "" = must succeed
}) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := f.run(t, tc.args...)
			if tc.wantErr != "" {
				require.Error(t, err, "expected error containing %q", tc.wantErr)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestFoodsCommands(t *testing.T) {
	f := newFakeAPI(t)
	tmp := filepath.Join(t.TempDir(), "out.json")
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"foods", "list"}, ""},
		{"list_search", []string{"foods", "list", "--search", "cheese"}, ""},
		{"list_all", []string{"foods", "list", "--all"}, ""},
		{"list_name_exact", []string{"foods", "list", "--name-exact", "Widget"}, ""},
		{"list_names_batch", []string{"foods", "list", "--name", "Widget", "--name", "Other"}, ""},
		{"list_category", []string{"foods", "list", "--category-id", "3", "--unit-id", "4"}, ""},
		{"list_jq", []string{"foods", "list", "--jq", ".count"}, ""},
		{"list_output_file", []string{"foods", "list", "--output-file", tmp}, ""},
		{"get", []string{"foods", "get", "1"}, ""},
		{"get_batch", []string{"foods", "get", "--id", "1", "--id", "2"}, ""},
		{"get_bad_id", []string{"foods", "get", "abc"}, "invalid ID"},
		{"get_no_id", []string{"foods", "get"}, "required"},
		{"create", []string{"foods", "create", "New Food"}, ""},
		{"create_dry_run", []string{"foods", "create", "New Food", "--dry-run"}, ""},
		{"create_no_name", []string{"foods", "create"}, "required"},
		{"update", []string{"foods", "update", "1", "--name", "X", "--fdc-id", "9", "--substitute-onhand"}, ""},
		{"update_clear_fdc", []string{"foods", "update", "1", "--clear-fdc-id"}, ""},
		{"patch", []string{"foods", "patch", "1", "--name", "X", "--shopping", "bulk"}, ""},
		{"patch_dry_run", []string{"foods", "patch", "1", "--name", "X", "--dry-run"}, ""},
		{"patch_clear_fdc", []string{"foods", "patch", "1", "--clear-fdc-id"}, ""},
		{"patch_no_fields", []string{"foods", "patch", "1"}, "no fields to patch"},
		{"delete", []string{"foods", "delete", "1"}, ""},
		{"merge", []string{"foods", "merge", "1", "2"}, ""},
		{"merge_missing_arg", []string{"foods", "merge", "1"}, "target_id"},
		{"ensure", []string{"foods", "ensure", "--name", "Widget"}, ""},
		{"ensure_dry_run", []string{"foods", "ensure", "--name", "Widget", "--dry-run"}, ""},
		{"ensure_force_create", []string{"foods", "ensure", "--name", "New", "--force-create"}, ""},
		{"ensure_output_file", []string{"foods", "ensure", "--name", "Widget", "--output-file", tmp}, ""},
		{"ensure_no_name", []string{"foods", "ensure"}, "name"},
		{"auto_conversions", []string{"foods", "auto-conversions", "--food-id", "1"}, ""},
		{"auto_conversions_dry", []string{"foods", "auto-conversions", "--food-id", "1", "--dry-run"}, ""},
		{"auto_conversions_no_id", []string{"foods", "auto-conversions"}, "food-id"},
	})
}

func TestFoodsAttachFDC(t *testing.T) {
	f := newFakeAPI(t)
	// No FDC key → error.
	err := f.run(t, "foods", "attach-fdc-properties", "--food-id", "1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FDC API key")

	// With key + fake FDC server.
	fdcSrv := fakeFDCServer(t)
	t.Setenv("FDC_API_KEY", "test")
	t.Setenv("FDC_BASE_URL", fdcSrv)
	err = f.run(t, "foods", "attach-fdc-properties", "--food-id", "1", "--fdc-id", "5", "--dry-run")
	require.NoError(t, err)
}

func TestRecipesCommands(t *testing.T) {
	f := newFakeAPI(t)
	// --flat uses /api/recipe/flat/ which returns a bare array.
	f.route("GET", "/api/recipe/flat/", func(w http.ResponseWriter, r *http.Request) {
		writeJSONBody(w, []map[string]any{{"id": 1, "name": "Widget"}})
	})
	tmp := filepath.Join(t.TempDir(), "out.json")
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"recipes", "list"}, ""},
		{"list_all", []string{"recipes", "list", "--all"}, ""},
		{"list_filters", []string{"recipes", "list", "--search", "pasta", "--flat", "--keyword-ids", "1,2", "--space-id", "1", "--recipe-book-id", "2", "--user-id", "3", "--is-favorite"}, ""},
		{"list_jq_file", []string{"recipes", "list", "--jq", ".count", "--output-file", tmp}, ""},
		{"get", []string{"recipes", "get", "1"}, ""},
		{"create", []string{"recipes", "create", "--name", "R", "--servings", "2", "--working-time", "5", "--keywords", "1,2", "--internal", "--private"}, ""},
		{"create_dry_run", []string{"recipes", "create", "--name", "R", "--dry-run"}, ""},
		{"create_no_name", []string{"recipes", "create"}, "name"},
		{"update", []string{"recipes", "update", "1", "--name", "X", "--description", "d"}, ""},
		{"patch", []string{"recipes", "patch", "1", "--name", "X"}, ""},
		{"delete", []string{"recipes", "delete", "1"}, ""},
		{"overview", []string{"recipes", "overview", "--search", "pasta"}, ""},
		{"batch_update", []string{"recipes", "batch-update", "--recipes", "1,2", "--keywords-add", "3", "--working-time", "10"}, ""},
		{"batch_update_no_recipes", []string{"recipes", "batch-update"}, "recipes"},
		{"set_image_url", []string{"recipes", "set-image", "1", "--url", "http://example.com/i.png"}, ""},
		{"related", []string{"recipes", "related", "1"}, ""},
	})
}

func TestShoppingCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"shopping", "list"}, ""},
		{"list_search", []string{"shopping", "list", "--search", "groceries"}, ""},
		{"get", []string{"shopping", "get", "1"}, ""},
		{"create", []string{"shopping", "create", "Groceries", "--color", "#fff"}, ""},
		{"create_no_name", []string{"shopping", "create"}, "required"},
		{"delete", []string{"shopping", "delete", "1"}, ""},
		{"entries", []string{"shopping", "entries", "1"}, ""},
		{"add_entry", []string{"shopping", "add-entry", "--list-id", "1", "--food-id", "2", "--amount", "1.5", "--note", "n"}, ""},
		{"add_entry_no_list", []string{"shopping", "add-entry", "--food-id", "2"}, "list-id"},
		{"bulk_update", []string{"shopping", "bulk-update", "--ids", "1,2", "--checked"}, ""},
		{"bulk_update_lists", []string{"shopping", "bulk-update", "--ids", "1", "--shopping-lists-add", "2", "--shopping-lists-remove", "3", "--shopping-lists-set", "4"}, ""},
		{"bulk_update_no_ids", []string{"shopping", "bulk-update"}, "ids"},
		{"add_recipe", []string{"shopping", "add-recipe", "5", "--list-id", "1", "--servings", "2"}, ""},
		{"create_entries_from_recipe", []string{"shopping", "create-entries-from-recipe", "7", "--shopping-list-ids", "1,2"}, ""},
	})
}

func TestMealPlanCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"mealplans", "list"}, ""},
		{"create_dates", []string{"mealplans", "create", "--title", "T", "--from-date", "2026-01-01", "--to-date", "2026-12-31"}, ""},
		{"create_bad_date", []string{"mealplans", "create", "--title", "T", "--from-date", "nope"}, "parse --from-date"},
		{"get", []string{"mealplans", "get", "1"}, ""},
		{"create", []string{"mealplans", "create", "--title", "T", "--recipe-id", "2", "--meal-type-id", "3"}, ""},
		{"update", []string{"mealplans", "update", "1", "--title", "X", "--note", "n"}, ""},
		{"delete", []string{"mealplans", "delete", "1"}, ""},
		{"ical", []string{"mealplans", "ical"}, ""},
		{"auto_plan", []string{"mealplans", "auto-plan", "--start-date", "2026-01-01", "--end-date", "2026-01-07", "--meal-type-id", "1", "--keywords", "2,3", "--keyword-mode", "and"}, ""},
		{"auto_plan_no_dates", []string{"mealplans", "auto-plan"}, "start"},
	})
}

func TestMealTypeCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"mealtypes", "list"}, ""},
		{"get", []string{"mealtypes", "get", "1"}, ""},
	})
}

func TestCookLogCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"cook-logs", "list"}, ""},
		{"create", []string{"cook-logs", "create", "--recipe-id", "1", "--servings", "4", "--rating", "5", "--comment", "good"}, ""},
		{"create_no_recipe", []string{"cook-logs", "create"}, "recipe-id"},
	})
}

func TestBookCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"books", "list"}, ""},
		{"get", []string{"books", "get", "1"}, ""},
		{"create", []string{"books", "create", "My Book", "--order", "2", "--shared", "1,2"}, ""},
		{"update", []string{"books", "update", "1", "--name", "X", "--description", "d"}, ""},
		{"delete", []string{"books", "delete", "1"}, ""},
		{"entries_list", []string{"book-entries", "list"}, ""},
		{"entries_list_book", []string{"book-entries", "list", "1"}, ""},
		{"entries_create", []string{"book-entries", "create", "--book-id", "1", "--recipe-id", "2"}, ""},
		{"entries_delete", []string{"book-entries", "delete", "1"}, ""},
	})
}

func TestImportCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"imports", "list"}, ""},
		{"get", []string{"imports", "get", "1"}, ""},
		{"import", []string{"imports", "import", "1"}, ""},
		{"import_all", []string{"imports", "import-all"}, ""},
		{"delete", []string{"imports", "delete", "1"}, ""},
		{"import_logs_list", []string{"import-logs", "list"}, ""},
		{"source_create_url", []string{"recipe-from-source", "create", "--url", "http://example.com/r"}, ""},
		{"source_create_data", []string{"recipe-from-source", "create", "--data", "<html>x</html>"}, ""},
		{"source_create_both", []string{"recipe-from-source", "create", "--url", "http://x", "--data", "d"}, ""},
		{"source_create_neither", []string{"recipe-from-source", "create"}, "url"},
		{"share_links_list", []string{"share-links", "list"}, ""},
		{"share_links_create", []string{"share-links", "create", "1"}, ""},
		{"share_links_delete", []string{"share-links", "delete", "1"}, ""},
	})
}

func TestUnitCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"units", "list"}, ""},
		{"list_all", []string{"units", "list", "--all"}, ""},
		{"get", []string{"units", "get", "1"}, ""},
		{"create", []string{"units", "create", "Gram", "--plural-name", "Grams"}, ""},
		{"create_no_name", []string{"units", "create"}, "required"},
		{"update", []string{"units", "update", "1", "--name", "X", "--base-unit", "y"}, ""},
		{"patch", []string{"units", "patch", "1", "--name", "X"}, ""},
		{"delete", []string{"units", "delete", "1"}, ""},
		{"merge", []string{"units", "merge", "1", "2"}, ""},
		{"merge_missing", []string{"units", "merge", "1"}, "target_id"},
	})
}

func TestConversionCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"conversions", "list"}, ""},
		{"list_food", []string{"conversions", "list", "--food-id", "1", "--all"}, ""},
		{"get", []string{"conversions", "get", "1"}, ""},
		{"create", []string{"conversions", "create", "--base-unit", "1", "--base-amount", "2", "--converted-unit", "3", "--converted-amount", "1", "--food", "4"}, ""},
		{"create_dry", []string{"conversions", "create", "--base-unit", "1", "--base-amount", "2", "--converted-unit", "3", "--converted-amount", "1", "--dry-run"}, ""},
		{"create_incomplete", []string{"conversions", "create", "--base-unit", "1", "--dry-run"}, "required"},
		{"update", []string{"conversions", "update", "1", "--base-amount", "5", "--open-data-slug", "s"}, ""},
		{"patch", []string{"conversions", "patch", "1", "--base-amount", "5"}, ""},
		{"delete", []string{"conversions", "delete", "1"}, ""},
	})
}

func TestIngredientCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"ingredients", "list"}, ""},
		{"list_filters", []string{"ingredients", "list", "--recipe-id", "1", "--food-id", "2", "--all"}, ""},
		{"get", []string{"ingredients", "get", "1"}, ""},
		{"create", []string{"ingredients", "create", "--food", "2", "--amount", "1", "--note", "n", "--order", "3"}, ""},
		{"create_dry", []string{"ingredients", "create", "--food", "2", "--dry-run"}, ""},
		{"update", []string{"ingredients", "update", "1", "--note", "x", "--checked"}, ""},
		{"patch", []string{"ingredients", "patch", "1", "--note", "x"}, ""},
		{"delete", []string{"ingredients", "delete", "1"}, ""},
	})
}

func TestKeywordCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"keywords", "list"}, ""},
		{"list_all", []string{"keywords", "list", "--all", "--parent-id", "1"}, ""},
		{"get", []string{"keywords", "get", "1"}, ""},
		{"create", []string{"keywords", "create", "Vegan", "--parent", "2"}, ""},
		{"update", []string{"keywords", "update", "1", "--name", "X", "--description", "d"}, ""},
		{"patch", []string{"keywords", "patch", "1", "--name", "X"}, ""},
		{"delete", []string{"keywords", "delete", "1"}, ""},
		{"merge", []string{"keywords", "merge", "1", "2"}, ""},
	})
}

func TestStepCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"steps", "list"}, ""},
		{"get", []string{"steps", "get", "1"}, ""},
		{"create", []string{"steps", "create", "Mix", "--time", "5", "--order", "1", "--show-as-header"}, ""},
		{"update", []string{"steps", "update", "1", "--name", "X", "--instruction", "i"}, ""},
		{"patch", []string{"steps", "patch", "1", "--name", "X"}, ""},
		{"delete", []string{"steps", "delete", "1"}, ""},
	})
}

func TestPropertyCommands(t *testing.T) {
	f := newFakeAPI(t)
	// Attach-by-ingredient resolves the ingredient's food.
	f.route("GET", "/api/ingredient/1/", func(w http.ResponseWriter, r *http.Request) {
		writeJSONBody(w, map[string]any{"id": 1, "food": map[string]any{"id": 2, "name": "Widget"}})
	})
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"list", []string{"properties", "list"}, ""},
		{"get", []string{"properties", "get", "1"}, ""},
		{"create", []string{"properties", "create", "--property-type-id", "1", "--property-amount", "2"}, ""},
		{"update", []string{"properties", "update", "1", "--property-amount", "3", "--clear-property-amount"}, ""},
		{"delete", []string{"properties", "delete", "1"}, ""},
		{"attach", []string{"properties", "attach", "--food-id", "1", "--property-type-id", "2", "--property-amount", "3"}, ""},
		{"attach_ingredient", []string{"properties", "attach", "--ingredient-id", "1", "--property-type-id", "2", "--property-amount", "3", "--dry-run"}, ""},
		{"attach_neither", []string{"properties", "attach", "--property-type-id", "2"}, "property-amount"},
		{"types_list", []string{"properties", "types", "list"}, ""},
		{"types_get", []string{"properties", "types", "get", "1"}, ""},
		{"pt_list", []string{"property-types", "list"}, ""},
		{"pt_create", []string{"property-types", "create", "Energy"}, ""},
		{"pt_update", []string{"property-types", "update", "1", "--name", "X", "--fdc-id", "9"}, ""},
		{"pt_delete", []string{"property-types", "delete", "1"}, ""},
	})
}

func TestAuthCommands(t *testing.T) {
	f := newFakeAPI(t)
	f.route("POST", "/api-token-auth/", func(w http.ResponseWriter, r *http.Request) {
		writeJSONBody(w, map[string]any{"token": "tda_test"})
	})
	err := f.run(t, "auth", "token", "--username", "u", "--password", "p")
	require.NoError(t, err)

	err = f.run(t, "auth", "token", "--username", "u")
	require.Error(t, err) // missing required password

	f.route("GET", "/api/access-token/", func(w http.ResponseWriter, r *http.Request) {
		writeJSONBody(w, []map[string]any{{"id": 1, "token": "tda_x"}})
	})
	require.NoError(t, f.run(t, "auth", "list-tokens"))
}

func TestFDCCommands(t *testing.T) {
	f := newFakeAPI(t)

	// Without key → error.
	err := f.run(t, "fdc", "search", "pasta")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FDC API key")

	t.Setenv("FDC_API_KEY", "test")
	fdcSrv := fakeFDCServer(t)
	t.Setenv("FDC_BASE_URL", fdcSrv)

	require.NoError(t, f.run(t, "fdc", "search", "pasta", "--limit", "5", "--type", "Branded"))
	require.NoError(t, f.run(t, "fdc", "search", "--query", "pasta"))
	require.NoError(t, f.run(t, "fdc", "get", "172828"))
	require.NoError(t, f.run(t, "fdc", "food", "1"))
}

func TestMCPCommandNoToken(t *testing.T) {
	f := newFakeAPI(t)
	err := f.run(t, "mcp")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no API token set")
}

func TestAuditCommands(t *testing.T) {
	f := newFakeAPI(t)
	runTable(t, f, []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"food_inspect", []string{"audit", "food", "inspect", "1"}, ""},
		{"food_suggest", []string{"audit", "food", "suggest", "1"}, ""},
		{"food_fix_dry", []string{"audit", "food", "fix", "1", "--dry-run"}, ""},
		{"food_fix_no_yes", []string{"audit", "food", "fix", "1"}, "--yes"},
		{"foods", []string{"audit", "foods"}, ""},
		{"foods_category", []string{"audit", "foods", "--category", "1"}, ""},
		{"duplicates", []string{"audit", "duplicates", "--threshold", "0.5"}, ""},
		{"connectors", []string{"audit", "connectors"}, ""},
	})
}

// fakeFDCServer returns the URL of a fake FDC API server.
func fakeFDCServer(t *testing.T) string {
	t.Helper()
	f := newFakeAPI(t)
	// FDC endpoints.
	f.route("GET", "/fdc/v1/foods/search", func(w http.ResponseWriter, r *http.Request) {
		writeJSONBody(w, map[string]any{
			"totalHits": 1, "currentPage": 1, "totalPages": 1,
			"foods": []map[string]any{{"fdcId": 1, "description": "Pasta"}},
		})
	})
	f.route("GET", "/fdc/v1/food/172828", func(w http.ResponseWriter, r *http.Request) {
		writeJSONBody(w, map[string]any{"fdcId": 172828, "description": "Pasta"})
	})
	return f.url + "/fdc"
}

// TestGlobalFlags checks base-url validation.
func TestGlobalFlags(t *testing.T) {
	err := runApp(t, "foods", "list")
	// base-url is set but unreachable → request error
	require.Error(t, err)

	// No base URL at all.
	app := newApp()
	err = app.Run(context.Background(), []string{"tandoor", "foods", "list"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no base URL set")
}

func TestConversionsCreateBatch(t *testing.T) {
	f := newFakeAPI(t)
	valid := `[{"base_unit":1,"base_amount":2,"converted_unit":3,"converted_amount":1}]`
	wrapped := `{"conversions":[{"base_unit":1,"base_amount":2,"converted_unit":3,"converted_amount":1}]}`
	bad := `[{"base_unit":1}]`
	notJSON := `{"foo": "bar"}`
	write := func(name, content string) string {
		p := filepath.Join(t.TempDir(), name)
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
		return p
	}

	require.NoError(t, f.run(t, "conversions", "create-batch", "--file", write("a.json", valid)))
	require.NoError(t, f.run(t, "conversions", "create-batch", "--file", write("b.json", wrapped)))
	err := f.run(t, "conversions", "create-batch", "--file", write("c.json", bad))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required fields")
	err = f.run(t, "conversions", "create-batch", "--file", write("d.json", notJSON))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected array")
	err = f.run(t, "conversions", "create-batch", "--file", write("e.json", `[]`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no conversions found")
	require.Error(t, f.run(t, "conversions", "create-batch"))
	require.Error(t, f.run(t, "conversions", "create-batch", "--file", "/nonexistent/x.json"))
}

func TestRecipesFromFiles(t *testing.T) {
	f := newFakeAPI(t)
	recipeJSON := `{"name": "File Recipe", "servings": 2, "steps": [{"instruction": "Mix"}]}`
	ingsJSON := `[{"food": {"id": 1}, "amount": 2}]`
	rp := filepath.Join(t.TempDir(), "recipe.json")
	require.NoError(t, os.WriteFile(rp, []byte(recipeJSON), 0o644))
	ip := filepath.Join(t.TempDir(), "ings.json")
	require.NoError(t, os.WriteFile(ip, []byte(ingsJSON), 0o644))

	require.NoError(t, f.run(t, "recipes", "create", "--file", rp, "--dry-run"))
	require.NoError(t, f.run(t, "recipes", "create", "--name", "R", "--ingredients-file", ip, "--dry-run"))
	err := f.run(t, "recipes", "create", "--file", "/nonexistent/r.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read recipe file")
	bad := filepath.Join(t.TempDir(), "bad.json")
	require.NoError(t, os.WriteFile(bad, []byte(`{`), 0o644))
	err = f.run(t, "recipes", "create", "--file", bad)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse recipe JSON")
	badIng := filepath.Join(t.TempDir(), "badings.json")
	require.NoError(t, os.WriteFile(badIng, []byte(`{`), 0o644))
	err = f.run(t, "recipes", "create", "--name", "R", "--ingredients-file", badIng)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ingredients file")
}

// TestErrorBranches runs the main CRUD commands against a server that
// returns 500 for every request, exercising the error-handling branches.
func TestErrorBranches(t *testing.T) {
	f := newFakeAPI(t)
	f.failAll(t, http.StatusInternalServerError)

	cmds := [][]string{
		{"foods", "list"},
		{"foods", "get", "1"},
		{"foods", "create", "X"},
		{"foods", "update", "1", "--name", "X"},
		{"foods", "patch", "1", "--name", "X"},
		{"foods", "delete", "1"},
		{"foods", "merge", "1", "2"},
		{"foods", "ensure", "--name", "X"},
		{"foods", "auto-conversions", "--food-id", "1"},
		{"recipes", "list"},
		{"recipes", "get", "1"},
		{"recipes", "create", "--name", "R"},
		{"recipes", "update", "1", "--name", "X"},
		{"recipes", "patch", "1", "--name", "X"},
		{"recipes", "delete", "1"},
		{"recipes", "overview", "--search", "pasta"},
		{"recipes", "batch-update", "--recipes", "1,2", "--keywords-add", "3"},
		{"recipes", "set-image", "1", "--url", "http://example.com/i.png"},
		{"recipes", "related", "1"},
		{"shopping", "list"},
		{"shopping", "get", "1"},
		{"shopping", "create", "X"},
		{"shopping", "delete", "1"},
		{"shopping", "entries", "1"},
		{"shopping", "add-entry", "--list-id", "1", "--food-id", "2"},
		{"shopping", "bulk-update", "--ids", "1", "--checked"},
		{"shopping", "add-recipe", "5", "--list-id", "1"},
		{"mealplans", "list"},
		{"mealplans", "get", "1"},
		{"mealplans", "create", "--title", "T"},
		{"mealplans", "update", "1", "--title", "X"},
		{"mealplans", "delete", "1"},
		{"mealplans", "ical"},
		{"mealtypes", "list"},
		{"mealtypes", "get", "1"},
		{"cook-logs", "list"},
		{"cook-logs", "create", "--recipe-id", "1"},
		{"books", "list"},
		{"books", "get", "1"},
		{"books", "create", "X"},
		{"books", "update", "1", "--name", "X"},
		{"books", "delete", "1"},
		{"book-entries", "list"},
		{"book-entries", "create", "--book-id", "1", "--recipe-id", "2"},
		{"book-entries", "delete", "1"},
		{"imports", "list"},
		{"imports", "get", "1"},
		{"imports", "import", "1"},
		{"imports", "import-all"},
		{"imports", "delete", "1"},
		{"import-logs", "list"},
		{"recipe-from-source", "create", "--url", "http://x"},
		{"share-links", "list"},
		{"share-links", "create", "1"},
		{"share-links", "delete", "1"},
		{"units", "list"},
		{"units", "get", "1"},
		{"units", "create", "X"},
		{"units", "update", "1", "--name", "X"},
		{"units", "patch", "1", "--name", "X"},
		{"units", "delete", "1"},
		{"units", "merge", "1", "2"},
		{"conversions", "list"},
		{"conversions", "get", "1"},
		{"conversions", "update", "1", "--base-amount", "5"},
		{"conversions", "patch", "1", "--base-amount", "5"},
		{"conversions", "delete", "1"},
		{"ingredients", "list"},
		{"ingredients", "get", "1"},
		{"ingredients", "create", "--food", "2"},
		{"ingredients", "update", "1", "--note", "x"},
		{"ingredients", "patch", "1", "--note", "x"},
		{"ingredients", "delete", "1"},
		{"keywords", "list"},
		{"keywords", "get", "1"},
		{"keywords", "create", "X"},
		{"keywords", "update", "1", "--name", "X"},
		{"keywords", "patch", "1", "--name", "X"},
		{"keywords", "delete", "1"},
		{"keywords", "merge", "1", "2"},
		{"steps", "list"},
		{"steps", "get", "1"},
		{"steps", "create", "Mix"},
		{"steps", "update", "1", "--name", "X"},
		{"steps", "patch", "1", "--name", "X"},
		{"steps", "delete", "1"},
		{"properties", "list"},
		{"properties", "get", "1"},
		{"properties", "create", "--property-type-id", "1", "--property-amount", "2"},
		{"properties", "attach", "--food-id", "1", "--property-type-id", "2", "--property-amount", "3"},
		{"properties", "update", "1", "--property-amount", "3"},
		{"properties", "delete", "1"},
		{"properties", "types", "list"},
		{"properties", "types", "get", "1"},
		{"property-types", "list"},
		{"property-types", "create", "Energy"},
		{"property-types", "update", "1", "--name", "X"},
		{"property-types", "delete", "1"},
	}
	for _, c := range cmds {
		t.Run(strings.Join(c, " "), func(t *testing.T) {
			require.Error(t, f.run(t, c...))
		})
	}
}

func TestAuditErrorBranches(t *testing.T) {
	f := newFakeAPI(t)
	f.failAll(t, http.StatusInternalServerError)
	for _, c := range [][]string{
		{"audit", "food", "inspect", "1"},
		{"audit", "food", "suggest", "1"},
		{"audit", "food", "fix", "1", "--dry-run"},
		{"audit", "foods"},
		{"audit", "duplicates"},
		{"audit", "connectors"},
	} {
		t.Run(strings.Join(c, " "), func(t *testing.T) {
			require.Error(t, f.run(t, c...))
		})
	}
}

func TestDateFormats(t *testing.T) {
	f := newFakeAPI(t)
	// RFC3339 datetime form.
	require.NoError(t, f.run(t, "mealplans", "create", "--title", "T",
		"--from-date", "2026-01-01T12:30:00Z", "--to-date", "2026-01-02T00:00:00+00:00"))
	// Invalid date.
	err := f.run(t, "mealplans", "update", "1", "--from-date", "not-a-date")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse --from-date")
}

// TestListAllAndJQ exercises the --all (multi-page) and --jq output branches
// on list commands that support them.
func TestListAllAndJQ(t *testing.T) {
	f := newFakeAPI(t)
	tmp := filepath.Join(t.TempDir(), "out.json")
	for _, c := range [][]string{
		{"ingredients", "list", "--all", "--recipe-id", "1", "--food-id", "2"},
		{"properties", "list", "--all", "--food-id", "1", "--search", "energy"},
		{"properties", "types", "list", "--all", "--search", "x"},
		{"conversions", "list", "--all", "--food-id", "1"},
		{"keywords", "list", "--all"},
		{"units", "list", "--all"},
		{"foods", "list", "--all", "--jq", "[.[].name]"},
		{"recipes", "list", "--all", "--jq", "[.[].name]", "--output-file", tmp},
	} {
		t.Run(strings.Join(c, " "), func(t *testing.T) {
			require.NoError(t, f.run(t, c...))
		})
	}
}
