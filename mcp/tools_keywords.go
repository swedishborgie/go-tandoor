package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/keyword"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func registerKeywordTools(d *deps) []toolDef {
	listOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List keywords. Filter by recipe book. Use all=true for the full set; jq projects fields to keep output small.")}
	listOpts = append(listOpts, baselineListParams()...)
	listOpts = append(listOpts,
		mcpgo.WithInteger("recipe_book_id", mcpgo.Description("Filter by recipe book ID")),
	)
	list := mcpgo.NewTool("keyword_list", listOpts...)
	listHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &keyword.ListOptions{
			ListOptions:  baseListOptions(req, d),
			RecipeBookID: req.GetInt("recipe_book_id", 0),
		}
		first, err := d.Tandoor.Keywords().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[keyword.Keyword], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Keywords().List(ctx, &o2)
		})
	}

	get := mcpgo.NewTool("keyword_get",
		mcpgo.WithDescription("Get a single keyword by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Keyword ID")),
		jqParam(),
	)
	getHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Keywords().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- writes ---
	kwCreate := mcpgo.NewTool("keyword_create",
		mcpgo.WithDescription("Create a keyword. Returns the created keyword."),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Keyword name")),
		mcpgo.WithString("description", mcpgo.Description("Keyword description")),
	)
	kwCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		k := &keyword.Keyword{Name: name, Description: req.GetString("description", "")}
		return runWrite(func() (any, error) {
			return d.Tandoor.Keywords().Create(ctx, k)
		})
	}

	kwUpdate := mcpgo.NewTool("keyword_update",
		mcpgo.WithDescription("Update a keyword (full replacement). Returns the updated keyword."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Keyword ID")),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Keyword name")),
		mcpgo.WithString("description", mcpgo.Description("Keyword description")),
	)
	kwUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		k := &keyword.Keyword{ID: id, Name: name, Description: req.GetString("description", "")}
		return runWrite(func() (any, error) {
			return d.Tandoor.Keywords().Update(ctx, k)
		})
	}

	kwPatch := mcpgo.NewTool("keyword_patch",
		mcpgo.WithDescription("Partially update a keyword. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Keyword ID")),
		mcpgo.WithString("name", mcpgo.Description("Keyword name")),
		mcpgo.WithString("description", mcpgo.Description("Keyword description")),
	)
	kwPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		k := &keyword.Keyword{ID: id}
		if v, err := req.RequireString("name"); err == nil {
			k.Name = v
		}
		if v, err := req.RequireString("description"); err == nil {
			k.Description = v
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.Keywords().Patch(ctx, k)
		})
	}

	kwDelete := mcpgo.NewTool("keyword_delete",
		mcpgo.WithDescription("Delete a keyword. Recipes keeping it are re-pointed to its parent."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Keyword ID")),
	)
	kwDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(fmt.Sprintf("api/keyword/%d/", id), func() error {
			return d.Tandoor.Keywords().Delete(ctx, id)
		})
	}

	kwMerge := mcpgo.NewTool("keyword_merge",
		mcpgo.WithDescription("Merge one keyword into another: recipes move to the target, then the source is deleted."),
		mcpgo.WithInteger("source_id", mcpgo.Required(), mcpgo.Description("Keyword to merge (deleted)")),
		mcpgo.WithInteger("target_id", mcpgo.Required(), mcpgo.Description("Keyword to keep")),
	)
	kwMergeHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		sourceID, err := req.RequireInt("source_id")
		if err != nil {
			return errResult(err), nil
		}
		targetID, err := req.RequireInt("target_id")
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.Keywords().Merge(ctx, sourceID, targetID)
		})
	}

	kwMove := mcpgo.NewTool("keyword_move",
		mcpgo.WithDescription("Move a keyword under a new parent (re-parent in the keyword tree)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Keyword ID")),
		mcpgo.WithInteger("parent_id", mcpgo.Required(), mcpgo.Description("New parent keyword ID (0 for top level)")),
	)
	kwMoveHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		parentID, err := req.RequireInt("parent_id")
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.Keywords().Move(ctx, id, parentID)
		})
	}

	return []toolDef{
		{tool: list, handler: listHandler},
		{tool: get, handler: getHandler},
		{tool: kwCreate, handler: kwCreateHandler, write: true},
		{tool: kwUpdate, handler: kwUpdateHandler, write: true},
		{tool: kwPatch, handler: kwPatchHandler, write: true},
		{tool: kwDelete, handler: kwDeleteHandler, write: true},
		{tool: kwMerge, handler: kwMergeHandler, write: true},
		{tool: kwMove, handler: kwMoveHandler, write: true},
	}
}
