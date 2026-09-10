package mealplan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

// --- Service (meal plans) ---

func TestService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "1", q.Get("space"))
		assert.Equal(t, "2", q.Get("user"))
		assert.Equal(t, "2025-01-01", q.Get("date_from"))
		assert.Equal(t, "2025-01-31", q.Get("date_to"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewService(testutil.NewMockExecutor(server.URL)).List(context.Background(), &ListOptions{
		SpaceID: 1, UserID: 2, DateFrom: "2025-01-01", DateTo: "2025-01-31",
	})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestService_List_Error(t *testing.T) {
	_, err := NewService(notFound(t)).List(context.Background(), nil)
	require.Error(t, err)
}

func TestService_Get_NotFound(t *testing.T) {
	_, err := NewService(notFound(t)).Get(context.Background(), 1)
	require.Error(t, err)
	var me *testutil.MockError
	require.ErrorAs(t, err, &me)
	assert.True(t, me.IsNotFound())
}

func TestService_Create_Validation(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"recipe":["This field is required."]}`)
	_, err := NewService(exec).Create(context.Background(), &MealPlan{})
	require.Error(t, err)
}

func TestService_Update_Unauthorized(t *testing.T) {
	exec := mockStatus(t, http.StatusUnauthorized, `{"detail":"No"}`)
	_, err := NewService(exec).Update(context.Background(), &MealPlan{ID: 1})
	require.Error(t, err)
}

func TestService_Patch_Error(t *testing.T) {
	_, err := NewService(notFound(t)).Patch(context.Background(), &MealPlan{ID: 99})
	require.Error(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	require.Error(t, NewService(notFound(t)).Delete(context.Background(), 99))
}

func TestService_ICAL_WithOpts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/meal-plan/ical/", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("space"))
		w.Write([]byte("BEGIN:VCALENDAR\nEND:VCALENDAR"))
	}))
	t.Cleanup(server.Close)

	out, err := NewService(testutil.NewMockExecutor(server.URL)).ICAL(
		context.Background(), &pagination.ListOptions{Extra: map[string]string{"space": "1"}})
	require.NoError(t, err)
	assert.Contains(t, out, "VCALENDAR")
}

func TestService_ICAL_Error(t *testing.T) {
	exec := mockStatus(t, http.StatusInternalServerError, "boom")
	_, err := NewService(exec).ICAL(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestService_ICAL_ErrorLongBody(t *testing.T) {
	long := strings.Repeat("x", 500)
	exec := mockStatus(t, http.StatusServiceUnavailable, long)
	_, err := NewService(exec).ICAL(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "...")
}

// --- MealTypeService ---

func TestMealTypeService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewMealTypeService(testutil.NewMockExecutor(server.URL)).List(
		context.Background(), &pagination.ListOptions{Page: 2})
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestMealTypeService_List_Error(t *testing.T) {
	_, err := NewMealTypeService(notFound(t)).List(context.Background(), nil)
	require.Error(t, err)
}

func TestMealTypeService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/meal-type/2/", r.URL.Path)
		w.Write([]byte(`{"id":2,"name":"Lunch"}`))
	}))
	t.Cleanup(server.Close)

	mt, err := NewMealTypeService(testutil.NewMockExecutor(server.URL)).Get(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, "Lunch", mt.Name)
}

func TestMealTypeService_Get_Error(t *testing.T) {
	_, err := NewMealTypeService(notFound(t)).Get(context.Background(), 99)
	require.Error(t, err)
}

func TestMealTypeService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.Write([]byte(`{"id":2,"name":"Dinner"}`))
	}))
	t.Cleanup(server.Close)

	mt, err := NewMealTypeService(testutil.NewMockExecutor(server.URL)).Update(context.Background(), &MealType{ID: 2, Name: "Dinner"})
	require.NoError(t, err)
	assert.Equal(t, "Dinner", mt.Name)
}

func TestMealTypeService_Update_Error(t *testing.T) {
	_, err := NewMealTypeService(notFound(t)).Update(context.Background(), &MealType{ID: 99})
	require.Error(t, err)
}

func TestMealTypeService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.Write([]byte(`{"id":2,"name":"Snack"}`))
	}))
	t.Cleanup(server.Close)

	mt, err := NewMealTypeService(testutil.NewMockExecutor(server.URL)).Patch(context.Background(), &MealType{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, "Snack", mt.Name)
}

func TestMealTypeService_Patch_Error(t *testing.T) {
	_, err := NewMealTypeService(notFound(t)).Patch(context.Background(), &MealType{ID: 99})
	require.Error(t, err)
}

func TestMealTypeService_Delete_Error(t *testing.T) {
	require.Error(t, NewMealTypeService(notFound(t)).Delete(context.Background(), 99))
}

func TestMealTypeService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/meal-type/2/cascading/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewMealTypeService(testutil.NewMockExecutor(server.URL)).Cascading(context.Background(), 2)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestMealTypeService_Cascading_Error(t *testing.T) {
	_, err := NewMealTypeService(notFound(t)).Cascading(context.Background(), 2)
	require.Error(t, err)
}

func TestMealTypeService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/meal-type/2/nulling/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewMealTypeService(testutil.NewMockExecutor(server.URL)).Nulling(context.Background(), 2)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestMealTypeService_Nulling_Error(t *testing.T) {
	_, err := NewMealTypeService(notFound(t)).Nulling(context.Background(), 2)
	require.Error(t, err)
}

func TestMealTypeService_Protecting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/meal-type/2/protecting/", r.URL.Path)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewMealTypeService(testutil.NewMockExecutor(server.URL)).Protecting(context.Background(), 2)
	require.NoError(t, err)
	require.NotNil(t, page)
}

func TestMealTypeService_Protecting_Error(t *testing.T) {
	_, err := NewMealTypeService(notFound(t)).Protecting(context.Background(), 2)
	require.Error(t, err)
}

// --- AutoPlanService ---

func TestAutoPlanService_Plan_Error(t *testing.T) {
	exec := mockStatus(t, http.StatusBadRequest, `{"keywords":["Provide at least one keyword."]}`)
	_, err := NewAutoPlanService(exec).Plan(context.Background(), &AutoMealPlanRequest{})
	require.Error(t, err)
}
