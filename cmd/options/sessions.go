//go:build with_http_session

package options

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type SessionStore interface {
	sessions.Store
}

type SessionOptions struct {
	SessionKey           string
	SessionName          string
	SessionMaxAge        int
	SessionHttpOnly      bool
	SessionDomain        string
	SessionSecureCookies bool
	SessionSameSite      http.SameSite
}
