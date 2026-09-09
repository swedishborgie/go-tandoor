package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/storage"
)

func registerStorageTools(d *deps) []toolDef {
	listOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List storages (legacy storage locations; prefer inventory locations for new data). Use all=true for the full set; jq projects fields to keep output small.")}
	listOpts = append(listOpts, baselineListParams()...)
	list := mcpgo.NewTool("storage_list", listOpts...)
	listHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &storage.ListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.Storages().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[storage.Storage], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Storages().List(ctx, &o2)
		})
	}

	get := mcpgo.NewTool("storage_get",
		mcpgo.WithDescription("Get a single storage by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Storage ID")),
		jqParam(),
	)
	getHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Storages().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- writes ---
	storageWriteFields := []mcpgo.ToolOption{
		mcpgo.WithString("username", mcpgo.Description("Storage username (if applicable)")),
		mcpgo.WithString("password", mcpgo.Description("Storage password (write-only, never returned)")),
		mcpgo.WithString("token", mcpgo.Description("API token (write-only, never returned)")),
		mcpgo.WithString("url", mcpgo.Description("Base URL (for webdav/rest storage)")),
		mcpgo.WithString("path", mcpgo.Description("Path on the storage (for webdav/rest storage)")),
	}

	stCreateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Create a storage backend (DB, NEXTCLOUD, or LOCAL). Returns the created storage."),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Storage name")),
		mcpgo.WithString("method", mcpgo.Required(), mcpgo.Description("Storage method: DB, NEXTCLOUD, or LOCAL")),
	}, storageWriteFields...)
	stCreate := mcpgo.NewTool("storage_create", stCreateOpts...)
	stCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		method, err := req.RequireString("method")
		if err != nil {
			return errResult(err), nil
		}
		st := &storage.Storage{Name: name, Method: method,
			Username: req.GetString("username", ""), Password: req.GetString("password", ""),
			Token: req.GetString("token", ""), URL: req.GetString("url", ""), Path: req.GetString("path", "")}
		return runWrite(func() (any, error) {
			return d.Tandoor.Storages().Create(ctx, st)
		})
	}

	stUpdateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Update a storage backend (full replacement). Returns the updated storage."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Storage ID")),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Storage name")),
		mcpgo.WithString("method", mcpgo.Required(), mcpgo.Description("Storage method: DB, NEXTCLOUD, or LOCAL")),
	}, storageWriteFields...)
	stUpdate := mcpgo.NewTool("storage_update", stUpdateOpts...)
	stUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		method, err := req.RequireString("method")
		if err != nil {
			return errResult(err), nil
		}
		st := &storage.Storage{ID: id, Name: name, Method: method,
			Username: req.GetString("username", ""), Password: req.GetString("password", ""),
			Token: req.GetString("token", ""), URL: req.GetString("url", ""), Path: req.GetString("path", "")}
		return runWrite(func() (any, error) {
			return d.Tandoor.Storages().Update(ctx, st)
		})
	}

	stPatchOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Partially update a storage backend. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Storage ID")),
		mcpgo.WithString("name", mcpgo.Description("Storage name")),
		mcpgo.WithString("method", mcpgo.Description("Storage method: DB, NEXTCLOUD, or LOCAL")),
	}, storageWriteFields...)
	stPatch := mcpgo.NewTool("storage_patch", stPatchOpts...)
	stPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		st := &storage.Storage{ID: id}
		if v, err := req.RequireString("name"); err == nil {
			st.Name = v
		}
		if v, err := req.RequireString("method"); err == nil {
			st.Method = v
		}
		for _, field := range []struct {
			key  string
			dest *string
		}{{"username", &st.Username}, {"password", &st.Password}, {"token", &st.Token}, {"url", &st.URL}, {"path", &st.Path}} {
			if v, err := req.RequireString(field.key); err == nil {
				*field.dest = v
			}
		}
		return runWrite(func() (any, error) {
			return d.Tandoor.Storages().Patch(ctx, st)
		})
	}

	stDelete := mcpgo.NewTool("storage_delete",
		mcpgo.WithDescription("Delete a storage backend. Fails if files still reference it."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Storage ID")),
	)
	stDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(fmt.Sprintf("api/storage/%d/", id), func() error {
			return d.Tandoor.Storages().Delete(ctx, id)
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
