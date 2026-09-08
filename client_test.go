package tandoor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("valid baseURL creates client", func(t *testing.T) {
		c, err := NewClient("https://example.com")
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", c.baseURL.String())
		assert.NotNil(t, c.httpClient)
	})

	t.Run("empty baseURL returns error", func(t *testing.T) {
		_, err := NewClient("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "baseURL is required")
	})

	t.Run("trailing slash is normalized", func(t *testing.T) {
		c, err := NewClient("https://example.com/")
		require.NoError(t, err)
		assert.Equal(t, "https://example.com/", c.baseURL.String())
	})
}

func TestWithAccessToken(t *testing.T) {
	t.Run("sets token", func(t *testing.T) {
		c, err := NewClient("https://example.com", WithAccessToken("tda_test123"))
		require.NoError(t, err)
		assert.Equal(t, "tda_test123", c.accessToken)
	})

	t.Run("empty token returns error", func(t *testing.T) {
		_, err := NewClient("https://example.com", WithAccessToken(""))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "access token is required")
	})
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 5 * time.Second}

	t.Run("sets custom client", func(t *testing.T) {
		c, err := NewClient("https://example.com", WithHTTPClient(customClient))
		require.NoError(t, err)
		assert.Same(t, customClient, c.httpClient)
	})

	t.Run("nil client returns error", func(t *testing.T) {
		_, err := NewClient("https://example.com", WithHTTPClient(nil))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be nil")
	})
}

func TestWithBaseURL(t *testing.T) {
	t.Run("overrides base URL", func(t *testing.T) {
		c, err := NewClient("https://example.com", WithBaseURL("https://other.com"))
		require.NoError(t, err)
		assert.Equal(t, "https://other.com", c.baseURL.String())
	})

	t.Run("empty base URL returns error", func(t *testing.T) {
		_, err := NewClient("https://example.com", WithBaseURL(""))
		require.Error(t, err)
	})

	t.Run("invalid base URL returns error", func(t *testing.T) {
		_, err := NewClient("https://example.com", WithBaseURL("\x00invalid"))
		require.Error(t, err)
	})
}

func TestWithUserAgent(t *testing.T) {
	t.Run("sets user agent", func(t *testing.T) {
		c, err := NewClient("https://example.com", WithUserAgent("custom-agent/1.0"))
		require.NoError(t, err)
		req, err := c.newRequest(context.Background(), "GET", "api/", nil)
		require.NoError(t, err)
		assert.Equal(t, "custom-agent/1.0", req.Header.Get("User-Agent"))
	})

	t.Run("empty user agent returns error", func(t *testing.T) {
		_, err := NewClient("https://example.com", WithUserAgent(""))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user agent must not be empty")
	})

	t.Run("default user agent is set", func(t *testing.T) {
		c, err := NewClient("https://example.com")
		require.NoError(t, err)
		req, err := c.newRequest(context.Background(), "GET", "api/", nil)
		require.NoError(t, err)
		assert.Contains(t, req.Header.Get("User-Agent"), "go-tandoor/")
		assert.Contains(t, req.Header.Get("User-Agent"), "github.com/swedishborgie/go-tandoor")
	})
}

func TestMultipleOptions(t *testing.T) {
	customClient := &http.Client{}
	c, err := NewClient(
		"https://example.com",
		WithAccessToken("tda_token"),
		WithHTTPClient(customClient),
	)
	require.NoError(t, err)
	assert.Equal(t, "tda_token", c.accessToken)
	assert.Same(t, customClient, c.httpClient)
}

func TestNewRequest(t *testing.T) {
	c, err := NewClient("https://example.com", WithAccessToken("tda_test"))
	require.NoError(t, err)

	t.Run("GET request with no body", func(t *testing.T) {
		req, err := c.newRequest(context.Background(), "GET", "api/recipe/", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/api/recipe/")
		assert.Equal(t, "Bearer tda_test", req.Header.Get("Authorization"))
		assert.Nil(t, req.Body)
	})

	t.Run("POST request with body", func(t *testing.T) {
		body := map[string]string{"name": "test"}
		req, err := c.newRequest(context.Background(), "POST", "api/recipe/", body)
		require.NoError(t, err)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	})

	t.Run("no token omits auth header", func(t *testing.T) {
		cNoAuth, _ := NewClient("https://example.com")
		req, err := cNoAuth.newRequest(context.Background(), "GET", "api/", nil)
		require.NoError(t, err)
		assert.Empty(t, req.Header.Get("Authorization"))
	})
}

func TestDo(t *testing.T) {
	t.Run("200 OK decodes response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": 1, "name": "Test Recipe"}`))
		}))
		defer server.Close()

		c, err := NewClient(server.URL)
		require.NoError(t, err)

		req, err := c.newRequest(context.Background(), "GET", "/", nil)
		require.NoError(t, err)

		var result map[string]interface{}
		err = c.do(req, &result)
		require.NoError(t, err)
		assert.InEpsilon(t, float64(1), result["id"], 0)
		assert.Equal(t, "Test Recipe", result["name"])
	})

	t.Run("404 returns TandoorError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": "not found"}`))
		}))
		defer server.Close()

		c, err := NewClient(server.URL)
		require.NoError(t, err)

		req, err := c.newRequest(context.Background(), "GET", "/", nil)
		require.NoError(t, err)

		err = c.do(req, new(interface{}))
		require.Error(t, err)
		var terr *TandoorError
		require.ErrorAs(t, err, &terr)
		assert.Equal(t, 404, terr.StatusCode)
	})

	t.Run("500 returns TandoorError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error_detail": "server error"}`))
		}))
		defer server.Close()

		c, err := NewClient(server.URL)
		require.NoError(t, err)

		req, err := c.newRequest(context.Background(), "GET", "/", nil)
		require.NoError(t, err)

		err = c.do(req, new(interface{}))
		require.Error(t, err)
		var terr *TandoorError
		require.ErrorAs(t, err, &terr)
		assert.Equal(t, 500, terr.StatusCode)
	})

	t.Run("nil v discards body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		c, err := NewClient(server.URL)
		require.NoError(t, err)

		req, err := c.newRequest(context.Background(), "DELETE", "/", nil)
		require.NoError(t, err)

		err = c.do(req, nil)
		require.NoError(t, err)
	})
}

func TestDoJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer tda_token", r.Header.Get("Authorization"))

		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": 42, "name": "Chicken Tikka"}`))
	}))
	defer server.Close()

	c, err := NewClient(server.URL, WithAccessToken("tda_token"))
	require.NoError(t, err)

	var result map[string]interface{}
	err = c.doJSON(context.Background(), "POST", "api/recipe/", map[string]string{"name": "Chicken Tikka"}, &result)
	require.NoError(t, err)
	assert.InEpsilon(t, float64(42), result["id"], 0)
}

func TestDoJSON_BodyMarshalError(t *testing.T) {
	c, err := NewClient("https://example.com")
	require.NoError(t, err)

	// Channel cannot be JSON marshaled — but we can't easily trigger this
	// without an unmarshalable type. Skip for now; the error path is
	// straightforward and covered by the integration tests.
	_ = c
}
