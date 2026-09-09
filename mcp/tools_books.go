package mcp

import (
	"context"
	"fmt"
	"strconv"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/recipe"
	"github.com/swedishborgie/go-tandoor/recipebook"
)

func registerBookTools(d *deps) []toolDef {
	// --- books ---
	bookListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List recipe books. Use all=true for the full set; jq projects fields to keep output small.")}
	bookListOpts = append(bookListOpts, baselineListParams(d)...)
	bookListOpts = append(bookListOpts,
		mcpgo.WithInteger("space_id", mcpgo.Description("Filter by space ID")),
	)
	bookList := mcpgo.NewTool("book_list", bookListOpts...)
	bookListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &recipebook.ListOptions{
			ListOptions: baseListOptions(req, d),
			SpaceID:     req.GetInt("space_id", 0),
		}
		first, err := d.Tandoor.RecipeBooks().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[recipebook.RecipeBook], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.RecipeBooks().List(ctx, &o2)
		})
	}

	bookGet := mcpgo.NewTool("book_get",
		mcpgo.WithDescription("Get a single recipe book by ID (includes its recipes)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe book ID")),
		jqParam(),
	)
	bookGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.RecipeBooks().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- book entries ---
	entryListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List recipe book entries (book↔recipe links). Filter by book. Use all=true for the full set; jq projects fields to keep output small.")}
	entryListOpts = append(entryListOpts, baselineListParams(d)...)
	entryListOpts = append(entryListOpts,
		mcpgo.WithInteger("book_id", mcpgo.Description("Filter by recipe book ID")),
	)
	entryList := mcpgo.NewTool("book_entry_list", entryListOpts...)
	entryListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		base := baseListOptions(req, d)
		if id := req.GetInt("book_id", 0); id != 0 {
			if base.Extra == nil {
				base.Extra = map[string]string{}
			}
			base.Extra["book"] = strconv.Itoa(id)
		}
		first, err := d.Tandoor.RecipeBookEntries().List(ctx, &base)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[recipebook.Entry], error) {
			o2 := base
			o2.Page = page
			return d.Tandoor.RecipeBookEntries().List(ctx, &o2)
		})
	}

	// --- writes ---
	bookCreate := mcpgo.NewTool("book_create",
		mcpgo.WithDescription("Create a recipe book. Returns the created book."),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Book name")),
		mcpgo.WithString("description", mcpgo.Description("Book description")),
		mcpgo.WithInteger("order", mcpgo.Description("Display order")),
		mcpgo.WithArray("shared_user_ids", mcpgo.WithIntegerItems(), mcpgo.Description("User IDs to share this book with (default: private)")),
		dryRunParam(),
	)
	bookCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		b := &recipebook.RecipeBook{Name: name, Description: req.GetString("description", ""), Shared: []recipe.User{}}
		if v, err := req.RequireInt("order"); err == nil {
			b.Order = v
		}
		if ids, err := req.RequireIntSlice("shared_user_ids"); err == nil {
			b.Shared = usersFromIDs(ids)
		}
		return runWrite(ctx, req, "POST", "api/recipe-book/", b, func() (any, error) {
			return d.Tandoor.RecipeBooks().Create(ctx, b)
		})
	}

	bookUpdate := mcpgo.NewTool("book_update",
		mcpgo.WithDescription("Update a recipe book (full replacement). Returns the updated book."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Book ID")),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Book name")),
		mcpgo.WithString("description", mcpgo.Description("Book description")),
		mcpgo.WithInteger("order", mcpgo.Description("Display order")),
		mcpgo.WithArray("shared_user_ids", mcpgo.WithIntegerItems(), mcpgo.Description("User IDs this book is shared with (empty to make private)")),
		dryRunParam(),
	)
	bookUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		b := &recipebook.RecipeBook{ID: id, Name: name, Description: req.GetString("description", ""), Shared: []recipe.User{}}
		if v, err := req.RequireInt("order"); err == nil {
			b.Order = v
		}
		if ids, err := req.RequireIntSlice("shared_user_ids"); err == nil {
			b.Shared = usersFromIDs(ids)
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/recipe-book/%d/", id), b, func() (any, error) {
			return d.Tandoor.RecipeBooks().Update(ctx, b)
		})
	}

	bookDelete := mcpgo.NewTool("book_delete",
		mcpgo.WithDescription("Delete a recipe book (its recipe links are removed)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Book ID")),
		dryRunParam(),
	)
	bookDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/recipe-book/%d/", id), func() error {
			return d.Tandoor.RecipeBooks().Delete(ctx, id)
		})
	}

	entryCreate := mcpgo.NewTool("book_entry_create",
		mcpgo.WithDescription("Add a recipe to a recipe book. Returns the created entry."),
		mcpgo.WithInteger("book_id", mcpgo.Required(), mcpgo.Description("Book ID")),
		mcpgo.WithInteger("recipe_id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
		dryRunParam(),
	)
	entryCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		bookID, err := req.RequireInt("book_id")
		if err != nil {
			return errResult(err), nil
		}
		recipeID, err := req.RequireInt("recipe_id")
		if err != nil {
			return errResult(err), nil
		}
		e := &recipebook.Entry{Book: bookID, Recipe: recipeID}
		return runWrite(ctx, req, "POST", "api/recipe-book-entry/", e, func() (any, error) {
			return d.Tandoor.RecipeBookEntries().Create(ctx, e)
		})
	}

	entryDelete := mcpgo.NewTool("book_entry_delete",
		mcpgo.WithDescription("Remove a recipe from a recipe book."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Book entry ID")),
		dryRunParam(),
	)
	entryDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/recipe-book-entry/%d/", id), func() error {
			return d.Tandoor.RecipeBookEntries().Delete(ctx, id)
		})
	}

	return []toolDef{
		{tool: bookList, handler: bookListHandler},
		{tool: bookGet, handler: bookGetHandler},
		{tool: entryList, handler: entryListHandler},
		{tool: bookCreate, handler: bookCreateHandler, write: true},
		{tool: bookUpdate, handler: bookUpdateHandler, write: true},
		{tool: bookDelete, handler: bookDeleteHandler, write: true},
		{tool: entryCreate, handler: entryCreateHandler, write: true},
		{tool: entryDelete, handler: entryDeleteHandler, write: true},
	}
}
