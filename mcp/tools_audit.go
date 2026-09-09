package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/auditfood"
)

// registerAuditTools registers the audit composite tools:
// food_audit_inspect, food_audit_fix, food_find_duplicates.
func registerAuditTools(d *deps) []toolDef {
	inspect := mcpgo.NewTool("food_audit_inspect",
		mcpgo.WithDescription("Deep-dive on one food: details, ingredient usage, naming issues, suggested canonical name, and FDC candidates when the food has no FDC ID. Read-only."),
		mcpgo.WithInteger("food_id", mcpgo.Required(), mcpgo.Description("Food ID")),
		jqParam(),
	)
	inspectHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		result, err := auditfood.Inspect(ctx, d.Tandoor, d.FDC, req.GetInt("food_id", 0), 50)
		if err != nil {
			return errResult(err), nil
		}
		if jq := req.GetString("jq", ""); jq != "" {
			s, err := applyJQ(ctx, jq, result)
			if err != nil {
				return errResult(err), nil
			}
			return mcpgo.NewToolResultText(s), nil
		}
		return jsonResult(result), nil
	}

	fix := mcpgo.NewTool("food_audit_fix",
		mcpgo.WithDescription("Rename a food to its normalized canonical name (Title Case, prep words stripped). On a name collision it merges into the existing food unless merge=false. Use dry_run first to preview."),
		mcpgo.WithInteger("food_id", mcpgo.Required(), mcpgo.Description("Food ID")),
		mcpgo.WithBoolean("merge", mcpgo.Description("Merge into an existing food with the same canonical name on collision (default true); false makes a collision an error")),
		dryRunParam(),
	)
	fixHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		dryRun := req.GetBool("dry_run", false)
		result, err := auditfood.Fix(ctx, d.Tandoor, &auditfood.FixOptions{
			FoodID:           req.GetInt("food_id", 0),
			MergeOnCollision: req.GetBool("merge", true),
		}, dryRun)
		if err != nil {
			return errResult(err), nil
		}
		if dryRun {
			return jsonResult(map[string]any{"dry_run": true, "result": result}), nil
		}
		return jsonResult(result), nil
	}

	dups := mcpgo.NewTool("food_find_duplicates",
		mcpgo.WithDescription("Scan all foods and group names that are likely duplicates (normalized Jaccard word similarity). Read-only."),
		mcpgo.WithNumber("threshold", mcpgo.Description("Jaccard similarity threshold, 0..1 (default 0.7)")),
		jqParam(),
	)
	dupsHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		result, err := auditfood.FindDuplicates(ctx, d.Tandoor, req.GetFloat("threshold", 0))
		if err != nil {
			return errResult(err), nil
		}
		if jq := req.GetString("jq", ""); jq != "" {
			s, err := applyJQ(ctx, jq, result)
			if err != nil {
				return errResult(err), nil
			}
			return mcpgo.NewToolResultText(s), nil
		}
		return jsonResult(result), nil
	}

	return []toolDef{
		{tool: inspect, handler: inspectHandler},
		{tool: fix, handler: fixHandler, write: true},
		{tool: dups, handler: dupsHandler},
	}
}

// fdcNotConfigured is the standard error for FDC-backed tools when no API
// key is set.
func fdcNotConfigured() error {
	return fmt.Errorf("FDC_API_KEY is not set — configure the server with an FDC API key to use FDC tools")
}
