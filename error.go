package tandoor

import (
	"fmt"
	"net/http"
	"strings"
)

// APIError represents a parsed DRF error response.
type APIError struct {
	Detail         string
	FieldErrors    map[string][]string
	NonFieldErrors []string
}

// TandoorError represents an API error returned by the Tandoor server.
//
// It implements the error interface and includes the HTTP status code
// and raw response body for debugging.
type TandoorError struct {
	// StatusCode is the HTTP response status code (e.g., 404, 500).
	StatusCode int

	// Body is the raw response body from the server.
	Body []byte

	// Message is a human-readable error message, parsed from the response
	// if available (e.g., DRF error_detail or non_field_errors).
	Message string

	// apiErr holds the parsed DRF error details, if any.
	apiErr *APIError
}

// Error implements the error interface.
func (e *TandoorError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = string(e.Body)
		msg = strings.TrimSpace(msg)
		if msg == "" || msg == "{}" || msg == "[]" {
			msg = httpStatusText(e.StatusCode)
		}
	}
	return fmt.Sprintf("tandoor: API error %d: %s", e.StatusCode, msg)
}

// Is implements errors.Is so callers can use errors.Is(err, &TandoorError{}).
func (e *TandoorError) Is(target error) bool {
	t, ok := target.(*TandoorError)
	if !ok {
		return false
	}
	if t.StatusCode == 0 {
		// Match any TandoorError.
		return true
	}
	return e.StatusCode == t.StatusCode
}

// As implements errors.As.
func (e *TandoorError) As(target any) bool {
	t, ok := target.(*TandoorError)
	if !ok {
		return false
	}
	*t = *e
	return true
}

// IsNotFound returns true if the error represents a 404 Not Found.
func (e *TandoorError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsForbidden returns true if the error represents a 403 Forbidden.
func (e *TandoorError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsUnauthorized returns true if the error represents a 401 Unauthorized.
func (e *TandoorError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// Details returns the parsed field errors from a DRF validation response.
// It returns an empty map if the error was not parsed or contains no field errors.
func (e *TandoorError) Details() map[string][]string {
	if e.apiErr != nil {
		return e.apiErr.FieldErrors
	}
	return nil
}

// NonFieldErrors returns non-field errors from a DRF validation response.
func (e *TandoorError) NonFieldErrors() []string {
	if e.apiErr != nil {
		return e.apiErr.NonFieldErrors
	}
	return nil
}

// httpStatusText returns a short text description for common HTTP status codes.
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
	case 409:
		return "Conflict"
	case 422:
		return "Unprocessable Entity"
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
