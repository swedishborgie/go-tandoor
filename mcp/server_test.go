package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpgoserver "github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"

	tandoor "github.com/swedishborgie/go-tandoor"
)

// fakeTandoor stands in for a Tandoor instance. The /api/recipe/ subtree is
// canned (see recipe tests); everything else is answered generically:
// flat-array endpoints return a one-element array, single-object paths
// (including /api/space/current/) return a small object, and all other
// list paths return an empty pagination envelope.
func fakeTandoor(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/recipe/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "fake"})
			return
		}
		switch r.URL.Path {
		case "/api/recipe/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count":    2,
				"next":     nil,
				"previous": nil,
				"results": []map[string]any{
					{"id": 1, "name": "Pancakes", "created": "2024-01-01"},
					{"id": 2, "name": "Toast", "created": "2024-01-02"},
				},
			})
		case "/api/recipe/flat/":
			// The flat endpoint responds with a bare array, not a paginated envelope.
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "name": "Pancakes"}})
		case "/api/recipe/1/related/":
			_ = json.NewEncoder(w).Encode(map[string]any{"count": 0, "next": nil, "previous": nil, "results": []any{}})
		case "/api/recipe/1/":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Pancakes"})
		default:
			http.Error(w, `{"detail": "not found"}`, http.StatusNotFound)
		}
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			// Writes are accepted generically so write tools can be tested
			// end-to-end: deletes get a bare 204, everything else a minimal
			// object (with an ID where callers extract one).
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if r.URL.Path == "/api/import-open-data/" {
				// The real endpoint returns a per-datatype result map, not an object.
				_ = json.NewEncoder(w).Encode(map[string]any{"sr_legacy": map[string]any{"total_created": 0, "total_updated": 0, "total_untouched": 0, "total_errored": 0}})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "fake"})
			return
		}
		path := r.URL.Path
		if path == "/api/meal-plan/ical/" {
			w.Header().Set("Content-Type", "text/calendar")
			fmt.Fprint(w, "BEGIN:VCALENDAR\nEND:VCALENDAR\n")
			return
		}
		if path == "/api/space/current/" || isObjectPath(path) {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "fake"})
			return
		}
		switch path {
		case "/api/user/", "/api/group/", "/api/access-token/", "/api/localization/", "/api/search-fields/", "/api/search-preference/":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "name": "fake"}})
			return
		case "/api/unit/":
			// property_attach resolves the per-100 basis unit by name when the
			// food has none, so the fake carries a "gram" unit.
			_ = json.NewEncoder(w).Encode(map[string]any{"count": 1, "next": nil, "previous": nil, "results": []map[string]any{{"id": 7, "name": "gram"}}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"count": 0, "next": nil, "previous": nil, "results": []any{}})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// isObjectPath reports whether path looks like /api/<resource>/<id>/.
func isObjectPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 {
		return false
	}
	for _, c := range parts[2] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return parts[2] != ""
}

// newTestClient builds the MCP server against the fake instance and
// returns an in-process MCP client, initialized and ready.
func newTestClient(t *testing.T, opts ...Option) *client.Client {
	t.Helper()
	fake := fakeTandoor(t)
	c, err := tandoor.NewClient(fake.URL)
	require.NoError(t, err)

	s := NewServer(c, nil, opts...)
	mc, err := client.NewInProcessClient(s)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, mc.Start(ctx))

	initReq := mcpgo.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpgo.Implementation{Name: "tandoor-mcp-test", Version: "0.0.0"}
	_, err = mc.Initialize(ctx, initReq)
	require.NoError(t, err)
	return mc
}

func callTool(t *testing.T, c *client.Client, name string, args map[string]any) *mcpgo.CallToolResult {
	t.Helper()
	req := mcpgo.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := c.CallTool(ctx, req)
	require.NoError(t, err)
	return res
}

func resultText(t *testing.T, res *mcpgo.CallToolResult) string {
	t.Helper()
	require.NotEmpty(t, res.Content)
	text, ok := res.Content[0].(mcpgo.TextContent)
	require.True(t, ok, "expected text content, got %T", res.Content[0])
	return text.Text
}

func toolNames(t *testing.T, c *client.Client) []string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := c.ListTools(ctx, mcpgo.ListToolsRequest{})
	require.NoError(t, err)
	names := make([]string, 0, len(res.Tools))
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	return names
}

func TestRegisteredTools(t *testing.T) {
	c := newTestClient(t)
	require.ElementsMatch(t, []string{
		"server_info",
		"space_list", "space_get", "space_current",
		"user_list", "user_get", "user_patch",
		"group_list", "group_get",
		"household_list", "household_get",
		"household_create", "household_update", "household_patch", "household_delete",
		"invite_link_list", "invite_link_get",
		"invite_link_create", "invite_link_delete",
		"auth_list_tokens", "auth_get_token", "auth_create_token", "auth_delete_token",
		"ingredient_list", "ingredient_get",
		"ingredient_create", "ingredient_update", "ingredient_patch", "ingredient_delete",
		"step_list", "step_get",
		"step_create", "step_update", "step_patch", "step_delete",
		"food_list", "food_get",
		"food_create", "food_update", "food_patch", "food_delete", "food_merge", "food_move", "food_batch_update",
		"keyword_list", "keyword_get",
		"keyword_create", "keyword_update", "keyword_patch", "keyword_delete", "keyword_merge", "keyword_move",
		"unit_list", "unit_get", "unit_conversion_list", "unit_conversion_get",
		"unit_create", "unit_update", "unit_patch", "unit_delete", "unit_merge",
		"unit_conversion_create", "unit_conversion_update", "unit_conversion_patch", "unit_conversion_delete",
		"property_list", "property_get", "property_type_list", "property_type_get",
		"property_create", "property_update", "property_patch", "property_delete", "property_attach",
		"property_type_create", "property_type_update", "property_type_patch", "property_type_delete",
		"book_list", "book_get", "book_entry_list",
		"book_create", "book_update", "book_delete", "book_entry_create", "book_entry_delete",
		"shopping_list_list", "shopping_list_get", "shopping_entry_list",
		"shopping_entry_get", "shopping_recipe_list",
		"shopping_list_create", "shopping_list_update", "shopping_list_delete",
		"shopping_entry_create", "shopping_entry_update", "shopping_entry_patch", "shopping_entry_delete",
		"shopping_entry_bulk_update", "shopping_list_add_recipe", "shopping_recipe_create_entries",
		"meal_plan_list", "meal_plan_get", "meal_plan_ical",
		"meal_type_list", "meal_type_get",
		"meal_plan_create", "meal_plan_update", "meal_plan_delete", "meal_plan_auto_plan",
		"meal_type_create", "meal_type_update", "meal_type_patch", "meal_type_delete",
		"cook_log_list", "cook_log_get",
		"cook_log_create",
		"inventory_location_list", "inventory_location_get",
		"inventory_location_create", "inventory_location_update", "inventory_location_patch", "inventory_location_delete",
		"inventory_entry_list", "inventory_entry_get", "inventory_log_list",
		"inventory_entry_create", "inventory_entry_update", "inventory_entry_patch", "inventory_entry_delete",
		"storage_list", "storage_get",
		"storage_create", "storage_update", "storage_patch", "storage_delete",
		"supermarket_list", "supermarket_get",
		"supermarket_category_list", "supermarket_category_get",
		"import_list", "import_get", "import_log_list",
		"import_import", "import_import_all", "import_delete",
		"export_log_list", "bookmarklet_import_list",
		"share_link_create",
		"open_data_metadata", "open_data_import",
		"sync_list", "sync_get", "sync_log_list",
		"sync_create", "sync_update", "sync_patch", "sync_delete", "sync_query_synced_folder",
		"view_log_list", "view_log_get",
		"user_file_list", "user_file_get",
		"automation_list", "automation_get",
		"custom_filter_list", "custom_filter_get",
		"connector_config_list", "connector_config_get",
		"search_field_list", "search_preference_list", "search_preference_patch",
		"ai_provider_list", "ai_provider_get",
		"ai_log_list", "ai_log_get",
		"localization_list", "server_settings_get",
		"recipe_list", "recipe_get", "recipe_overview", "recipe_related",
		"recipe_create", "recipe_update", "recipe_patch", "recipe_delete",
		"recipe_batch_update", "recipe_add_to_shopping",
		"recipe_add_ingredients", "recipe_add_step",
		"recipe_upload_image", "recipe_ai_properties", "recipe_delete_external",
		"recipe_from_source_create",
		"food_fdc_import", "food_fdc_attach", "food_auto_conversions", "food_ai_properties", "food_ensure",
		"food_audit_inspect", "food_audit_fix", "food_audit_fix_preview", "food_find_duplicates",
		"fdc_search", "fdc_get_food",
	}, toolNames(t, c))
}

// TestAllToolsRespond calls every registered tool against the fake instance
// and requires a non-error result. The fake answers any path for any method
// (writes included), so write tools exercise their real HTTP path and
// argument plumbing is verified end-to-end. Together this catches path and
// argument-plumbing regressions across the whole catalog.
func TestAllToolsRespond(t *testing.T) {
	c := newTestClient(t)
	requiredArgs := map[string]map[string]any{
		"recipe_get":                     {"id": 1},
		"space_get":                      {"id": 1},
		"user_get":                       {"id": 1},
		"group_get":                      {"id": 1},
		"household_get":                  {"id": 1},
		"invite_link_get":                {"id": 1},
		"auth_get_token":                 {"id": 1},
		"ingredient_get":                 {"id": 1},
		"step_get":                       {"id": 1},
		"food_get":                       {"id": 1},
		"keyword_get":                    {"id": 1},
		"unit_get":                       {"id": 1},
		"unit_conversion_get":            {"id": 1},
		"property_get":                   {"id": 1},
		"property_type_get":              {"id": 1},
		"book_get":                       {"id": 1},
		"shopping_list_get":              {"id": 1},
		"shopping_entry_get":             {"id": 1},
		"meal_plan_get":                  {"id": 1},
		"meal_type_get":                  {"id": 1},
		"cook_log_get":                   {"id": 1},
		"inventory_location_get":         {"id": 1},
		"inventory_entry_get":            {"id": 1},
		"storage_get":                    {"id": 1},
		"supermarket_get":                {"id": 1},
		"supermarket_category_get":       {"id": 1},
		"import_get":                     {"id": 1},
		"sync_get":                       {"id": 1},
		"view_log_get":                   {"id": 1},
		"user_file_get":                  {"id": 1},
		"automation_get":                 {"id": 1},
		"custom_filter_get":              {"id": 1},
		"connector_config_get":           {"id": 1},
		"ai_provider_get":                {"id": 1},
		"ai_log_get":                     {"id": 1},
		"recipe_related":                 {"id": 1},
		"import_import":                  {"id": 1},
		"import_delete":                  {"id": 1},
		"share_link_create":              {"recipe_id": 1},
		"recipe_from_source_create":      {"url": "http://example.com/recipe"},
		"open_data_import":               {"selected_version": "2024.02", "selected_datatypes": []any{"sr_legacy"}},
		"sync_create":                    {"data": map[string]any{"path": "/"}},
		"sync_update":                    {"id": 1, "data": map[string]any{"active": true}},
		"sync_patch":                     {"id": 1, "data": map[string]any{"active": true}},
		"sync_delete":                    {"id": 1},
		"sync_query_synced_folder":       {"id": 1},
		"auth_create_token":              {"expires": "2027-01-01T00:00:00Z"},
		"auth_delete_token":              {"id": 1},
		"household_create":               {"name": "Test Household"},
		"household_update":               {"id": 1, "name": "Renamed"},
		"household_patch":                {"id": 1, "name": "Renamed"},
		"household_delete":               {"id": 1},
		"invite_link_create":             {"email": "a@b.c", "group_id": 1},
		"invite_link_delete":             {"id": 1},
		"user_patch":                     {"id": 1, "first_name": "New"},
		"search_preference_patch":        {"user_id": 1, "data": map[string]any{"search": ""}},
		"recipe_upload_image":            {"id": 1, "image_url": "http://example.com/img.jpg"},
		"recipe_ai_properties":           {"id": 1},
		"recipe_delete_external":         {"id": 1},
		"recipe_add_ingredients":         {"recipe_id": 1, "ingredients": []any{map[string]any{"food_id": 1}}},
		"recipe_add_step":                {"recipe_id": 1, "name": "Step"},
		"food_fdc_import":                {"id": 1},
		"food_fdc_attach":                {"food_id": 1},
		"food_auto_conversions":          {"food_id": 1},
		"food_ai_properties":             {"id": 1},
		"food_ensure":                    {"names": []any{"Test Food"}},
		"property_attach":                {"food_id": 1, "property_type_id": 1},
		"food_audit_fix":                 {"food_id": 1},
		"food_audit_inspect":             {"food_id": 1},
		"meal_plan_auto_plan":            {"start_date": "2026-01-01", "end_date": "2026-01-07", "meal_type_id": 1},
		"shopping_recipe_create_entries": {"recipe_id": 1},
		"fdc_search":                     {"query": "test"},
		"fdc_get_food":                   {"fdc_id": 1},
	}
	// The FDC tools require a real FDC API instance; the in-process test has
	// no FDC client, so they are expected to return a configuration error.
	fdcSkip := map[string]bool{"fdc_search": true, "fdc_get_food": true, "food_fdc_attach": true}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	listRes, err := c.ListTools(ctx, mcpgo.ListToolsRequest{})
	require.NoError(t, err)

	for _, tool := range listRes.Tools {
		args := map[string]any{}
		for k, v := range requiredArgs[tool.Name] {
			args[k] = v
		}

		// Fill any schema-required argument we have not provided with a
		// type-appropriate placeholder.
		for _, reqName := range tool.InputSchema.Required {
			if _, ok := args[reqName]; ok {
				continue
			}
			schema, _ := tool.InputSchema.Properties[reqName].(map[string]any)
			if strings.HasSuffix(reqName, "_date") {
				args[reqName] = "2026-01-01"
				continue
			}
			switch schema["type"] {
			case "string":
				args[reqName] = "test"
			case "integer":
				args[reqName] = 1
			case "number":
				args[reqName] = 1.5
			case "boolean":
				args[reqName] = true
			case "array":
				args[reqName] = []any{1}
			default:
				args[reqName] = map[string]any{"name": "test"}
			}
		}

		if fdcSkip[tool.Name] {
			res := callTool(t, c, tool.Name, args)
			require.Truef(t, res.IsError, "tool %s should report FDC not configured", tool.Name)
			continue
		}

		res := callTool(t, c, tool.Name, args)
		require.Falsef(t, res.IsError, "tool %s returned error: %s", tool.Name, resultText(t, res))
	}
}

func TestServerInfo(t *testing.T) {
	c := newTestClient(t)
	res := callTool(t, c, "server_info", nil)
	require.False(t, res.IsError)
	var info map[string]any
	require.NoError(t, json.Unmarshal([]byte(resultText(t, res)), &info))
	require.Equal(t, "http", info["base_url"].(string)[:4])
	require.Equal(t, false, info["read_only"])
	require.Equal(t, false, info["fdc_configured"])
}

func TestRecipeList(t *testing.T) {
	c := newTestClient(t)
	res := callTool(t, c, "recipe_list", nil)
	require.False(t, res.IsError)
	var page map[string]any
	require.NoError(t, json.Unmarshal([]byte(resultText(t, res)), &page))
	require.InDelta(t, 2.0, page["count"], 0.001)
	results := page["results"].([]any)
	require.Len(t, results, 2)
	require.Equal(t, "Pancakes", results[0].(map[string]any)["name"])
}

func TestRecipeListQueryParam(t *testing.T) {
	// The fake records the query string so we can assert parameter plumbing.
	var gotQuery string
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("query")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"count": 0, "results": []any{}})
	}))
	t.Cleanup(fake.Close)

	c, err := tandoor.NewClient(fake.URL)
	require.NoError(t, err)
	s := NewServer(c, nil)
	mc, err := client.NewInProcessClient(s)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, mc.Start(ctx))
	initReq := mcpgo.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpgo.Implementation{Name: "test", Version: "0.0.0"}
	_, err = mc.Initialize(ctx, initReq)
	require.NoError(t, err)

	res := callTool(t, mc, "recipe_list", map[string]any{"query": "pancake", "ordering": "-name"})
	require.False(t, res.IsError)
	require.Equal(t, "pancake", gotQuery)
}

func TestRecipeListJQ(t *testing.T) {
	c := newTestClient(t)
	res := callTool(t, c, "recipe_list", map[string]any{"jq": ".results[].name"})
	require.False(t, res.IsError)
	text := resultText(t, res)
	require.Containsf(t, text, "Pancakes", "got %q", text)
	require.Containsf(t, text, "Toast", "got %q", text)
}

func TestRecipeListAll(t *testing.T) {
	c := newTestClient(t)
	res := callTool(t, c, "recipe_list", map[string]any{"all": true})
	require.False(t, res.IsError)
	var items []map[string]any
	require.NoError(t, json.Unmarshal([]byte(resultText(t, res)), &items))
	require.Len(t, items, 2)
}

func TestRecipeGet(t *testing.T) {
	c := newTestClient(t)
	res := callTool(t, c, "recipe_get", map[string]any{"id": 1})
	require.False(t, res.IsError)
	var r map[string]any
	require.NoError(t, json.Unmarshal([]byte(resultText(t, res)), &r))
	require.Equal(t, "Pancakes", r["name"])
}

func TestRecipeGetMissingID(t *testing.T) {
	c := newTestClient(t)
	res := callTool(t, c, "recipe_get", nil)
	require.True(t, res.IsError, "missing required id must be an error")
}

func TestRecipeGetNotFound(t *testing.T) {
	c := newTestClient(t)
	res := callTool(t, c, "recipe_get", map[string]any{"id": 999})
	require.True(t, res.IsError)
	text := resultText(t, res)
	require.Contains(t, text, "404", "error should carry the HTTP status, got %q", text)
}

func TestToolFilter(t *testing.T) {
	allRecipe := []string{
		"recipe_list", "recipe_get", "recipe_overview", "recipe_related",
		"recipe_create", "recipe_update", "recipe_patch", "recipe_delete",
		"recipe_batch_update", "recipe_add_to_shopping",
		"recipe_add_ingredients", "recipe_add_step",
		"recipe_upload_image", "recipe_ai_properties", "recipe_delete_external",
		"recipe_from_source_create",
	}
	c := newTestClient(t, WithToolFilter("recipe_*"))
	require.ElementsMatch(t, allRecipe, toolNames(t, c))

	c = newTestClient(t, WithToolFilter("recipe_*", "-recipe_get"))
	require.ElementsMatch(t, []string{
		"recipe_list", "recipe_overview", "recipe_related",
		"recipe_create", "recipe_update", "recipe_patch", "recipe_delete",
		"recipe_batch_update", "recipe_add_to_shopping",
		"recipe_add_ingredients", "recipe_add_step",
		"recipe_upload_image", "recipe_ai_properties", "recipe_delete_external",
		"recipe_from_source_create",
	}, toolNames(t, c))

	c = newTestClient(t, WithToolFilter("+server_info"))
	require.Equal(t, []string{"server_info"}, toolNames(t, c))
}

// TestReadOnlyMode verifies that WithReadOnly hides every write tool
// (registered with write: true) while keeping the read catalog intact.
func TestReadOnlyMode(t *testing.T) {
	full := toolNames(t, newTestClient(t))
	readOnly := toolNames(t, newTestClient(t, WithReadOnly()))

	require.NotEmpty(t, full)
	require.Len(t, readOnly, 92, "read-only mode should keep the read catalog (M1 + M3 read tools)")

	// Write-tool name shapes: CRUD verbs plus the M3 action verbs (import,
	// ensure, attach, fix, auto_plan, create_entries, ...).
	writeName := regexp.MustCompile(`_(create|update|patch|delete|merge|move|batch_update|import|import_all|query_synced_folder|upload_image|ai_properties|delete_external|ensure|attach|fix|auto_plan|auto_conversions|create_entries|create_token|delete_token)$|add_recipe|add_to_shopping|add_ingredients|add_step$`)
	for _, name := range readOnly {
		require.NotRegexp(t, writeName, name, "read-only mode must not expose write tool %s", name)
	}
	for _, name := range full {
		if !contains(readOnly, name) {
			require.Regexp(t, writeName, name, "tool %s was hidden by read-only mode but looks like a read tool", name)
		}
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestHTTPTransport(t *testing.T) {
	// Exercises the streamable-HTTP wiring (session handling included).
	fake := fakeTandoor(t)
	c, err := tandoor.NewClient(fake.URL)
	require.NoError(t, err)
	s := NewServer(c, nil)
	hs := mcpgoserver.NewStreamableHTTPServer(s, mcpgoserver.WithEndpointPath("/mcp"))
	ts := httptest.NewServer(hs)
	t.Cleanup(ts.Close)

	mc, err := client.NewStreamableHttpClient(ts.URL + "/mcp")
	require.NoError(t, err)
	t.Cleanup(func() { _ = mc.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initReq := mcpgo.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpgo.Implementation{Name: "http-test", Version: "0.0.0"}
	initRes, err := mc.Initialize(ctx, initReq)
	require.NoError(t, err)
	require.Equal(t, ServerName, initRes.ServerInfo.Name)

	res := callTool(t, mc, "server_info", nil)
	require.False(t, res.IsError)
	require.Contains(t, resultText(t, res), "tandoor-mcp")
}
