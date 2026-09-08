package keyword

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

func TestKeywordService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/keyword/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1,"name":"Dinner","label":"Dinner"},{"id":2,"name":"Lunch","label":"Lunch"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
	assert.Equal(t, "Dinner", page.Results[0].Name)
}

func TestKeywordService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/keyword/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Appetizer","label":"Appetizer","parent":0,"numchild":2}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	k, err := NewService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Appetizer", k.Name)
	assert.Equal(t, 2, k.NumChild)
}

func TestKeywordService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"Brunch","label":"Brunch"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	k, err := NewService(mock).Create(context.Background(), &Keyword{Name: "Brunch"})
	require.NoError(t, err)
	assert.Equal(t, 10, k.ID)
}

func TestKeywordService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Updated Keyword"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	k, err := NewService(mock).Update(context.Background(), &Keyword{ID: 1, Name: "Updated Keyword"})
	require.NoError(t, err)
	assert.Equal(t, "Updated Keyword", k.Name)
}

func TestKeywordService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"description":"New description"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	k, err := NewService(mock).Patch(context.Background(), &Keyword{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "New description", k.Description)
}

func TestKeywordService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/keyword/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestKeywordService_Merge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/keyword/1/merge/2/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":2,"name":"Merged Keyword"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	k, err := NewService(mock).Merge(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 2, k.ID)
}

func TestKeywordService_Move(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/keyword/1/move/10/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"parent":10,"full_name":"Category > Sub"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	k, err := NewService(mock).Move(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 10, k.Parent)
}
