package inventory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func okServer(t *testing.T, body string) *testutil.MockExecutor {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return testutil.NewMockExecutor(server.URL)
}

// TestInventoryServiceHappyPaths covers success returns plus the query
// string branches of the list/cascade endpoints.
func TestInventoryServiceHappyPaths(t *testing.T) {
	ctx := context.Background()
	e := okServer(t, `{"id":1,"name":"pantry"}`)

	loc := NewLocationService(e)
	_, err := loc.List(ctx, &LocationListOptions{ListOptions: pagination.ListOptions{Page: 2}})
	require.NoError(t, err)
	_, err = loc.Get(ctx, 1)
	require.NoError(t, err)
	_, err = loc.Create(ctx, &Location{Name: "pantry"})
	require.NoError(t, err)
	_, err = loc.Update(ctx, &Location{ID: 1})
	require.NoError(t, err)
	_, err = loc.Patch(ctx, &Location{ID: 1})
	require.NoError(t, err)
	require.NoError(t, loc.Delete(ctx, 1))
	lo := &pagination.ListOptions{Page: 2}
	_, err = loc.Cascading(ctx, 1, lo)
	require.NoError(t, err)
	_, err = loc.Nulling(ctx, 1, lo)
	require.NoError(t, err)
	_, err = loc.Protecting(ctx, 1, lo)
	require.NoError(t, err)

	entry := NewEntryService(e)
	_, err = entry.List(ctx, &EntryListOptions{ListOptions: &pagination.ListOptions{Page: 2}})
	require.NoError(t, err)
	_, err = entry.Get(ctx, 1)
	require.NoError(t, err)
	_, err = entry.Create(ctx, &Entry{})
	require.NoError(t, err)
	_, err = entry.Update(ctx, &Entry{ID: 1})
	require.NoError(t, err)
	_, err = entry.Patch(ctx, &Entry{ID: 1})
	require.NoError(t, err)
	require.NoError(t, entry.Delete(ctx, 1))

	log := NewLogService(e)
	_, err = log.List(ctx, &LogListOptions{ListOptions: &pagination.ListOptions{Page: 2}})
	require.NoError(t, err)
	_, err = log.Get(ctx, 1)
	require.NoError(t, err)
}
