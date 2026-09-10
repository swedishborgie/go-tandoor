package step

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

func TestStepService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "5", r.URL.Query().Get("recipe"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).List(context.Background(), &ListOptions{RecipeID: 5})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestStepService_List_Error(t *testing.T) {
	_, err := NewService(notFound(t)).List(context.Background(), nil)
	require.Error(t, err)
}

func TestStepService_Get_NotFound(t *testing.T) {
	_, err := NewService(notFound(t)).Get(context.Background(), 1)
	require.Error(t, err)
	var me *testutil.MockError
	require.ErrorAs(t, err, &me)
	assert.True(t, me.IsNotFound())
}

func TestStepService_Create_Validation(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"step":["This field may not be blank."]}`)
	_, err := NewService(exec).Create(context.Background(), &Step{})
	require.Error(t, err)
}

func TestStepService_Update_Unauthorized(t *testing.T) {
	exec := mockStatus(t, http.StatusUnauthorized, `{"detail":"No"}`)
	_, err := NewService(exec).Update(context.Background(), &Step{ID: 1})
	require.Error(t, err)
}

func TestStepService_Patch_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Patch(context.Background(), &Step{ID: 99})
	require.Error(t, err)
}

func TestStepService_Delete_Error(t *testing.T) {
	require.Error(t, NewService(notFound(t)).Delete(context.Background(), 99))
}

func TestStepService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/step/1/cascading/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Cascading(
		context.Background(), 1, &pagination.ListOptions{Page: 1})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestStepService_Cascading_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Cascading(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestStepService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/step/1/nulling/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Nulling(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestStepService_Nulling_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Nulling(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestStepService_Protecting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/step/1/protecting/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Protecting(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestStepService_Protecting_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Protecting(context.Background(), 1, nil)
	require.Error(t, err)
}
