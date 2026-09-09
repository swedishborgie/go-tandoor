package mcp

import (
	"context"
	"strconv"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/misc"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func registerCookLogTools(d *deps) []toolDef {
	listOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List cook logs (recorded times a recipe was cooked). Filter by recipe. Use all=true for the full set; jq projects fields to keep output small.")}
	listOpts = append(listOpts, baselineListParams(d)...)
	listOpts = append(listOpts,
		mcpgo.WithInteger("recipe_id", mcpgo.Description("Filter by recipe ID")),
	)
	list := mcpgo.NewTool("cook_log_list", listOpts...)
	listHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		base := baseListOptions(req, d)
		if id := req.GetInt("recipe_id", 0); id != 0 {
			if base.Extra == nil {
				base.Extra = map[string]string{}
			}
			base.Extra["recipe"] = strconv.Itoa(id)
		}
		opts := &misc.CookLogListOptions{ListOptions: base}
		first, err := d.Tandoor.CookLogs().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[misc.CookLog], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.CookLogs().List(ctx, &o2)
		})
	}

	get := mcpgo.NewTool("cook_log_get",
		mcpgo.WithDescription("Get a single cook log entry by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Cook log ID")),
		jqParam(),
	)
	getHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.CookLogs().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- writes ---
	clCreate := mcpgo.NewTool("cook_log_create",
		mcpgo.WithDescription("Log a cooked meal (cook log entry). Returns the created cook log."),
		mcpgo.WithInteger("recipe_id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		mcpgo.WithInteger("servings", mcpgo.Description("Servings cooked (default 1)")),
		mcpgo.WithInteger("rating", mcpgo.Description("Rating 0-5")),
		mcpgo.WithString("comment", mcpgo.Description("Comment about the cooking")),
		dryRunParam(),
	)
	clCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		recipeID, err := req.RequireInt("recipe_id")
		if err != nil {
			return errResult(err), nil
		}
		log := &misc.CookLog{Recipe: &recipeID, Servings: intPtr(req.GetInt("servings", 1))}
		if v, err := req.RequireInt("rating"); err == nil {
			log.Rating = &v
		}
		if v, err := req.RequireString("comment"); err == nil {
			log.Comment = &v
		}
		return runWrite(ctx, req, "POST", "api/cook-log/", log, func() (any, error) {
			return d.Tandoor.CookLogs().Create(ctx, log)
		})
	}

	return []toolDef{
		{tool: list, handler: listHandler},
		{tool: get, handler: getHandler},
		{tool: clCreate, handler: clCreateHandler, write: true},
	}
}

// intPtr returns a pointer to v.
func intPtr(v int) *int { return &v }
