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
