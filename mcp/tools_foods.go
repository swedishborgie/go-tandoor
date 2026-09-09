package mcp

import (
	"context"
	"fmt"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/auditfood"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func registerFoodTools(d *deps) []toolDef {
	listOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List foods. Use query for fuzzy name search, name_exact for a case-insensitive exact name, or names for a batch exact lookup returning a {name: food|null} map. Filters: category_id, unit_id. Use all=true for the full set; jq projects fields to keep output small.")}
	listOpts = append(listOpts, baselineListParams()...)
	listOpts = append(listOpts,
		mcpgo.WithInteger("category_id", mcpgo.Description("Filter by food category ID")),
		mcpgo.WithInteger("unit_id", mcpgo.Description("Filter by unit ID")),
		mcpgo.WithString("name_exact", mcpgo.Description("Case-insensitive exact name match (returns matching foods as an array)")),
		mcpgo.WithArray("names", mcpgo.WithStringItems(), mcpgo.Description("Batch exact lookup; returns a {name: food|null} map")),
	)
	list := mcpgo.NewTool("food_list", listOpts...)
	listHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		// Batch exact lookup: name -> food (or null).
		if names := req.GetStringSlice("names", nil); len(names) > 0 {
			results := make(map[string]any, len(names))
			for _, name := range names {
				page, err := d.Tandoor.Foods().List(ctx, &food.ListOptions{
					ListOptions: pagination.ListOptions{Page: 1, PageSize: 100, Search: name},
				})
				if err != nil {
					return errResult(err), nil
				}
				var found *food.Food
				for i := range page.Results {
					if strings.EqualFold(page.Results[i].Name, name) {
						found = &page.Results[i]
						break
					}
				}
				results[name] = found
			}
			if jq := req.GetString("jq", ""); jq != "" {
				s, err := applyJQ(ctx, jq, results)
				if err != nil {
					return errResult(err), nil
				}
				return mcpgo.NewToolResultText(s), nil
			}
			return jsonResult(results), nil
		}

		// Exact-name match: narrow server-side, enforce client-side.
		if nameExact := req.GetString("name_exact", ""); nameExact != "" {
			matches := make([]food.Food, 0)
			pageNum := 1
			for {
				page, err := d.Tandoor.Foods().List(ctx, &food.ListOptions{
					ListOptions: pagination.ListOptions{
						Page:     pageNum,
						PageSize: d.Cfg.DefaultPageSize,
						Extra:    map[string]string{"query": nameExact},
					},
				})
				if err != nil {
					return errResult(err), nil
				}
				for _, f := range page.Results {
					if strings.EqualFold(f.Name, nameExact) {
						matches = append(matches, f)
					}
				}
				if !page.HasNext() {
					break
				}
				pageNum++
			}
			if jq := req.GetString("jq", ""); jq != "" {
				s, err := applyJQ(ctx, jq, matches)
				if err != nil {
					return errResult(err), nil
				}
				return mcpgo.NewToolResultText(s), nil
			}
			return jsonResult(matches), nil
		}

		opts := &food.ListOptions{
			ListOptions: baseListOptions(req, d),
			CategoryID:  req.GetInt("category_id", 0),
			UnitID:      req.GetInt("unit_id", 0),
		}
		first, err := d.Tandoor.Foods().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[food.Food], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Foods().List(ctx, &o2)
		})
	}

	get := mcpgo.NewTool("food_get",
		mcpgo.WithDescription("Get a single food by ID (includes category, unit, and properties)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Food ID")),
		jqParam(),
	)
	getHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Foods().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- writes ---
	foodWriteFields := []mcpgo.ToolOption{
		mcpgo.WithString("plural_name", mcpgo.Description("Plural name (defaults to name)")),
		mcpgo.WithString("description", mcpgo.Description("Food description")),
		mcpgo.WithInteger("fdc_id", mcpgo.Description("USDA FoodData Central ID")),
		mcpgo.WithInteger("parent_id", mcpgo.Description("Parent food ID for the food tree (0 for top level)")),
		mcpgo.WithBoolean("ignore_shopping", mcpgo.Description("Exclude this food from shopping lists")),
		mcpgo.WithString("open_data_slug", mcpgo.Description("Open data slug (only for open data foods)")),
	}
	buildFood := func(id int, req mcpgo.CallToolRequest) (*food.Food, error) {
		f := &food.Food{ID: id}
		if v, err := req.RequireString("name"); err == nil {
			f.Name = v
		}
		if v, err := req.RequireString("plural_name"); err == nil {
			f.PluralName = v
		}
		if v, err := req.RequireString("description"); err == nil {
			f.Description = v
		}
		if v, err := req.RequireInt("fdc_id"); err == nil && v != 0 {
			f.FDCID = &v
		}
		if v, err := req.RequireInt("parent_id"); err == nil {
			f.Parent = v
		}
		if v, err := req.RequireBool("ignore_shopping"); err == nil {
			f.IgnoreShopping = v
		}
		if v, err := req.RequireString("open_data_slug"); err == nil {
			f.OpenDataSlug = v
		}
		return f, nil
	}

	foodCreateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Create a food. Returns the created food."),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Food name")),
		dryRunParam(),
	}, foodWriteFields...)
	foodCreate := mcpgo.NewTool("food_create", foodCreateOpts...)
	foodCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		f, err := buildFood(0, req)
		if err != nil {
			return errResult(err), nil
		}
		f.Name = name
		return runWrite(ctx, req, "POST", "api/food/", f, func() (any, error) {
			return d.Tandoor.Foods().Create(ctx, f)
		})
	}

	foodUpdateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Update a food (full replacement). Returns the updated food."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Food ID")),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Food name")),
		dryRunParam(),
	}, foodWriteFields...)
	foodUpdate := mcpgo.NewTool("food_update", foodUpdateOpts...)
	foodUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		f, err := buildFood(id, req)
		if err != nil {
			return errResult(err), nil
		}
		f.Name = name
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/food/%d/", id), f, func() (any, error) {
			return d.Tandoor.Foods().Update(ctx, f)
		})
	}

	foodPatchOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Partially update a food. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Food ID")),
		dryRunParam(),
	}, foodWriteFields...)
	foodPatch := mcpgo.NewTool("food_patch", foodPatchOpts...)
	foodPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		f, err := buildFood(id, req)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/food/%d/", id), f, func() (any, error) {
			return d.Tandoor.Foods().Patch(ctx, f)
		})
	}

	foodDelete := mcpgo.NewTool("food_delete",
		mcpgo.WithDescription("Delete a food. Fails if the food is still in use."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Food ID")),
		dryRunParam(),
	)
	foodDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/food/%d/", id), func() error {
			return d.Tandoor.Foods().Delete(ctx, id)
		})
	}

	foodMerge := mcpgo.NewTool("food_merge",
		mcpgo.WithDescription("Merge one food into another: usages re-point to the target, then the source is deleted."),
		mcpgo.WithInteger("source_id", mcpgo.Required(), mcpgo.Description("Food to merge (deleted)")),
		mcpgo.WithInteger("target_id", mcpgo.Required(), mcpgo.Description("Food to keep")),
		dryRunParam(),
	)
	foodMergeHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		sourceID, err := req.RequireInt("source_id")
		if err != nil {
			return errResult(err), nil
		}
		targetID, err := req.RequireInt("target_id")
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/food/%d/merge/%d/", sourceID, targetID), nil, func() (any, error) {
			return d.Tandoor.Foods().Merge(ctx, sourceID, targetID)
		})
	}

	foodMove := mcpgo.NewTool("food_move",
		mcpgo.WithDescription("Move a food under a new parent in the food tree."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Food ID")),
		mcpgo.WithInteger("parent_id", mcpgo.Required(), mcpgo.Description("New parent food ID (0 for top level)")),
		dryRunParam(),
	)
	foodMoveHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		parentID, err := req.RequireInt("parent_id")
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/food/%d/move/%d/", id, parentID), nil, func() (any, error) {
			return d.Tandoor.Foods().Move(ctx, id, parentID)
		})
	}

	foodBatchUpdate := mcpgo.NewTool("food_batch_update",
		mcpgo.WithDescription("Batch update foods: add, remove, or replace substitutes for a set of foods at once."),
		mcpgo.WithArray("foods", mcpgo.WithIntegerItems(), mcpgo.Required(), mcpgo.Description("Food IDs to update")),
		mcpgo.WithArray("substitute_add", mcpgo.WithIntegerItems(), mcpgo.Description("Substitute food IDs to add")),
		mcpgo.WithArray("substitute_remove", mcpgo.WithIntegerItems(), mcpgo.Description("Substitute food IDs to remove")),
		mcpgo.WithArray("substitute_set", mcpgo.WithIntegerItems(), mcpgo.Description("Substitute food IDs to set (replaces all)")),
		dryRunParam(),
	)
	foodBatchUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		foods, err := req.RequireIntSlice("foods")
		if err != nil {
			return errResult(err), nil
		}
		update := &food.BatchUpdate{Foods: foods}
		if s, err := req.RequireIntSlice("substitute_add"); err == nil {
			update.SubstituteAdd = s
		}
		if s, err := req.RequireIntSlice("substitute_remove"); err == nil {
			update.SubstituteRemove = s
		}
		if s, err := req.RequireIntSlice("substitute_set"); err == nil {
			update.SubstituteSet = s
		}
		return runWrite(ctx, req, "POST", "api/food/batch_update/", update, func() (any, error) {
			return d.Tandoor.Foods().BatchUpdate(ctx, update)
		})
	}

	fdcImport := mcpgo.NewTool("food_fdc_import",
		mcpgo.WithDescription("Pull USDA FDC data into a food that already has an fdc_id set (populates properties and conversions server-side)."),
		mcpgo.WithInteger("food_id", mcpgo.Required(), mcpgo.Description("Food ID")),
		dryRunParam(),
		jqParam(),
	)
	fdcImportHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id := req.GetInt("food_id", 0)
		return runWrite(ctx, req, "POST", fmt.Sprintf("api/food/%d/fdc/", id), nil, func() (any, error) {
			return d.Tandoor.Foods().FdcImport(ctx, id)
		})
	}

	aiProps := mcpgo.NewTool("food_ai_properties",
		mcpgo.WithDescription("Trigger server-side AI to generate properties for a food. Requires an AI provider configured on the instance."),
		mcpgo.WithInteger("food_id", mcpgo.Required(), mcpgo.Description("Food ID")),
		dryRunParam(),
		jqParam(),
	)
	aiPropsHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id := req.GetInt("food_id", 0)
		return runWrite(ctx, req, "POST", fmt.Sprintf("api/food/%d/aiproperties/", id), nil, func() (any, error) {
			return d.Tandoor.Foods().AiProperties(ctx, id)
		})
	}

	ensure := mcpgo.NewTool("food_ensure",
		mcpgo.WithDescription("Ensure foods exist by exact name, creating missing ones (composite of food_list + food_create + FDC candidate lookup). Use dry_run to preview without writing."),
		mcpgo.WithArray("names", mcpgo.WithStringItems(), mcpgo.Required(), mcpgo.Description("Food names to ensure")),
		mcpgo.WithBoolean("force_create", mcpgo.Description("Create when no exact match exists (default true); false only reports")),
		mcpgo.WithNumber("threshold", mcpgo.Description("Jaccard similarity floor for near-match reporting (default 0.6)")),
		mcpgo.WithInteger("fdc_limit", mcpgo.Description("FDC candidates per name when FDC is configured (default 5)")),
		dryRunParam(),
		jqParam(),
	)
	ensureHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		names := req.GetStringSlice("names", nil)
		if len(names) == 0 {
			return errResult(fmt.Errorf("names is required")), nil
		}
		report, err := auditfood.Ensure(ctx, d.Tandoor, &auditfood.EnsureOptions{
			Names:       names,
			FDC:         d.FDC,
			Threshold:   req.GetFloat("threshold", 0),
			FDCLimit:    req.GetInt("fdc_limit", 0),
			ForceCreate: req.GetBool("force_create", true),
		}, req.GetBool("dry_run", false))
		if err != nil {
			return errResult(err), nil
		}
		if jq := req.GetString("jq", ""); jq != "" {
			s, err := applyJQ(ctx, jq, report)
			if err != nil {
				return errResult(err), nil
			}
			return mcpgo.NewToolResultText(s), nil
		}
		if req.GetBool("dry_run", false) {
			return jsonResult(map[string]any{"dry_run": true, "report": report}), nil
		}
		return jsonResult(report), nil
	}

	return []toolDef{
		{tool: list, handler: listHandler},
		{tool: get, handler: getHandler},
		{tool: foodCreate, handler: foodCreateHandler, write: true},
		{tool: foodUpdate, handler: foodUpdateHandler, write: true},
		{tool: foodPatch, handler: foodPatchHandler, write: true},
		{tool: foodDelete, handler: foodDeleteHandler, write: true},
		{tool: foodMerge, handler: foodMergeHandler, write: true},
		{tool: foodMove, handler: foodMoveHandler, write: true},
		{tool: foodBatchUpdate, handler: foodBatchUpdateHandler, write: true},
		{tool: fdcImport, handler: fdcImportHandler, write: true},
		{tool: aiProps, handler: aiPropsHandler, write: true},
		{tool: ensure, handler: ensureHandler, write: true},
	}
}
