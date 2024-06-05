package cli

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/securecookie"
	"github.com/morphy76/g-fe-server/internal/options"
)

type serveOptionsBuilder func() (*options.ServeOptions, error)

var errInvalidContextRoot = errors.New("invalid context root")
var errInvalidStaticPath = errors.New("invalid static path")
var errInvalidSessionSameSite = errors.New("invalid session same site")

func IsInvalidContextRoot(err error) bool {
	return err == errInvalidContextRoot
}

func IsInvalidStaticPath(err error) bool {
	return err == errInvalidStaticPath
}

func IsInvalidSessionSameSite(err error) bool {
	return err == errInvalidSessionSameSite
}

const (
	ENV_CONTEXT_ROOT      = "CONTEXT_ROOT"
	ENV_STATIC_PATH       = "STATIC_PATH"
	ENV_PORT              = "SERVE_PORT"
	ENV_HOST              = "SERVE_HOST"
	ENV_SESSION_KEY       = "SESSION_KEY"
	ENV_SESSION_NAME      = "SESSION_NAME"
	ENV_SESSION_MAX_AGE   = "SESSION_MAX_AGE"
	ENV_SESSION_HTTP_ONLY = "SESSION_HTTP_ONLY"
	ENV_SESSION_DOMAIN    = "SESSION_DOMAIN"
	ENV_SESSION_SECURE    = "SESSION_SECURE"
	ENV_SESSION_SAME_SITE = "SESSION_SAME_SITE"
)

func ServeOptionsBuilder() serveOptionsBuilder {

	ctxRootArg := flag.String("ctx", "", "presentation server context root. Environment: "+ENV_CONTEXT_ROOT)
	staticPathArg := flag.String("static", "/static", "static path of the served application. Environment: "+ENV_STATIC_PATH)
	portArg := flag.String("port", "8080", "binding port of the presentation server. Environment: "+ENV_PORT)
	hostArg := flag.String("host", "0.0.0.0", "binding host of the presentation server. Environment: "+ENV_HOST)
	sessionKeyArg := flag.String("session-key", "", "session key. Environment: "+ENV_SESSION_KEY)
	sessionNameArg := flag.String("session-name", "gofe.sid", "session name. Environment: "+ENV_SESSION_NAME)
	sessionMaxAgeArg := flag.Int("session-max-age", 0, "session max age. Environment: "+ENV_SESSION_MAX_AGE)
	sessionHttpOnlyArg := flag.Bool("session-http-only", false, "session http only. Environment: "+ENV_SESSION_HTTP_ONLY)
	sessionDomainArg := flag.String("session-domain", "", "session domain. Environment: "+ENV_SESSION_DOMAIN)
	sessionSecureArg := flag.Bool("session-secure", false, "session secure. Environment: "+ENV_SESSION_SECURE)
	sessionSameSiteArg := flag.String("session-same-site", "Lax", "session same site: Default, Lax, Strict or None. Environment: "+ENV_SESSION_SAME_SITE)

	rv := func() (*options.ServeOptions, error) {

		ctxRoot, found := os.LookupEnv(ENV_CONTEXT_ROOT)
		if !found {
			ctxRoot = *ctxRootArg
		}
		if len(ctxRoot) == 0 || strings.Contains(ctxRoot, " ") || !strings.HasPrefix(ctxRoot, "/") {
			return nil, errInvalidContextRoot
		}

		staticPath, found := os.LookupEnv(ENV_STATIC_PATH)
		if !found {
			staticPath = *staticPathArg
		}
		if len(staticPath) == 0 || strings.Contains(staticPath, " ") {
			return nil, errInvalidStaticPath
		}

		usePort, found := os.LookupEnv(ENV_PORT)
		if !found {
			usePort = *portArg
		}

		useHost, found := os.LookupEnv(ENV_HOST)
		if !found {
			useHost = *hostArg
		}

		useSessionKey, found := os.LookupEnv(ENV_SESSION_KEY)
		if !found {
			useSessionKey = *sessionKeyArg
		}
		if len(useSessionKey) == 0 {
			useSessionKey = string(securecookie.GenerateRandomKey(32))
		}

		useSessionName, found := os.LookupEnv(ENV_SESSION_NAME)
		if !found {
			useSessionName = *sessionNameArg
		}
		if len(useSessionName) == 0 {
			useSessionName = "gofe.sid"
		}

		var useSessionMaxAge int
		strSessionMaxAge, found := os.LookupEnv(ENV_SESSION_MAX_AGE)
		if !found {
			useSessionMaxAge = *sessionMaxAgeArg
		} else {
			maxAge, err := strconv.Atoi(strSessionMaxAge)
			if err != nil {
				return nil, err
			}
			useSessionMaxAge = maxAge
		}

		var useSessionHttpOnly bool
		strSessionHttpOnly, found := os.LookupEnv(ENV_SESSION_HTTP_ONLY)
		if !found {
			useSessionHttpOnly = *sessionHttpOnlyArg
		} else {
			useSessionHttpOnly = strSessionHttpOnly == "true"
		}

		useSessionDomain, found := os.LookupEnv(ENV_SESSION_DOMAIN)
		if !found {
			useSessionDomain = *sessionDomainArg
		}

		var useSessionSecure bool
		strSessionSecure, found := os.LookupEnv(ENV_SESSION_SECURE)
		if !found {
			useSessionSecure = *sessionSecureArg
		} else {
			useSessionSecure = strSessionSecure == "true"
		}

		var useSessionSameSite http.SameSite
		strSessionSameSite, found := os.LookupEnv(ENV_SESSION_SAME_SITE)
		if !found {
			strSessionSameSite = *sessionSameSiteArg
		}
		if strSessionSameSite == "Lax" {
			useSessionSameSite = http.SameSiteLaxMode
		} else if strSessionSameSite == "Strict" {
			useSessionSameSite = http.SameSiteStrictMode
		} else if strSessionSameSite == "None" {
			useSessionSameSite = http.SameSiteNoneMode
		} else if strSessionSameSite == "Default" {
			useSessionSameSite = http.SameSiteDefaultMode
		} else {
			return nil, errInvalidSessionSameSite
		}

		return &options.ServeOptions{
			ContextRoot:          ctxRoot,
			StaticPath:           staticPath,
			Protocol:             "http",
			Port:                 usePort,
			Host:                 useHost,
			SessionKey:           useSessionKey,
			SessionName:          useSessionName,
			SessionMaxAge:        useSessionMaxAge,
			SessionHttpOnly:      useSessionHttpOnly,
			SessionDomain:        useSessionDomain,
			SessionSecureCookies: useSessionSecure,
			SessionSameSite:      useSessionSameSite,
		}, nil
	}

	return rv
}
