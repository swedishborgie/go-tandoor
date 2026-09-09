package mcp

import (
	"context"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

func registerServerTools(d *deps) []toolDef {
	info := mcpgo.NewTool("server_info",
		mcpgo.WithDescription("Server status: Tandoor base URL, tool count, and read-only/FDC configuration. Call first to understand the environment."),
		jqParam(),
	)
	return []toolDef{{
		tool: info,
		handler: func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			return jsonResultJQ(ctx, req, map[string]any{
				"server":         ServerName,
				"version":        ServerVersion,
				"base_url":       d.Tandoor.BaseURLOrigin(),
				"read_only":      d.Cfg.ReadOnly,
				"fdc_configured": d.FDC != nil,
			})
		},
	}}
}
