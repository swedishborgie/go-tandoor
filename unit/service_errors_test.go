package unit

import (
	"context"
	"encoding/json"
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

func TestService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/unit/1/cascading/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).Cascading(context.Background(), 1, nil)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_Cascading_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Cascading(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/unit/1/nulling/", r.URL.Path)
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
		assert.Equal(t, "/api/unit/1/protecting/", r.URL.Path)
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

func TestConversionService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.Write([]byte(`{"id":1,"name":"Cup","base_amount":1,"converted_amount":2.5}`))
	}))
	t.Cleanup(server.Close)

	c, err := NewConversionService(testutil.NewMockExecutor(server.URL)).Patch(context.Background(), &Conversion{ID: 1})
	require.NoError(t, err)
	assert.InDelta(t, 2.5, c.ConvertedAmount, 1e-9)
}

func TestConversionService_Patch_Error(t *testing.T) {
	_, err := NewConversionService(notFound(t)).Patch(context.Background(), &Conversion{ID: 99})
	require.Error(t, err)
}

func TestRef_MarshalJSON_IDOnly(t *testing.T) {
	b, err := json.Marshal(&Ref{ID: 3})
	require.NoError(t, err)
	assert.JSONEq(t, `3`, string(b))
}

func TestRef_MarshalJSON_Full(t *testing.T) {
	b, err := json.Marshal(&Ref{ID: 3, Name: "g"})
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "g", m["name"])
}

func TestRef_UnmarshalJSON_BareID(t *testing.T) {
	var r Ref
	require.NoError(t, json.Unmarshal([]byte(`3`), &r))
	assert.Equal(t, 3, r.ID)
}

func TestFoodRef_MarshalJSON_IDOnly(t *testing.T) {
	b, err := json.Marshal(&FoodRef{ID: 5})
	require.NoError(t, err)
	assert.JSONEq(t, `5`, string(b))
}

func TestFoodRef_MarshalJSON_Full(t *testing.T) {
	b, err := json.Marshal(&FoodRef{ID: 5, Name: "Pasta"})
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "Pasta", m["name"])
}

func TestFoodRef_UnmarshalJSON_BareID(t *testing.T) {
	var f FoodRef
	require.NoError(t, json.Unmarshal([]byte(`5`), &f))
	assert.Equal(t, 5, f.ID)
}

func TestFoodRef_UnmarshalJSON_Object(t *testing.T) {
	var f FoodRef
	require.NoError(t, json.Unmarshal([]byte(`{"id":5,"name":"Pasta"}`), &f))
	assert.Equal(t, 5, f.ID)
	assert.Equal(t, "Pasta", f.Name)
}

func TestListOptions_Values(t *testing.T) {
	v := ListOptions{ListOptions: pagination.ListOptions{Page: 2}}.Values()
	assert.Equal(t, "2", v.Get("page"))
}

func TestListOptions_ToPaginationOptions(t *testing.T) {
	p := ListOptions{ListOptions: pagination.ListOptions{Page: 3, PageSize: 50}}.ToPaginationOptions()
	assert.Equal(t, 3, p.Page)
	assert.Equal(t, 50, p.PageSize)
}
