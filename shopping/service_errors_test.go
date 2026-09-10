package shopping

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
	return mockStatus(t, http.StatusNotFound, `{"detail":"Not found."}`)
}

func refPage() string {
	return `{"count":1,"results":[{"id":2,"model":"recipe","name":"Pasta"}]}`
}

// --- ListService ---

func TestListService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "groceries", r.URL.Query().Get("query"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewListService(testutil.NewMockExecutor(server.URL)).List(context.Background(),
		&ListListOptions{ListOptions: pagination.ListOptions{Search: "groceries"}})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestListService_List_Error(t *testing.T) {
	_, err := NewListService(notFound(t)).List(context.Background(), nil)
	require.Error(t, err)
}

func TestListService_Get_NotFound(t *testing.T) {
	_, err := NewListService(notFound(t)).Get(context.Background(), 1)
	require.Error(t, err)
	var me *testutil.MockError
	require.ErrorAs(t, err, &me)
	assert.True(t, me.IsNotFound())
}

func TestListService_Create_Validation(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"name":["This field is required."]}`)
	_, err := NewListService(exec).Create(context.Background(), &List{})
	require.Error(t, err)
}

func TestListService_Update_Error(t *testing.T) {
	_, err := NewListService(notFound(t)).Update(context.Background(), &List{ID: 99})
	require.Error(t, err)
}

func TestListService_Patch_Error(t *testing.T) {
	_, err := NewListService(notFound(t)).Patch(context.Background(), &List{ID: 99})
	require.Error(t, err)
}

func TestListService_Delete_Error(t *testing.T) {
	require.Error(t, NewListService(notFound(t)).Delete(context.Background(), 99))
}

func TestListService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/shopping-list/1/cascading/", r.URL.Path)
		w.Write([]byte(refPage()))
	}))
	t.Cleanup(server.Close)

	page, err := NewListService(testutil.NewMockExecutor(server.URL)).Cascading(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, page.Results, 1)
	assert.Equal(t, "recipe", page.Results[0].Model)
}

func TestListService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/shopping-list/1/nulling/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewListService(testutil.NewMockExecutor(server.URL)).Nulling(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestListService_Protecting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/shopping-list/1/protecting/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewListService(testutil.NewMockExecutor(server.URL)).Protecting(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestListService_Cascading_Error(t *testing.T) {
	_, err := NewListService(notFound(t)).Cascading(context.Background(), 1)
	require.Error(t, err)
}

func TestListService_Nulling_Error(t *testing.T) {
	_, err := NewListService(notFound(t)).Nulling(context.Background(), 1)
	require.Error(t, err)
}

func TestListService_Protecting_Error(t *testing.T) {
	_, err := NewListService(notFound(t)).Protecting(context.Background(), 1)
	require.Error(t, err)
}

// --- EntryService ---

func TestListEntryService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewEntryService(testutil.NewMockExecutor(server.URL)).List(context.Background(),
		&ListEntryListOptions{ListOptions: pagination.ListOptions{Page: 2}})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestListEntryService_List_Error(t *testing.T) {
	_, err := NewEntryService(notFound(t)).List(context.Background(), nil)
	require.Error(t, err)
}

func TestListEntryService_Get_NotFound(t *testing.T) {
	_, err := NewEntryService(notFound(t)).Get(context.Background(), 1)
	require.Error(t, err)
}

func TestListEntryService_Create_Validation(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"list":["This field is required."]}`)
	_, err := NewEntryService(exec).Create(context.Background(), &ListEntry{})
	require.Error(t, err)
}

func TestListEntryService_Update_Error(t *testing.T) {
	_, err := NewEntryService(notFound(t)).Update(context.Background(), &ListEntry{ID: 99})
	require.Error(t, err)
}

func TestListEntryService_Patch_Error(t *testing.T) {
	_, err := NewEntryService(notFound(t)).Patch(context.Background(), &ListEntry{ID: 99})
	require.Error(t, err)
}

func TestListEntryService_Delete_Error(t *testing.T) {
	require.Error(t, NewEntryService(notFound(t)).Delete(context.Background(), 99))
}

func TestListEntryService_BulkUpdate_Error(t *testing.T) {
	_, err := NewEntryService(notFound(t)).BulkUpdate(context.Background(), &ListEntryBulk{})
	require.Error(t, err)
}

// --- RecipeService ---

func TestListRecipeService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "pasta", r.URL.Query().Get("query"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewRecipeService(testutil.NewMockExecutor(server.URL)).List(context.Background(),
		&ListRecipeListOptions{ListOptions: pagination.ListOptions{Search: "pasta"}})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestListRecipeService_List_Error(t *testing.T) {
	_, err := NewRecipeService(notFound(t)).List(context.Background(), nil)
	require.Error(t, err)
}

func TestListRecipeService_Get_NotFound(t *testing.T) {
	_, err := NewRecipeService(notFound(t)).Get(context.Background(), 1)
	require.Error(t, err)
}

func TestListRecipeService_Create_Validation(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"recipe":["This field is required."]}`)
	_, err := NewRecipeService(exec).Create(context.Background(), &ListRecipe{})
	require.Error(t, err)
}

func TestListRecipeService_Update_Error(t *testing.T) {
	_, err := NewRecipeService(notFound(t)).Update(context.Background(), &ListRecipe{ID: 99})
	require.Error(t, err)
}

func TestListRecipeService_Patch_Error(t *testing.T) {
	_, err := NewRecipeService(notFound(t)).Patch(context.Background(), &ListRecipe{ID: 99})
	require.Error(t, err)
}

func TestListRecipeService_Delete_Error(t *testing.T) {
	require.Error(t, NewRecipeService(notFound(t)).Delete(context.Background(), 99))
}

func TestListRecipeService_BulkCreateEntries_Error(t *testing.T) {
	_, err := NewRecipeService(notFound(t)).BulkCreateEntries(context.Background(), 1, &ListEntryBulkCreate{})
	require.Error(t, err)
}
