package step

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

func TestStepService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/step/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1,"name":"Prep"},{"id":2,"name":"Cook"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
	assert.Equal(t, "Prep", page.Results[0].Name)
}

func TestStepService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/step/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Mix","instruction":"Combine ingredients"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	s, err := NewService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Combine ingredients", s.Instruction)
}

func TestStepService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"Bake","instruction":"350F for 20min"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	s, err := NewService(mock).Create(context.Background(), &Step{Name: "Bake"})
	require.NoError(t, err)
	assert.Equal(t, 10, s.ID)
}

func TestStepService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Updated Step"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	s, err := NewService(mock).Update(context.Background(), &Step{ID: 1, Name: "Updated Step"})
	require.NoError(t, err)
	assert.Equal(t, "Updated Step", s.Name)
}

func TestStepService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"instruction":"New instruction"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	s, err := NewService(mock).Patch(context.Background(), &Step{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "New instruction", s.Instruction)
}

func TestStepService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/step/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).Delete(context.Background(), 1)
	require.NoError(t, err)
}
