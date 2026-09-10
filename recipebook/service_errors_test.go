package recipebook

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

// --- Service (recipe books) ---

func TestService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "1", r.URL.Query().Get("space"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).List(context.Background(), &ListOptions{SpaceID: 1})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_List_Error(t *testing.T) {
	_, err := NewService(notFound(t)).List(context.Background(), nil)
	require.Error(t, err)
}

func TestService_Get_NotFound(t *testing.T) {
	_, err := NewService(notFound(t)).Get(context.Background(), 1)
	require.Error(t, err)
	var me *testutil.MockError
	require.ErrorAs(t, err, &me)
	assert.True(t, me.IsNotFound())
}

func TestService_Create_Validation(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"name":["This field is required."]}`)
	_, err := NewService(exec).Create(context.Background(), &RecipeBook{})
	require.Error(t, err)
}

func TestService_Update_Unauthorized(t *testing.T) {
	exec := mockStatus(t, http.StatusUnauthorized, `{"detail":"No"}`)
	_, err := NewService(exec).Update(context.Background(), &RecipeBook{ID: 1})
	require.Error(t, err)
}

func TestService_Patch_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Patch(context.Background(), &RecipeBook{ID: 99})
	require.Error(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	require.Error(t, NewService(notFound(t)).Delete(context.Background(), 99))
}

func TestService_Cascading_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Cascading(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/recipe-book/1/nulling/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Nulling(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_Nulling_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Nulling(
		context.Background(), 1, &pagination.ListOptions{Page: 1})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_Nulling_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Nulling(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestService_Protecting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/recipe-book/1/protecting/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Protecting(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_Protecting_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Protecting(context.Background(), 1, nil)
	require.Error(t, err)
}

// --- EntryService ---

func TestEntryService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "3", r.URL.Query().Get("page"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewEntryService(testutil.NewMockExecutor(server.URL)).List(
		context.Background(), &pagination.ListOptions{Page: 3})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestEntryService_List_Error(t *testing.T) {
	_, err := NewEntryService(notFound(t)).List(context.Background(), nil)
	require.Error(t, err)
}

func TestEntryService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/recipe-book-entry/9/", r.URL.Path)
		w.Write([]byte(`{"id":9,"book":1,"recipe":2}`))
	}))
	t.Cleanup(server.Close)

	e, err := NewEntryService(testutil.NewMockExecutor(server.URL)).Get(context.Background(), 9)
	require.NoError(t, err)
	assert.Equal(t, 9, e.ID)
}

func TestEntryService_Get_NotFound(t *testing.T) {
	_, err := NewEntryService(notFound(t)).Get(context.Background(), 99)
	require.Error(t, err)
}

func TestEntryService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.Write([]byte(`{"id":9,"book":3,"recipe":2}`))
	}))
	t.Cleanup(server.Close)

	e, err := NewEntryService(testutil.NewMockExecutor(server.URL)).Update(context.Background(), &Entry{ID: 9, Book: 3})
	require.NoError(t, err)
	assert.Equal(t, 3, e.Book)
}

func TestEntryService_Update_Error(t *testing.T) {
	_, err := NewEntryService(notFound(t)).Update(context.Background(), &Entry{ID: 99})
	require.Error(t, err)
}

func TestEntryService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.Write([]byte(`{"id":9,"book":4,"recipe":2}`))
	}))
	t.Cleanup(server.Close)

	e, err := NewEntryService(testutil.NewMockExecutor(server.URL)).Patch(context.Background(), &Entry{ID: 9})
	require.NoError(t, err)
	assert.Equal(t, 4, e.Book)
}

func TestEntryService_Patch_Error(t *testing.T) {
	_, err := NewEntryService(notFound(t)).Patch(context.Background(), &Entry{ID: 99})
	require.Error(t, err)
}

func TestEntryService_Delete_Error(t *testing.T) {
	require.Error(t, NewEntryService(notFound(t)).Delete(context.Background(), 99))
}
