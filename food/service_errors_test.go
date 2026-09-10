package food

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

// mockStatus starts an httptest server that responds to every request with
// the given status and body, and returns a MockExecutor pointed at it.
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

func TestService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "flour", q.Get("query"))
		assert.Equal(t, "2", q.Get("page"))
		assert.Equal(t, "30", q.Get("page_size"))
		assert.Equal(t, "-name", q.Get("ordering"))
		assert.Equal(t, "5", q.Get("category"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	svc := NewService(testutil.NewMockExecutor(server.URL))
	page, err := svc.List(context.Background(), &ListOptions{
		ListOptions: pagination.ListOptions{Page: 2, PageSize: 30, Search: "flour", OrderBy: "-name"},
		CategoryID:  5,
	})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_List_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.List(context.Background(), nil)
	require.Error(t, err)
	var me *testutil.MockError
	require.ErrorAs(t, err, &me)
	assert.True(t, me.IsNotFound())
}

func TestService_ListAll_MultiPage(t *testing.T) {
	var page2 string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			page2 = r.URL.Query().Get("page_size")
			w.Write([]byte(`{"count":3,"results":[{"id":2,"name":"Sugar"}]}`))
			return
		}
		next := r.URL.Scheme + "://" + r.Host + "/api/food/?page=2"
		w.Write([]byte(`{"count":3,"next":"` + next + `","results":[{"id":1,"name":"Flour"}]}`))
	}))
	t.Cleanup(server.Close)

	svc := NewService(testutil.NewMockExecutor(server.URL))
	foods, err := svc.ListAll(context.Background(), &ListOptions{
		ListOptions: pagination.ListOptions{PageSize: 1},
	})
	require.NoError(t, err)
	require.Len(t, foods, 2)
	assert.Equal(t, "Flour", foods[0].Name)
	assert.Equal(t, "Sugar", foods[1].Name)
	assert.Equal(t, "1", page2) // page_size carried onto follow-up pages
}

func TestService_ListAll_DefaultsAndEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		// opts==nil path: defaults Page=1, PageSize=100 must be sent.
		assert.Equal(t, "1", q.Get("page"))
		assert.Equal(t, "100", q.Get("page_size"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	svc := NewService(testutil.NewMockExecutor(server.URL))
	foods, err := svc.ListAll(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, foods)
}

func TestService_ListAll_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.ListAll(context.Background(), nil)
	require.Error(t, err)
}

func TestService_Get_NotFound(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.Get(context.Background(), 1)
	require.Error(t, err)
	var me *testutil.MockError
	require.ErrorAs(t, err, &me)
	assert.True(t, me.IsNotFound())
}

func TestService_Create_Validation(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"name":["This field is required."]}`)
	_, err := NewService(exec).Create(context.Background(), &Food{})
	require.Error(t, err)
	var me *testutil.MockError
	require.ErrorAs(t, err, &me)
	assert.Equal(t, http.StatusBadRequest, me.StatusCode)
}

func TestService_Update_Unauthorized(t *testing.T) {
	exec := mockStatus(t, http.StatusUnauthorized, `{"detail":"Authentication credentials were not provided."}`)
	_, err := NewService(exec).Update(context.Background(), &Food{ID: 1})
	require.Error(t, err)
	var me *testutil.MockError
	require.ErrorAs(t, err, &me)
	assert.True(t, me.IsUnauthorized())
}

func TestService_Patch_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.Patch(context.Background(), &Food{ID: 99})
	require.Error(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	svc := NewService(notFound(t))
	require.Error(t, svc.Delete(context.Background(), 99))
}

func TestService_Merge_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.Merge(context.Background(), 1, 2)
	require.Error(t, err)
}

func TestService_Move_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.Move(context.Background(), 1, 2)
	require.Error(t, err)
}

func TestService_BatchUpdate_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.BatchUpdate(context.Background(), &BatchUpdate{})
	require.Error(t, err)
}

func TestService_UpdateShopping_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.UpdateShopping(context.Background(), 1, &ShoppingUpdate{})
	require.Error(t, err)
}

func TestService_FdcImport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/food/7/fdc/", r.URL.Path)
		w.Write([]byte(`{"id":7,"name":"Rice"}`))
	}))
	t.Cleanup(server.Close)

	f, err := NewService(testutil.NewMockExecutor(server.URL)).FdcImport(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, "Rice", f.Name)
}

func TestService_FdcImport_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.FdcImport(context.Background(), 7)
	require.Error(t, err)
}

func TestService_AiProperties(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/food/7/aiproperties/", r.URL.Path)
		w.Write([]byte(`{"id":7,"name":"Rice"}`))
	}))
	t.Cleanup(server.Close)

	f, err := NewService(testutil.NewMockExecutor(server.URL)).AiProperties(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, "Rice", f.Name)
}

func TestService_AiProperties_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.AiProperties(context.Background(), 7)
	require.Error(t, err)
}

func TestService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/food/1/cascading/", r.URL.Path)
		w.Write([]byte(`{"count":1,"results":[{"id":9,"name":"Flour (derived)"}]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Cascading(context.Background(), 1, nil)
	require.NoError(t, err)
	require.Len(t, page.Results, 1)
}

func TestService_Cascading_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Cascading(
		context.Background(), 1, &pagination.ListOptions{Page: 1})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_Cascading_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.Cascading(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/food/1/nulling/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Nulling(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_Nulling_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.Nulling(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestService_Protecting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/food/1/protecting/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Protecting(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_Protecting_Error(t *testing.T) {
	svc := NewService(notFound(t))
	_, err := svc.Protecting(context.Background(), 1, nil)
	require.Error(t, err)
}
