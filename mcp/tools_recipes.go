package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/recipe"
)

func registerRecipeTools(d *deps) []toolDef {
	listOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List recipes. Supports text query, keyword/space/book/user filters, ordering, and pagination. Use all=true for the full set; jq projects fields to keep output small.")}
	listOpts = append(listOpts, baselineListParams()...)
	listOpts = append(listOpts,
		mcpgo.WithArray("keyword_ids", mcpgo.WithIntegerItems(), mcpgo.Description("Filter by keyword IDs")),
		mcpgo.WithInteger("space_id", mcpgo.Description("Filter by space ID")),
		mcpgo.WithInteger("recipe_book_id", mcpgo.Description("Filter by recipe book ID")),
		mcpgo.WithInteger("user_id", mcpgo.Description("Filter by user ID")),
		mcpgo.WithBoolean("is_favorite", mcpgo.Description("Filter by favorite status (omit for all)")),
	)
	list := mcpgo.NewTool("recipe_list", listOpts...)
	listHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &recipe.ListOptions{
			ListOptions:  baseListOptions(req, d),
			KeywordIDs:   req.GetIntSlice("keyword_ids", nil),
			SpaceID:      req.GetInt("space_id", 0),
			RecipeBookID: req.GetInt("recipe_book_id", 0),
			UserID:       req.GetInt("user_id", 0),
		}
		if fav, ok := boolArg(req, "is_favorite"); ok {
			opts.IsFavorite = &fav
		}
		first, err := d.Tandoor.Recipes().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[recipe.Recipe], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Recipes().List(ctx, &o2)
		})
	}

	get := mcpgo.NewTool("recipe_get",
		mcpgo.WithDescription("Get a single recipe by ID, including ingredients, steps, and keywords."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		jqParam(),
	)
	getHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		recipe, err := d.Tandoor.Recipes().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, recipe)
	}

	// Lightweight flat list (id, name, image) — cheap for big catalogs.
	// The flat endpoint returns the whole set as one array (no pagination).
	overviewOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List a lightweight overview of recipes (id, name, image only — no ingredients/steps). Supports text query; returns the full set as a single array. Use jq to project fields.")}
	overviewOpts = append(overviewOpts,
		mcpgo.WithString("query", mcpgo.Description("Search query (name match)")),
		mcpgo.WithString("jq", mcpgo.Description("jq filter to project/transform the result")),
	)
	overview := mcpgo.NewTool("recipe_overview", overviewOpts...)
	overviewHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &recipe.ListOptions{ListOptions: baseListOptions(req, d)}
		res, err := d.Tandoor.Recipes().Flat(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		if jq := req.GetString("jq", ""); jq != "" {
			s, err := applyJQ(ctx, jq, res)
			if err != nil {
				return errResult(err), nil
			}
			return mcpgo.NewToolResultText(s), nil
		}
		return jsonResult(res), nil
	}

	// Recipes related to one recipe (shared keywords/foods).
	relatedOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List recipes related to the given recipe. Use all=true for the full set; jq projects fields to keep output small.")}
	relatedOpts = append(relatedOpts,
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		mcpgo.WithInteger("page", mcpgo.Description("Page number (1-indexed)")),
		mcpgo.WithInteger("page_size", mcpgo.Description(fmt.Sprintf("Results per page (default %d)", d.Cfg.DefaultPageSize))),
		mcpgo.WithBoolean("all", mcpgo.Description("Collect all pages into a single array")),
		mcpgo.WithString("jq", mcpgo.Description("jq filter to project/transform the result")),
	)
	related := mcpgo.NewTool("recipe_related", relatedOpts...)
	relatedHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		opts := &recipe.ListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.Recipes().Related(ctx, id, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[recipe.Recipe], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Recipes().Related(ctx, id, &o2)
		})
	}

	// --- writes ---
	// Recipe writes pass the `data` object through to Tandoor verbatim (no
	// struct round-trip), so any field of the recipe serializer can be used.
	dataParam := mcpgo.WithAny("data", mcpgo.Required(), mcpgo.Description("Recipe data as a JSON object (Tandoor recipe serializer fields: name, description, servings, servings_text, working_time, waiting_time, steps, keywords, image, source_url, private, archived, tags, nutrition, etc.)"))

	recCreate := mcpgo.NewTool("recipe_create",
		mcpgo.WithDescription("Create a recipe from a raw Tandoor recipe payload. A top-level ingredients array is merged into the first step (Tandoor ignores it there); each step is given an ingredients array when missing. Returns the created recipe."),
		dataParam,
	)
	recCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		data, err := recipeData(req)
		if err != nil {
			return errResult(err), nil
		}
		if err := validateRecipeData(data, true); err != nil {
			return errResult(err), nil
		}
		normalizeCreateData(data)
		return runWrite(func() (any, error) {
			var out any
			if err := d.Tandoor.DoJSON(ctx, "POST", "api/recipe/", data, &out); err != nil {
				return nil, err
			}
			return out, nil
		})
	}

	recUpdate := mcpgo.NewTool("recipe_update",
		mcpgo.WithDescription("Replace a recipe with the given data (full replacement; use recipe_patch for partial changes). Returns the updated recipe."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		dataParam,
	)
	recUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		data, err := recipeData(req)
		if err != nil {
			return errResult(err), nil
		}
		if err := validateRecipeData(data, false); err != nil {
			return errResult(err), nil
		}
		path := fmt.Sprintf("api/recipe/%d/", id)
		return runWrite(func() (any, error) {
			var out any
			if err := d.Tandoor.DoJSON(ctx, "PUT", path, data, &out); err != nil {
				return nil, err
			}
			return out, nil
		})
	}

	recPatch := mcpgo.NewTool("recipe_patch",
		mcpgo.WithDescription("Partially update a recipe. Only fields present in data change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		dataParam,
	)
	recPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		data, err := recipeData(req)
		if err != nil {
			return errResult(err), nil
		}
		if err := validateRecipeData(data, false); err != nil {
			return errResult(err), nil
		}
		path := fmt.Sprintf("api/recipe/%d/", id)
		return runWrite(func() (any, error) {
			var out any
			if err := d.Tandoor.DoJSON(ctx, "PATCH", path, data, &out); err != nil {
				return nil, err
			}
			return out, nil
		})
	}

	// addIng is the idempotent append path for recipe ingredients: it fetches
	// the recipe, appends new entries to one step (skipping entries already
	// present), and PUTs the full step list back — no full-replacement
	// ceremony and no risk of dropping unsent ingredients.
	addIng := mcpgo.NewTool("recipe_add_ingredients",
		mcpgo.WithDescription("Append ingredients to a recipe step without touching anything else (idempotent: entries already in the step are skipped unless allow_duplicates). When the recipe has no steps, the first step is created (step_index 0). Returns the added/skipped counts and the updated recipe."),
		mcpgo.WithInteger("recipe_id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		mcpgo.WithInteger("step_index", mcpgo.Description("0-based index of the step to append to (default 0)")),
		mcpgo.WithArray("ingredients", mcpgo.Required(), mcpgo.Description("Ingredients to append: objects with food_id (required) and optional unit_id, amount, note, no_amount, order")),
		mcpgo.WithBoolean("allow_duplicates", mcpgo.Description("Append even when an identical entry already exists in the step (default false)")),
	)
	addIngHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("recipe_id")
		if err != nil {
			return errResult(err), nil
		}
		ingredients, err := parseIngredientObjects(req.GetArguments()["ingredients"], true)
		if err != nil {
			return errResult(err), nil
		}
		stepIndex := req.GetInt("step_index", 0)
		if stepIndex < 0 {
			return errResult(fmt.Errorf("step_index must be >= 0")), nil
		}
		allowDup := req.GetBool("allow_duplicates", false)

		var data map[string]any
		if err := d.Tandoor.DoJSON(ctx, "GET", fmt.Sprintf("api/recipe/%d/", id), nil, &data); err != nil {
			return errResult(err), nil
		}
		steps, _ := data["steps"].([]any)
		if len(steps) == 0 {
			if stepIndex != 0 {
				return errResult(fmt.Errorf("recipe %d has no steps; only step_index 0 (which creates the first step) is valid", id)), nil
			}
			steps = []any{map[string]any{"name": "Instructions", "ingredients": []any{}}}
		}
		if stepIndex >= len(steps) {
			return errResult(fmt.Errorf("step_index %d out of range (recipe %d has %d steps)", stepIndex, id, len(steps))), nil
		}
		step, ok := steps[stepIndex].(map[string]any)
		if !ok {
			return errResult(fmt.Errorf("step %d is not an object", stepIndex)), nil
		}
		existing, _ := step["ingredients"].([]any)
		added, skipped := 0, 0
		for _, ing := range ingredients {
			key := ingredientKey(ing)
			if !allowDup {
				dup := false
				for _, e := range existing {
					if em, ok := e.(map[string]any); ok && ingredientKey(em) == key {
						dup = true
						break
					}
				}
				if dup {
					skipped++
					continue
				}
			}
			existing = append(existing, ing)
			added++
		}
		step["ingredients"] = existing
		data["steps"] = steps
		path := fmt.Sprintf("api/recipe/%d/", id)
		return runWrite(func() (any, error) {
			var out any
			if err := d.Tandoor.DoJSON(ctx, "PUT", path, data, &out); err != nil {
				return nil, err
			}
			return map[string]any{"added": added, "skipped": skipped, "recipe": out}, nil
		})
	}

	addStep := mcpgo.NewTool("recipe_add_step",
		mcpgo.WithDescription("Append a step to a recipe, optionally with ingredients; other steps and fields are untouched. When the recipe has no steps, this creates the first one. Returns the updated recipe."),
		mcpgo.WithInteger("recipe_id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Step name")),
		mcpgo.WithArray("ingredients", mcpgo.Description("Ingredients for the step: objects with food_id and optional unit_id, amount, note, no_amount, order")),
	)
	addStepHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("recipe_id")
		if err != nil {
			return errResult(err), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		ingredients, err := parseIngredientObjects(req.GetArguments()["ingredients"], false)
		if err != nil {
			return errResult(err), nil
		}
		var data map[string]any
		if err := d.Tandoor.DoJSON(ctx, "GET", fmt.Sprintf("api/recipe/%d/", id), nil, &data); err != nil {
			return errResult(err), nil
		}
		steps, _ := data["steps"].([]any)
		steps = append(steps, map[string]any{"name": name, "ingredients": ingredients})
		data["steps"] = steps
		path := fmt.Sprintf("api/recipe/%d/", id)
		return runWrite(func() (any, error) {
			var out any
			if err := d.Tandoor.DoJSON(ctx, "PUT", path, data, &out); err != nil {
				return nil, err
			}
			return out, nil
		})
	}

	recDelete := mcpgo.NewTool("recipe_delete",
		mcpgo.WithDescription("Delete a recipe."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
	)
	recDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(fmt.Sprintf("api/recipe/%d/", id), func() error {
			return d.Tandoor.Recipes().Delete(ctx, id)
		})
	}

	recBatchUpdate := mcpgo.NewTool("recipe_batch_update",
		mcpgo.WithDescription("Update multiple recipes at once: set keywords and/or working/waiting time. Returns the updated recipes."),
		mcpgo.WithArray("recipes", mcpgo.WithIntegerItems(), mcpgo.Required(), mcpgo.Description("Recipe IDs")),
		mcpgo.WithArray("keywords_add", mcpgo.WithIntegerItems(), mcpgo.Description("Keyword IDs to add")),
		mcpgo.WithArray("keywords_remove", mcpgo.WithIntegerItems(), mcpgo.Description("Keyword IDs to remove")),
		mcpgo.WithArray("keywords_set", mcpgo.WithIntegerItems(), mcpgo.Description("Keyword IDs to set (replaces all)")),
		mcpgo.WithBoolean("keywords_remove_all", mcpgo.Description("Remove all keywords")),
		mcpgo.WithInteger("working_time", mcpgo.Description("Working time in minutes")),
		mcpgo.WithInteger("waiting_time", mcpgo.Description("Waiting time in minutes")),
	)
	recBatchUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		recipes, err := req.RequireIntSlice("recipes")
		if err != nil {
			return errResult(err), nil
		}
		update := &recipe.BatchUpdate{Recipes: recipes}
		if v, err := req.RequireIntSlice("keywords_add"); err == nil {
			update.KeywordsAdd = v
		}
		if v, err := req.RequireIntSlice("keywords_remove"); err == nil {
			update.KeywordsRemove = v
		}
		if v, err := req.RequireIntSlice("keywords_set"); err == nil {
			update.KeywordsSet = v
		}
		if v, err := req.RequireBool("keywords_remove_all"); err == nil {
			update.KeywordsRemoveAll = v
		}
		if v, err := req.RequireInt("working_time"); err == nil {
			update.WorkingTime = &v
		}
		if v, err := req.RequireInt("waiting_time"); err == nil {
			update.WaitingTime = &v
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.Recipes().BatchUpdate(ctx, update)
		})
	}

	recAddToShopping := mcpgo.NewTool("recipe_add_to_shopping",
		mcpgo.WithDescription("Add a recipe's ingredients to a shopping list. With list_recipe, edits that existing entry instead; servings 0 with list_recipe deletes it."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		mcpgo.WithInteger("servings", mcpgo.Description("Servings (default 1; 0 with list_recipe deletes the entry)")),
		mcpgo.WithInteger("list_recipe", mcpgo.Description("Existing shopping list recipe ID to edit")),
		mcpgo.WithArray("ingredients", mcpgo.WithIntegerItems(), mcpgo.Description("Recipe ingredient IDs to include (all if omitted)")),
	)
	recAddToShoppingHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		update := &recipe.ShoppingUpdate{Ingredients: []int{}}
		if v, err := req.RequireInt("servings"); err == nil {
			update.Servings = &v
		}
		if v, err := req.RequireInt("list_recipe"); err == nil {
			update.ListRecipeID = &v
		}
		if v, err := req.RequireIntSlice("ingredients"); err == nil {
			update.Ingredients = v
		}
		return runWrite(func() (any, error) {
			if err := d.Tandoor.Recipes().AddToShopping(ctx, id, update); err != nil {
				return nil, err
			}
			return map[string]any{"added": true, "recipe": id}, nil
		})
	}

	uploadImage := mcpgo.NewTool("recipe_upload_image",
		mcpgo.WithDescription("Set or replace a recipe's image from a URL or base64 data."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		mcpgo.WithString("image_url", mcpgo.Description("Image URL to fetch (or image_base64; exactly one required)")),
		mcpgo.WithString("image_base64", mcpgo.Description("Base64-encoded image data (or image_url; exactly one required)")),
	)
	uploadImageHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		imageURL := req.GetString("image_url", "")
		imageB64 := req.GetString("image_base64", "")
		if (imageURL == "") == (imageB64 == "") {
			return errResult(fmt.Errorf("exactly one of image_url or image_base64 is required")), nil
		}
		img := &recipe.Image{Image: imageB64, ImageURL: imageURL}
		return runWrite(func() (any, error) {
			return d.Tandoor.Recipes().UploadImage(ctx, id, img)
		})
	}

	aiProps := mcpgo.NewTool("recipe_ai_properties",
		mcpgo.WithDescription("Trigger server-side AI to generate keywords, servings, and times for a recipe. Requires an AI provider configured on the instance."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
	)
	aiPropsHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.Recipes().AiProperties(ctx, id)
		})
	}

	delExternal := mcpgo.NewTool("recipe_delete_external",
		mcpgo.WithDescription("Remove the external file reference from a recipe (keeps the recipe)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
	)
	delExternalHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.Recipes().DeleteExternal(ctx, id)
		})
	}

	return []toolDef{
		{tool: list, handler: listHandler},
		{tool: get, handler: getHandler},
		{tool: overview, handler: overviewHandler},
		{tool: related, handler: relatedHandler},
		{tool: recCreate, handler: recCreateHandler, write: true},
		{tool: recUpdate, handler: recUpdateHandler, write: true},
		{tool: recPatch, handler: recPatchHandler, write: true},
		{tool: recDelete, handler: recDeleteHandler, write: true},
		{tool: recBatchUpdate, handler: recBatchUpdateHandler, write: true},
		{tool: addIng, handler: addIngHandler, write: true},
		{tool: addStep, handler: addStepHandler, write: true},
		{tool: recAddToShopping, handler: recAddToShoppingHandler, write: true},
		{tool: uploadImage, handler: uploadImageHandler, write: true},
		{tool: aiProps, handler: aiPropsHandler, write: true},
		{tool: delExternal, handler: delExternalHandler, write: true},
	}
}

// recipeData extracts the free-form `data` object from a tool call.
func recipeData(req mcpgo.CallToolRequest) (map[string]any, error) {
	data, ok := req.GetArguments()["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data must be a JSON object")
	}
	return data, nil
}

// validateRecipeData rejects payload keys Tandoor would silently drop, so a
// dropped field fails loudly here instead of producing a recipe that looks
// created but is missing data.
func validateRecipeData(data map[string]any, create bool) error {
	if _, ok := data["source"]; ok {
		return fmt.Errorf(`data: "source" is silently ignored by Tandoor — use "source_url" instead`)
	}
	if _, has := data["ingredients"]; has && !create {
		return fmt.Errorf(`data: top-level "ingredients" is ignored by Tandoor on update — nest them in steps[].ingredients, or use recipe_add_ingredients to append`)
	}
	return nil
}

// normalizeCreateData adapts a create payload to Tandoor's writable
// serializer: a top-level ingredients list is merged into the first step
// (Tandoor drops it there), and every step is given an ingredients array
// (Tandoor 400s on steps without one).
func normalizeCreateData(data map[string]any) {
	steps, _ := data["steps"].([]any)
	if top, ok := data["ingredients"].([]any); ok {
		delete(data, "ingredients")
		if len(steps) == 0 {
			steps = []any{map[string]any{"name": "Instructions"}}
		}
		if step, ok := steps[0].(map[string]any); ok {
			existing, _ := step["ingredients"].([]any)
			step["ingredients"] = append(existing, top...)
		}
	}
	for _, s := range steps {
		if step, ok := s.(map[string]any); ok {
			if _, has := step["ingredients"]; !has {
				step["ingredients"] = []any{}
			}
		}
	}
	if len(steps) > 0 {
		data["steps"] = steps
	}
}

// parseIngredientObjects converts a raw ingredients argument into validated
// ingredient objects (food_id required on each).
func parseIngredientObjects(v any, required bool) ([]map[string]any, error) {
	raw, _ := v.([]any)
	if required && len(raw) == 0 {
		return nil, fmt.Errorf("ingredients must be a non-empty array of objects")
	}
	out := make([]map[string]any, 0, len(raw))
	for i, e := range raw {
		m, ok := e.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("ingredients[%d] must be an object", i)
		}
		if _, has := m["food_id"]; !has {
			return nil, fmt.Errorf("ingredients[%d] requires food_id", i)
		}
		out = append(out, m)
	}
	return out, nil
}

// ingredientKey is the idempotency key for recipe_add_ingredients dedup:
// food + unit + amount + no_amount + note.
func ingredientKey(m map[string]any) string {
	food, _ := m["food_id"].(float64)
	unit, _ := m["unit_id"].(float64)
	amount, _ := m["amount"].(float64)
	noAmount, _ := m["no_amount"].(bool)
	note, _ := m["note"].(string)
	return fmt.Sprintf("%g|%g|%v|%v|%s", food, unit, amount, noAmount, note)
}
