package mcp

import "strings"

// options is the mutable state accumulated by Option values.
type options struct {
	cfg   *Config
	allow []string // allow-list specs; if non-empty, only these register
	deny  []string // deny-list specs; always win
}

// Option configures the MCP server.
type Option func(*options)

// WithReadOnly disables write tools: they are not registered, so a
// deployment with a read-only token cannot issue writes through MCP.
func WithReadOnly() Option {
	return func(o *options) {
		o.cfg.ReadOnly = true
	}
}

// WithDefaultPageSize sets the default page size for list tools.
// The zero value keeps the default of 50.
func WithDefaultPageSize(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.cfg.DefaultPageSize = n
		}
	}
}

// WithToolFilter filters registered tools by name or "prefix_*" pattern.
//
// Spec entries:
//   - "name" or "+name" — allow entry. If any allow entry is present, only
//     allowed tools are registered.
//   - "-name" — deny entry. Deny always wins over allow.
//
// Example: WithToolFilter("recipe_*", "fdc_*", "-shopping_*")
func WithToolFilter(specs ...string) Option {
	return func(o *options) {
		for _, spec := range specs {
			switch {
			case strings.HasPrefix(spec, "-"):
				o.deny = append(o.deny, strings.TrimPrefix(spec, "-"))
			case strings.HasPrefix(spec, "+"):
				o.allow = append(o.allow, strings.TrimPrefix(spec, "+"))
			default:
				o.allow = append(o.allow, spec)
			}
		}
	}
}
