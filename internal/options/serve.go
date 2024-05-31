package options

import "net/http"

type ServeOptions struct {
	ContextRoot          string
	StaticPath           string
	Protocol             string
	Port                 string
	Host                 string
	SessionKey           string
	SessionName          string
	SessionMaxAge        int
	SessionHttpOnly      bool
	SessionDomain        string
	SessionSecureCookies bool
	SessionSameSite      http.SameSite
}
