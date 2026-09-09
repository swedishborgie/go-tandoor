package mcp

import (
	"context"
	"fmt"
	"time"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/recipe"
	"github.com/swedishborgie/go-tandoor/shopping"
)

// shoppingListsFromIDs converts list IDs to minimal list references
// (marshaled as bare integers, which Tandoor's writable nested serializer accepts).
func shoppingListsFromIDs(ids []int) []shopping.List {
	lists := make([]shopping.List, 0, len(ids))
	for _, id := range ids {
		lists = append(lists, shopping.List{ID: id})
	}
	return lists
}

// usersFromIDs converts user IDs to minimal user references (marshaled as
// bare integers, which Tandoor's writable nested serializer accepts).
func usersFromIDs(ids []int) []recipe.User {
	users := make([]recipe.User, 0, len(ids))
	for _, id := range ids {
		users = append(users, recipe.User{ID: id})
	}
	return users
}

// dryRunParam is the shared schema option for all write tools.
func dryRunParam() mcpgo.ToolOption {
	return mcpgo.WithBoolean("dry_run", mcpgo.Description("Return the exact request (method, path, body) that would be sent, without sending it"))
}

// runWrite executes a write tool call, honoring the dry_run contract.
// method is HTTP (POST/PUT/PATCH) and path is the service path without a
// leading slash. When dry_run is set, the request preview is returned
// instead of calling exec.
func runWrite(ctx context.Context, req mcpgo.CallToolRequest, method, path string, body any, exec func() (any, error)) (*mcpgo.CallToolResult, error) {
	if req.GetBool("dry_run", false) {
		return jsonResult(map[string]any{
			"dry_run": true,
			"method":  method,
			"path":    "/" + path,
			"body":    body,
		}), nil
	}
	out, err := exec()
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(out), nil
}

// deleteResult handles delete tools (no body, 204 response).
func deleteResult(ctx context.Context, req mcpgo.CallToolRequest, path string, exec func() error) (*mcpgo.CallToolResult, error) {
	if req.GetBool("dry_run", false) {
		return jsonResult(map[string]any{"dry_run": true, "method": "DELETE", "path": "/" + path}), nil
	}
	if err := exec(); err != nil {
		return errResult(err), nil
	}
	return jsonResult(map[string]any{"deleted": true, "path": "/" + path}), nil
}

// parseDateOrDateTime accepts RFC3339 or YYYY-MM-DD, matching the CLI.
func parseDateOrDateTime(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date %q (use RFC3339 or YYYY-MM-DD)", value)
}
