package mcp

import (
	"context"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/supermarket"
)

func registerSupermarketTools(d *deps) []toolDef {
	listOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List supermarkets. Use all=true for the full set; jq projects fields to keep output small.")}
	listOpts = append(listOpts, baselineListParams()...)
	list := mcpgo.NewTool("supermarket_list", listOpts...)
	listHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &supermarket.ListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.Supermarkets().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[supermarket.Supermarket], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Supermarkets().List(ctx, &o2)
		})
	}

	get := mcpgo.NewTool("supermarket_get",
		mcpgo.WithDescription("Get a single supermarket by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Supermarket ID")),
		jqParam(),
	)
	getHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Supermarkets().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	catListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List supermarket categories. Use all=true for the full set; jq projects fields to keep output small.")}
	catListOpts = append(catListOpts, baselineListParams()...)
	catList := mcpgo.NewTool("supermarket_category_list", catListOpts...)
	catListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		base := baseListOptions(req, d)
		first, err := d.Tandoor.SupermarketCategories().List(ctx, &base)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[supermarket.Category], error) {
			o2 := base
			o2.Page = page
			return d.Tandoor.SupermarketCategories().List(ctx, &o2)
		})
	}

	catGet := mcpgo.NewTool("supermarket_category_get",
		mcpgo.WithDescription("Get a single supermarket category by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Category ID")),
		jqParam(),
	)
	catGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.SupermarketCategories().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	return []toolDef{
		{tool: list, handler: listHandler},
		{tool: get, handler: getHandler},
		{tool: catList, handler: catListHandler},
		{tool: catGet, handler: catGetHandler},
	}
}
