// Package mcprun serves a pre-built Tandoor MCP server over the stdio or
// streamable-HTTP transport. It backs the "tandoor-cli mcp" subcommand.
package mcprun

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

	tandoor "github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/fdc"
	mcpsrv "github.com/swedishborgie/go-tandoor/mcp"
)

// Options controls how the MCP server is served.
type Options struct {
	// PageSize is the default page size for list tools (0 keeps the
	// server default of 50).
	PageSize int
	// ReadOnly registers read tools only; write tools are not exposed.
	ReadOnly bool
	// Tools are tool filter specs: name or prefix_* (allow), -name (deny).
	Tools []string
	// Transport is "stdio" or "http" (empty defaults to stdio).
	Transport string
	// HTTPAddr is the listen address for the http transport
	// (empty defaults to 127.0.0.1:8090).
	HTTPAddr string
}

// Serve builds the MCP server from the given clients and runs it on the
// configured transport. It blocks until the transport closes (stdio) or the
// process receives SIGINT/SIGTERM (http).
//
// All diagnostics go to stderr; stdout is reserved for MCP frames in stdio
// mode. fdcClient may be nil; FDC-backed tools then return a clear
// "FDC_API_KEY not set" error.
func Serve(ctx context.Context, c *tandoor.Client, fdcClient *fdc.Client, opts Options) error {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	var srvOpts []mcpsrv.Option
	if opts.PageSize > 0 {
		srvOpts = append(srvOpts, mcpsrv.WithDefaultPageSize(opts.PageSize))
	}
	if opts.ReadOnly {
		srvOpts = append(srvOpts, mcpsrv.WithReadOnly())
	}
	if len(opts.Tools) > 0 {
		srvOpts = append(srvOpts, mcpsrv.WithToolFilter(opts.Tools...))
	}

	s := mcpsrv.NewServer(c, fdcClient, srvOpts...)

	switch opts.Transport {
	case "", "stdio":
		log.Printf("tandoor-cli mcp: stdio transport, base URL %s", c.BaseURLOrigin())
		return mcpgoserver.ServeStdio(s)
	case "http":
		addr := opts.HTTPAddr
		if addr == "" {
			addr = "127.0.0.1:8090"
		}
		log.Printf("tandoor-cli mcp: http transport on %s, base URL %s", addr, c.BaseURLOrigin())
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
			log.Printf("tandoor-cli mcp: received %s, shutting down", sig)
			// The caller's context is still live while the action runs, so
			// the shutdown inherits it instead of starting from Background.
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return httpServer.Shutdown(shutdownCtx)
		}
	default:
		return fmt.Errorf("unknown transport %q (use stdio or http)", opts.Transport)
	}
}
