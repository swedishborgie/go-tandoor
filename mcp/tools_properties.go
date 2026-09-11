package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/auditfood"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/property"
)

func registerPropertyTools(d *deps) []toolDef {
	// --- properties (property values on foods) ---
	propListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List property values attached to foods (e.g. calories per 100 g). Filter by food. Use all=true for the full set; jq projects fields to keep output small.")}
	propListOpts = append(propListOpts, baselineListParams()...)
	propListOpts = append(propListOpts,
		mcpgo.WithInteger("food_id", mcpgo.Description("Filter by food ID")),
	)
	propList := mcpgo.NewTool("property_list", propListOpts...)
	propListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		base := baseListOptions(req, d)
		opts := &property.ListOptions{ListOptions: &base}
		if id := req.GetInt("food_id", 0); id != 0 {
			opts.FoodID = &id
		}
		first, err := d.Tandoor.Properties().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[property.Property], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Properties().List(ctx, &o2)
		})
	}

	propGet := mcpgo.NewTool("property_get",
		mcpgo.WithDescription("Get a single property value by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Property ID")),
		jqParam(),
	)
	propGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Properties().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- property types ---
	typeListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List property types (e.g. Calories, Protein). Types define which properties foods can have; their IDs are instance-specific — enumerate before referencing them in writes.")}
	typeListOpts = append(typeListOpts, baselineListParams()...)
	typeList := mcpgo.NewTool("property_type_list", typeListOpts...)
	typeListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &property.TypeListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.PropertyTypes().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[property.Type], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.PropertyTypes().List(ctx, &o2)
		})
	}

	typeGet := mcpgo.NewTool("property_type_get",
		mcpgo.WithDescription("Get a single property type by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Property type ID")),
		jqParam(),
	)
	typeGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.PropertyTypes().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- property writes (standalone values; food attachment is via property_attach) ---
	propCreate := mcpgo.NewTool("property_create",
		mcpgo.WithDescription("Create a property value. Returns the created property. Property type IDs are instance-specific — use property_type_list first."),
		mcpgo.WithInteger("property_type_id", mcpgo.Required(), mcpgo.Description("Property type ID")),
		mcpgo.WithNumber("property_amount", mcpgo.Description("Property amount")),
	)
	propCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		typeID, err := req.RequireInt("property_type_id")
		if err != nil {
			return errResult(err), nil
		}
		p := &property.Property{Type: property.Type{ID: typeID}}
		if v, err := req.RequireFloat("property_amount"); err == nil {
			p.PropertyAmount = &v
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.Properties().Create(ctx, p)
		})
	}

	propUpdate := mcpgo.NewTool("property_update",
		mcpgo.WithDescription("Update a property value (full replacement). Returns the updated property."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Property ID")),
		mcpgo.WithInteger("property_type_id", mcpgo.Required(), mcpgo.Description("Property type ID")),
		mcpgo.WithNumber("property_amount", mcpgo.Description("Property amount")),
	)
	propUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		typeID, err := req.RequireInt("property_type_id")
		if err != nil {
			return errResult(err), nil
		}
		p := &property.Property{ID: id, Type: property.Type{ID: typeID}}
		if v, err := req.RequireFloat("property_amount"); err == nil {
			p.PropertyAmount = &v
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.Properties().Update(ctx, p)
		})
	}

	propPatch := mcpgo.NewTool("property_patch",
		mcpgo.WithDescription("Partially update a property value. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Property ID")),
		mcpgo.WithInteger("property_type_id", mcpgo.Description("Property type ID")),
		mcpgo.WithNumber("property_amount", mcpgo.Description("Property amount")),
	)
	propPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		p := &property.Property{ID: id}
		if v, err := req.RequireInt("property_type_id"); err == nil {
			p.Type.ID = v
		}
		if v, err := req.RequireFloat("property_amount"); err == nil {
			p.PropertyAmount = &v
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.Properties().Patch(ctx, p)
		})
	}

	propDelete := mcpgo.NewTool("property_delete",
		mcpgo.WithDescription("Delete a property value."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Property ID")),
	)
	propDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(fmt.Sprintf("api/property/%d/", id), func() error {
			return d.Tandoor.Properties().Delete(ctx, id)
		})
	}

	// --- property type writes ---
	typeCreate := mcpgo.NewTool("property_type_create",
		mcpgo.WithDescription("Create a property type (e.g. calories, protein). Returns the created type."),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Type name")),
		mcpgo.WithString("unit", mcpgo.Description("Unit (e.g. g, %)")),
		mcpgo.WithString("description", mcpgo.Description("Description")),
		mcpgo.WithInteger("order", mcpgo.Description("Display order")),
		mcpgo.WithInteger("fdc_id", mcpgo.Description("USDA FoodData Central nutrient ID")),
	)
	typeCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		t := &property.Type{Name: name, Unit: req.GetString("unit", ""), Description: req.GetString("description", "")}
		if v, err := req.RequireInt("order"); err == nil {
			t.Order = v
		}
		if v, err := req.RequireInt("fdc_id"); err == nil && v != 0 {
			t.FDCID = &v
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.PropertyTypes().Create(ctx, t)
		})
	}

	typeUpdate := mcpgo.NewTool("property_type_update",
		mcpgo.WithDescription("Update a property type (full replacement). Returns the updated type."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Property type ID")),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Type name")),
		mcpgo.WithString("unit", mcpgo.Description("Unit")),
		mcpgo.WithString("description", mcpgo.Description("Description")),
		mcpgo.WithInteger("order", mcpgo.Description("Display order")),
		mcpgo.WithInteger("fdc_id", mcpgo.Description("USDA FoodData Central nutrient ID")),
	)
	typeUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		t := &property.Type{ID: id, Name: name, Unit: req.GetString("unit", ""), Description: req.GetString("description", "")}
		if v, err := req.RequireInt("order"); err == nil {
			t.Order = v
		}
		if v, err := req.RequireInt("fdc_id"); err == nil && v != 0 {
			t.FDCID = &v
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.PropertyTypes().Update(ctx, t)
		})
	}

	typePatch := mcpgo.NewTool("property_type_patch",
		mcpgo.WithDescription("Partially update a property type. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Property type ID")),
		mcpgo.WithString("name", mcpgo.Description("Type name")),
		mcpgo.WithString("unit", mcpgo.Description("Unit")),
		mcpgo.WithString("description", mcpgo.Description("Description")),
		mcpgo.WithInteger("order", mcpgo.Description("Display order")),
		mcpgo.WithInteger("fdc_id", mcpgo.Description("USDA FoodData Central nutrient ID")),
	)
	typePatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		t := &property.Type{ID: id}
		if v, err := req.RequireString("name"); err == nil {
			t.Name = v
		}
		if v, err := req.RequireString("unit"); err == nil {
			t.Unit = v
		}
		if v, err := req.RequireString("description"); err == nil {
			t.Description = v
		}
		if v, err := req.RequireInt("order"); err == nil {
			t.Order = v
		}
		if v, err := req.RequireInt("fdc_id"); err == nil && v != 0 {
			t.FDCID = &v
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.PropertyTypes().Patch(ctx, t)
		})
	}

	typeDelete := mcpgo.NewTool("property_type_delete",
		mcpgo.WithDescription("Delete a property type. Fails if property values still use it."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Property type ID")),
	)
	typeDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(fmt.Sprintf("api/property-type/%d/", id), func() error {
			return d.Tandoor.PropertyTypes().Delete(ctx, id)
		})
	}

	attach := mcpgo.NewTool("property_attach",
		mcpgo.WithDescription("Attach a nutrient property to a food per 100 g (composite: resolves the food, sets the per-100 unit, creates or updates the property). Idempotent — re-running updates the existing property."),
		mcpgo.WithInteger("food_id", mcpgo.Description("Target food ID (or ingredient_id; exactly one required)")),
		mcpgo.WithInteger("ingredient_id", mcpgo.Description("Ingredient whose food is the target (or food_id; exactly one required)")),
		mcpgo.WithInteger("property_type_id", mcpgo.Required(), mcpgo.Description("Property type ID (e.g. 1 = calories; see property_type_list)")),
		mcpgo.WithNumber("property_amount", mcpgo.Required(), mcpgo.Description("Amount per 100 g")),
		mcpgo.WithNumber("per_100_amount", mcpgo.Description("Basis amount for the food's properties_food_amount (default 100)")),
		mcpgo.WithInteger("per_100_unit_id", mcpgo.Description("Basis unit ID for properties_food_unit (default: food's existing unit, else the 'gram' unit)")),
		jqParam(),
	)
	attachHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		foodID := req.GetInt("food_id", 0)
		ingredientID := req.GetInt("ingredient_id", 0)
		if (foodID == 0) == (ingredientID == 0) {
			return errResult(fmt.Errorf("exactly one of food_id or ingredient_id is required")), nil
		}
		result, err := auditfood.Attach(ctx, d.Tandoor, &auditfood.AttachOptions{
			FoodID:         foodID,
			IngredientID:   ingredientID,
			PropertyTypeID: req.GetInt("property_type_id", 0),
			Amount:         req.GetFloat("property_amount", 0),
			Per100Amount:   req.GetFloat("per_100_amount", 0),
			Per100UnitID:   req.GetInt("per_100_unit_id", 0),
		}, false)
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

	// attachMany is the batch counterpart of attach: several per-100-g
	// properties in one food PATCH, resolved by type ID or name.
	attachMany := mcpgo.NewTool("property_attach_many",
		mcpgo.WithDescription("Attach or update several per-100-g properties on a food in a single PATCH (idempotent). Each entry needs amount plus property_type_id or property_type_name (resolved case-insensitively against the instance's property types — see instance_vocab or property_type_list). Duplicate types collapse, last entry wins. Use for FDC nutrients the API is missing (e.g. fiber, minerals, vitamins)."),
		mcpgo.WithInteger("food_id", mcpgo.Required(), mcpgo.Description("Target food")),
		mcpgo.WithArray("properties", mcpgo.Required(),
			mcpgo.Description(`Array of {property_type_id | property_type_name, amount}`),
			func(schema map[string]any) {
				schema["items"] = map[string]any{
					"type": "object",
					"properties": map[string]any{
						"property_type_id":   map[string]any{"type": "integer"},
						"property_type_name": map[string]any{"type": "string"},
						"amount":             map[string]any{"type": "number"},
					},
					"required": []string{"amount"},
				}
			},
		),
		mcpgo.WithNumber("per_100_amount", mcpgo.Description("Basis amount for the food's properties_food_amount (default 100)")),
		mcpgo.WithInteger("per_100_unit_id", mcpgo.Description("Basis unit ID for properties_food_unit (default: food's existing unit, else the instance's gram unit)")),
		jqParam(),
	)
	attachManyHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		list, _ := req.GetArguments()["properties"].([]any)
		props := make([]auditfood.ManyProperty, 0, len(list))
		for i, e := range list {
			m, ok := e.(map[string]any)
			if !ok {
				return errResult(fmt.Errorf("properties[%d] must be an object", i)), nil
			}
			p := auditfood.ManyProperty{}
			if v, ok := m["property_type_id"].(float64); ok {
				p.PropertyTypeID = int(v)
			}
			p.PropertyTypeName, _ = m["property_type_name"].(string)
			v, ok := m["amount"].(float64)
			if !ok {
				return errResult(fmt.Errorf("properties[%d] requires a numeric amount", i)), nil
			}
			p.Amount = v
			props = append(props, p)
		}
		result, err := auditfood.AttachMany(ctx, d.Tandoor, &auditfood.AttachManyOptions{
			FoodID:       req.GetInt("food_id", 0),
			Properties:   props,
			Per100Amount: req.GetFloat("per_100_amount", 0),
			Per100UnitID: req.GetInt("per_100_unit_id", 0),
		})
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
		{tool: propList, handler: propListHandler},
		{tool: propGet, handler: propGetHandler},
		{tool: typeList, handler: typeListHandler},
		{tool: typeGet, handler: typeGetHandler},
		{tool: propCreate, handler: propCreateHandler, write: true},
		{tool: propUpdate, handler: propUpdateHandler, write: true},
		{tool: propPatch, handler: propPatchHandler, write: true},
		{tool: propDelete, handler: propDeleteHandler, write: true},
		{tool: typeCreate, handler: typeCreateHandler, write: true},
		{tool: typeUpdate, handler: typeUpdateHandler, write: true},
		{tool: typePatch, handler: typePatchHandler, write: true},
		{tool: typeDelete, handler: typeDeleteHandler, write: true},
		{tool: attach, handler: attachHandler, write: true},
		{tool: attachMany, handler: attachManyHandler, write: true},
	}
}
