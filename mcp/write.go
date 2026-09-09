package mcp

import (
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

// runWrite executes a write tool call, wrapping the outcome: errors become
// tool errors, results are JSON-encoded.
func runWrite(exec func() (any, error)) (*mcpgo.CallToolResult, error) {
	out, err := exec()
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(out), nil
}

// deleteResult handles delete tools (no body, 204 response).
func deleteResult(path string, exec func() error) (*mcpgo.CallToolResult, error) {
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
