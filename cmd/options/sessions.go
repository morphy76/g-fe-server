// go:build with-http-session

package options

import "net/http"

type SessionOptions struct {
	SessionKey           string
	SessionName          string
	SessionMaxAge        int
	SessionHttpOnly      bool
	SessionDomain        string
	SessionSecureCookies bool
	SessionSameSite      http.SameSite
}
