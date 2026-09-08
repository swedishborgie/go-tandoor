package supermarket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

func TestSupermarketService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/supermarket/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Test Market","description":"A test store"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "Test Market", page.Results[0].Name)
}

func TestSupermarketService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/supermarket/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Test Market","description":"A test store"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sm, err := NewService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Test Market", sm.Name)
}

func TestSupermarketService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"New Market","description":""}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sm, err := NewService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "New Market", sm.Name)
}

func TestSupermarketService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Updated Market"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sm, err := NewService(mock).Update(context.Background(), &Supermarket{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Updated Market", sm.Name)
}

func TestSupermarketService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sm, err := NewService(mock).Patch(context.Background(), &Supermarket{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, sm.ID)
}

func TestSupermarketService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/supermarket/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestCategoryService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/supermarket-category/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Dairy"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewCategoryService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "Dairy", page.Results[0].Name)
}

func TestCategoryService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/supermarket-category/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Produce","description":"Fresh produce"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cat, err := NewCategoryService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Produce", cat.Name)
}

func TestCategoryService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"Frozen Foods"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cat, err := NewCategoryService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "Frozen Foods", cat.Name)
}

func TestCategoryService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Updated Category"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cat, err := NewCategoryService(mock).Update(context.Background(), &Category{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Updated Category", cat.Name)
}

func TestCategoryService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cat, err := NewCategoryService(mock).Patch(context.Background(), &Category{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, cat.ID)
}

func TestCategoryService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/supermarket-category/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewCategoryService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestCategoryService_Merge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/supermarket-category/1/merge/", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "other=2")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Merged Category"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cat, err := NewCategoryService(mock).Merge(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, "Merged Category", cat.Name)
}

func TestCategoryRelationService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/supermarket-category-relation/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"supermarket":1,"order":0}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewCategoryRelationService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
}

func TestCategoryRelationService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/supermarket-category-relation/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"supermarket":2,"order":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	rel, err := NewCategoryRelationService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1, rel.Order)
}

func TestCategoryRelationService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"supermarket":3,"order":0}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	rel, err := NewCategoryRelationService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 10, rel.ID)
}

func TestCategoryRelationService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"order":5}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	rel, err := NewCategoryRelationService(mock).Update(context.Background(), &CategoryRelation{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 5, rel.Order)
}

func TestCategoryRelationService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	rel, err := NewCategoryRelationService(mock).Patch(context.Background(), &CategoryRelation{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, rel.ID)
}

func TestCategoryRelationService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/supermarket-category-relation/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewCategoryRelationService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}
