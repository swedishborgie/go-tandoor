// Package storage contains types for Tandoor Storage resources.
//
// Storage defines external storage backends for recipe imports (e.g. local
// filesystem, S3-compatible endpoints).
package storage

// Storage represents an external storage backend.
type Storage struct {
	ID int `json:"id,omitempty"`

	// Name is a human-readable label for the storage backend.
	Name string `json:"name,omitempty"`

	// Method is the storage method identifier.
	Method string `json:"method,omitempty"`

	// Username is the auth username (if applicable).
	Username string `json:"username,omitempty"`

	// Password is the auth password (write-only, never returned).
	Password string `json:"password,omitempty"`

	// Token is an auth token (write-only, never returned).
	Token string `json:"token,omitempty"`

	// URL is the base URL or path.
	URL string `json:"url,omitempty"`

	// Path is the sub-path or bucket.
	Path string `json:"path,omitempty"`

	// CreatedBy is the user who created this storage (read-only).
	CreatedBy int `json:"created_by,omitempty"`
}
