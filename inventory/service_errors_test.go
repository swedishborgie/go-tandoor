package inventory

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

// TestInventoryErrorBranches drives every inventory service method against a
// 404 server to cover the error branches.
func TestInventoryErrorBranches(t *testing.T) {
	e := errExecutor(t)
	ctx := context.Background()

	loc := NewLocationService(e)
	if _, err := loc.List(ctx, &LocationListOptions{}); err == nil {
		t.Error("Location List: expected error")
	}
	if _, err := loc.Get(ctx, 1); err == nil {
		t.Error("Location Get: expected error")
	}
	if _, err := loc.Create(ctx, &Location{}); err == nil {
		t.Error("Location Create: expected error")
	}
	if _, err := loc.Update(ctx, &Location{ID: 1}); err == nil {
		t.Error("Location Update: expected error")
	}
	if _, err := loc.Patch(ctx, &Location{ID: 1}); err == nil {
		t.Error("Location Patch: expected error")
	}
	if err := loc.Delete(ctx, 1); err == nil {
		t.Error("Location Delete: expected error")
	}
	if _, err := loc.Cascading(ctx, 1, &pagination.ListOptions{}); err == nil {
		t.Error("Location Cascading: expected error")
	}
	if _, err := loc.Nulling(ctx, 1, &pagination.ListOptions{}); err == nil {
		t.Error("Location Nulling: expected error")
	}
	if _, err := loc.Protecting(ctx, 1, &pagination.ListOptions{}); err == nil {
		t.Error("Location Protecting: expected error")
	}

	entry := NewEntryService(e)
	if _, err := entry.List(ctx, &EntryListOptions{}); err == nil {
		t.Error("Entry List: expected error")
	}
	if _, err := entry.Get(ctx, 1); err == nil {
		t.Error("Entry Get: expected error")
	}
	if _, err := entry.Create(ctx, &Entry{}); err == nil {
		t.Error("Entry Create: expected error")
	}
	if _, err := entry.Update(ctx, &Entry{ID: 1}); err == nil {
		t.Error("Entry Update: expected error")
	}
	if _, err := entry.Patch(ctx, &Entry{ID: 1}); err == nil {
		t.Error("Entry Patch: expected error")
	}
	if err := entry.Delete(ctx, 1); err == nil {
		t.Error("Entry Delete: expected error")
	}

	log := NewLogService(e)
	if _, err := log.List(ctx, &LogListOptions{}); err == nil {
		t.Error("Log List: expected error")
	}
	if _, err := log.Get(ctx, 1); err == nil {
		t.Error("Log Get: expected error")
	}
}
