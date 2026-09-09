// Package mcp provides an embeddable Model Context Protocol (MCP) server
// for the Tandoor Recipes API.
//
// The server is built from an already-constructed *tandoor.Client (and an
// optional *fdc.Client) so all configuration and authentication lives with
// the caller (see the "tandoor-cli mcp" subcommand), keeping this package trivially
// embeddable and testable.
package mcp

import (
	"fmt"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	tandoor "github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/fdc"
)

const (
	// ServerName is the MCP server implementation name.
	ServerName = "tandoor-mcp"
	// ServerVersion is the MCP server implementation version.
	ServerVersion = "0.1.0"
)

// Config holds per-server configuration shared by all tools.
type Config struct {
	// DefaultPageSize is used by list tools when page_size is not given.
	DefaultPageSize int
	// ReadOnly disables write tools (they are not registered).
	ReadOnly bool
}

// deps bundles the clients and configuration shared by all tool handlers.
type deps struct {
	Tandoor *tandoor.Client
	FDC     *fdc.Client // nil when FDC_API_KEY is not set
	Cfg     Config
}

// toolDef is a registered tool plus registration metadata.
type toolDef struct {
	tool    mcpgo.Tool
	handler server.ToolHandlerFunc
	write   bool // write tools are skipped in read-only mode
}

//go:generate go test -count=1 -run TestCatalog . -args -gen

// NewServer builds the tandoor MCP server.
//
// The tandoor client must already be constructed (base URL, auth). fdcClient
// may be nil; FDC-backed tools then return a clear "FDC_API_KEY not set"
// error.
func NewServer(tandoorClient *tandoor.Client, fdcClient *fdc.Client, opts ...Option) *server.MCPServer {
	cfg := Config{DefaultPageSize: 50}
	o := options{cfg: &cfg}
	for _, opt := range opts {
		opt(&o)
	}

	d := &deps{Tandoor: tandoorClient, FDC: fdcClient, Cfg: cfg}

	// Register all tool groups, then filter.
	var defs []toolDef
	defs = append(defs, registerServerTools(d)...)
	defs = append(defs, registerSpaceTools(d)...)
	defs = append(defs, registerAuthTools(d)...)
	defs = append(defs, registerIngredientTools(d)...)
	defs = append(defs, registerStepTools(d)...)
	defs = append(defs, registerFoodTools(d)...)
	defs = append(defs, registerKeywordTools(d)...)
	defs = append(defs, registerUnitTools(d)...)
	defs = append(defs, registerPropertyTools(d)...)
	defs = append(defs, registerBookTools(d)...)
	defs = append(defs, registerShoppingTools(d)...)
	defs = append(defs, registerMealplanTools(d)...)
	defs = append(defs, registerCookLogTools(d)...)
	defs = append(defs, registerInventoryTools(d)...)
	defs = append(defs, registerStorageTools(d)...)
	defs = append(defs, registerSupermarketTools(d)...)
	defs = append(defs, registerImportTools(d)...)
	defs = append(defs, registerSyncTools(d)...)
	defs = append(defs, registerMiscTools(d)...)
	defs = append(defs, registerRecipeTools(d)...)
	defs = append(defs, registerAuditTools(d)...)
	defs = append(defs, registerFdcTools(d)...)

	var filtered []toolDef
	for _, td := range defs {
		if cfg.ReadOnly && td.write {
			continue
		}
		if !toolAllowed(td.tool.Name, o.allow, o.deny) {
			continue
		}
		filtered = append(filtered, td)
	}

	s := server.NewMCPServer(ServerName, ServerVersion,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
		server.WithInstructions(serverInstructions(d, len(filtered))),
	)
	for i := range filtered {
		applyAnnotations(&filtered[i].tool, filtered[i].write)
		s.AddTool(filtered[i].tool, filtered[i].handler)
	}
	return s
}

// applyAnnotations sets the MCP tool annotations clients use to gate
// writes: read tools are flagged read-only, write tools destructive.
func applyAnnotations(tool *mcpgo.Tool, write bool) {
	readOnly := !write
	destructive := write
	openWorld := true // all tools hit a remote Tandoor instance
	tool.Annotations.ReadOnlyHint = &readOnly
	tool.Annotations.DestructiveHint = &destructive
	tool.Annotations.OpenWorldHint = &openWorld
}

// serverInstructions is the operational brief given to the model at
// initialization.
func serverInstructions(d *deps, toolCount int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Tandoor Recipes API at %s. %d tools registered. IDs are integers. ",
		d.Tandoor.BaseURLOrigin(), toolCount)
	b.WriteString("Prefer *_list tools with query/filters; add jq to keep output small; all=true for full results. ")
	if d.Cfg.ReadOnly {
		b.WriteString("Read-only mode: write tools are not registered. ")
	} else {
		b.WriteString("Write tools perform real writes; destructive ones are annotated for client approval. ")
	}
	if d.FDC == nil {
		b.WriteString("FDC_API_KEY is not set, so fdc_* tools return an error until it is configured. ")
	} else {
		b.WriteString("fdc_* tools are available. ")
	}
	b.WriteString("Property type and unit IDs are instance-specific — enumerate them (property_type_list, unit_list) before using them.")
	return b.String()
}

// toolAllowed applies the allow/deny tool filter.
//
// Spec entries are exact names or "prefix_*" patterns. If any allow entries
// are present, only allowed tools are registered; deny entries always win.
func toolAllowed(name string, allow, deny []string) bool {
	match := func(specs []string) bool {
		for _, spec := range specs {
			if strings.HasSuffix(spec, "*") {
				if strings.HasPrefix(name, strings.TrimSuffix(spec, "*")) {
					return true
				}
				continue
			}
			if name == spec {
				return true
			}
		}
		return false
	}
	if match(deny) {
		return false
	}
	if len(allow) == 0 {
		return true
	}
	return match(allow)
}
