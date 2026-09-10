package fdc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*Client, string) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := NewClient(
		WithAPIKey("test-key"),
		WithBaseURL(srv.URL),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client, srv.URL
}

// --- Constructor tests ---

func TestNewClientRequiresAPIKey(t *testing.T) {
	_, err := NewClient()
	if err == nil {
		t.Fatal("expected error when API key is missing")
	}
}

func TestNewClientSuccess(t *testing.T) {
	client, err := NewClient(WithAPIKey("my-key"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if client.apiKey != "my-key" {
		t.Fatalf("expected apiKey=my-key, got %s", client.apiKey)
	}
}

func TestWithHTTPClientNil(t *testing.T) {
	_, err := NewClient(WithAPIKey("key"), WithHTTPClient(nil))
	if err == nil {
		t.Fatal("expected error for nil http.Client")
	}
}

func TestWithBaseURLEmpty(t *testing.T) {
	_, err := NewClient(WithAPIKey("key"), WithBaseURL(""))
	if err == nil {
		t.Fatal("expected error for empty base URL")
	}
}

// --- Error tests ---

func TestFdcError(t *testing.T) {
	err := &Error{StatusCode: 404, Body: []byte(`"not found"`)}
	if !err.IsNotFound() {
		t.Fatal("expected IsNotFound() == true")
	}
	if err.IsBadRequest() {
		t.Fatal("expected IsBadRequest() == false")
	}
	msg := err.Error()
	if msg == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestFdcErrorEmptyBody(t *testing.T) {
	err := &Error{StatusCode: 404, Body: nil}
	msg := err.Error()
	if msg == "" {
		t.Fatal("expected non-empty error message")
	}
	// Should fall back to httpStatusText
	if msg != "fdc: API error 404: Not Found" {
		t.Fatalf("unexpected message: %s", msg)
	}
}

func TestFdcErrorBadRequest(t *testing.T) {
	err := &Error{StatusCode: 400}
	if !err.IsBadRequest() {
		t.Fatal("expected IsBadRequest() == true")
	}
	if err.IsNotFound() {
		t.Fatal("expected IsNotFound() == false")
	}
}

// --- splitQuery tests ---

func TestSplitQuery(t *testing.T) {
	tests := []struct {
		input    string
		wantPath string
		wantQ    string
	}{
		{"v1/food/123", "v1/food/123", ""},
		{"v1/foods?format=abridged", "v1/foods", "format=abridged"},
		{"v1/foods/list", "v1/foods/list", ""},
		{"", "", ""},
	}
	for _, tt := range tests {
		p, q := splitQuery(tt.input)
		if p != tt.wantPath {
			t.Errorf("splitQuery(%q) path = %q, want %q", tt.input, p, tt.wantPath)
		}
		if q != tt.wantQ {
			t.Errorf("splitQuery(%q) query = %q, want %q", tt.input, q, tt.wantQ)
		}
	}
}

// --- do API key injection test ---

func TestDoInjectsAPIKey(t *testing.T) {
	var capturedKey string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedKey = r.URL.Query().Get("api_key")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"fdcId": 1}`))
	})

	_, err := client.GetFood(context.Background(), 123, "", nil)
	if err != nil {
		t.Fatalf("GetFood: %v", err)
	}
	if capturedKey != "test-key" {
		t.Fatalf("expected api_key=test-key, got %s", capturedKey)
	}
}

func TestDoReturnsErrorOn4xx(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"bad request"`))
	})

	_, err := client.GetFood(context.Background(), 123, "", nil)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	fdcErr := &Error{}
	ok := errors.As(err, &fdcErr)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if fdcErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", fdcErr.StatusCode)
	}
}

func TestDoReturnsErrorOn404(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{}`))
	})

	_, err := client.GetFood(context.Background(), 123, "", nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	fdcErr := &Error{}
	ok := errors.As(err, &fdcErr)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if !fdcErr.IsNotFound() {
		t.Fatal("expected IsNotFound() == true")
	}
}

// --- Context cancellation test ---

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Slow response
		w.WriteHeader(http.StatusOK)
	})

	_, err := client.GetFood(ctx, 123, "", nil)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// --- POST body Content-Type test ---

func TestDoSetsContentTypeForPOST(t *testing.T) {
	var contentType string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	})

	_, err := client.PostFoods(context.Background(), &FoodsCriteria{FDCIDs: []int{123}})
	if err != nil {
		t.Fatalf("PostFoods: %v", err)
	}
	if contentType != "application/json" {
		t.Fatalf("expected Content-Type=application/json, got %s", contentType)
	}
}

// --- JSON unmarshal test ---

func TestDoUnmarshalsJSON(t *testing.T) {
	payload := map[string]interface{}{
		"fdcId":       534358,
		"dataType":    "Branded",
		"description": "NUT 'N BERRY MIX",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	food, err := client.GetFood(context.Background(), 534358, "", nil)
	if err != nil {
		t.Fatalf("GetFood: %v", err)
	}
	if food.FDCID != 534358 {
		t.Fatalf("expected fdcId=534358, got %d", food.FDCID)
	}
	if food.DataType != "Branded" {
		t.Fatalf("expected dataType=Branded, got %s", food.DataType)
	}
}

// --- Remaining client method tests ---

func TestGetFoods(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query()["fdcIds"]; len(got) != 2 {
			t.Errorf("expected 2 fdcIds, got %v", got)
		}
		if r.URL.Query().Get("format") != "abridged" {
			t.Errorf("expected format=abridged, got %q", r.URL.Query().Get("format"))
		}
		w.Write([]byte(`[{"fdcId": 1}, {"fdcId": 2}]`))
	})
	foods, err := client.GetFoods(context.Background(), []int{1, 2}, "abridged", nil)
	if err != nil {
		t.Fatalf("GetFoods: %v", err)
	}
	if len(foods) != 2 {
		t.Fatalf("expected 2 foods, got %d", len(foods))
	}
}

func TestPostFoods(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Write([]byte(`[{"fdcId": 1}]`))
	})
	foods, err := client.PostFoods(context.Background(), &FoodsCriteria{FDCIDs: []int{1}})
	if err != nil {
		t.Fatalf("PostFoods: %v", err)
	}
	if len(foods) != 1 {
		t.Fatalf("expected 1 food, got %d", len(foods))
	}
	if _, err := client.PostFoods(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil criteria")
	}
}

func TestListFoods(t *testing.T) {
	ps, pn := 10, 2
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("pageSize") != "10" || q.Get("pageNumber") != "2" {
			t.Errorf("unexpected pagination: %v", q)
		}
		if q.Get("dataType") == "" {
			t.Error("expected dataType filter")
		}
		w.Write([]byte(`[{"fdcId": 1}]`))
	})
	foods, err := client.ListFoods(context.Background(), []DataType{DataTypeBranded}, &ps, &pn, "description", "asc")
	if err != nil {
		t.Fatalf("ListFoods: %v", err)
	}
	if len(foods) != 1 {
		t.Fatalf("expected 1 food, got %d", len(foods))
	}
}

func TestPostFoodsList(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"fdcId": 1}]`))
	})
	foods, err := client.PostFoodsList(context.Background(), &FoodListCriteria{DataType: []DataType{DataTypeBranded}})
	if err != nil {
		t.Fatalf("PostFoodsList: %v", err)
	}
	if len(foods) != 1 {
		t.Fatalf("expected 1 food, got %d", len(foods))
	}
	if _, err := client.PostFoodsList(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil criteria")
	}
}

func TestSearchFoods(t *testing.T) {
	ps, pn := 5, 1
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("query") != "pasta" || q.Get("brandOwner") != "Acme" {
			t.Errorf("unexpected query: %v", q)
		}
		w.Write([]byte(`{"totalHits": 1, "currentPage": 1, "totalPages": 1, "foods": [{"fdcId": 1, "description": "Pasta"}]}`))
	})
	resp, err := client.SearchFoods(context.Background(), "pasta", []DataType{DataTypeBranded}, &ps, &pn, "description", "asc", "Acme")
	if err != nil {
		t.Fatalf("SearchFoods: %v", err)
	}
	if resp.TotalHits != 1 || len(resp.Foods) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestPostFoodsSearch(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"totalHits": 1, "foods": [{"fdcId": 1, "description": "Pasta"}]}`))
	})
	resp, err := client.PostFoodsSearch(context.Background(), &FoodSearchCriteria{Query: "pasta"})
	if err != nil {
		t.Fatalf("PostFoodsSearch: %v", err)
	}
	if resp.TotalHits != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if _, err := client.PostFoodsSearch(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil criteria")
	}
	if _, err := client.PostFoodsSearch(context.Background(), &FoodSearchCriteria{}); err == nil {
		t.Fatal("expected error for empty query")
	}
}
