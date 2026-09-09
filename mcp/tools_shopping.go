package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/shopping"
	"github.com/swedishborgie/go-tandoor/unit"
)

func registerShoppingTools(d *deps) []toolDef {
	// --- shopping lists ---
	listListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List shopping lists. Use all=true for the full set; jq projects fields to keep output small.")}
	listListOpts = append(listListOpts, baselineListParams()...)
	listList := mcpgo.NewTool("shopping_list_list", listListOpts...)
	listListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &shopping.ListListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.ShoppingLists().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[shopping.List], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.ShoppingLists().List(ctx, &o2)
		})
	}

	listGet := mcpgo.NewTool("shopping_list_get",
		mcpgo.WithDescription("Get a single shopping list by ID (includes its entries and recipes)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Shopping list ID")),
		jqParam(),
	)
	listGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.ShoppingLists().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- shopping entries ---
	entryListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List shopping entries (items to buy). Use all=true for the full set; jq projects fields to keep output small.")}
	entryListOpts = append(entryListOpts, baselineListParams()...)
	entryList := mcpgo.NewTool("shopping_entry_list", entryListOpts...)
	entryListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &shopping.ListEntryListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.ShoppingEntries().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[shopping.ListEntry], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.ShoppingEntries().List(ctx, &o2)
		})
	}

	entryGet := mcpgo.NewTool("shopping_entry_get",
		mcpgo.WithDescription("Get a single shopping entry by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Shopping entry ID")),
		jqParam(),
	)
	entryGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.ShoppingEntries().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- shopping list recipes ---
	recipeListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List shopping list recipes (recipes contributing ingredients to shopping lists). Use all=true for the full set; jq projects fields to keep output small.")}
	recipeListOpts = append(recipeListOpts, baselineListParams()...)
	recipeList := mcpgo.NewTool("shopping_recipe_list", recipeListOpts...)
	recipeListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &shopping.ListRecipeListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.ShoppingRecipes().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[shopping.ListRecipe], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.ShoppingRecipes().List(ctx, &o2)
		})
	}

	// --- list writes ---
	listCreate := mcpgo.NewTool("shopping_list_create",
		mcpgo.WithDescription("Create a shopping list. Returns the created list."),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("List name")),
		mcpgo.WithString("description", mcpgo.Description("List description")),
		mcpgo.WithString("color", mcpgo.Description("List color (e.g. #3498db)")),
		dryRunParam(),
	)
	listCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		sl := &shopping.List{Name: name, Description: req.GetString("description", ""), Color: req.GetString("color", "")}
		return runWrite(ctx, req, "POST", "api/shopping-list/", sl, func() (any, error) {
			return d.Tandoor.ShoppingLists().Create(ctx, sl)
		})
	}

	listUpdate := mcpgo.NewTool("shopping_list_update",
		mcpgo.WithDescription("Update a shopping list (full replacement). Returns the updated list."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("List ID")),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("List name")),
		mcpgo.WithString("description", mcpgo.Description("List description")),
		mcpgo.WithString("color", mcpgo.Description("List color")),
		dryRunParam(),
	)
	listUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		sl := &shopping.List{ID: id, Name: name, Description: req.GetString("description", ""), Color: req.GetString("color", "")}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/shopping-list/%d/", id), sl, func() (any, error) {
			return d.Tandoor.ShoppingLists().Update(ctx, sl)
		})
	}

	listDelete := mcpgo.NewTool("shopping_list_delete",
		mcpgo.WithDescription("Delete a shopping list (its entries are removed)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("List ID")),
		dryRunParam(),
	)
	listDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/shopping-list/%d/", id), func() error {
			return d.Tandoor.ShoppingLists().Delete(ctx, id)
		})
	}

	// --- entry writes ---
	entryWriteFields := []mcpgo.ToolOption{
		mcpgo.WithInteger("food_id", mcpgo.Description("Food ID (0 for headers)")),
		mcpgo.WithInteger("unit_id", mcpgo.Description("Unit ID (0 for none)")),
		mcpgo.WithNumber("amount", mcpgo.Description("Amount (default 0)")),
		mcpgo.WithString("note", mcpgo.Description("Note")),
		mcpgo.WithInteger("order", mcpgo.Description("Display order")),
		mcpgo.WithBoolean("checked", mcpgo.Description("Whether the item is purchased")),
		mcpgo.WithArray("list_ids", mcpgo.WithIntegerItems(), mcpgo.Description("Shopping list IDs to attach the entry to")),
	}
	buildListEntry := func(id int, req mcpgo.CallToolRequest, amountDefault bool) (*shopping.ListEntry, error) {
		e := &shopping.ListEntry{ID: id}
		if v, err := req.RequireInt("food_id"); err == nil && v != 0 {
			e.Food = &food.Shopping{ID: v}
		}
		if v, err := req.RequireInt("unit_id"); err == nil && v != 0 {
			e.Unit = &unit.Unit{ID: v}
		}
		if v, err := req.RequireFloat("amount"); err == nil || amountDefault {
			e.Amount = &v
		}
		if v, err := req.RequireString("note"); err == nil {
			e.Note = v
		}
		if v, err := req.RequireInt("order"); err == nil {
			e.Order = v
		}
		if v, err := req.RequireBool("checked"); err == nil {
			e.Checked = &v
		}
		if ids, err := req.RequireIntSlice("list_ids"); err == nil && len(ids) > 0 {
			e.Lists = shoppingListsFromIDs(ids)
		}
		return e, nil
	}

	einCreateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Create a shopping list entry. Returns the created entry."),
		dryRunParam(),
	}, entryWriteFields...)
	einCreate := mcpgo.NewTool("shopping_entry_create", einCreateOpts...)
	einCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		e, err := buildListEntry(0, req, true)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "POST", "api/shopping-list-entry/", e, func() (any, error) {
			return d.Tandoor.ShoppingEntries().Create(ctx, e)
		})
	}

	einUpdateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Update a shopping list entry (full replacement). Returns the updated entry."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Entry ID")),
		dryRunParam(),
	}, entryWriteFields...)
	einUpdate := mcpgo.NewTool("shopping_entry_update", einUpdateOpts...)
	einUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		e, err := buildListEntry(id, req, true)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/shopping-list-entry/%d/", id), e, func() (any, error) {
			return d.Tandoor.ShoppingEntries().Update(ctx, e)
		})
	}

	einPatchOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Partially update a shopping list entry. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Entry ID")),
		dryRunParam(),
	}, entryWriteFields...)
	einPatch := mcpgo.NewTool("shopping_entry_patch", einPatchOpts...)
	einPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		e, err := buildListEntry(id, req, false)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/shopping-list-entry/%d/", id), e, func() (any, error) {
			return d.Tandoor.ShoppingEntries().Patch(ctx, e)
		})
	}

	einDelete := mcpgo.NewTool("shopping_entry_delete",
		mcpgo.WithDescription("Delete a shopping list entry."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Entry ID")),
		dryRunParam(),
	)
	einDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/shopping-list-entry/%d/", id), func() error {
			return d.Tandoor.ShoppingEntries().Delete(ctx, id)
		})
	}

	einBulkUpdate := mcpgo.NewTool("shopping_entry_bulk_update",
		mcpgo.WithDescription("Bulk update shopping list entries: check/uncheck and move between lists in one call."),
		mcpgo.WithArray("ids", mcpgo.WithIntegerItems(), mcpgo.Required(), mcpgo.Description("Entry IDs")),
		mcpgo.WithBoolean("checked", mcpgo.Description("Set checked state")),
		mcpgo.WithArray("lists_add", mcpgo.WithIntegerItems(), mcpgo.Description("List IDs to add to")),
		mcpgo.WithArray("lists_remove", mcpgo.WithIntegerItems(), mcpgo.Description("List IDs to remove from")),
		mcpgo.WithArray("lists_set", mcpgo.WithIntegerItems(), mcpgo.Description("List IDs to set (replaces all)")),
		mcpgo.WithBoolean("lists_remove_all", mcpgo.Description("Remove entries from all lists")),
		dryRunParam(),
	)
	einBulkUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		ids, err := req.RequireIntSlice("ids")
		if err != nil {
			return errResult(err), nil
		}
		bulk := &shopping.ListEntryBulk{IDs: ids}
		if v, err := req.RequireBool("checked"); err == nil {
			bulk.Checked = &v
		}
		if v, err := req.RequireIntSlice("lists_add"); err == nil {
			bulk.ListsAdd = v
		}
		if v, err := req.RequireIntSlice("lists_remove"); err == nil {
			bulk.ListsRemove = v
		}
		if v, err := req.RequireIntSlice("lists_set"); err == nil {
			bulk.ListsSet = v
		}
		if v, err := req.RequireBool("lists_remove_all"); err == nil {
			bulk.ListsRemoveAll = v
		}
		return runWrite(ctx, req, "POST", "api/shopping-list-entry/bulk_update/", bulk, func() (any, error) {
			return d.Tandoor.ShoppingEntries().BulkUpdate(ctx, bulk)
		})
	}

	// --- composite: add recipe to a shopping list ---
	addRecipe := mcpgo.NewTool("shopping_list_add_recipe",
		mcpgo.WithDescription("Add a recipe's ingredients to a shopping list (creates a shopping list recipe and entries). Amounts scale with servings vs the recipe's servings."),
		mcpgo.WithInteger("list_id", mcpgo.Required(), mcpgo.Description("Shopping list ID")),
		mcpgo.WithInteger("recipe_id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		mcpgo.WithNumber("servings", mcpgo.Description("Target servings (defaults to the recipe's servings)")),
		dryRunParam(),
	)
	addRecipeHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		listID, err := req.RequireInt("list_id")
		if err != nil {
			return errResult(err), nil
		}
		recipeID, err := req.RequireInt("recipe_id")
		if err != nil {
			return errResult(err), nil
		}

		r, err := d.Tandoor.Recipes().Get(ctx, recipeID)
		if err != nil {
			return errResult(err), nil
		}
		servings := req.GetFloat("servings", 0)
		if servings <= 0 {
			servings = float64(max(r.Servings, 1))
		}
		factor := servings / float64(max(r.Servings, 1))

		var entries []shopping.ListEntryCreate
		for _, st := range r.Steps {
			for _, ing := range st.Ingredients {
				if ing.Food == nil {
					continue
				}
				e := shopping.ListEntryCreate{FoodID: intPtr(ing.Food.ID), Amount: ing.Amount * factor}
				if ing.Unit != nil && ing.Unit.ID != 0 {
					e.UnitID = intPtr(ing.Unit.ID)
				}
				entries = append(entries, e)
			}
		}
		if entries == nil {
			entries = []shopping.ListEntryCreate{}
		}

		steps := []map[string]any{
			{"method": "POST", "path": fmt.Sprintf("api/shopping-list-recipe/%d/", 0), "body": map[string]any{"recipe": recipeID, "servings": servings}},
			{"method": "POST", "path": "api/shopping-list-recipe/{id}/bulk_create_entries/", "body": map[string]any{"entries": entries, "shopping_lists_ids": []int{listID}}},
		}
		if req.GetBool("dry_run", false) {
			return jsonResult(map[string]any{"dry_run": true, "steps": steps}), nil
		}

		var slr shopping.ListRecipe
		if err := d.Tandoor.DoJSON(ctx, "POST", "api/shopping-list-recipe/", map[string]any{"recipe": recipeID, "servings": servings}, &slr); err != nil {
			return errResult(err), nil
		}
		result, err := d.Tandoor.ShoppingRecipes().BulkCreateEntries(ctx, slr.ID, &shopping.ListEntryBulkCreate{Entries: entries, ListIDs: []int{listID}})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result), nil
	}

	// --- composite: create entries from a recipe into shopping lists ---
	createEntries := mcpgo.NewTool("shopping_recipe_create_entries",
		mcpgo.WithDescription("Create shopping entries from a recipe's ingredients and attach them to shopping lists (composite: derives scaled entries, creates a shopping list recipe when needed). Pass shopping_list_recipe_id to reuse an existing link, or recipe_id to create a new one. Use dry_run to preview."),
		mcpgo.WithInteger("shopping_list_recipe_id", mcpgo.Description("Existing shopping list recipe ID to create entries for (or recipe_id; exactly one required)")),
		mcpgo.WithInteger("recipe_id", mcpgo.Description("Recipe ID to build a new shopping list recipe from (or shopping_list_recipe_id; exactly one required)")),
		mcpgo.WithNumber("servings", mcpgo.Description("Target servings, amounts scale vs the recipe's servings (default: recipe's servings)")),
		mcpgo.WithArray("shopping_list_ids", mcpgo.WithIntegerItems(), mcpgo.Description("Shopping lists to attach the entries to (default: each food's default lists)")),
		dryRunParam(),
	)
	createEntriesHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		slrID := req.GetInt("shopping_list_recipe_id", 0)
		recipeID := req.GetInt("recipe_id", 0)
		if (slrID == 0) == (recipeID == 0) {
			return errResult(fmt.Errorf("exactly one of shopping_list_recipe_id or recipe_id is required")), nil
		}
		if slrID != 0 {
			// Reuse an existing shopping list recipe: resolve its recipe.
			slr, err := d.Tandoor.ShoppingRecipes().Get(ctx, slrID)
			if err != nil {
				return errResult(err), nil
			}
			if slr.Recipe == nil || slr.Recipe.ID == 0 {
				return errResult(fmt.Errorf("shopping list recipe %d has no linked recipe", slrID)), nil
			}
			recipeID = slr.Recipe.ID
		}

		r, err := d.Tandoor.Recipes().Get(ctx, recipeID)
		if err != nil {
			return errResult(err), nil
		}
		servings := req.GetFloat("servings", 0)
		if servings <= 0 {
			servings = float64(max(r.Servings, 1))
		}
		factor := servings / float64(max(r.Servings, 1))

		var entries []shopping.ListEntryCreate
		for _, st := range r.Steps {
			for _, ing := range st.Ingredients {
				if ing.Food == nil {
					continue
				}
				e := shopping.ListEntryCreate{FoodID: intPtr(ing.Food.ID), Amount: ing.Amount * factor}
				if ing.Unit != nil && ing.Unit.ID != 0 {
					e.UnitID = intPtr(ing.Unit.ID)
				}
				if ing.ID != 0 {
					e.IngredientID = intPtr(ing.ID)
				}
				entries = append(entries, e)
			}
		}
		if entries == nil {
			entries = []shopping.ListEntryCreate{}
		}
		listIDs := req.GetIntSlice("shopping_list_ids", nil)

		body := shopping.ListEntryBulkCreate{Entries: entries, ListIDs: listIDs}
		newSlr := slrID == 0
		if newSlr {
			// dry_run: show both the link creation and the bulk entry create.
			if req.GetBool("dry_run", false) {
				return jsonResult(map[string]any{
					"dry_run": true,
					"steps": []map[string]any{
						{"method": "POST", "path": "api/shopping-list-recipe/", "body": map[string]any{"recipe": recipeID, "servings": servings}},
						{"method": "POST", "path": "api/shopping-list-recipe/{id}/bulk_create_entries/", "body": body},
					},
				}), nil
			}
			var slr shopping.ListRecipe
			if err := d.Tandoor.DoJSON(ctx, "POST", "api/shopping-list-recipe/", map[string]any{"recipe": recipeID, "servings": servings}, &slr); err != nil {
				return errResult(err), nil
			}
			slrID = slr.ID
		}
		if req.GetBool("dry_run", false) {
			return jsonResult(map[string]any{
				"dry_run": true,
				"steps": []map[string]any{
					{"method": "POST", "path": fmt.Sprintf("api/shopping-list-recipe/%d/bulk_create_entries/", slrID), "body": body},
				},
			}), nil
		}
		result, err := d.Tandoor.ShoppingRecipes().BulkCreateEntries(ctx, slrID, &body)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result), nil
	}

	return []toolDef{
		{tool: listList, handler: listListHandler},
		{tool: listGet, handler: listGetHandler},
		{tool: entryList, handler: entryListHandler},
		{tool: entryGet, handler: entryGetHandler},
		{tool: recipeList, handler: recipeListHandler},
		{tool: listCreate, handler: listCreateHandler, write: true},
		{tool: listUpdate, handler: listUpdateHandler, write: true},
		{tool: listDelete, handler: listDeleteHandler, write: true},
		{tool: einCreate, handler: einCreateHandler, write: true},
		{tool: einUpdate, handler: einUpdateHandler, write: true},
		{tool: einPatch, handler: einPatchHandler, write: true},
		{tool: einDelete, handler: einDeleteHandler, write: true},
		{tool: einBulkUpdate, handler: einBulkUpdateHandler, write: true},
		{tool: addRecipe, handler: addRecipeHandler, write: true},
		{tool: createEntries, handler: createEntriesHandler, write: true},
	}
}
