package ingredient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

func TestIngredientService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/ingredient/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1,"amount":2.5},{"id":2,"amount":1.0}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
	assert.InEpsilon(t, 2.5, *page.Results[0].Amount, 1e-9)
}

func TestIngredientService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/ingredient/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"amount":2.0,"note":"chopped"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	i, err := NewService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "chopped", i.Note)
}

func TestIngredientService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"amount":0.5,"note":"minced"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	amt := 0.5
	i, err := NewService(mock).Create(context.Background(), &Ingredient{Amount: &amt})
	require.NoError(t, err)
	assert.Equal(t, 10, i.ID)
}

func TestIngredientService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"amount":3.0}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	amt := 3.0
	i, err := NewService(mock).Update(context.Background(), &Ingredient{ID: 1, Amount: &amt})
	require.NoError(t, err)
	assert.InEpsilon(t, 3.0, *i.Amount, 1e-9)
}

func TestIngredientService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"note":"diced"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	i, err := NewService(mock).Patch(context.Background(), &Ingredient{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "diced", i.Note)
}

func TestIngredientService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/ingredient/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).Delete(context.Background(), 1)
	require.NoError(t, err)
}
