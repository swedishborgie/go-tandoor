package inventory

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

func strPtr(s string) *string { return &s }

// LocationService tests

func TestLocationService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/inventory-location/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1,"name":"Pantry"},{"id":2,"name":"Freezer","is_freezer":true}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewLocationService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
	assert.Equal(t, "Pantry", page.Results[0].Name)
	assert.True(t, page.Results[1].IsFreezer)
}

func TestLocationService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/inventory-location/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Pantry","is_freezer":false}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	loc, err := NewLocationService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Pantry", loc.Name)
}

func TestLocationService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"Cellar"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	loc, err := NewLocationService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "Cellar", loc.Name)
}

func TestLocationService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Dry Storage"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	loc, err := NewLocationService(mock).Update(context.Background(), &Location{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Dry Storage", loc.Name)
}

func TestLocationService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	loc, err := NewLocationService(mock).Patch(context.Background(), &Location{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, loc.ID)
}

func TestLocationService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewLocationService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestLocationService_Cascading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/inventory-location/1/cascading/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Pantry"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	locs, err := NewLocationService(mock).Cascading(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, locs.Count)
}

func TestLocationService_Nulling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/inventory-location/1/nulling/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":0,"results":[]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	locs, err := NewLocationService(mock).Nulling(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, locs.Count)
}

func TestLocationService_Protecting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/inventory-location/1/protecting/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	locs, err := NewLocationService(mock).Protecting(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, locs.Count)
}

// EntryService tests

func TestEntryService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/inventory-entry/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"amount":5.5,"code":"ABC123"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewEntryService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "ABC123", page.Results[0].Code)
}

func TestEntryService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/inventory-entry/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"amount":5.5,"food":1,"unit":2}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	entry, err := NewEntryService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.InEpsilon(t, 5.5, entry.Amount, 1e-9)
}

func TestEntryService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"code":"DEF456","amount":10}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	entry, err := NewEntryService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "DEF456", entry.Code)
}

func TestEntryService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"amount":20}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	entry, err := NewEntryService(mock).Update(context.Background(), &Entry{ID: 1})
	require.NoError(t, err)
	assert.InEpsilon(t, 20.0, entry.Amount, 1e-9)
}

func TestEntryService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	entry, err := NewEntryService(mock).Patch(context.Background(), &Entry{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, entry.ID)
}

func TestEntryService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewEntryService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// LogService tests

func TestLogService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/inventory-log/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"booking_type":"B_ADD","old_amount":0,"new_amount":5}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewLogService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "B_ADD", page.Results[0].BookingType)
}

func TestLogService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/inventory-log/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"booking_type":"B_MOVE","note":"relocated"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewLogService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "B_MOVE", log.BookingType)
	assert.Equal(t, "relocated", log.Note)
}

// EntryListOptions tests

func TestEntryListOptions_QueryString(t *testing.T) {
	trueVal := true
	foodID := 5
	locID := 3
	opts := &EntryListOptions{
		ListOptions:  &pagination.ListOptions{Page: 1, PageSize: 10},
		IncludeEmpty: &trueVal,
		Code:         strPtr("ABC123"),
		FoodID:       &foodID,
		LocationID:   &locID,
	}
	qs := opts.QueryString()
	assert.Contains(t, qs, "page=1")
	assert.Contains(t, qs, "page_size=10")
	assert.Contains(t, qs, "empty=true")
	assert.Contains(t, qs, "code=ABC123")
	assert.Contains(t, qs, "food_id=5")
	assert.Contains(t, qs, "inventory_location_id=3")
}

func TestEntryListOptions_Empty(t *testing.T) {
	assert.True(t, (&EntryListOptions{}).Empty())
	assert.True(t, (*EntryListOptions)(nil).Empty())
	foodID := 0
	opts := &EntryListOptions{FoodID: &foodID}
	assert.False(t, opts.Empty())
}

// LogListOptions tests

func TestLogListOptions_QueryString(t *testing.T) {
	entryID := 5
	foodID := 3
	opts := &LogListOptions{
		ListOptions: &pagination.ListOptions{Page: 1, PageSize: 10},
		EntryID:     &entryID,
		FoodID:      &foodID,
	}
	qs := opts.QueryString()
	assert.Contains(t, qs, "page=1")
	assert.Contains(t, qs, "page_size=10")
	assert.Contains(t, qs, "entry_id=5")
	assert.Contains(t, qs, "food_id=3")
}

func TestLogListOptions_Empty(t *testing.T) {
	assert.Empty(t, (&LogListOptions{}).QueryString())
	entryID := 0
	opts := &LogListOptions{EntryID: &entryID}
	assert.NotEmpty(t, opts.QueryString())
}
