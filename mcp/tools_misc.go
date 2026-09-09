package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

func registerMiscTools(d *deps) []toolDef {
	// --- view logs ---
	viewLogList := mcpgo.NewTool("view_log_list", append(
		[]mcpgo.ToolOption{mcpgo.WithDescription("List view logs (recipe view history). Use all=true for the full set; jq projects fields to keep output small.")},
		baselineListParams()...,
	)...)
	viewLogListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		return plainPaginatedList(ctx, req, d, d.Tandoor.ViewLogs().List)
	}

	viewLogGet := mcpgo.NewTool("view_log_get",
		mcpgo.WithDescription("Get a single view log entry by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("View log ID")),
		jqParam(),
	)
	viewLogGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.ViewLogs().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- user files ---
	userFileList := mcpgo.NewTool("user_file_list", append(
		[]mcpgo.ToolOption{mcpgo.WithDescription("List user files (uploaded attachments). Use all=true for the full set; jq projects fields to keep output small.")},
		baselineListParams()...,
	)...)
	userFileListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		return plainPaginatedList(ctx, req, d, d.Tandoor.UserFiles().List)
	}

	userFileGet := mcpgo.NewTool("user_file_get",
		mcpgo.WithDescription("Get a single user file's metadata by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("User file ID")),
		jqParam(),
	)
	userFileGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.UserFiles().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- automations ---
	automationList := mcpgo.NewTool("automation_list", append(
		[]mcpgo.ToolOption{mcpgo.WithDescription("List automations (scheduled tasks). Use all=true for the full set; jq projects fields to keep output small.")},
		baselineListParams()...,
	)...)
	automationListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		return plainPaginatedList(ctx, req, d, d.Tandoor.Automations().List)
	}

	automationGet := mcpgo.NewTool("automation_get",
		mcpgo.WithDescription("Get a single automation by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Automation ID")),
		jqParam(),
	)
	automationGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Automations().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- custom filters ---
	customFilterList := mcpgo.NewTool("custom_filter_list", append(
		[]mcpgo.ToolOption{mcpgo.WithDescription("List custom filters (saved recipe search filters). Use all=true for the full set; jq projects fields to keep output small.")},
		baselineListParams()...,
	)...)
	customFilterListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		return plainPaginatedList(ctx, req, d, d.Tandoor.CustomFilters().List)
	}

	customFilterGet := mcpgo.NewTool("custom_filter_get",
		mcpgo.WithDescription("Get a single custom filter by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Custom filter ID")),
		jqParam(),
	)
	customFilterGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.CustomFilters().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- connector configs ---
	connectorList := mcpgo.NewTool("connector_config_list", append(
		[]mcpgo.ToolOption{mcpgo.WithDescription("List connector configurations (external service integrations). Use all=true for the full set; jq projects fields to keep output small.")},
		baselineListParams()...,
	)...)
	connectorListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		return plainPaginatedList(ctx, req, d, d.Tandoor.ConnectorConfigs().List)
	}

	connectorGet := mcpgo.NewTool("connector_config_get",
		mcpgo.WithDescription("Get a single connector configuration by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Connector config ID")),
		jqParam(),
	)
	connectorGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.ConnectorConfigs().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- search fields (flat-array endpoint) ---
	searchFieldList := mcpgo.NewTool("search_field_list",
		mcpgo.WithDescription("List custom search fields (returns the full set as a single array). Use jq to project fields."),
		mcpgo.WithString("jq", mcpgo.Description("jq filter to project/transform the result")),
	)
	searchFieldListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		res, err := d.Tandoor.SearchFields().List(ctx, nil)
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

	// --- search preference (flat-array endpoint) ---
	searchPrefList := mcpgo.NewTool("search_preference_list",
		mcpgo.WithDescription("List search preferences (per-user default search settings; returns the full set as a single array). Use jq to project fields."),
		mcpgo.WithString("jq", mcpgo.Description("jq filter to project/transform the result")),
	)
	searchPrefListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		res, err := d.Tandoor.SearchPreference().List(ctx, nil)
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

	// --- AI providers ---
	aiProviderList := mcpgo.NewTool("ai_provider_list", append(
		[]mcpgo.ToolOption{mcpgo.WithDescription("List AI provider configurations. Use all=true for the full set; jq projects fields to keep output small.")},
		baselineListParams()...,
	)...)
	aiProviderListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		return plainPaginatedList(ctx, req, d, d.Tandoor.AiProviders().List)
	}

	aiProviderGet := mcpgo.NewTool("ai_provider_get",
		mcpgo.WithDescription("Get a single AI provider configuration by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("AI provider ID")),
		jqParam(),
	)
	aiProviderGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.AiProviders().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- AI logs ---
	aiLogList := mcpgo.NewTool("ai_log_list", append(
		[]mcpgo.ToolOption{mcpgo.WithDescription("List AI logs (AI feature usage history). Use all=true for the full set; jq projects fields to keep output small.")},
		baselineListParams()...,
	)...)
	aiLogListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		return plainPaginatedList(ctx, req, d, d.Tandoor.AiLogs().List)
	}

	aiLogGet := mcpgo.NewTool("ai_log_get",
		mcpgo.WithDescription("Get a single AI log entry by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("AI log ID")),
		jqParam(),
	)
	aiLogGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.AiLogs().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- localization ---
	localization := mcpgo.NewTool("localization_list",
		mcpgo.WithDescription("List available localizations (language/locale settings); returns the full set as a single array."),
		jqParam(),
	)
	localizationHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		res, err := d.Tandoor.Localization().List(ctx)
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

	// --- server settings ---
	serverSettings := mcpgo.NewTool("server_settings_get",
		mcpgo.WithDescription("Get the current server settings (read-only view of instance-wide settings)."),
		jqParam(),
	)
	serverSettingsHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		res, err := d.Tandoor.ServerSettings().Settings(ctx)
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

	searchPrefPatch := mcpgo.NewTool("search_preference_patch",
		mcpgo.WithDescription("Update a user's search preferences (data passes through verbatim; fields: search, lookup, trigram_threshold, unaccent/icontains/istartswith/trigram/fulltext as lists of {name, field} objects)."),
		mcpgo.WithInteger("user_id", mcpgo.Required(), mcpgo.Description("User ID whose preferences to update (search preferences are per-user)")),
		mcpgo.WithAny("data", mcpgo.Required(), mcpgo.Description("Preference fields to set as a JSON object (only fields present are changed)")),
		dryRunParam(),
	)
	searchPrefPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		userID, err := req.RequireInt("user_id")
		if err != nil {
			return errResult(fmt.Errorf("user_id is required: %w", err)), nil
		}
		data, ok := req.GetArguments()["data"].(map[string]any)
		if !ok {
			return errResult(fmt.Errorf("data must be a JSON object")), nil
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/search-preference/%d/", userID), data, func() (any, error) {
			var out any
			if err := d.Tandoor.DoJSON(ctx, "PATCH", fmt.Sprintf("api/search-preference/%d/", userID), data, &out); err != nil {
				return nil, err
			}
			return out, nil
		})
	}

	return []toolDef{
		{tool: viewLogList, handler: viewLogListHandler},
		{tool: viewLogGet, handler: viewLogGetHandler},
		{tool: userFileList, handler: userFileListHandler},
		{tool: userFileGet, handler: userFileGetHandler},
		{tool: automationList, handler: automationListHandler},
		{tool: automationGet, handler: automationGetHandler},
		{tool: customFilterList, handler: customFilterListHandler},
		{tool: customFilterGet, handler: customFilterGetHandler},
		{tool: connectorList, handler: connectorListHandler},
		{tool: connectorGet, handler: connectorGetHandler},
		{tool: searchFieldList, handler: searchFieldListHandler},
		{tool: searchPrefList, handler: searchPrefListHandler},
		{tool: searchPrefPatch, handler: searchPrefPatchHandler, write: true},
		{tool: aiProviderList, handler: aiProviderListHandler},
		{tool: aiProviderGet, handler: aiProviderGetHandler},
		{tool: aiLogList, handler: aiLogListHandler},
		{tool: aiLogGet, handler: aiLogGetHandler},
		{tool: localization, handler: localizationHandler},
		{tool: serverSettings, handler: serverSettingsHandler},
	}
}
