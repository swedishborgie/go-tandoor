package property

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
	return mockStatus(t, http.StatusNotFound, `{"detail":"Not found."}`)
}

func TestService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.Write([]byte(`{"id":1,"property_type":{"id":1,"name":"Calories"}}`))
	}))
	t.Cleanup(server.Close)

	p, err := NewService(testutil.NewMockExecutor(server.URL)).Update(context.Background(), &Property{ID: 1, Type: Type{ID: 1}})
	require.NoError(t, err)
	assert.Equal(t, "Calories", p.Type.Name)
}

func TestService_Update_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Update(context.Background(), &Property{ID: 99})
	require.Error(t, err)
}

func TestService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/property/1/cascading/", r.URL.Path)
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
		assert.Equal(t, "/api/property/1/nulling/", r.URL.Path)
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
		assert.Equal(t, "/api/property/1/protecting/", r.URL.Path)
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

func TestFoodRef_MarshalJSON_IDOnly(t *testing.T) {
	b, err := json.Marshal(&FoodRef{ID: 7})
	require.NoError(t, err)
	assert.JSONEq(t, `7`, string(b))
}

func TestFoodRef_UnmarshalJSON_BareID(t *testing.T) {
	var f FoodRef
	require.NoError(t, json.Unmarshal([]byte(`7`), &f))
	assert.Equal(t, 7, f.ID)
}

func TestFoodRef_UnmarshalJSON_Object(t *testing.T) {
	var f FoodRef
	require.NoError(t, json.Unmarshal([]byte(`{"id":7}`), &f))
	assert.Equal(t, 7, f.ID)
}

func TestTypeListOptions_Values(t *testing.T) {
	v := TypeListOptions{ListOptions: pagination.ListOptions{Page: 2}}.Values()
	assert.Equal(t, "2", v.Get("page"))
}
