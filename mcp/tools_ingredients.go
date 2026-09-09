package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/ingredient"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func registerIngredientTools(d *deps) []toolDef {
	listOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List ingredients. Filter by recipe, food, or space. Use all=true for the full set; jq projects fields to keep output small.")}
	listOpts = append(listOpts, baselineListParams(d)...)
	listOpts = append(listOpts,
		mcpgo.WithInteger("recipe_id", mcpgo.Description("Filter by recipe ID")),
		mcpgo.WithInteger("food_id", mcpgo.Description("Filter by food ID")),
		mcpgo.WithInteger("space_id", mcpgo.Description("Filter by space ID")),
	)
	list := mcpgo.NewTool("ingredient_list", listOpts...)
	listHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &ingredient.ListOptions{
			ListOptions: baseListOptions(req, d),
			FoodID:      req.GetInt("food_id", 0),
			RecipeID:    req.GetInt("recipe_id", 0),
			SpaceID:     req.GetInt("space_id", 0),
		}
		first, err := d.Tandoor.Ingredients().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[ingredient.Ingredient], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Ingredients().List(ctx, &o2)
		})
	}

	get := mcpgo.NewTool("ingredient_get",
		mcpgo.WithDescription("Get a single ingredient by ID (includes its food, unit, and amount)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Ingredient ID")),
		jqParam(),
	)
	getHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Ingredients().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- writes ---
	ingredientWriteFields := []mcpgo.ToolOption{
		mcpgo.WithInteger("food_id", mcpgo.Description("Food ID (0 for header-only ingredients)")),
		mcpgo.WithInteger("unit_id", mcpgo.Description("Unit ID (0 for none)")),
		mcpgo.WithNumber("amount", mcpgo.Description("Amount (default 0)")),
		mcpgo.WithString("note", mcpgo.Description("Note (e.g. chopped, optional)")),
		mcpgo.WithInteger("order", mcpgo.Description("Order within the step")),
		mcpgo.WithBoolean("is_header", mcpgo.Description("Mark as a section header")),
		mcpgo.WithBoolean("no_amount", mcpgo.Description("Ingredient has no amount (e.g. pinch)")),
		mcpgo.WithString("original_text", mcpgo.Description("Original ingredient text (for imports)")),
	}
	// buildIngredient constructs an ingredient write payload. full is true
	// for create/update, where the API requires the food, unit, and amount
	// keys (null food/unit = no food/unit; amount defaults to 0). For patch
	// the keys are omitted unless provided.
	buildIngredient := func(id int, req mcpgo.CallToolRequest, full bool) (*ingredient.Ingredient, error) {
		i := &ingredient.Ingredient{ID: id, ForceRefs: full}
		if v, err := req.RequireInt("food_id"); err == nil && v != 0 {
			i.FoodID = v
		}
		if v, err := req.RequireInt("unit_id"); err == nil && v != 0 {
			i.UnitID = v
		}
		if v, err := req.RequireFloat("amount"); err == nil || full {
			i.Amount = &v
		}
		if v, err := req.RequireString("note"); err == nil {
			i.Note = v
		}
		if v, err := req.RequireInt("order"); err == nil {
			i.Order = v
		}
		if v, err := req.RequireBool("is_header"); err == nil {
			i.IsHeader = v
		}
		if v, err := req.RequireBool("no_amount"); err == nil {
			i.NoAmount = v
		}
		if v, err := req.RequireString("original_text"); err == nil {
			i.OriginalText = v
		}
		return i, nil
	}

	ingCreateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Create a standalone ingredient. Note: the API only shows ingredients attached to a visible recipe's steps, so prefer creating ingredients inside a recipe (recipe_create/update); use ingredient_patch/delete on recipe ingredients. Returns the created ingredient."),
		dryRunParam(),
	}, ingredientWriteFields...)
	ingCreate := mcpgo.NewTool("ingredient_create", ingCreateOpts...)
	ingCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		i, err := buildIngredient(0, req, true)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "POST", "api/ingredient/", i, func() (any, error) {
			return d.Tandoor.Ingredients().Create(ctx, i)
		})
	}

	ingUpdateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Update an ingredient (full replacement). Returns the updated ingredient."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Ingredient ID")),
		dryRunParam(),
	}, ingredientWriteFields...)
	ingUpdate := mcpgo.NewTool("ingredient_update", ingUpdateOpts...)
	ingUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		i, err := buildIngredient(id, req, true)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/ingredient/%d/", id), i, func() (any, error) {
			return d.Tandoor.Ingredients().Update(ctx, i)
		})
	}

	ingPatchOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Partially update an ingredient. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Ingredient ID")),
		dryRunParam(),
	}, ingredientWriteFields...)
	ingPatch := mcpgo.NewTool("ingredient_patch", ingPatchOpts...)
	ingPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		i, err := buildIngredient(id, req, false)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/ingredient/%d/", id), i, func() (any, error) {
			return d.Tandoor.Ingredients().Patch(ctx, i)
		})
	}

	ingDelete := mcpgo.NewTool("ingredient_delete",
		mcpgo.WithDescription("Delete an ingredient. Fails if it is still in use."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Ingredient ID")),
		dryRunParam(),
	)
	ingDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/ingredient/%d/", id), func() error {
			return d.Tandoor.Ingredients().Delete(ctx, id)
		})
	}

	return []toolDef{
		{tool: list, handler: listHandler},
		{tool: get, handler: getHandler},
		{tool: ingCreate, handler: ingCreateHandler, write: true},
		{tool: ingUpdate, handler: ingUpdateHandler, write: true},
		{tool: ingPatch, handler: ingPatchHandler, write: true},
		{tool: ingDelete, handler: ingDeleteHandler, write: true},
	}
}
