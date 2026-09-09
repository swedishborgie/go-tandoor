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
		mcpgo.WithDescription("Create a recipe from a raw Tandoor recipe payload. Returns the created recipe."),
		dataParam,
		dryRunParam(),
	)
	recCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		data, err := recipeData(req)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "POST", "api/recipe/", data, func() (any, error) {
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
		dryRunParam(),
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
		path := fmt.Sprintf("api/recipe/%d/", id)
		return runWrite(ctx, req, "PUT", path, data, func() (any, error) {
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
		dryRunParam(),
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
		path := fmt.Sprintf("api/recipe/%d/", id)
		return runWrite(ctx, req, "PATCH", path, data, func() (any, error) {
			var out any
			if err := d.Tandoor.DoJSON(ctx, "PATCH", path, data, &out); err != nil {
				return nil, err
			}
			return out, nil
		})
	}

	recDelete := mcpgo.NewTool("recipe_delete",
		mcpgo.WithDescription("Delete a recipe."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		dryRunParam(),
	)
	recDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/recipe/%d/", id), func() error {
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
		dryRunParam(),
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
		return runWrite(ctx, req, "PUT", "api/recipe/batch_update/", update, func() (any, error) {
			return d.Tandoor.Recipes().BatchUpdate(ctx, update)
		})
	}

	recAddToShopping := mcpgo.NewTool("recipe_add_to_shopping",
		mcpgo.WithDescription("Add a recipe's ingredients to a shopping list. With list_recipe, edits that existing entry instead; servings 0 with list_recipe deletes it."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		mcpgo.WithInteger("servings", mcpgo.Description("Servings (default 1; 0 with list_recipe deletes the entry)")),
		mcpgo.WithInteger("list_recipe", mcpgo.Description("Existing shopping list recipe ID to edit")),
		mcpgo.WithArray("ingredients", mcpgo.WithIntegerItems(), mcpgo.Description("Recipe ingredient IDs to include (all if omitted)")),
		dryRunParam(),
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
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/recipe/%d/shopping/", id), update, func() (any, error) {
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
		dryRunParam(),
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
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/recipe/%d/image/", id), img, func() (any, error) {
			return d.Tandoor.Recipes().UploadImage(ctx, id, img)
		})
	}

	aiProps := mcpgo.NewTool("recipe_ai_properties",
		mcpgo.WithDescription("Trigger server-side AI to generate keywords, servings, and times for a recipe. Requires an AI provider configured on the instance."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		dryRunParam(),
	)
	aiPropsHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "POST", fmt.Sprintf("api/recipe/%d/aiproperties/", id), nil, func() (any, error) {
			return d.Tandoor.Recipes().AiProperties(ctx, id)
		})
	}

	delExternal := mcpgo.NewTool("recipe_delete_external",
		mcpgo.WithDescription("Remove the external file reference from a recipe (keeps the recipe)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		dryRunParam(),
	)
	delExternalHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/recipe/%d/delete_external/", id), nil, func() (any, error) {
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
