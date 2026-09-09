package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/importexport"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func registerSyncTools(d *deps) []toolDef {
	listOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List sync configurations (external recipe sync sources). Use all=true for the full set; jq projects fields to keep output small.")}
	listOpts = append(listOpts, baselineListParams()...)
	list := mcpgo.NewTool("sync_list", listOpts...)
	listHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &importexport.SyncListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.Syncs().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[importexport.Sync], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Syncs().List(ctx, &o2)
		})
	}

	get := mcpgo.NewTool("sync_get",
		mcpgo.WithDescription("Get a single sync configuration by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Sync ID")),
		jqParam(),
	)
	getHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Syncs().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	logListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List sync logs (history of sync runs). Use all=true for the full set; jq projects fields to keep output small.")}
	logListOpts = append(logListOpts, baselineListParams()...)
	logList := mcpgo.NewTool("sync_log_list", logListOpts...)
	logListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &importexport.SyncLogListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.SyncLogs().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[importexport.SyncLog], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.SyncLogs().List(ctx, &o2)
		})
	}

	// Sync writes pass the `data` object through verbatim: Tandoor's sync
	// serializer fields are storage (storage ID), path, and active.
	syncDataParam := mcpgo.WithAny("data", mcpgo.Required(), mcpgo.Description("Sync data as a JSON object (Tandoor sync serializer fields: storage (storage ID), path, active)"))

	create := mcpgo.NewTool("sync_create",
		mcpgo.WithDescription("Create a sync configuration (external recipe sync source). Returns the created sync configuration."),
		syncDataParam,
		dryRunParam(),
	)
	createHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		data, err := syncData(req)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "POST", "api/sync/", data, func() (any, error) {
			var out any
			if err := d.Tandoor.DoJSON(ctx, "POST", "api/sync/", data, &out); err != nil {
				return nil, err
			}
			return out, nil
		})
	}

	update := mcpgo.NewTool("sync_update",
		mcpgo.WithDescription("Replace a sync configuration with the given data (full replacement). Returns the updated sync configuration."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Sync ID")),
		syncDataParam,
		dryRunParam(),
	)
	updateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		data, err := syncData(req)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/sync/%d/", id), data, func() (any, error) {
			var out any
			if err := d.Tandoor.DoJSON(ctx, "PUT", fmt.Sprintf("api/sync/%d/", id), data, &out); err != nil {
				return nil, err
			}
			return out, nil
		})
	}

	patch := mcpgo.NewTool("sync_patch",
		mcpgo.WithDescription("Partially update a sync configuration. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Sync ID")),
		syncDataParam,
		dryRunParam(),
	)
	patchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		data, err := syncData(req)
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/sync/%d/", id), data, func() (any, error) {
			var out any
			if err := d.Tandoor.DoJSON(ctx, "PATCH", fmt.Sprintf("api/sync/%d/", id), data, &out); err != nil {
				return nil, err
			}
			return out, nil
		})
	}

	syncDelete := mcpgo.NewTool("sync_delete",
		mcpgo.WithDescription("Delete a sync configuration."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Sync ID")),
		dryRunParam(),
	)
	syncDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/sync/%d/", id), func() error {
			return d.Tandoor.Syncs().Delete(ctx, id)
		})
	}

	querySynced := mcpgo.NewTool("sync_query_synced_folder",
		mcpgo.WithDescription("Trigger a sync run for a sync configuration's folder and return the resulting sync log."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Sync ID")),
		dryRunParam(),
	)
	querySyncedHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return runWrite(ctx, req, "POST", fmt.Sprintf("api/sync/%d/query_synced_folder/", id), nil, func() (any, error) {
			return d.Tandoor.Syncs().QuerySyncedFolder(ctx, id)
		})
	}

	return []toolDef{
		{tool: list, handler: listHandler},
		{tool: get, handler: getHandler},
		{tool: logList, handler: logListHandler},
		{tool: create, handler: createHandler, write: true},
		{tool: update, handler: updateHandler, write: true},
		{tool: patch, handler: patchHandler, write: true},
		{tool: syncDelete, handler: syncDeleteHandler, write: true},
		{tool: querySynced, handler: querySyncedHandler, write: true},
	}
}

// syncData extracts the free-form `data` object from a sync tool call.
func syncData(req mcpgo.CallToolRequest) (map[string]any, error) {
	data, ok := req.GetArguments()["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data must be a JSON object")
	}
	return data, nil
}
