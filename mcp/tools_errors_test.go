package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	tandoor "github.com/swedishborgie/go-tandoor"
)

// newTestClient500 builds an MCP client whose Tandoor server returns 500 on
// every request, so every tool's API-error branch is reachable.
func newTestClient500(t *testing.T, opts ...Option) *client.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"detail":"boom"}`))
	}))
	t.Cleanup(server.Close)

	c, err := tandoor.NewClient(server.URL)
	require.NoError(t, err)

	s := NewServer(c, nil, opts...)
	mc, err := client.NewInProcessClient(s)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, mc.Start(ctx))

	initReq := mcpgo.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpgo.Implementation{Name: "tandoor-mcp-500", Version: "0.0.0"}
	_, err = mc.Initialize(ctx, initReq)
	require.NoError(t, err)
	return mc
}

func toolArgsForErrorTest(toolName string, args map[string]any) map[string]any {
	// Same exactly-one-of exclusions as the full-args test.
	if toolName == "property_attach" {
		delete(args, "ingredient_id")
	}
	if toolName == "recipe_upload_image" {
		delete(args, "image_base64")
	}
	if toolName == "shopping_recipe_create_entries" {
		delete(args, "shopping_list_recipe_id")
	}
	return args
}

// TestAllToolsAPIErrorBranches calls every tool with full args against a
// 500-returning Tandoor, covering the errResult branch after the first API
// call in each handler.
func TestAllToolsAPIErrorBranches(t *testing.T) {
	c := newTestClient500(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	listRes, err := c.ListTools(ctx, mcpgo.ListToolsRequest{})
	require.NoError(t, err)

	for _, tool := range listRes.Tools {
		if tool.Name == "server_info" {
			continue // local tool: no Tandoor API call
		}
		res := callTool(t, c, tool.Name, toolArgsForErrorTest(tool.Name, fullArgs(tool)))
		require.True(t, res.IsError, "tool %s: expected error result against 500 server", tool.Name)
	}
}

// TestAllToolsMissingArgs calls every tool with no arguments, covering the
// first RequireX failure branch in each handler (list tools fall through to
// the 500 API error, which is also an error result).
func TestAllToolsMissingArgs(t *testing.T) {
	c := newTestClient500(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	listRes, err := c.ListTools(ctx, mcpgo.ListToolsRequest{})
	require.NoError(t, err)

	for _, tool := range listRes.Tools {
		if tool.Name == "server_info" {
			continue // local tool: no Tandoor API call
		}
		res := callTool(t, c, tool.Name, map[string]any{})
		require.True(t, res.IsError, "tool %s: expected error result with missing args", tool.Name)
	}
}
