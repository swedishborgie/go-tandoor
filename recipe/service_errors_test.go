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

func TestService_ListAll_MultiPage(t *testing.T) {
	var sawPage2 bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			sawPage2 = true
			w.Write([]byte(`{"count":2,"results":[{"id":2,"name":"Soup"}]}`))
			return
		}
		next := r.URL.Scheme + "://" + r.Host + "/api/recipe/?page=2"
		w.Write([]byte(`{"count":2,"next":"` + next + `","results":[{"id":1,"name":"Pasta"}]}`))
	}))
	t.Cleanup(server.Close)

	recipes, err := NewService(testutil.NewMockExecutor(server.URL)).ListAll(context.Background(), &ListOptions{})
	require.NoError(t, err)
	require.Len(t, recipes, 2)
	assert.True(t, sawPage2)
}

func TestService_ListAll_Error(t *testing.T) {
	_, err := NewService(notFound(t)).ListAll(context.Background(), nil)
	require.Error(t, err)
}

func TestService_List_Error(t *testing.T) {
	_, err := NewService(notFound(t)).List(context.Background(), nil)
	require.Error(t, err)
}

func TestService_Flat_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Flat(context.Background(), nil)
	require.Error(t, err)
}

func TestService_Create_Validation(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"name":["This field is required."]}`)
	_, err := NewService(exec).Create(context.Background(), &Recipe{})
	require.Error(t, err)
}

func TestService_Update_Unauthorized(t *testing.T) {
	exec := mockStatus(t, http.StatusUnauthorized, `{"detail":"No"}`)
	_, err := NewService(exec).Update(context.Background(), &Recipe{ID: 1})
	require.Error(t, err)
}

func TestService_Patch_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Patch(context.Background(), &Recipe{ID: 99})
	require.Error(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	require.Error(t, NewService(notFound(t)).Delete(context.Background(), 99))
}

func TestService_Related_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/recipe/1/related/", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Related(
		context.Background(), 1, &ListOptions{ListOptions: pagination.ListOptions{Page: 1}})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_Related_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Related(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestService_BatchUpdate_Error(t *testing.T) {
	_, err := NewService(notFound(t)).BatchUpdate(context.Background(), &BatchUpdate{Recipes: []int{1}})
	require.Error(t, err)
}

func TestService_UploadImage_Error(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"image":["No file was provided."]}`)
	_, err := NewService(exec).UploadImage(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/recipe/1/cascading/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Cascading(
		context.Background(), 1, &ListOptions{ListOptions: pagination.ListOptions{Page: 1}})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_Cascading_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Cascading(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/recipe/1/nulling/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Nulling(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_Nulling_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Nulling(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestService_Protecting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/recipe/1/protecting/", r.URL.Path)
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

func TestService_AiProperties(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/recipe/3/aiproperties/", r.URL.Path)
		w.Write([]byte(`{"id":3,"name":"Pasta"}`))
	}))
	t.Cleanup(server.Close)

	r, err := NewService(testutil.NewMockExecutor(server.URL)).AiProperties(context.Background(), 3)
	require.NoError(t, err)
	assert.Equal(t, "Pasta", r.Name)
}

func TestService_AiProperties_Error(t *testing.T) {
	_, err := NewService(notFound(t)).AiProperties(context.Background(), 3)
	require.Error(t, err)
}

func TestService_DeleteExternal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/api/recipe/3/delete_external/", r.URL.Path)
		w.Write([]byte(`{"id":3,"name":"Pasta"}`))
	}))
	t.Cleanup(server.Close)

	r, err := NewService(testutil.NewMockExecutor(server.URL)).DeleteExternal(context.Background(), 3)
	require.NoError(t, err)
	assert.Equal(t, 3, r.ID)
}

func TestService_DeleteExternal_Error(t *testing.T) {
	_, err := NewService(notFound(t)).DeleteExternal(context.Background(), 3)
	require.Error(t, err)
}
