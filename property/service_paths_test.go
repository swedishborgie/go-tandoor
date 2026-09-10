package property

import (
	"context"
	"encoding/json"
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

// TestPropertyServiceHappyPaths covers the success returns for the property
// and property-type services.
func TestPropertyServiceHappyPaths(t *testing.T) {
	ctx := context.Background()

	s := NewService(okServer(t, `{"id":1,"type":2,"amount":1}`))
	p, err := s.Get(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, p)
	_, err = s.Create(ctx, &Property{Type: Type{ID: 2}})
	require.NoError(t, err)
	_, err = s.Update(ctx, &Property{ID: 1})
	require.NoError(t, err)
	_, err = s.Patch(ctx, &Property{ID: 1})
	require.NoError(t, err)
	require.NoError(t, s.Delete(ctx, 1))

	ts := NewTypeService(okServer(t, `{"id":1,"name":"salt"}`))
	_, err = ts.Get(ctx, 1)
	require.NoError(t, err)
	_, err = ts.Create(ctx, &Type{Name: "salt"})
	require.NoError(t, err)
	_, err = ts.Update(ctx, &Type{ID: 1})
	require.NoError(t, err)
	_, err = ts.Patch(ctx, &Type{ID: 1})
	require.NoError(t, err)
	require.NoError(t, ts.Delete(ctx, 1))
}

// TestPropertyServiceQueryStrings covers the "path += ?query" branches.
func TestPropertyServiceQueryStrings(t *testing.T) {
	ctx := context.Background()
	e := okServer(t, `{"count":0,"results":[]}`)

	s := NewService(e)
	_, err := s.List(ctx, &ListOptions{ListOptions: &pagination.ListOptions{Page: 2}})
	require.NoError(t, err)
	lo := &pagination.ListOptions{Page: 2}
	_, err = s.Cascading(ctx, 1, lo)
	require.NoError(t, err)
	_, err = s.Nulling(ctx, 1, lo)
	require.NoError(t, err)
	_, err = s.Protecting(ctx, 1, lo)
	require.NoError(t, err)

	ts := NewTypeService(e)
	_, err = ts.List(ctx, &TypeListOptions{ListOptions: pagination.ListOptions{Page: 2}})
	require.NoError(t, err)
	_, err = ts.Cascading(ctx, 1, lo)
	require.NoError(t, err)
	_, err = ts.Nulling(ctx, 1, lo)
	require.NoError(t, err)
	_, err = ts.Protecting(ctx, 1, lo)
	require.NoError(t, err)
}

// TestPropertyUnmarshalErrors covers the UnmarshalJSON error branches.
func TestPropertyUnmarshalErrors(t *testing.T) {
	var ty Type
	require.Error(t, json.Unmarshal([]byte(`{`), &ty))
	var fr FoodRef
	require.Error(t, json.Unmarshal([]byte(`{`), &fr))
}
