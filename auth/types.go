// Package auth contains types for Tandoor authentication.
package auth

import "time"

// Request is the body sent to POST /api-token-auth/.
type Request struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Token is the response from POST /api-token-auth/.
type Token struct {
	Token  string    `json:"token"`
	Scope  string    `json:"scope"`
	Expiry time.Time `json:"expires"`
	UserID int       `json:"user_id"`
}

// AccessToken is the OAuth2 token managed via /api/access-token/.
type AccessToken struct {
	ID      int       `json:"id"`
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
	Scope   string    `json:"scope"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}
