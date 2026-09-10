package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/swedishborgie/go-tandoor/internal/testutil"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func errExecutor(t *testing.T) *testutil.MockExecutor {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"Not found."}`))
	}))
	t.Cleanup(server.Close)
	return testutil.NewMockExecutor(server.URL)
}

// TestStorageErrorBranches drives every storage service method against a 404
// server to cover the error branches.
func TestStorageErrorBranches(t *testing.T) {
	e := errExecutor(t)
	ctx := context.Background()

	s := NewService(e)
	if _, err := s.List(ctx, &ListOptions{}); err == nil {
		t.Error("Storage List: expected error")
	}
	if _, err := s.Get(ctx, 1); err == nil {
		t.Error("Storage Get: expected error")
	}
	if _, err := s.Create(ctx, &Storage{}); err == nil {
		t.Error("Storage Create: expected error")
	}
	if _, err := s.Update(ctx, &Storage{ID: 1}); err == nil {
		t.Error("Storage Update: expected error")
	}
	if _, err := s.Patch(ctx, &Storage{ID: 1}); err == nil {
		t.Error("Storage Patch: expected error")
	}
	if err := s.Delete(ctx, 1); err == nil {
		t.Error("Storage Delete: expected error")
	}
	if _, err := s.Cascading(ctx, 1, &pagination.ListOptions{}); err == nil {
		t.Error("Storage Cascading: expected error")
	}
	if _, err := s.Nulling(ctx, 1, &pagination.ListOptions{}); err == nil {
		t.Error("Storage Nulling: expected error")
	}
	if _, err := s.Protecting(ctx, 1, &pagination.ListOptions{}); err == nil {
		t.Error("Storage Protecting: expected error")
	}
}
