package mealplan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

func TestMealPlanService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/meal-plan/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"from_date":"2024-01-01T00:00:00Z"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, 1, page.Results[0].ID)
}

func TestMealPlanService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/meal-plan/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"from_date":"2024-01-01T00:00:00Z","servings":4}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	mp, err := NewService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1, mp.ID)
	assert.InEpsilon(t, float64(4), mp.Servings, 1e-9)
}

func TestMealPlanService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":5,"from_date":"2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	mp, err := NewService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 5, mp.ID)
}

func TestMealPlanService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/meal-plan/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"servings":6}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	mp, err := NewService(mock).Update(context.Background(), &MealPlan{ID: 1})
	require.NoError(t, err)
	assert.InEpsilon(t, float64(6), mp.Servings, 1e-9)
}

func TestMealPlanService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	mp, err := NewService(mock).Patch(context.Background(), &MealPlan{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, mp.ID)
}

func TestMealPlanService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/meal-plan/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestMealPlanService_ICAL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/meal-plan/ical/", r.URL.Path)
		w.Header().Set("Content-Type", "text/calendar")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("BEGIN:VCALENDAR\r\nEND:VCALENDAR"))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cal, err := NewService(mock).ICAL(context.Background(), nil)
	require.NoError(t, err)
	assert.Contains(t, cal, "VCALENDAR")
}

func TestMealTypeService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/meal-type/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":3,"results":[{"id":1,"name":"breakfast"},{"id":2,"name":"dinner"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewMealTypeService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 3, page.Count)
	assert.Equal(t, "breakfast", page.Results[0].Name)
}

func TestMealTypeService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"snack","order":3}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	mt, err := NewMealTypeService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "snack", mt.Name)
}

func TestMealTypeService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/meal-type/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewMealTypeService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestAutoPlanService_Plan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/auto-meal-plan/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"start_date":"2024-01-01T00:00:00Z","end_date":"2024-01-07T00:00:00Z","meal_type_id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	result, err := NewAutoPlanService(mock).Plan(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, result.MealTypeID)
}
