package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/unit"
)

func registerUnitTools(d *deps) []toolDef {
	// --- units ---
	unitListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List units of measurement. Use all=true for the full set; jq projects fields to keep output small. Unit IDs are instance-specific — enumerate before referencing them in writes.")}
	unitListOpts = append(unitListOpts, baselineListParams()...)
	unitList := mcpgo.NewTool("unit_list", unitListOpts...)
	unitListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &unit.ListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.Units().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[unit.Unit], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Units().List(ctx, &o2)
		})
	}

	unitGet := mcpgo.NewTool("unit_get",
		mcpgo.WithDescription("Get a single unit by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Unit ID")),
		jqParam(),
	)
	unitGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Units().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- unit conversions ---
	convListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List unit conversions (how much of a unit a food weighs/volumes). Filter by food. Use all=true for the full set; jq projects fields to keep output small.")}
	convListOpts = append(convListOpts, baselineListParams()...)
	convListOpts = append(convListOpts,
		mcpgo.WithInteger("food_id", mcpgo.Description("Filter by food ID")),
	)
	convList := mcpgo.NewTool("unit_conversion_list", convListOpts...)
	convListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		base := baseListOptions(req, d)
		opts := &unit.ConversionListOptions{ListOptions: &base}
		if id := req.GetInt("food_id", 0); id != 0 {
			opts.FoodID = &id
		}
		first, err := d.Tandoor.UnitConversions().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[unit.Conversion], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.UnitConversions().List(ctx, &o2)
		})
	}

	convGet := mcpgo.NewTool("unit_conversion_get",
		mcpgo.WithDescription("Get a single unit conversion by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Conversion ID")),
		jqParam(),
	)
	convGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.UnitConversions().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- unit writes ---
	unitCreate := mcpgo.NewTool("unit_create",
		mcpgo.WithDescription("Create a unit of measurement. Returns the created unit."),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Unit name")),
		mcpgo.WithString("plural_name", mcpgo.Description("Plural name (defaults to name)")),
		mcpgo.WithString("description", mcpgo.Description("Unit description")),
		mcpgo.WithString("base_unit", mcpgo.Description("Base unit slug (e.g. gram, milliliter)")),
		dryRunParam(),
	)
	unitCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		u := &unit.Unit{Name: name, PluralName: req.GetString("plural_name", ""), Description: req.GetString("description", ""), BaseUnit: req.GetString("base_unit", "")}
		return runWrite(ctx, req, "POST", "api/unit/", u, func() (any, error) {
			return d.Tandoor.Units().Create(ctx, u)
		})
	}

	unitUpdate := mcpgo.NewTool("unit_update",
		mcpgo.WithDescription("Update a unit (full replacement). Returns the updated unit."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Unit ID")),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Unit name")),
		mcpgo.WithString("plural_name", mcpgo.Description("Plural name")),
		mcpgo.WithString("description", mcpgo.Description("Unit description")),
		mcpgo.WithString("base_unit", mcpgo.Description("Base unit slug")),
		dryRunParam(),
	)
	unitUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		u := &unit.Unit{ID: id, Name: name, PluralName: req.GetString("plural_name", ""), Description: req.GetString("description", ""), BaseUnit: req.GetString("base_unit", "")}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/unit/%d/", id), u, func() (any, error) {
			return d.Tandoor.Units().Update(ctx, u)
		})
	}

	unitPatch := mcpgo.NewTool("unit_patch",
		mcpgo.WithDescription("Partially update a unit. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Unit ID")),
		mcpgo.WithString("name", mcpgo.Description("Unit name")),
		mcpgo.WithString("plural_name", mcpgo.Description("Plural name")),
		mcpgo.WithString("description", mcpgo.Description("Unit description")),
		mcpgo.WithString("base_unit", mcpgo.Description("Base unit slug")),
		dryRunParam(),
	)
	unitPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		u := &unit.Unit{ID: id}
		if v, err := req.RequireString("name"); err == nil {
			u.Name = v
		}
		if v, err := req.RequireString("plural_name"); err == nil {
			u.PluralName = v
		}
		if v, err := req.RequireString("description"); err == nil {
			u.Description = v
		}
		if v, err := req.RequireString("base_unit"); err == nil {
			u.BaseUnit = v
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/unit/%d/", id), u, func() (any, error) {
			return d.Tandoor.Units().Patch(ctx, u)
		})
	}

	unitDelete := mcpgo.NewTool("unit_delete",
		mcpgo.WithDescription("Delete a unit. Fails if the unit is still in use."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Unit ID")),
		dryRunParam(),
	)
	unitDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/unit/%d/", id), func() error {
			return d.Tandoor.Units().Delete(ctx, id)
		})
	}

	unitMerge := mcpgo.NewTool("unit_merge",
		mcpgo.WithDescription("Merge one unit into another: usages re-point to the target, then the source is deleted."),
		mcpgo.WithInteger("source_id", mcpgo.Required(), mcpgo.Description("Unit to merge (deleted)")),
		mcpgo.WithInteger("target_id", mcpgo.Required(), mcpgo.Description("Unit to keep")),
		dryRunParam(),
	)
	unitMergeHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		sourceID, err := req.RequireInt("source_id")
		if err != nil {
			return errResult(err), nil
		}
		targetID, err := req.RequireInt("target_id")
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/unit/%d/merge/%d/", sourceID, targetID), nil, func() (any, error) {
			return d.Tandoor.Units().Merge(ctx, sourceID, targetID)
		})
	}

	// --- unit conversion writes ---
	convRequiredFields := []mcpgo.ToolOption{
		mcpgo.WithInteger("base_unit_id", mcpgo.Required(), mcpgo.Description("Base unit ID")),
		mcpgo.WithNumber("base_amount", mcpgo.Required(), mcpgo.Description("Amount in the base unit")),
		mcpgo.WithInteger("converted_unit_id", mcpgo.Required(), mcpgo.Description("Converted unit ID")),
		mcpgo.WithNumber("converted_amount", mcpgo.Required(), mcpgo.Description("Amount in the converted unit")),
		mcpgo.WithInteger("food_id", mcpgo.Description("Food this conversion applies to (0 for generic)")),
	}
	buildConversion := func(id int, req mcpgo.CallToolRequest) (*unit.Conversion, error) {
		baseUnit, err := req.RequireInt("base_unit_id")
		if err != nil {
			return nil, err
		}
		convertedUnit, err := req.RequireInt("converted_unit_id")
		if err != nil {
			return nil, err
		}
		baseAmount, err := req.RequireFloat("base_amount")
		if err != nil {
			return nil, err
		}
		convertedAmount, err := req.RequireFloat("converted_amount")
		if err != nil {
			return nil, err
		}
		conv := &unit.Conversion{
			ID:              id,
			BaseUnit:        &unit.Ref{ID: baseUnit},
			ConvertedUnit:   &unit.Ref{ID: convertedUnit},
			BaseAmount:      baseAmount,
			ConvertedAmount: convertedAmount,
		}
		if foodID, err := req.RequireInt("food_id"); err == nil && foodID != 0 {
			conv.Food = &unit.FoodRef{ID: foodID}
		}
		return conv, nil
	}

	convCreateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Create a unit conversion. Returns the created conversion."),
		dryRunParam(),
	}, convRequiredFields...)
	convCreate := mcpgo.NewTool("unit_conversion_create", convCreateOpts...)
	convCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		conv, err := buildConversion(0, req)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "POST", "api/unit-conversion/", conv, func() (any, error) {
			return d.Tandoor.UnitConversions().Create(ctx, conv)
		})
	}

	convUpdateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Update a unit conversion. All conversion fields are required by the API. Returns the updated conversion."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Conversion ID")),
		dryRunParam(),
	}, convRequiredFields...)
	convUpdate := mcpgo.NewTool("unit_conversion_update", convUpdateOpts...)
	convUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		conv, err := buildConversion(id, req)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/unit-conversion/%d/", id), conv, func() (any, error) {
			return d.Tandoor.UnitConversions().Update(ctx, conv)
		})
	}

	convPatchOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Partially update a unit conversion. Note: the API still requires base/converted unit and amounts."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Conversion ID")),
		dryRunParam(),
	}, convRequiredFields...)
	convPatch := mcpgo.NewTool("unit_conversion_patch", convPatchOpts...)
	convPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		conv, err := buildConversion(id, req)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/unit-conversion/%d/", id), conv, func() (any, error) {
			return d.Tandoor.UnitConversions().Patch(ctx, conv)
		})
	}

	convDelete := mcpgo.NewTool("unit_conversion_delete",
		mcpgo.WithDescription("Delete a unit conversion."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Conversion ID")),
		dryRunParam(),
	)
	convDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/unit-conversion/%d/", id), func() error {
			return d.Tandoor.UnitConversions().Delete(ctx, id)
		})
	}

	return []toolDef{
		{tool: unitList, handler: unitListHandler},
		{tool: unitGet, handler: unitGetHandler},
		{tool: convList, handler: convListHandler},
		{tool: convGet, handler: convGetHandler},
		{tool: unitCreate, handler: unitCreateHandler, write: true},
		{tool: unitUpdate, handler: unitUpdateHandler, write: true},
		{tool: unitPatch, handler: unitPatchHandler, write: true},
		{tool: unitDelete, handler: unitDeleteHandler, write: true},
		{tool: unitMerge, handler: unitMergeHandler, write: true},
		{tool: convCreate, handler: convCreateHandler, write: true},
		{tool: convUpdate, handler: convUpdateHandler, write: true},
		{tool: convPatch, handler: convPatchHandler, write: true},
		{tool: convDelete, handler: convDeleteHandler, write: true},
	}
}
