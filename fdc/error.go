package fdc

import (
	"fmt"
	"net/http"
	"strings"
)

// Error represents an API error returned by the FDC server.
type Error struct {
	// StatusCode is the HTTP response status code (e.g., 400, 404).
	StatusCode int

	// Body is the raw response body from the server.
	Body []byte
}

// Error implements the error interface.
func (e *Error) Error() string {
	msg := strings.TrimSpace(string(e.Body))
	if msg == "" || msg == "{}" || msg == "[]" {
		msg = httpStatusText(e.StatusCode)
	}
	return fmt.Sprintf("fdc: API error %d: %s", e.StatusCode, msg)
}

// IsNotFound returns true if the error represents a 404 Not Found.
func (e *Error) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsBadRequest returns true if the error represents a 400 Bad Request.
func (e *Error) IsBadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

func httpStatusText(code int) string {
	switch code {
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 429:
		return "Too Many Requests"
	case 500:
		return "Internal Server Error"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	default:
		return fmt.Sprintf("HTTP %d", code)
	}
}
