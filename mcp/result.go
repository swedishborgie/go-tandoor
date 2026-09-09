package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/itchyny/gojq"
	mcpgo "github.com/mark3labs/mcp-go/mcp"

	tandoor "github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// jsonResult marshals v into a JSON tool result. Service return values are
// marshaled directly so field names match the library's JSON tags.
func jsonResult(v any) *mcpgo.CallToolResult {
	data, err := json.Marshal(v)
	if err != nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("internal: marshal result: %v", err))
	}
	return mcpgo.NewToolResultText(string(data))
}

// jqParam is the shared schema option for the optional jq filter. Every
// JSON-returning read tool exposes it so callers can keep output small.
func jqParam() mcpgo.ToolOption {
	return mcpgo.WithString("jq", mcpgo.Description("gojq filter applied to the result before returning"))
}

// jsonResultJQ is jsonResult with optional jq projection: if the request
// carries a non-empty "jq" argument, the value is filtered instead of
// returned verbatim. Used by single-object and flat-array read tools
// (paginated list tools use listAllResult, which handles jq itself).
func jsonResultJQ(ctx context.Context, req mcpgo.CallToolRequest, v any) (*mcpgo.CallToolResult, error) {
	if jq := req.GetString("jq", ""); jq != "" {
		s, err := applyJQ(ctx, jq, v)
		if err != nil {
			return errResult(err), nil
		}
		return mcpgo.NewToolResultText(s), nil
	}
	return jsonResult(v), nil
}

// errResult converts an error into a tool error result. Tandoor API errors
// keep their HTTP status and message so the model can diagnose
// (404 vs 400 vs 500).
func errResult(err error) *mcpgo.CallToolResult {
	if te, ok := err.(*tandoor.TandoorError); ok {
		return mcpgo.NewToolResultError(te.Error())
	}
	return mcpgo.NewToolResultError(err.Error())
}

// listAllResult shapes the first page of a list call into a tool result,
// honoring all=true (auto-paginate via pagination.CollectAll) and jq.
// fetchNext retrieves subsequent pages by 1-indexed page number.
func listAllResult[T any](ctx context.Context, req mcpgo.CallToolRequest,
	first *pagination.Paginated[T], fetchNext func(page int) (*pagination.Paginated[T], error),
) (*mcpgo.CallToolResult, error) {
	out := any(first)
	if req.GetBool("all", false) {
		items, err := pagination.CollectAll(ctx, first, fetchNext)
		if err != nil {
			return errResult(err), nil
		}
		out = items
	}
	if jq := req.GetString("jq", ""); jq != "" {
		s, err := applyJQ(ctx, jq, out)
		if err != nil {
			return errResult(err), nil
		}
		return mcpgo.NewToolResultText(s), nil
	}
	return jsonResult(out), nil
}

// plainPaginatedList runs a list call that takes plain pagination options
// (no typed filters), honoring all=true and jq like listAllResult.
func plainPaginatedList[T any](ctx context.Context, req mcpgo.CallToolRequest,
	d *deps, listFn func(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[T], error),
) (*mcpgo.CallToolResult, error) {
	base := baseListOptions(req, d)
	first, err := listFn(ctx, &base)
	if err != nil {
		return errResult(err), nil
	}
	return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[T], error) {
		o2 := base
		o2.Page = page
		return listFn(ctx, &o2)
	})
}

// boolArg returns the boolean argument value and whether it was present.
// This distinguishes "not provided" from an explicit false, which matters
// for pointer filters like is_favorite.
func boolArg(req mcpgo.CallToolRequest, key string) (bool, bool) {
	args := req.GetArguments()
	if args == nil {
		return false, false
	}
	v, ok := args[key]
	if !ok {
		return false, false
	}
	b, _ := v.(bool)
	return b, true
}

// applyJQ runs a gojq filter over v and returns the results as one JSON
// value per line (mirrors the CLI --jq behavior). An empty result set
// yields "null".
func applyJQ(ctx context.Context, filter string, v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("jq: %w", err)
	}
	// Decode with UseNumber so large integers and exact decimals survive.
	var input any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&input); err != nil {
		return "", fmt.Errorf("jq: invalid JSON input: %w", err)
	}
	query, err := gojq.Parse(filter)
	if err != nil {
		return "", fmt.Errorf("jq: %w", err)
	}
	var buf bytes.Buffer
	iter := query.RunWithContext(ctx, input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if hErr, ok := err.(*gojq.HaltError); ok && hErr.Value() == nil {
				break
			}
			return "", fmt.Errorf("jq: %w", err)
		}
		out, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("jq: %w", err)
		}
		buf.Write(out)
		buf.WriteByte('\n')
	}
	if buf.Len() == 0 {
		return "null", nil
	}
	return buf.String(), nil
}
