package keyword

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

func TestKeywordService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "4", r.URL.Query().Get("recipe_book"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).List(context.Background(), &ListOptions{RecipeBookID: 4})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestKeywordService_List_Error(t *testing.T) {
	_, err := NewService(notFound(t)).List(context.Background(), nil)
	require.Error(t, err)
}

func TestKeywordService_ListAll_MultiPage(t *testing.T) {
	var sawPage2 bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			sawPage2 = true
			w.Write([]byte(`{"count":2,"results":[{"id":2,"name":"Dinner"}]}`))
			return
		}
		next := r.URL.Scheme + "://" + r.Host + "/api/keyword/?page=2"
		w.Write([]byte(`{"count":2,"next":"` + next + `","results":[{"id":1,"name":"Lunch"}]}`))
	}))
	t.Cleanup(server.Close)

	keywords, err := NewService(testutil.NewMockExecutor(server.URL)).ListAll(context.Background(), &ListOptions{})
	require.NoError(t, err)
	require.Len(t, keywords, 2)
	assert.True(t, sawPage2)
}

func TestKeywordService_ListAll_Error(t *testing.T) {
	_, err := NewService(notFound(t)).ListAll(context.Background(), nil)
	require.Error(t, err)
}

func TestKeywordService_Get_NotFound(t *testing.T) {
	_, err := NewService(notFound(t)).Get(context.Background(), 1)
	require.Error(t, err)
	var me *testutil.MockError
	require.ErrorAs(t, err, &me)
	assert.True(t, me.IsNotFound())
}

func TestKeywordService_Create_Validation(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"name":["This field is required."]}`)
	_, err := NewService(exec).Create(context.Background(), &Keyword{})
	require.Error(t, err)
}

func TestKeywordService_Update_Unauthorized(t *testing.T) {
	exec := mockStatus(t, http.StatusUnauthorized, `{"detail":"No"}`)
	_, err := NewService(exec).Update(context.Background(), &Keyword{ID: 1})
	require.Error(t, err)
}

func TestKeywordService_Patch_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Patch(context.Background(), &Keyword{ID: 99})
	require.Error(t, err)
}

func TestKeywordService_Delete_Error(t *testing.T) {
	require.Error(t, NewService(notFound(t)).Delete(context.Background(), 99))
}

func TestKeywordService_Merge_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Merge(context.Background(), 1, 2)
	require.Error(t, err)
}

func TestKeywordService_Move_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Move(context.Background(), 1, 2)
	require.Error(t, err)
}

func TestKeywordService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/keyword/1/cascading/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Cascading(
		context.Background(), 1, &pagination.ListOptions{Page: 1})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestKeywordService_Cascading_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Cascading(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestKeywordService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/keyword/1/nulling/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Nulling(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestKeywordService_Nulling_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Nulling(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestKeywordService_Protecting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/keyword/1/protecting/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Protecting(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestKeywordService_Protecting_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Protecting(context.Background(), 1, nil)
	require.Error(t, err)
}
