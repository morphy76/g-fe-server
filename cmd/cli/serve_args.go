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

// ServeOptionsBuilderFn is a function that returns ServeOptions
type ServeOptionsBuilderFn func() (*options.ServeOptions, error)

var errInvalidContextRoot = errors.New("invalid context root")
var errInvalidStaticPath = errors.New("invalid static path")
var errInvalidSessionSameSite = errors.New("invalid session same site")

// IsInvalidContextRoot checks if the error is due to an invalid context root
func IsInvalidContextRoot(err error) bool {
	return err == errInvalidContextRoot
}

// IsInvalidStaticPath checks if the error is due to an invalid static path
func IsInvalidStaticPath(err error) bool {
	return err == errInvalidStaticPath
}

// IsInvalidSessionSameSite checks if the error is due to an invalid session same site
func IsInvalidSessionSameSite(err error) bool {
	return err == errInvalidSessionSameSite
}

const (
	envNonFnRoot       = "NON_FUNCTIONAL_ROOT"
	envCtxRoot         = "CONTEXT_ROOT"
	envStaticPath      = "STATIC_PATH"
	envPort            = "SERVE_PORT"
	envHost            = "SERVE_HOST"
	envSessionKey      = "SESSION_KEY"
	envSessionName     = "SESSION_NAME"
	envSessionMaxAge   = "SESSION_MAX_AGE"
	envSessionHTTPOnly = "SESSION_HTTP_ONLY"
	envSessionDomain   = "SESSION_DOMAIN"
	envSessionSecure   = "SESSION_SECURE"
	envSessionSameSite = "SESSION_SAME_SITE"
)

// ServeOptionsBuilder returns a function that builds ServeOptions from the command line arguments and environment variables
func ServeOptionsBuilder() ServeOptionsBuilderFn {

	nonFnRootArg := flag.String("non-fn", "/g", "presentation server non functional root. Environment: "+envNonFnRoot)
	ctxRootArg := flag.String("ctx", "", "presentation server context root. Environment: "+envCtxRoot)
	staticPathArg := flag.String("static", "/static", "static path of the served application. Environment: "+envStaticPath)
	portArg := flag.String("port", "8080", "binding port of the presentation server. Environment: "+envPort)
	hostArg := flag.String("host", "0.0.0.0", "binding host of the presentation server. Environment: "+envHost)
	sessionKeyArg := flag.String("session-key", "", "session key. Environment: "+envSessionKey)
	sessionNameArg := flag.String("session-name", "gofe.sid", "session name. Environment: "+envSessionName)
	sessionMaxAgeArg := flag.Int("session-max-age", 0, "session max age. Environment: "+envSessionMaxAge)
	sessionHTTPOnlyArg := flag.Bool("session-http-only", false, "session http only. Environment: "+envSessionHTTPOnly)
	sessionDomainArg := flag.String("session-domain", "", "session domain. Environment: "+envSessionDomain)
	sessionSecureArg := flag.Bool("session-secure", false, "session secure. Environment: "+envSessionSecure)
	sessionSameSiteArg := flag.String("session-same-site", "Lax", "session same site: Default, Lax, Strict or None. Environment: "+envSessionSameSite)

	rv := func() (*options.ServeOptions, error) {

		nonFnPath, found := os.LookupEnv(envNonFnRoot)
		if !found {
			nonFnPath = *nonFnRootArg
		}
		if len(nonFnPath) == 0 || strings.Contains(nonFnPath, " ") {
			return nil, errInvalidStaticPath
		}

		ctxRoot, found := os.LookupEnv(envCtxRoot)
		if !found {
			ctxRoot = *ctxRootArg
		}
		if len(ctxRoot) == 0 || strings.Contains(ctxRoot, " ") || !strings.HasPrefix(ctxRoot, "/") {
			return nil, errInvalidContextRoot
		}

		staticPath, found := os.LookupEnv(envStaticPath)
		if !found {
			staticPath = *staticPathArg
		}
		if len(staticPath) == 0 || strings.Contains(staticPath, " ") {
			return nil, errInvalidStaticPath
		}

		usePort, found := os.LookupEnv(envPort)
		if !found {
			usePort = *portArg
		}

		useHost, found := os.LookupEnv(envHost)
		if !found {
			useHost = *hostArg
		}

		useSessionKey, found := os.LookupEnv(envSessionKey)
		if !found {
			useSessionKey = *sessionKeyArg
		}
		if len(useSessionKey) == 0 {
			useSessionKey = string(securecookie.GenerateRandomKey(32))
		}

		useSessionName, found := os.LookupEnv(envSessionName)
		if !found {
			useSessionName = *sessionNameArg
		}
		if len(useSessionName) == 0 {
			useSessionName = "gofe.sid"
		}

		var useSessionMaxAge int
		strSessionMaxAge, found := os.LookupEnv(envSessionMaxAge)
		if !found {
			useSessionMaxAge = *sessionMaxAgeArg
		} else {
			maxAge, err := strconv.Atoi(strSessionMaxAge)
			if err != nil {
				return nil, err
			}
			useSessionMaxAge = maxAge
		}

		var useSessionHTTPOnly bool
		strSessionHTTPOnly, found := os.LookupEnv(envSessionHTTPOnly)
		if !found {
			useSessionHTTPOnly = *sessionHTTPOnlyArg
		} else {
			useSessionHTTPOnly = strSessionHTTPOnly == "true"
		}

		useSessionDomain, found := os.LookupEnv(envSessionDomain)
		if !found {
			useSessionDomain = *sessionDomainArg
		}

		var useSessionSecure bool
		strSessionSecure, found := os.LookupEnv(envSessionSecure)
		if !found {
			useSessionSecure = *sessionSecureArg
		} else {
			useSessionSecure = strSessionSecure == "true"
		}

		var useSessionSameSite http.SameSite
		strSessionSameSite, found := os.LookupEnv(envSessionSameSite)
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
			NonFunctionalRoot:    nonFnPath,
			ContextRoot:          ctxRoot,
			StaticPath:           staticPath,
			Protocol:             "http",
			Port:                 usePort,
			Host:                 useHost,
			SessionKey:           useSessionKey,
			SessionName:          useSessionName,
			SessionMaxAge:        useSessionMaxAge,
			SessionHttpOnly:      useSessionHTTPOnly,
			SessionDomain:        useSessionDomain,
			SessionSecureCookies: useSessionSecure,
			SessionSameSite:      useSessionSameSite,
		}, nil
	}

	return rv
}
