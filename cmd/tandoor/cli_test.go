// cmd/tandoor/cli_test.go
//
// Shared harness for CLI tests: runs the full command tree (newApp) against
// an httptest server that emulates the Tandoor API.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// fakeAPI is a permissive Tandoor API emulator. Unrouted paths get
// reasonable generic responses so simple pass-through commands work without
// per-endpoint setup; tests can override behavior via routes.
type fakeAPI struct {
	mu     sync.Mutex
	routes map[string]http.HandlerFunc // key: "METHOD /path/" or "METHOD /path"
	hits   []string                    // recorded "METHOD path"
	fail   int                         // if non-zero, respond with this status for all requests
	url    string
}

// failAll makes every request return the given status code (for error-branch
// tests). Restored automatically at test cleanup.
func (f *fakeAPI) failAll(t *testing.T, status int) {
	t.Helper()
	f.mu.Lock()
	f.fail = status
	f.mu.Unlock()
	t.Cleanup(func() {
		f.mu.Lock()
		f.fail = 0
		f.mu.Unlock()
	})
}

func newFakeAPI(t *testing.T) *fakeAPI {
	t.Helper()
	f := &fakeAPI{routes: map[string]http.HandlerFunc{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.hits = append(f.hits, r.Method+" "+r.URL.Path)
		fail := f.fail
		f.mu.Unlock()
		if fail != 0 {
			w.WriteHeader(fail)
			return
		}
		key := r.Method + " " + r.URL.Path
		if h, ok := f.routes[key]; ok {
			h(w, r)
			return
		}
		f.generic(w, r)
	}))
	t.Cleanup(server.Close)
	f.url = server.URL
	return f
}

func (f *fakeAPI) generic(w http.ResponseWriter, r *http.Request) {
	// Unit lists include a "gram" unit so name-based unit resolution works.
	if r.Method == http.MethodGet && r.URL.Path == "/api/unit/" {
		writeJSONBody(w, map[string]any{
			"count": 1, "next": nil, "previous": nil,
			"results": []map[string]any{{"id": 1, "name": "gram"}},
		})
		return
	}
	// Body for mutating requests: echo it back (plus an id).
	if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body == nil {
			body = map[string]any{}
		}
		if _, ok := body["id"]; !ok {
			body["id"] = 1
		}
		writeJSONBody(w, body)
		return
	}
	if r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// GET: list endpoint (trailing slash, no id) vs object endpoint.
	trimmed := strings.TrimSuffix(r.URL.Path, "/")
	if id := trailingID(trimmed); id > 0 {
		writeJSONBody(w, map[string]any{"id": id, "name": "Widget"})
		return
	}
	writeJSONBody(w, map[string]any{
		"count": 1, "next": nil, "previous": nil,
		"results": []map[string]any{{"id": 1, "name": "Widget"}},
	})
}

func trailingID(path string) int {
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return 0
	}
	var id int
	if _, err := fmt.Sscanf(path[idx+1:], "%d", &id); err != nil {
		return 0
	}
	return id
}

func writeJSONBody(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (f *fakeAPI) route(method, path string, h http.HandlerFunc) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.routes[method+" "+path] = h
}

// run executes the CLI with the given args (after the program name) against
// the fake server, returning the error (nil on success).
func (f *fakeAPI) run(t *testing.T, args ...string) error {
	t.Helper()
	all := append([]string{"tandoor", "--base-url", f.url}, args...)
	app := newApp()
	return app.Run(context.Background(), all)
}

func (f *fakeAPI) hitCount(method, path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, h := range f.hits {
		if h == method+" "+path {
			n++
		}
	}
	return n
}

// runApp runs the CLI without a fake server (for arg-validation tests).
func runApp(t *testing.T, args ...string) error {
	t.Helper()
	app := newApp()
	return app.Run(context.Background(), append([]string{"tandoor", "--base-url", "http://127.0.0.1:1"}, args...))
}
