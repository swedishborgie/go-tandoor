// tandoor-mcp is a Model Context Protocol (MCP) server for the Tandoor
// Recipes API. It exposes the breadth of the Tandoor API as MCP tools with
// sensible defaults, optional parameters, and dry-run support on writes.
//
// In stdio mode the process speaks MCP over stdin/stdout and must never
// write anything else to stdout; diagnostics go to stderr.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mcpgoserver "github.com/mark3labs/mcp-go/server"
	"github.com/urfave/cli/v3"

	tandoor "github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/fdc"
	mcpsrv "github.com/swedishborgie/go-tandoor/mcp"
)

const appVersion = "0.1.0"

func main() {
	app := &cli.Command{
		Name:    "tandoor-mcp",
		Usage:   "MCP server for the Tandoor Recipes API",
		Version: appVersion,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "base-url",
				Usage:   "Tandoor instance base URL (required; or set TANDOOR_BASE_URL)",
				Sources: cli.EnvVars("TANDOOR_BASE_URL"),
			},
			&cli.StringFlag{
				Name:    "token",
				Usage:   "Tandoor API access token (required; or set TANDOOR_TOKEN)",
				Sources: cli.EnvVars("TANDOOR_TOKEN"),
			},
			&cli.StringFlag{
				Name:    "fdc-api-key",
				Usage:   "USDA FoodData Central API key (optional; enables fdc_* tools)",
				Sources: cli.EnvVars("FDC_API_KEY"),
			},
			&cli.IntFlag{
				Name:    "page-size",
				Usage:   "Default page size for list tools",
				Value:   50,
				Sources: cli.EnvVars("TANDOOR_PAGE_SIZE"),
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
			baseURL := cmd.String("base-url")
			if baseURL == "" {
				return fmt.Errorf("no base URL set: use --base-url or TANDOOR_BASE_URL")
			}
			token := cmd.String("token")
			if token == "" {
				return fmt.Errorf("no API token set: use --token or TANDOOR_TOKEN")
			}

			c, err := tandoor.NewClient(baseURL, tandoor.WithAccessToken(token))
			if err != nil {
				return fmt.Errorf("create tandoor client: %w", err)
			}

			var fdcClient *fdc.Client
			if key := cmd.String("fdc-api-key"); key != "" {
				if fdcClient, err = fdc.NewClient(fdc.WithAPIKey(key)); err != nil {
					return fmt.Errorf("create fdc client: %w", err)
				}
			}

			var srvOpts []mcpsrv.Option
			if n := cmd.Int("page-size"); n > 0 {
				srvOpts = append(srvOpts, mcpsrv.WithDefaultPageSize(n))
			}
			if cmd.Bool("read-only") {
				srvOpts = append(srvOpts, mcpsrv.WithReadOnly())
			}
			if specs := cmd.StringSlice("tools"); len(specs) > 0 {
				srvOpts = append(srvOpts, mcpsrv.WithToolFilter(specs...))
			}

			s := mcpsrv.NewServer(c, fdcClient, srvOpts...)

			// All diagnostics go to stderr; stdout is reserved for MCP
			// frames in stdio mode.
			log.SetOutput(os.Stderr)
			log.SetFlags(0)

			switch cmd.String("transport") {
			case "stdio":
				log.Printf("tandoor-mcp: stdio transport, base URL %s", c.BaseURLOrigin())
				return mcpgoserver.ServeStdio(s)
			case "http":
				addr := cmd.String("http-addr")
				log.Printf("tandoor-mcp: http transport on %s, base URL %s", addr, c.BaseURLOrigin())
				hs := mcpgoserver.NewStreamableHTTPServer(s, mcpgoserver.WithEndpointPath("/mcp"))
				httpServer := &http.Server{Addr: addr, Handler: hs}
				errCh := make(chan error, 1)
				go func() { errCh <- httpServer.ListenAndServe() }()
				stop := make(chan os.Signal, 1)
				signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
				select {
				case err := <-errCh:
					return err
				case sig := <-stop:
					log.Printf("tandoor-mcp: received %s, shutting down", sig)
					shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					return httpServer.Shutdown(shutdownCtx)
				}
			default:
				return fmt.Errorf("unknown transport %q (use stdio or http)", cmd.String("transport"))
			}
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
