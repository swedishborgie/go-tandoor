package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/action"
	"github.com/swedishborgie/go-tandoor/importexport"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func registerImportTools(d *deps) []toolDef {
	// --- recipe imports (staged imports awaiting execution) ---
	importListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List recipe imports (staged imports waiting to be executed). Use all=true for the full set; jq projects fields to keep output small.")}
	importListOpts = append(importListOpts, baselineListParams()...)
	importList := mcpgo.NewTool("import_list", importListOpts...)
	importListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &importexport.RecipeImportListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.RecipeImports().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[importexport.RecipeImport], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.RecipeImports().List(ctx, &o2)
		})
	}

	importGet := mcpgo.NewTool("import_get",
		mcpgo.WithDescription("Get a single recipe import by ID (the staged payload)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Import ID")),
		jqParam(),
	)
	importGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.RecipeImports().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- logs ---
	importLogListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List import logs (history of executed imports). Use all=true for the full set; jq projects fields to keep output small.")}
	importLogListOpts = append(importLogListOpts, baselineListParams()...)
	importLogList := mcpgo.NewTool("import_log_list", importLogListOpts...)
	importLogListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &importexport.ImportLogListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.ImportLogs().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[importexport.ImportLog], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.ImportLogs().List(ctx, &o2)
		})
	}

	exportLogListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List export logs (history of recipe exports). Use all=true for the full set; jq projects fields to keep output small.")}
	exportLogListOpts = append(exportLogListOpts, baselineListParams()...)
	exportLogList := mcpgo.NewTool("export_log_list", exportLogListOpts...)
	exportLogListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &importexport.ExportLogListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.ExportLogs().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[importexport.ExportLog], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.ExportLogs().List(ctx, &o2)
		})
	}

	bookmarkletListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List bookmarklet imports (recipes captured via the browser bookmarklet). Use all=true for the full set; jq projects fields to keep output small.")}
	bookmarkletListOpts = append(bookmarkletListOpts, baselineListParams()...)
	bookmarkletList := mcpgo.NewTool("bookmarklet_import_list", bookmarkletListOpts...)
	bookmarkletListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &importexport.BookmarkletImportListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.BookmarkletImports().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[importexport.BookmarkletImport], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.BookmarkletImports().List(ctx, &o2)
		})
	}

	importImport := mcpgo.NewTool("import_import",
		mcpgo.WithDescription("Execute a single staged recipe import (creates the recipe)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe import ID (from import_list)")),
		dryRunParam(),
	)
	importImportHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "POST", fmt.Sprintf("api/recipe-import/%d/import_recipe/", id), nil, func() (any, error) {
			return d.Tandoor.RecipeImports().ImportRecipe(ctx, id)
		})
	}

	importAll := mcpgo.NewTool("import_import_all",
		mcpgo.WithDescription("Execute all pending staged recipe imports."),
		dryRunParam(),
	)
	importAllHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		return runWrite(ctx, req, "POST", "api/recipe-import/import_all/", nil, func() (any, error) {
			return d.Tandoor.RecipeImports().ImportAll(ctx)
		})
	}

	importDelete := mcpgo.NewTool("import_delete",
		mcpgo.WithDescription("Delete a staged recipe import without executing it."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Recipe import ID")),
		dryRunParam(),
	)
	importDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/recipe-import/%d/", id), func() error {
			return d.Tandoor.RecipeImports().Delete(ctx, id)
		})
	}

	// --- share links ---
	shareCreate := mcpgo.NewTool("share_link_create",
		mcpgo.WithDescription("Create a public share link for a recipe. Returns the share URL."),
		mcpgo.WithInteger("recipe_id", mcpgo.Required(), mcpgo.Description("Recipe ID")),
	)
	shareCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("recipe_id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.ShareLinks().Create(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// Note: Tandoor (2.6.13, current release branch) exposes share links
	// only via GET /api/share-link/<recipe_id>/ (create). There is no
	// DELETE endpoint, so no share_link_delete tool exists.

	// --- recipe from source (URL/payload scrape) ---
	fromSource := mcpgo.NewTool("recipe_from_source_create",
		mcpgo.WithDescription("Create a recipe from a source URL or raw HTML/JSON payload (server-side scrape). The response includes detected duplicates when the URL was imported before."),
		mcpgo.WithString("url", mcpgo.Description("Source URL to scrape (or data; exactly one required)")),
		mcpgo.WithString("data", mcpgo.Description("Raw source data (or url; exactly one required)")),
		mcpgo.WithInteger("bookmarklet_id", mcpgo.Description("Bookmarklet import ID to use instead (overrides url/data)")),
		dryRunParam(),
	)
	fromSourceHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		body := &action.RecipeFromSourceRequest{
			URL:  req.GetString("url", ""),
			Data: req.GetString("data", ""),
		}
		if v := req.GetInt("bookmarklet_id", 0); v != 0 {
			body.Bookmarklet = &v
		}
		if body.URL == "" && body.Data == "" && body.Bookmarklet == nil {
			return errResult(fmt.Errorf("one of url, data, or bookmarklet_id is required")), nil
		}
		return runWrite(ctx, req, "POST", "api/recipe-from-source/", body, func() (any, error) {
			return d.Tandoor.RecipeFromSource().Import(ctx, body)
		})
	}

	// --- open data (USDA) import ---
	openDataMeta := mcpgo.NewTool("open_data_metadata",
		mcpgo.WithDescription("List available USDA open data versions and datatypes (call before open_data_import)."),
		jqParam(),
	)
	openDataMetaHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		res, err := d.Tandoor.ImportOpenData().GetMetadata(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	openDataImport := mcpgo.NewTool("open_data_import",
		mcpgo.WithDescription("Import USDA open data (foods, units, conversions) for a version and datatypes. Check open_data_metadata for valid values first."),
		mcpgo.WithString("selected_version", mcpgo.Required(), mcpgo.Description("Data version (from open_data_metadata)")),
		mcpgo.WithArray("selected_datatypes", mcpgo.WithStringItems(), mcpgo.Required(), mcpgo.Description("Datatypes to import (from open_data_metadata)")),
		mcpgo.WithBoolean("update_existing", mcpgo.Description("Update already-imported objects (default false)")),
		mcpgo.WithBoolean("use_metric", mcpgo.Description("Import metric units (default false)")),
		dryRunParam(),
	)
	openDataImportHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		version, err := req.RequireString("selected_version")
		if err != nil {
			return errResult(err), nil
		}
		datatypes := req.GetStringSlice("selected_datatypes", nil)
		if len(datatypes) == 0 {
			return errResult(fmt.Errorf("selected_datatypes is required")), nil
		}
		body := &action.ImportOpenDataRequest{
			SelectedVersion:   version,
			SelectedDataTypes: datatypes,
			UpdateExisting:    req.GetBool("update_existing", false),
			UseMetric:         req.GetBool("use_metric", false),
		}
		return runWrite(ctx, req, "POST", "api/import-open-data/", body, func() (any, error) {
			return d.Tandoor.ImportOpenData().Import(ctx, body)
		})
	}

	return []toolDef{
		{tool: importList, handler: importListHandler},
		{tool: importGet, handler: importGetHandler},
		{tool: importImport, handler: importImportHandler, write: true},
		{tool: importAll, handler: importAllHandler, write: true},
		{tool: importDelete, handler: importDeleteHandler, write: true},
		{tool: importLogList, handler: importLogListHandler},
		{tool: exportLogList, handler: exportLogListHandler},
		{tool: bookmarkletList, handler: bookmarkletListHandler},
		{tool: shareCreate, handler: shareCreateHandler, write: true},
		{tool: fromSource, handler: fromSourceHandler, write: true},
		{tool: openDataMeta, handler: openDataMetaHandler},
		{tool: openDataImport, handler: openDataImportHandler, write: true},
	}
}
