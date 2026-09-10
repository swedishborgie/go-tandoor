package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	tandoor "github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/fdc"
)

// fullArgs builds an argument map that fills every schema property with a
// type-appropriate placeholder, so one call per tool exercises every
// optional-argument branch (jq, all, typed filters, ...).
func fullArgs(tool mcpgo.Tool) map[string]any {
	args := map[string]any{}
	for name, prop := range tool.InputSchema.Properties {
		schema, _ := prop.(map[string]any)
		if schema == nil {
			continue
		}
		args[name] = placeholder(tool.Name, name, schema)
	}
	return args
}

func placeholder(toolName, name string, schema map[string]any) any {
	// Special-cased properties first.
	switch name {
	case "jq":
		return "." // identity filter: valid on any JSON shape
	case "all":
		return true
	case "types":
		return []any{"srlegacy"}
	case "data_types":
		return []any{"srlegacy"}
	case "selected_datatypes":
		return []any{"sr_legacy"}
	case "names":
		return []any{"Test Food"}
	case "expires":
		return "2027-01-01T00:00:00Z"
	case "data":
		return map[string]any{"name": "test"}
	}
	items, _ := schema["items"].(map[string]any)
	switch schema["type"] {
	case "boolean":
		return true
	case "integer":
		return 1
	case "number":
		return 1.5
	case "array":
		if items != nil && items["type"] == "string" {
			return []any{"test"}
		}
		return []any{1}
	case "object":
		return map[string]any{"name": "test"}
	default: // string
		switch {
		case strings.HasSuffix(name, "_date"):
			return "2026-01-01"
		case strings.HasSuffix(name, "_url"), name == "image_url":
			return "http://example.com/x"
		case name == "ordering":
			return "id"
		case name == "data_type":
			return "Branded"
		}
		return "test"
	}
}

// TestAllToolsWithFullArgs calls every tool with every optional argument
// filled. The fake instance answers generically, so any failure is a real
// argument-plumbing bug.
func TestAllToolsWithFullArgs(t *testing.T) {
	c := newTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	listRes, err := c.ListTools(ctx, mcpgo.ListToolsRequest{})
	require.NoError(t, err)

	for _, tool := range listRes.Tools {
		args := fullArgs(tool)
		if tool.Name == "property_attach" {
			// Exactly one of food_id / ingredient_id is allowed.
			delete(args, "ingredient_id")
		}
		if tool.Name == "recipe_upload_image" {
			// Exactly one of image_url / image_base64 is allowed.
			delete(args, "image_base64")
		}
		if tool.Name == "shopping_recipe_create_entries" {
			// Exactly one of shopping_list_recipe_id / recipe_id is allowed.
			delete(args, "shopping_list_recipe_id")
		}
		res := callTool(t, c, tool.Name, args)
		if strings.HasPrefix(tool.Name, "fdc_") {
			// No FDC client in this server: configuration error is correct.
			require.Truef(t, res.IsError, "tool %s without FDC should error", tool.Name)
			continue
		}
		require.Falsef(t, res.IsError, "tool %s with full args: %s", tool.Name, resultText(t, res))
	}
}

// fakeFDCCanned serves the FDC endpoints the client uses.
func fakeFDCCanned(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/food/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"fdcId": 123, "description": "Test Food", "dataType": "Branded",
		})
	})
	mux.HandleFunc("/v1/foods/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("query") != "" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"totalHits": 1, "currentPage": 1, "totalPages": 1,
				"foods": []map[string]any{{"fdcId": 123, "description": "Test Food", "dataType": "Branded"}},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{{"fdcId": 123, "description": "Test Food", "dataType": "Branded"}},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// newTestClientWithFDC builds the test client with a working FDC client so
// the FDC tools exercise their success paths.
func newTestClientWithFDC(t *testing.T, opts ...Option) *client.Client {
	t.Helper()
	fake := fakeTandoor(t)
	c, err := tandoor.NewClient(fake.URL)
	require.NoError(t, err)

	fdcSrv := fakeFDCCanned(t)
	fc, err := fdc.NewClient(fdc.WithAPIKey("test-key"), fdc.WithBaseURL(fdcSrv.URL))
	require.NoError(t, err)

	s := NewServer(c, fc, opts...)
	mc, err := client.NewInProcessClient(s)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, mc.Start(ctx))

	initReq := mcpgo.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpgo.Implementation{Name: "tandoor-mcp-fdc-test", Version: "0.0.0"}
	_, err = mc.Initialize(ctx, initReq)
	require.NoError(t, err)
	return mc
}

func TestFdcToolsConfigured(t *testing.T) {
	c := newTestClientWithFDC(t)

	res := callTool(t, c, "fdc_search", map[string]any{"query": "test", "limit": 5, "types": []any{"branded", "foundation"}})
	require.False(t, res.IsError, resultText(t, res))
	require.Contains(t, resultText(t, res), "Test Food")

	res = callTool(t, c, "fdc_get_food", map[string]any{"fdc_id": 123})
	require.False(t, res.IsError, resultText(t, res))
	require.Contains(t, resultText(t, res), "123")

	// Invalid data type name must be a tool error.
	res = callTool(t, c, "fdc_search", map[string]any{"query": "test", "types": []any{"bogus"}})
	require.True(t, res.IsError)
	require.Contains(t, resultText(t, res), "unknown data type")
}

// TestToolErrorBranches exercises validation and API-error branches:
// missing required args, bad jq filters, and 404s.
func TestToolErrorBranches(t *testing.T) {
	c := newTestClient(t)

	// Missing required id.
	for _, name := range []string{"recipe_get", "food_get", "unit_get", "meal_plan_get"} {
		res := callTool(t, c, name, nil)
		require.Truef(t, res.IsError, "%s without id must fail", name)
	}

	// Bad jq filter.
	res := callTool(t, c, "recipe_list", map[string]any{"jq": ".["})
	require.True(t, res.IsError)
	require.Contains(t, resultText(t, res), "jq")

	// 404 on object fetch (the canned recipe handler 404s unknown IDs).
	res = callTool(t, c, "recipe_get", map[string]any{"id": 999})
	require.True(t, res.IsError)

	// food_merge / keyword_merge / unit_merge missing second arg.
	res = callTool(t, c, "food_merge", map[string]any{"source_id": 1})
	require.True(t, res.IsError)
}

// TestWithDefaultPageSize verifies the option is wired through without
// breaking tool registration or calls.
func TestWithDefaultPageSize(t *testing.T) {
	c := newTestClient(t, WithDefaultPageSize(7))
	res := callTool(t, c, "recipe_list", nil)
	require.False(t, res.IsError)
}
