package cli

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/securecookie"
	"github.com/morphy76/g-fe-server/internal/http/session"
)

// SessionOptionsBuilderFn is a function that returns SessionOptions
type SessionOptionsBuilderFn func() (*session.SessionOptions, error)

// ErrInvalidSessionSameSite is an invalid session same site error
var ErrInvalidSessionSameSite = errors.New("invalid session same site")

const (
	envSessionKey      = "SESSION_KEY"
	envSessionName     = "SESSION_NAME"
	envSessionMaxAge   = "SESSION_MAX_AGE"
	envSessionHTTPOnly = "SESSION_HTTP_ONLY"
	envSessionDomain   = "SESSION_DOMAIN"
	envSessionSecure   = "SESSION_SECURE"
	envSessionSameSite = "SESSION_SAME_SITE"
)

// SessionOptionsBuilder returns a function that builds SessionOptions from the command line arguments and environment variables
func SessionOptionsBuilder() SessionOptionsBuilderFn {

	sessionKeyArg := flag.String("session-key", "", "session key. Environment: "+envSessionKey)
	sessionNameArg := flag.String("session-name", "gofe.sid", "session name. Environment: "+envSessionName)
	sessionMaxAgeArg := flag.Int("session-max-age", 0, "session max age. Environment: "+envSessionMaxAge)
	sessionHTTPOnlyArg := flag.Bool("session-http-only", false, "session http only. Environment: "+envSessionHTTPOnly)
	sessionDomainArg := flag.String("session-domain", "", "session domain. Environment: "+envSessionDomain)
	sessionSecureArg := flag.Bool("session-secure", false, "session secure. Environment: "+envSessionSecure)
	sessionSameSiteArg := flag.String("session-same-site", "Lax", "session same site: Default, Lax, Strict or None. Environment: "+envSessionSameSite)

	return func() (*session.SessionOptions, error) {
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
			return nil, ErrInvalidSessionSameSite
		}

		return &session.SessionOptions{
			SessionKey:           useSessionKey,
			SessionName:          useSessionName,
			SessionMaxAge:        useSessionMaxAge,
			SessionHttpOnly:      useSessionHTTPOnly,
			SessionDomain:        useSessionDomain,
			SessionSecureCookies: useSessionSecure,
			SessionSameSite:      useSessionSameSite,
		}, nil
	}
}
