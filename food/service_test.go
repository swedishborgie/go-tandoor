package food

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

func TestService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/food/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1,"name":"Flour"},{"id":2,"name":"Sugar"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
	assert.Equal(t, "Flour", page.Results[0].Name)
}

func TestService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/food/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Flour","parent":0,"numchild":0,"full_name":"Flour"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	f, err := NewService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Flour", f.Name)
}

func TestService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":5,"name":"Salt","full_name":"Salt"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	f, err := NewService(mock).Create(context.Background(), &Food{Name: "Salt"})
	require.NoError(t, err)
	assert.Equal(t, 5, f.ID)
}

func TestService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"All Purpose Flour"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	f, err := NewService(mock).Update(context.Background(), &Food{ID: 1, Name: "All Purpose Flour"})
	require.NoError(t, err)
	assert.Equal(t, "All Purpose Flour", f.Name)
}

func TestService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Flour"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	f, err := NewService(mock).Patch(context.Background(), &Food{ID: 1, Name: "Flour"})
	require.NoError(t, err)
	assert.Equal(t, "Flour", f.Name)
}

func TestService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/food/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	err := NewService(mock).Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestService_Merge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/food/1/merge/2/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":2,"name":"Merged Food"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	f, err := NewService(mock).Merge(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 2, f.ID)
}

func TestService_Move(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/food/1/move/10/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"parent":10,"full_name":"Dairy > Milk"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	f, err := NewService(mock).Move(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 10, f.Parent)
}

func TestService_BatchUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/food/batch_update/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1},{"id":2}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	page, err := NewService(mock).BatchUpdate(context.Background(), &BatchUpdate{Foods: []int{1, 2}})
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
}

func TestService_UpdateShopping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/food/1/shopping/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"ignore_shopping":false}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)
	f, err := NewService(mock).UpdateShopping(context.Background(), 1, &ShoppingUpdate{ShoppingListID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, f.ID)
}
