package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestUnitService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/unit/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1,"name":"cup"},{"id":2,"name":"tablespoon"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
	assert.Equal(t, "cup", page.Results[0].Name)
}

func TestUnitService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/unit/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"cup","plural_name":"cups"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	u, err := NewService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "cups", u.PluralName)
}

func TestUnitService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"pinch"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	u, err := NewService(mock).Create(context.Background(), &Unit{Name: "pinch"})
	require.NoError(t, err)
	assert.Equal(t, 10, u.ID)
}

func TestUnitService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"US cup"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	u, err := NewService(mock).Update(context.Background(), &Unit{ID: 1, Name: "US cup"})
	require.NoError(t, err)
	assert.Equal(t, "US cup", u.Name)
}

func TestUnitService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"description":"240ml"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	u, err := NewService(mock).Patch(context.Background(), &Unit{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "240ml", u.Description)
}

func TestUnitService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/unit/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestUnitService_Merge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/unit/1/merge/2/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":2,"name":"merged unit"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	u, err := NewService(mock).Merge(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 2, u.ID)
}

func TestConversionService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/unit-conversion/", r.URL.Path)
		assert.Equal(t, "796", r.URL.Query().Get("food_id"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"base_amount":1,"base_unit":{"id":27,"name":"cup"},"converted_amount":140,"converted_unit":{"id":17,"name":"g"},"food":{"id":796,"name":"Ground Beef"}}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	foodID := 796
	page, err := NewConversionService(mock).List(context.Background(), &ConversionListOptions{
		ListOptions: &pagination.ListOptions{PageSize: 10},
		FoodID:      &foodID,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	require.NotNil(t, page.Results[0].BaseUnit)
	assert.Equal(t, 27, page.Results[0].BaseUnit.ID)
}

func TestConversionService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/unit-conversion/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"base_amount":1,"base_unit":{"id":27,"name":"cup"},"converted_amount":140,"converted_unit":{"id":17,"name":"g"},"food":{"id":796,"name":"Ground Beef"}}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	conv, err := NewConversionService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1, conv.ID)
}

func TestConversionService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/unit-conversion/", r.URL.Path)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":2,"base_amount":1,"base_unit":{"id":27,"name":"cup"},"converted_amount":140,"converted_unit":{"id":17,"name":"g"},"food":{"id":796,"name":"Ground Beef"}}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	created, err := NewConversionService(mock).Create(context.Background(), &Conversion{
		BaseAmount:      1,
		BaseUnit:        &Ref{ID: 27},
		ConvertedAmount: 140,
		ConvertedUnit:   &Ref{ID: 17},
		Food:            &FoodRef{ID: 796},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, created.ID)
}

func TestConversionService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/unit-conversion/2/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":2,"base_amount":1,"base_unit":{"id":27,"name":"cup"},"converted_amount":155,"converted_unit":{"id":17,"name":"g"},"food":{"id":1231,"name":"Frozen Mixed Vegetables"}}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	updated, err := NewConversionService(mock).Update(context.Background(), &Conversion{
		ID:              2,
		BaseAmount:      1,
		BaseUnit:        &Ref{ID: 27},
		ConvertedAmount: 155,
		ConvertedUnit:   &Ref{ID: 17},
		Food:            &FoodRef{ID: 1231},
	})
	require.NoError(t, err)
	assert.InEpsilon(t, 155.0, updated.ConvertedAmount, 1e-9)
}

func TestConversionService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/unit-conversion/2/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	err := NewConversionService(mock).Delete(context.Background(), 2)
	require.NoError(t, err)
}
