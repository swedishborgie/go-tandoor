// cmd/tandoor/mcp.go

package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/fdc"
	"github.com/swedishborgie/go-tandoor/internal/mcprun"
)

// GetMCPCommand returns the "mcp" subcommand, which runs this binary as a
// Model Context Protocol server. It lets a single binary operate either as
// the tandoor or as an MCP server:
//
//	TANDOOR_BASE_URL=... TANDOOR_TOKEN=... tandoor mcp
func GetMCPCommand() *cli.Command {
	return &cli.Command{
		Name:  "mcp",
		Usage: "Run this binary as an MCP server (stdio or http transport)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "fdc-api-key",
				Usage:   "USDA FoodData Central API key (optional; enables fdc_* tools)",
				Sources: cli.EnvVars("FDC_API_KEY"),
			},
			&cli.StringFlag{
				Name:  "transport",
				Usage: "MCP transport: stdio or http",
				Value: "stdio",
			},
			&cli.StringFlag{
				Name:    "http-addr",
				Usage:   "Listen address for the http transport",
				Value:   "127.0.0.1:8090",
				Sources: cli.EnvVars("TANDOOR_MCP_HTTP_ADDR"),
			},
			&cli.BoolFlag{
				Name:    "read-only",
				Usage:   "Register read tools only (write tools are not exposed)",
				Sources: cli.EnvVars("TANDOOR_MCP_READ_ONLY"),
			},
			&cli.StringSliceFlag{
				Name:    "tools",
				Usage:   "Tool filter specs: name or prefix_* (allow); -name to deny",
				Sources: cli.EnvVars("TANDOOR_MCP_TOOLS"),
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			token := cmd.String("token")
			if token == "" {
				return fmt.Errorf("no API token set: use --token or TANDOOR_TOKEN")
			}
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)

			var fdcClient *fdc.Client
			if key := cmd.String("fdc-api-key"); key != "" {
				var err error
				if fdcClient, err = fdc.NewClient(fdc.WithAPIKey(key)); err != nil {
					return fmt.Errorf("create fdc client: %w", err)
				}
			}

			return mcprun.Serve(ctx, c, fdcClient, mcprun.Options{
				PageSize:  cmd.Int("page-size"),
				ReadOnly:  cmd.Bool("read-only"),
				Tools:     cmd.StringSlice("tools"),
				Transport: cmd.String("transport"),
				HTTPAddr:  cmd.String("http-addr"),
			})
		},
	}
}
