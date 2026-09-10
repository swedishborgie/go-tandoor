package fdc

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetFoodWithFormatAndNutrients(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("format") != "full" || len(q["nutrients"]) != 2 {
			t.Errorf("unexpected query: %v", q)
		}
		w.Write([]byte(`{"fdcId": 1}`))
	})
	if _, err := client.GetFood(context.Background(), 1, FormatFull, []int{208, 205}); err != nil {
		t.Fatalf("GetFood: %v", err)
	}
}

func TestGetFoodsWithNutrients(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Query()["nutrients"]) != 1 {
			t.Errorf("expected nutrients, got %v", r.URL.Query())
		}
		w.Write([]byte(`[{"fdcId": 1}]`))
	})
	if _, err := client.GetFoods(context.Background(), []int{1}, "", []int{208}); err != nil {
		t.Fatalf("GetFoods: %v", err)
	}
}

func TestHttpStatusText(t *testing.T) {
	cases := map[int]string{
		400: "Bad Request",
		401: "Unauthorized",
		403: "Forbidden",
		404: "Not Found",
		429: "Too Many Requests",
		500: "Internal Server Error",
		502: "Bad Gateway",
		503: "Service Unavailable",
		599: "HTTP 599",
	}
	for code, want := range cases {
		if got := httpStatusText(code); got != want {
			t.Errorf("httpStatusText(%d) = %q, want %q", code, got, want)
		}
	}
}

func TestWithBaseURLInvalid(t *testing.T) {
	_, err := NewClient(WithAPIKey("k"), WithBaseURL("://bad"))
	if err == nil {
		t.Fatal("expected error for invalid base URL")
	}
}

func TestWithHTTPClientSuccess(t *testing.T) {
	client, err := NewClient(WithAPIKey("k"), WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if client.httpClient == nil {
		t.Fatal("expected http.Client to be set")
	}
}

func TestFlexibleIntUnmarshal(t *testing.T) {
	var f FlexibleInt
	if err := json.Unmarshal([]byte("null"), &f); err != nil || f.HasValue {
		t.Errorf("null: err=%v hasValue=%v", err, f.HasValue)
	}
	if err := json.Unmarshal([]byte("42"), &f); err != nil || f.Value != 42 || !f.HasValue {
		t.Errorf("int: err=%v val=%d", err, f.Value)
	}
	if err := json.Unmarshal([]byte(`"7"`), &f); err != nil || f.Value != 7 || !f.HasValue {
		t.Errorf("string: err=%v val=%d", err, f.Value)
	}
	if err := json.Unmarshal([]byte(`"abc"`), &f); err == nil {
		t.Error("expected error for non-numeric string")
	}
	if err := json.Unmarshal([]byte("true"), &f); err == nil {
		t.Error("expected error for bool")
	}
}

func TestFlexibleIntMarshal(t *testing.T) {
	b, err := json.Marshal(FlexibleInt{})
	if err != nil || string(b) != "null" {
		t.Errorf("no value: %s %v", b, err)
	}
	b, err = json.Marshal(FlexibleInt{Value: 7, HasValue: true})
	if err != nil || string(b) != "7" {
		t.Errorf("value: %s %v", b, err)
	}
}

func TestFlexibleStringUnmarshal(t *testing.T) {
	var f FlexibleString
	if err := json.Unmarshal([]byte("null"), &f); err != nil || f.HasValue {
		t.Errorf("null: err=%v hasValue=%v", err, f.HasValue)
	}
	if err := json.Unmarshal([]byte(`"x"`), &f); err != nil || f.Value != "x" {
		t.Errorf("string: err=%v val=%q", err, f.Value)
	}
	if err := json.Unmarshal([]byte("3.5"), &f); err != nil || f.Value != "3.5" {
		t.Errorf("number: err=%v val=%q", err, f.Value)
	}
	if err := json.Unmarshal([]byte("true"), &f); err == nil {
		t.Error("expected error for bool")
	}
}

func TestFlexibleStringMarshal(t *testing.T) {
	b, err := json.Marshal(FlexibleString{})
	if err != nil || string(b) != "null" {
		t.Errorf("no value: %s %v", b, err)
	}
	b, err = json.Marshal(FlexibleString{Value: "hi", HasValue: true})
	if err != nil || string(b) != `"hi"` {
		t.Errorf("value: %s %v", b, err)
	}
}

// TestFdcDoTransportError verifies the request-failure branch of do.
func TestFdcDoTransportError(t *testing.T) {
	client, err := NewClient(WithAPIKey("k"), WithBaseURL("http://127.0.0.1:1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetFood(context.Background(), 1, "", nil); err == nil {
		t.Fatal("expected transport error")
	}
}

// TestFdcErrorBodyFormats exercises Error() with parsed JSON and raw bodies.
func TestFdcErrorBodyFormats(t *testing.T) {
	e := &Error{StatusCode: 400, Body: []byte(`{"detail":"nope"}`)}
	if e.Error() == "" {
		t.Fatal("expected non-empty error")
	}
	e2 := &Error{StatusCode: 418, Body: []byte(`weird`)}
	_ = e2
	if e2.Error() == "" {
		t.Fatal("expected non-empty error")
	}
}
