package tandoor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c, err := NewClient(server.URL)
	require.NoError(t, err)
	return c
}

// TestDoNonJSONErrorBody covers the branch where the error body is not JSON
// (apiErr stays nil) and the Details/NonFieldErrors nil paths.
func TestDoNonJSONErrorBody(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("plain text failure"))
	})
	err := c.DoJSON(context.Background(), "GET", "api/x/", nil, nil)
	require.Error(t, err)
	tErr := &TandoorError{}
	require.ErrorAs(t, err, &tErr)
	assert.Nil(t, tErr.Details())
	assert.Nil(t, tErr.NonFieldErrors())
}

// TestDoErrorBodyOddShapes covers non-string detail and non-string
// non_field_errors entries.
func TestDoErrorBodyOddShapes(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":42,"non_field_errors":["ok",7]}`))
	})
	err := c.DoJSON(context.Background(), "GET", "api/x/", nil, nil)
	require.Error(t, err)
	tErr := &TandoorError{}
	require.ErrorAs(t, err, &tErr)
	assert.Equal(t, []string{"ok"}, tErr.NonFieldErrors())
}

// TestDoTransportError covers the httpClient.Do failure branch.
func TestDoTransportError(t *testing.T) {
	c, err := NewClient("http://127.0.0.1:1")
	require.NoError(t, err)
	err = c.DoJSON(context.Background(), "GET", "api/x/", nil, nil)
	require.Error(t, err)
}

// TestDoJSONWithQueryString covers splitQuery's query branch.
func TestDoJSONWithQueryString(t *testing.T) {
	var gotQuery string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"count":0,"results":[]}`))
	})
	err := c.DoJSON(context.Background(), "GET", "api/food/?page=2&page_size=5", nil, nil)
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "page=2")
}

// TestDoRawHappyPaths covers DoRaw and DoRawNoAuth success paths.
func TestDoRawHappyPaths(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("raw-bytes"))
	})
	resp, err := c.DoRaw(context.Background(), "GET", "api/raw/", nil)
	require.NoError(t, err)
	resp.Body.Close()
	resp, err = c.DoRawNoAuth(context.Background(), "GET", "api/raw/", nil)
	require.NoError(t, err)
	resp.Body.Close()
}

// TestTandoorErrorAs covers both branches of As.
func TestTandoorErrorAs(t *testing.T) {
	e := &TandoorError{StatusCode: 404}
	// As expects a *TandoorError target (not a pointer to one).
	out := &TandoorError{}
	assert.True(t, e.As(out))
	require.NotNil(t, out)
	assert.Equal(t, 404, out.StatusCode)
	var notErr int
	assert.False(t, e.As(&notErr))
}

// TestTandoorErrorDetailsNil covers Details/NonFieldErrors on a manually
// constructed error without parsed apiErr.
func TestTandoorErrorDetailsNil(t *testing.T) {
	e := &TandoorError{StatusCode: 400, Body: []byte("x")}
	assert.Nil(t, e.Details())
	assert.Nil(t, e.NonFieldErrors())
}
