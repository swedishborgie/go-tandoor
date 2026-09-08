// Package executor defines the internal interface that the tandoor Client implements.
//
// This interface is placed in an internal package so that sub-packages (auth, food,
// recipe, etc.) can reference it without creating import cycles with the root package.
//
// The interface exposes only what services need: JSON API calls, raw HTTP responses
// for non-JSON endpoints, and the base URL origin for manual URL construction.
package executor

import (
	"context"
	"net/http"
)

// Executor is the interface implemented by the tandoor Client.
// It is placed in an internal package so callers outside the module
// cannot implement it, but sub-packages within the module can reference it.
type Executor interface {
	// DoJSON creates a request, executes it, and decodes the JSON response into v.
	DoJSON(ctx context.Context, method, path string, body any, v any) error

	// DoRaw creates a request and executes it, returning the raw *http.Response.
	// The request includes standard headers (Authorization, Content-Type).
	// The caller is responsible for reading and closing the response body.
	DoRaw(ctx context.Context, method, path string, body any) (*http.Response, error)

	// DoRawNoAuth creates a request WITHOUT the Authorization header and
	// executes it, returning the raw *http.Response.
	// Used for auth endpoints that must not carry a Bearer token.
	DoRawNoAuth(ctx context.Context, method, path string, body any) (*http.Response, error)

	// BaseURLOrigin returns the origin URL (scheme + host).
	BaseURLOrigin() string
}
