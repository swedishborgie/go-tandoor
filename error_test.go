package tandoor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTandoorError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *TandoorError
		wantSub string
	}{
		{
			name:    "JSON error body",
			err:     &TandoorError{StatusCode: 400, Body: []byte(`{"name": ["This field is required."]}`)},
			wantSub: "API error 400",
		},
		{
			name:    "empty body falls back to status text",
			err:     &TandoorError{StatusCode: 404, Body: []byte("")},
			wantSub: "Not Found",
		},
		{
			name:    "empty object body falls back to status text",
			err:     &TandoorError{StatusCode: 500, Body: []byte("{}")},
			wantSub: "Internal Server Error",
		},
		{
			name:    "empty array body falls back to status text",
			err:     &TandoorError{StatusCode: 503, Body: []byte("[]")},
			wantSub: "Service Unavailable",
		},
		{
			name:    "unrecognized status code",
			err:     &TandoorError{StatusCode: 599, Body: []byte("")},
			wantSub: "HTTP 599",
		},
		{
			name:    "message overrides body",
			err:     &TandoorError{StatusCode: 400, Body: []byte(`{"detail":"bad request"}`), Message: "bad request"},
			wantSub: "bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tt.err.Error(), tt.wantSub)
		})
	}
}

func TestTandoorError_Is(t *testing.T) {
	err := &TandoorError{StatusCode: 404}

	t.Run("matches any TandoorError", func(t *testing.T) {
		assert.ErrorIs(t, err, &TandoorError{})
	})

	t.Run("matches same status code", func(t *testing.T) {
		assert.ErrorIs(t, err, &TandoorError{StatusCode: 404})
	})

	t.Run("does not match different status code", func(t *testing.T) {
		assert.NotErrorIs(t, err, &TandoorError{StatusCode: 500})
	})

	t.Run("does not match non-TandoorError", func(t *testing.T) {
		assert.NotErrorIs(t, err, assert.AnError)
	})
}

func TestTandoorError_IsNotFound(t *testing.T) {
	assert.True(t, (&TandoorError{StatusCode: 404}).IsNotFound())
	assert.False(t, (&TandoorError{StatusCode: 403}).IsNotFound())
}

func TestTandoorError_IsForbidden(t *testing.T) {
	assert.True(t, (&TandoorError{StatusCode: 403}).IsForbidden())
	assert.False(t, (&TandoorError{StatusCode: 404}).IsForbidden())
}

func TestTandoorError_IsUnauthorized(t *testing.T) {
	assert.True(t, (&TandoorError{StatusCode: 401}).IsUnauthorized())
	assert.False(t, (&TandoorError{StatusCode: 404}).IsUnauthorized())
}

func TestTandoorError_As(t *testing.T) {
	err := &TandoorError{StatusCode: 403, Body: []byte(`{"detail":"forbidden"}`)}

	var target *TandoorError
	require.ErrorAs(t, err, &target)
	assert.Equal(t, 403, target.StatusCode)
}

func TestTandoorError_Details(t *testing.T) {
	err := &TandoorError{StatusCode: 400}
	assert.Nil(t, err.Details())
	assert.Nil(t, err.NonFieldErrors())
}

func TestHttpStatusText(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{400, "Bad Request"},
		{401, "Unauthorized"},
		{403, "Forbidden"},
		{404, "Not Found"},
		{409, "Conflict"},
		{422, "Unprocessable Entity"},
		{429, "Too Many Requests"},
		{500, "Internal Server Error"},
		{502, "Bad Gateway"},
		{503, "Service Unavailable"},
		{999, "HTTP 999"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, httpStatusText(tt.code), "status %d", tt.code)
	}
}
