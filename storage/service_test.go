package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

// StorageService tests

func TestStorageService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/storage/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Local","method":"local"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "Local", page.Results[0].Name)
}

func TestStorageService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/storage/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"S3 Bucket","method":"s3","url":"https://s3.example.com","path":"recipes"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	st, err := NewService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "s3", st.Method)
	assert.Equal(t, "recipes", st.Path)
}

func TestStorageService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"New Storage","created_by":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	st, err := NewService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "New Storage", st.Name)
}

func TestStorageService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Updated Storage"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	st, err := NewService(mock).Update(context.Background(), &Storage{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Updated Storage", st.Name)
}

func TestStorageService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	st, err := NewService(mock).Patch(context.Background(), &Storage{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, st.ID)
}

func TestStorageService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestStorageService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/storage/1/cascading/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Local","method":"local"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sts, err := NewService(mock).Cascading(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, sts.Count)
}

func TestStorageService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/storage/1/nulling/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sts, err := NewService(mock).Nulling(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, sts.Count)
}

func TestStorageService_Protecting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/storage/1/protecting/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Storage","method":"s3"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sts, err := NewService(mock).Protecting(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, sts.Count)
}
