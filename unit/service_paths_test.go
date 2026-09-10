package unit

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

// TestUnitServiceHappyPaths covers the success returns for the unit and
// conversion services.
func TestUnitServiceHappyPaths(t *testing.T) {
	ctx := context.Background()

	s := NewService(okServer(t, `{"id":1,"name":"gram"}`))
	_, err := s.Get(ctx, 1)
	require.NoError(t, err)
	_, err = s.Create(ctx, &Unit{Name: "gram"})
	require.NoError(t, err)
	_, err = s.Update(ctx, &Unit{ID: 1})
	require.NoError(t, err)
	_, err = s.Patch(ctx, &Unit{ID: 1})
	require.NoError(t, err)
	require.NoError(t, s.Delete(ctx, 1))
	_, err = s.Merge(ctx, 1, 2)
	require.NoError(t, err)

	cs := NewConversionService(okServer(t, `{"id":1,"name":"c"}`))
	_, err = cs.Get(ctx, 1)
	require.NoError(t, err)
	_, err = cs.Create(ctx, &Conversion{
		BaseAmount: 1, ConvertedAmount: 1000,
		BaseUnit: &Ref{ID: 1}, ConvertedUnit: &Ref{ID: 2},
		Food: &FoodRef{ID: 3}, OpenDataSlug: "slug",
	})
	require.NoError(t, err)
	// Minimal conversion covers the nil-ref branches of convToPayload.
	_, err = cs.Create(ctx, &Conversion{})
	require.NoError(t, err)
	_, err = cs.Update(ctx, &Conversion{ID: 1})
	require.NoError(t, err)
	_, err = cs.Patch(ctx, &Conversion{ID: 1})
	require.NoError(t, err)
	require.NoError(t, cs.Delete(ctx, 1))
}

// TestUnitServiceQueryStrings covers the "path += ?query" branches.
func TestUnitServiceQueryStrings(t *testing.T) {
	ctx := context.Background()
	e := okServer(t, `{"count":0,"results":[]}`)

	s := NewService(e)
	_, err := s.List(ctx, &ListOptions{ListOptions: pagination.ListOptions{Page: 2}})
	require.NoError(t, err)
	lo := &pagination.ListOptions{Page: 2}
	_, err = s.Cascading(ctx, 1, lo)
	require.NoError(t, err)
	_, err = s.Nulling(ctx, 1, lo)
	require.NoError(t, err)
	_, err = s.Protecting(ctx, 1, lo)
	require.NoError(t, err)

	fid := 7
	cs := NewConversionService(e)
	_, err = cs.List(ctx, &ConversionListOptions{
		ListOptions: &pagination.ListOptions{Page: 2},
		FoodID:      &fid,
	})
	require.NoError(t, err)
}

// TestUnitUnmarshalErrors covers the UnmarshalJSON error branches.
func TestUnitUnmarshalErrors(t *testing.T) {
	var u Unit
	require.Error(t, json.Unmarshal([]byte(`{`), &u))
	var r Ref
	require.Error(t, json.Unmarshal([]byte(`{`), &r))
	var fr FoodRef
	require.Error(t, json.Unmarshal([]byte(`{`), &fr))
}
