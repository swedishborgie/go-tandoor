package shopping

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

func TestListService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/shopping-list/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1,"name":"Weekly"},{"id":2,"name":"Party"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewListService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
	assert.Equal(t, "Weekly", page.Results[0].Name)
}

func TestListService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/shopping-list/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Weekly","color":"#ff0000"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sl, err := NewListService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "#ff0000", sl.Color)
}

func TestListService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":5,"name":"Weekend"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sl, err := NewListService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 5, sl.ID)
}

func TestListService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Updated"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sl, err := NewListService(mock).Update(context.Background(), &List{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Updated", sl.Name)
}

func TestListService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/shopping-list/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewListService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestListEntryService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/shopping-list-entry/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"amount":2.5,"checked":false}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewEntryService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.InEpsilon(t, 2.5, page.Results[0].Amount, 1e-9)
}

func TestListEntryService_BulkUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/shopping-list-entry/bulk/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ids":[1,2],"timestamp":"2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	result, err := NewEntryService(mock).BulkUpdate(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, result.IDs, 2)
}

func TestListRecipeService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/shopping-list-recipe/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"servings":2}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewRecipeService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.InEpsilon(t, float64(2), page.Results[0].Servings, 1e-9)
}

func TestListRecipeService_BulkCreateEntries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/shopping-list-recipe/1/bulk_create_entries/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"entries":[]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	result, err := NewRecipeService(mock).BulkCreateEntries(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.NotNil(t, result)
}
