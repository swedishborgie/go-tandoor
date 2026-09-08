package recipe

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

func TestRecipeService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/recipe/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1,"name":"Pizza"},{"id":2,"name":"Pasta"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
	assert.Equal(t, "Pizza", page.Results[0].Name)
}

func TestRecipeService_ListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "search", r.URL.Query().Get("query"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":10,"results":[{"id":3,"name":"Soup"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), &ListOptions{ListOptions: pagination.ListOptions{Page: 2, Search: "search"}})
	require.NoError(t, err)
	assert.Equal(t, 10, page.Count)
	assert.Equal(t, "Soup", page.Results[0].Name)
}

func TestRecipeService_Flat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/api/recipe/flat/")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Pizza","image":"http://example.com/img.jpg"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).Flat(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "Pizza", page.Results[0].Name)
	assert.Equal(t, "http://example.com/img.jpg", page.Results[0].Image)
}

func TestRecipeService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/recipe/42/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":42,"name":"Tacos","working_time":30,"waiting_time":10}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	r, err := NewService(mock).Get(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, 42, r.ID)
	assert.Equal(t, "Tacos", r.Name)
	assert.Equal(t, 30, r.WorkingTime)
}

func TestRecipeService_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"detail":"Not found."}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	_, err := NewService(mock).Get(context.Background(), 999)
	require.Error(t, err)
	var terr *testutil.MockError
	require.ErrorAs(t, err, &terr)
	assert.True(t, terr.IsNotFound())
}

func TestRecipeService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"New Recipe","working_time":15}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	r, err := NewService(mock).Create(context.Background(), &Recipe{Name: "New Recipe"})
	require.NoError(t, err)
	assert.Equal(t, 10, r.ID)
}

func TestRecipeService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/recipe/10/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":10,"name":"Updated Recipe"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	r, err := NewService(mock).Update(context.Background(), &Recipe{ID: 10, Name: "Updated Recipe"})
	require.NoError(t, err)
	assert.Equal(t, "Updated Recipe", r.Name)
}

func TestRecipeService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":10,"name":"Patched Recipe"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	r, err := NewService(mock).Patch(context.Background(), &Recipe{ID: 10, Name: "Patched Recipe"})
	require.NoError(t, err)
	assert.Equal(t, "Patched Recipe", r.Name)
}

func TestRecipeService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/recipe/10/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).Delete(context.Background(), 10)
	require.NoError(t, err)
}

func TestRecipeService_BatchUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/recipe/batch_update/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1},{"id":2}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	workingTime := 20
	page, err := NewService(mock).BatchUpdate(context.Background(), &BatchUpdate{
		Recipes:     []int{1, 2},
		WorkingTime: &workingTime,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
}

func TestRecipeService_Related(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/recipe/1/related/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":2,"name":"Related"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).Related(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, "Related", page.Results[0].Name)
}

func TestRecipeService_UploadImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/recipe/1/image/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"image":"http://example.com/new.jpg"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	r, err := NewService(mock).UploadImage(context.Background(), 1, &Image{ImageURL: "http://example.com/new.jpg"})
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/new.jpg", r.Image)
}

func TestRecipeService_AddToShopping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/recipe/1/shopping/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).AddToShopping(context.Background(), 1, &ShoppingUpdate{
		Ingredients: []int{1, 2, 3},
	})
	require.NoError(t, err)
}
