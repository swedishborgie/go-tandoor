package property

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

// TypeService tests

func TestTypeService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/property-type/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1,"name":"Calories","unit":"kcal"},{"id":2,"name":"Protein","unit":"g"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewTypeService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
	assert.Equal(t, "Calories", page.Results[0].Name)
}

func TestTypeService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/property-type/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Calories","unit":"kcal","order":0}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	pt, err := NewTypeService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "kcal", pt.Unit)
}

func TestTypeService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":5,"name":"Sodium","unit":"mg"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	pt, err := NewTypeService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "Sodium", pt.Name)
}

func TestTypeService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Energy"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	pt, err := NewTypeService(mock).Update(context.Background(), &Type{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Energy", pt.Name)
}

func TestTypeService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	pt, err := NewTypeService(mock).Patch(context.Background(), &Type{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, pt.ID)
}

func TestTypeService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewTypeService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestTypeService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/property-type/1/cascading/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Calories"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	pts, err := NewTypeService(mock).Cascading(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, pts.Count)
}

func TestTypeService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/property-type/1/nulling/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	pts, err := NewTypeService(mock).Nulling(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, pts.Count)
}

func TestTypeService_Protecting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/property-type/1/protecting/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Calories"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	pts, err := NewTypeService(mock).Protecting(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, pts.Count)
}

// PropertyService tests

func TestPropertyService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/property/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"property_amount":250}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.InEpsilon(t, 250.0, *page.Results[0].PropertyAmount, 1e-9)
}

func TestPropertyService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/property/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"property_type":{"id":1,"name":"Calories"}}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	p, err := NewService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Calories", p.Type.Name)
}

func TestPropertyService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"property_amount":12}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	p, err := NewService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.InEpsilon(t, 12.0, *p.PropertyAmount, 1e-9)
}

func TestPropertyService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	p, err := NewService(mock).Patch(context.Background(), &Property{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, p.ID)
}

func TestPropertyService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// ListOptions tests

func TestListOptions_QueryString(t *testing.T) {
	foodID := 5
	opts := &ListOptions{ListOptions: &pagination.ListOptions{Page: 1, PageSize: 20}, FoodID: &foodID}
	qs := opts.QueryString()
	assert.Contains(t, qs, "page=1")
	assert.Contains(t, qs, "page_size=20")
	assert.Contains(t, qs, "food_id=5")
}

func TestListOptions_Empty(t *testing.T) {
	assert.Empty(t, (&ListOptions{}).QueryString())
	assert.Empty(t, (*ListOptions)(nil).QueryString())
	foodID := 0
	opts := &ListOptions{FoodID: &foodID}
	assert.NotEmpty(t, opts.QueryString())
}
