package mcp

import (
	"context"
	"fmt"
	"strconv"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/mealplan"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/recipe"
)

func registerMealplanTools(d *deps) []toolDef {
	// --- meal plans ---
	planListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List meal plans. Filter by space, user, and/or a date range (date_from/date_to, YYYY-MM-DD). Use all=true for the full set; jq projects fields to keep output small.")}
	planListOpts = append(planListOpts, baselineListParams(d)...)
	planListOpts = append(planListOpts,
		mcpgo.WithInteger("space_id", mcpgo.Description("Filter by space ID")),
		mcpgo.WithInteger("user_id", mcpgo.Description("Filter by user ID")),
		mcpgo.WithString("date_from", mcpgo.Description("Start date (YYYY-MM-DD)")),
		mcpgo.WithString("date_to", mcpgo.Description("End date (YYYY-MM-DD)")),
	)
	planList := mcpgo.NewTool("meal_plan_list", planListOpts...)
	planListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &mealplan.ListOptions{
			ListOptions: baseListOptions(req, d),
			SpaceID:     req.GetInt("space_id", 0),
			UserID:      req.GetInt("user_id", 0),
			DateFrom:    req.GetString("date_from", ""),
			DateTo:      req.GetString("date_to", ""),
		}
		first, err := d.Tandoor.MealPlans().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[mealplan.MealPlan], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.MealPlans().List(ctx, &o2)
		})
	}

	planGet := mcpgo.NewTool("meal_plan_get",
		mcpgo.WithDescription("Get a single meal plan entry by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Meal plan ID")),
		jqParam(),
	)
	planGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.MealPlans().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- iCal export ---
	icalOpts := []mcpgo.ToolOption{mcpgo.WithDescription("Export meal plans as an iCal/ICS calendar (plain text). Filter by space, user, and/or date range.")}
	icalOpts = append(icalOpts,
		mcpgo.WithInteger("space_id", mcpgo.Description("Filter by space ID")),
		mcpgo.WithInteger("user_id", mcpgo.Description("Filter by user ID")),
		mcpgo.WithString("date_from", mcpgo.Description("Start date (YYYY-MM-DD)")),
		mcpgo.WithString("date_to", mcpgo.Description("End date (YYYY-MM-DD)")),
	)
	ical := mcpgo.NewTool("meal_plan_ical", icalOpts...)
	icalHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &pagination.ListOptions{
			Extra: map[string]string{},
		}
		if id := req.GetInt("space_id", 0); id != 0 {
			opts.Extra["space"] = strconv.Itoa(id)
		}
		if id := req.GetInt("user_id", 0); id != 0 {
			opts.Extra["user"] = strconv.Itoa(id)
		}
		if s := req.GetString("date_from", ""); s != "" {
			opts.Extra["date_from"] = s
		}
		if s := req.GetString("date_to", ""); s != "" {
			opts.Extra["date_to"] = s
		}
		ics, err := d.Tandoor.MealPlans().ICAL(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return mcpgo.NewToolResultText(ics), nil
	}

	// --- meal types ---
	typeListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List meal types (Breakfast, Lunch, Dinner, ...). Their IDs are instance-specific — enumerate before referencing them in writes.")}
	typeListOpts = append(typeListOpts, baselineListParams(d)...)
	typeList := mcpgo.NewTool("meal_type_list", typeListOpts...)
	typeListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		base := baseListOptions(req, d)
		opts := &base
		first, err := d.Tandoor.MealTypes().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[mealplan.MealType], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.MealTypes().List(ctx, &o2)
		})
	}

	typeGet := mcpgo.NewTool("meal_type_get",
		mcpgo.WithDescription("Get a single meal type by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Meal type ID")),
		jqParam(),
	)
	typeGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.MealTypes().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- meal plan writes ---
	planCreate := mcpgo.NewTool("meal_plan_create",
		mcpgo.WithDescription("Create a meal plan entry. Returns the created entry. Meal type IDs are instance-specific — use meal_type_list first."),
		mcpgo.WithString("from_date", mcpgo.Required(), mcpgo.Description("Start date (RFC3339 or YYYY-MM-DD)")),
		mcpgo.WithInteger("meal_type_id", mcpgo.Required(), mcpgo.Description("Meal type ID (breakfast/lunch/dinner)")),
		mcpgo.WithInteger("recipe_id", mcpgo.Description("Recipe ID (optional for free-text entries)")),
		mcpgo.WithString("title", mcpgo.Description("Entry title")),
		mcpgo.WithString("note", mcpgo.Description("Note")),
		mcpgo.WithNumber("servings", mcpgo.Description("Servings (default 1)")),
		mcpgo.WithString("to_date", mcpgo.Description("End date (RFC3339 or YYYY-MM-DD)")),
		mcpgo.WithBoolean("add_shopping", mcpgo.Description("Add this meal plan to the shopping list")),
		dryRunParam(),
	)
	planCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		from, err := parseDateOrDateTime(req.GetString("from_date", ""))
		if err != nil {
			return errResult(err), nil
		}
		mealTypeID, err := req.RequireInt("meal_type_id")
		if err != nil {
			return errResult(err), nil
		}
		mp := &mealplan.MealPlan{FromDate: &from, MealTypeID: mealTypeID, Servings: req.GetFloat("servings", 1)}
		if v, err := req.RequireInt("recipe_id"); err == nil && v != 0 {
			mp.Recipe = &recipe.Overview{ID: v}
		}
		if v, err := req.RequireString("title"); err == nil {
			mp.Title = v
		}
		if v, err := req.RequireString("note"); err == nil {
			mp.Note = v
		}
		if v, err := req.RequireString("to_date"); err == nil {
			t, err := parseDateOrDateTime(v)
			if err != nil {
				return errResult(err), nil
			}
			mp.ToDate = &t
		}
		if v, err := req.RequireBool("add_shopping"); err == nil {
			mp.AddShopping = v
		}
		return runWrite(ctx, req, "POST", "api/meal-plan/", mp, func() (any, error) {
			return d.Tandoor.MealPlans().Create(ctx, mp)
		})
	}

	planUpdate := mcpgo.NewTool("meal_plan_update",
		mcpgo.WithDescription("Update a meal plan entry (full replacement). Returns the updated entry."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Meal plan ID")),
		mcpgo.WithString("from_date", mcpgo.Required(), mcpgo.Description("Start date (RFC3339 or YYYY-MM-DD)")),
		mcpgo.WithInteger("meal_type_id", mcpgo.Required(), mcpgo.Description("Meal type ID")),
		mcpgo.WithInteger("recipe_id", mcpgo.Description("Recipe ID (0 to clear)")),
		mcpgo.WithString("title", mcpgo.Description("Entry title")),
		mcpgo.WithString("note", mcpgo.Description("Note")),
		mcpgo.WithNumber("servings", mcpgo.Description("Servings (default 1)")),
		mcpgo.WithString("to_date", mcpgo.Description("End date (RFC3339 or YYYY-MM-DD)")),
		mcpgo.WithBoolean("add_shopping", mcpgo.Description("Add this meal plan to the shopping list")),
		dryRunParam(),
	)
	planUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		from, err := parseDateOrDateTime(req.GetString("from_date", ""))
		if err != nil {
			return errResult(err), nil
		}
		mealTypeID, err := req.RequireInt("meal_type_id")
		if err != nil {
			return errResult(err), nil
		}
		mp := &mealplan.MealPlan{ID: id, FromDate: &from, MealTypeID: mealTypeID, Servings: req.GetFloat("servings", 1)}
		if v, err := req.RequireInt("recipe_id"); err == nil && v != 0 {
			mp.Recipe = &recipe.Overview{ID: v}
		}
		if v, err := req.RequireString("title"); err == nil {
			mp.Title = v
		}
		if v, err := req.RequireString("note"); err == nil {
			mp.Note = v
		}
		if v, err := req.RequireString("to_date"); err == nil {
			t, err := parseDateOrDateTime(v)
			if err != nil {
				return errResult(err), nil
			}
			mp.ToDate = &t
		}
		if v, err := req.RequireBool("add_shopping"); err == nil {
			mp.AddShopping = v
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/meal-plan/%d/", id), mp, func() (any, error) {
			return d.Tandoor.MealPlans().Update(ctx, mp)
		})
	}

	planDelete := mcpgo.NewTool("meal_plan_delete",
		mcpgo.WithDescription("Delete a meal plan entry."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Meal plan ID")),
		dryRunParam(),
	)
	planDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/meal-plan/%d/", id), func() error {
			return d.Tandoor.MealPlans().Delete(ctx, id)
		})
	}

	// --- meal type writes ---
	typeCreate := mcpgo.NewTool("meal_type_create",
		mcpgo.WithDescription("Create a meal type (e.g. breakfast). Returns the created meal type."),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Meal type name")),
		mcpgo.WithInteger("order", mcpgo.Description("Display order")),
		mcpgo.WithString("time", mcpgo.Description("Default time (HH:MM)")),
		mcpgo.WithString("color", mcpgo.Description("Display color (e.g. #ff0000)")),
		dryRunParam(),
	)
	typeCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		mt := &mealplan.MealType{Name: name}
		if v, err := req.RequireInt("order"); err == nil {
			mt.Order = v
		}
		if v, err := req.RequireString("time"); err == nil {
			mt.Time = &v
		}
		if v, err := req.RequireString("color"); err == nil {
			mt.Color = &v
		}
		return runWrite(ctx, req, "POST", "api/meal-type/", mt, func() (any, error) {
			return d.Tandoor.MealTypes().Create(ctx, mt)
		})
	}

	typeUpdate := mcpgo.NewTool("meal_type_update",
		mcpgo.WithDescription("Update a meal type (full replacement). Returns the updated meal type."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Meal type ID")),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Meal type name")),
		mcpgo.WithInteger("order", mcpgo.Description("Display order")),
		mcpgo.WithString("time", mcpgo.Description("Default time (HH:MM)")),
		mcpgo.WithString("color", mcpgo.Description("Display color")),
		dryRunParam(),
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
		mt := &mealplan.MealType{ID: id, Name: name}
		if v, err := req.RequireInt("order"); err == nil {
			mt.Order = v
		}
		if v, err := req.RequireString("time"); err == nil {
			mt.Time = &v
		}
		if v, err := req.RequireString("color"); err == nil {
			mt.Color = &v
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/meal-type/%d/", id), mt, func() (any, error) {
			return d.Tandoor.MealTypes().Update(ctx, mt)
		})
	}

	typePatch := mcpgo.NewTool("meal_type_patch",
		mcpgo.WithDescription("Partially update a meal type. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Meal type ID")),
		mcpgo.WithString("name", mcpgo.Description("Meal type name")),
		mcpgo.WithInteger("order", mcpgo.Description("Display order")),
		mcpgo.WithString("time", mcpgo.Description("Default time (HH:MM)")),
		mcpgo.WithString("color", mcpgo.Description("Display color")),
		dryRunParam(),
	)
	typePatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		mt := &mealplan.MealType{ID: id}
		if v, err := req.RequireString("name"); err == nil {
			mt.Name = v
		}
		if v, err := req.RequireInt("order"); err == nil {
			mt.Order = v
		}
		if v, err := req.RequireString("time"); err == nil {
			mt.Time = &v
		}
		if v, err := req.RequireString("color"); err == nil {
			mt.Color = &v
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/meal-type/%d/", id), mt, func() (any, error) {
			return d.Tandoor.MealTypes().Patch(ctx, mt)
		})
	}

	typeDelete := mcpgo.NewTool("meal_type_delete",
		mcpgo.WithDescription("Delete a meal type. Fails if meal plans still use it."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Meal type ID")),
		dryRunParam(),
	)
	typeDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/meal-type/%d/", id), func() error {
			return d.Tandoor.MealTypes().Delete(ctx, id)
		})
	}

	autoPlan := mcpgo.NewTool("meal_plan_auto_plan",
		mcpgo.WithDescription("Auto-generate meal plans for a date range by picking recipes matching keywords. Use dry_run to preview the request body."),
		mcpgo.WithString("start_date", mcpgo.Required(), mcpgo.Description("Start date (RFC3339 or YYYY-MM-DD)")),
		mcpgo.WithString("end_date", mcpgo.Required(), mcpgo.Description("End date (RFC3339 or YYYY-MM-DD)")),
		mcpgo.WithInteger("meal_type_id", mcpgo.Required(), mcpgo.Description("Meal type ID to assign")),
		mcpgo.WithArray("keyword_ids", mcpgo.WithIntegerItems(), mcpgo.Description("Keyword IDs to match")),
		mcpgo.WithString("keyword_mode", mcpgo.Description("How multiple keywords combine: 'or' or 'and' (default or)")),
		mcpgo.WithNumber("servings", mcpgo.Description("Planned servings (default 1)")),
		mcpgo.WithArray("shared_user_ids", mcpgo.WithIntegerItems(), mcpgo.Description("User IDs to share the plans with")),
		mcpgo.WithBoolean("add_shopping", mcpgo.Description("Add planned recipes to shopping lists")),
		dryRunParam(),
	)
	autoPlanHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		start, err := parseDateOrDateTime(req.GetString("start_date", ""))
		if err != nil {
			return errResult(fmt.Errorf("invalid start_date: %w", err)), nil
		}
		end, err := parseDateOrDateTime(req.GetString("end_date", ""))
		if err != nil {
			return errResult(fmt.Errorf("invalid end_date: %w", err)), nil
		}
		mealTypeID, err := req.RequireInt("meal_type_id")
		if err != nil {
			return errResult(fmt.Errorf("meal_type_id is required")), nil
		}
		var shared []mealplan.SharedUser
		for _, uid := range req.GetIntSlice("shared_user_ids", nil) {
			shared = append(shared, mealplan.SharedUser{ID: uid})
		}
		body := &mealplan.AutoMealPlanRequest{
			StartDate:   start,
			EndDate:     end,
			MealTypeID:  mealTypeID,
			Keywords:    req.GetIntSlice("keyword_ids", nil),
			KeywordMode: req.GetString("keyword_mode", ""),
			Servings:    req.GetFloat("servings", 1),
			Shared:      shared,
			AddShopping: req.GetBool("add_shopping", false),
		}
		return runWrite(ctx, req, "POST", "api/auto-plan/", body, func() (any, error) {
			return d.Tandoor.AutoPlan().Plan(ctx, body)
		})
	}

	return []toolDef{
		{tool: planList, handler: planListHandler},
		{tool: planGet, handler: planGetHandler},
		{tool: ical, handler: icalHandler},
		{tool: typeList, handler: typeListHandler},
		{tool: typeGet, handler: typeGetHandler},
		{tool: planCreate, handler: planCreateHandler, write: true},
		{tool: planUpdate, handler: planUpdateHandler, write: true},
		{tool: planDelete, handler: planDeleteHandler, write: true},
		{tool: typeCreate, handler: typeCreateHandler, write: true},
		{tool: typeUpdate, handler: typeUpdateHandler, write: true},
		{tool: typePatch, handler: typePatchHandler, write: true},
		{tool: typeDelete, handler: typeDeleteHandler, write: true},
		{tool: autoPlan, handler: autoPlanHandler, write: true},
	}
}
