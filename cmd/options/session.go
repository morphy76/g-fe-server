package options

import (
	"net/http"
)

// SessionOptions holds the options for the session management.
type SessionOptions struct {
	Key           string
	Name          string
	Path          string
	MaxAge        int
	Domain        string
	SecureCookies bool
	SameSite      http.SameSite
	Partitioned   bool
}
