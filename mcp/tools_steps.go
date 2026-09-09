package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/step"
)

func registerStepTools(d *deps) []toolDef {
	listOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List steps. Filter by recipe. Use all=true for the full set; jq projects fields to keep output small.")}
	listOpts = append(listOpts, baselineListParams()...)
	listOpts = append(listOpts,
		mcpgo.WithInteger("recipe_id", mcpgo.Description("Filter by recipe ID")),
	)
	list := mcpgo.NewTool("step_list", listOpts...)
	listHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &step.ListOptions{
			ListOptions: baseListOptions(req, d),
			RecipeID:    req.GetInt("recipe_id", 0),
		}
		first, err := d.Tandoor.Steps().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[step.Step], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Steps().List(ctx, &o2)
		})
	}

	get := mcpgo.NewTool("step_get",
		mcpgo.WithDescription("Get a single step by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Step ID")),
		jqParam(),
	)
	getHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Steps().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- writes ---
	stepWriteFields := []mcpgo.ToolOption{
		mcpgo.WithString("name", mcpgo.Description("Step name (optional, for titled steps)")),
		mcpgo.WithString("instruction", mcpgo.Description("Step instruction text")),
		mcpgo.WithInteger("time", mcpgo.Description("Step time in minutes")),
		mcpgo.WithInteger("order", mcpgo.Description("Order within the recipe")),
		mcpgo.WithBoolean("show_as_header", mcpgo.Description("Render the step name as a header")),
		mcpgo.WithArray("ingredient_ids", mcpgo.WithIntegerItems(), mcpgo.Description("IDs of ingredients attached to the step (replaces the current set)")),
	}
	// buildStep constructs a step write payload. requireIngredients is true
	// for create/update, where the API mandates the ingredients key (empty
	// list = no ingredients); false for patch, where nil omits the key.
	buildStep := func(id int, req mcpgo.CallToolRequest, requireIngredients bool) (*step.Step, error) {
		s := &step.Step{ID: id}
		if v, err := req.RequireString("name"); err == nil {
			s.Name = v
		}
		if v, err := req.RequireString("instruction"); err == nil {
			s.Instruction = v
		}
		if v, err := req.RequireInt("time"); err == nil {
			s.Time = v
		}
		if v, err := req.RequireInt("order"); err == nil {
			s.Order = v
		}
		if v, err := req.RequireBool("show_as_header"); err == nil {
			s.ShowAsHeader = v
		}
		if ids := req.GetIntSlice("ingredient_ids", nil); len(ids) > 0 {
			s.Ingredients = ids
		}
		if requireIngredients && s.Ingredients == nil {
			s.Ingredients = []int{}
		}
		return s, nil
	}

	stCreate := mcpgo.NewTool("step_create",
		mcpgo.WithDescription("Create a standalone step. Note: the API only shows steps attached to a visible recipe, so use step_update/patch/delete on steps created inside a recipe (via recipe_create/update). Returns the created step."),
		mcpgo.WithString("instruction", mcpgo.Description("Step instruction text")),
		mcpgo.WithString("name", mcpgo.Description("Step name (optional)")),
		mcpgo.WithInteger("time", mcpgo.Description("Step time in minutes")),
		mcpgo.WithInteger("order", mcpgo.Description("Order within the recipe")),
		mcpgo.WithBoolean("show_as_header", mcpgo.Description("Render the step name as a header")),
		dryRunParam(),
	)
	stCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		s, err := buildStep(0, req, true)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "POST", "api/step/", s, func() (any, error) {
			return d.Tandoor.Steps().Create(ctx, s)
		})
	}

	stUpdateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Update a step (full replacement). Returns the updated step."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Step ID")),
		dryRunParam(),
	}, stepWriteFields...)
	stUpdate := mcpgo.NewTool("step_update", stUpdateOpts...)
	stUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		s, err := buildStep(id, req, true)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/step/%d/", id), s, func() (any, error) {
			return d.Tandoor.Steps().Update(ctx, s)
		})
	}

	stPatchOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Partially update a step. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Step ID")),
		dryRunParam(),
	}, stepWriteFields...)
	stPatch := mcpgo.NewTool("step_patch", stPatchOpts...)
	stPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		s, err := buildStep(id, req, false)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/step/%d/", id), s, func() (any, error) {
			return d.Tandoor.Steps().Patch(ctx, s)
		})
	}

	stDelete := mcpgo.NewTool("step_delete",
		mcpgo.WithDescription("Delete a step (removes it from its recipe)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Step ID")),
		dryRunParam(),
	)
	stDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/step/%d/", id), func() error {
			return d.Tandoor.Steps().Delete(ctx, id)
		})
	}

	return []toolDef{
		{tool: list, handler: listHandler},
		{tool: get, handler: getHandler},
		{tool: stCreate, handler: stCreateHandler, write: true},
		{tool: stUpdate, handler: stUpdateHandler, write: true},
		{tool: stPatch, handler: stPatchHandler, write: true},
		{tool: stDelete, handler: stDeleteHandler, write: true},
	}
}
