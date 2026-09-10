package ingredient

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

func mockStatus(t *testing.T, status int, body string) *testutil.MockExecutor {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return testutil.NewMockExecutor(server.URL)
}

func notFound(t *testing.T) *testutil.MockExecutor {
	t.Helper()
	return mockStatus(t, http.StatusNotFound, `{"detail":"Not found."}`)
}

func TestIngredientService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "3", q.Get("food"))
		assert.Equal(t, "8", q.Get("recipe"))
		assert.Equal(t, "1", q.Get("space"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).List(context.Background(), &ListOptions{
		FoodID: 3, RecipeID: 8, SpaceID: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestIngredientService_List_Error(t *testing.T) {
	_, err := NewService(notFound(t)).List(context.Background(), nil)
	require.Error(t, err)
}

func TestIngredientService_Get_NotFound(t *testing.T) {
	_, err := NewService(notFound(t)).Get(context.Background(), 1)
	require.Error(t, err)
	var me *testutil.MockError
	require.ErrorAs(t, err, &me)
	assert.True(t, me.IsNotFound())
}

func TestIngredientService_Create_Validation(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"food":["This field is required."]}`)
	_, err := NewService(exec).Create(context.Background(), &Ingredient{})
	require.Error(t, err)
}

func TestIngredientService_Update_Unauthorized(t *testing.T) {
	exec := mockStatus(t, http.StatusUnauthorized, `{"detail":"No"}`)
	_, err := NewService(exec).Update(context.Background(), &Ingredient{ID: 1})
	require.Error(t, err)
}

func TestIngredientService_Patch_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Patch(context.Background(), &Ingredient{ID: 99})
	require.Error(t, err)
}

func TestIngredientService_Delete_Error(t *testing.T) {
	require.Error(t, NewService(notFound(t)).Delete(context.Background(), 99))
}

func TestIngredientService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/ingredient/1/cascading/", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Cascading(
		context.Background(), 1, &pagination.ListOptions{Page: 1})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestIngredientService_Cascading_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Cascading(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestIngredientService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/ingredient/1/nulling/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Nulling(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestIngredientService_Nulling_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Nulling(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestIngredientService_Protecting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/ingredient/1/protecting/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Protecting(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestIngredientService_Protecting_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Protecting(context.Background(), 1, nil)
	require.Error(t, err)
}
