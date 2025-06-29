package cli

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/securecookie"
	"github.com/morphy76/g-fe-server/cmd/options"
)

// SessionOptionsBuilderFn is a function that returns SessionOptions
type SessionOptionsBuilderFn func() (*options.SessionOptions, error)

// ErrInvalidSessionSameSite is an invalid session same site error
var ErrInvalidSessionSameSite = errors.New("invalid session same site")

const (
	envSessionKey         = "SESSION_KEY"
	envSessionName        = "SESSION_NAME"
	envSessionMaxAge      = "SESSION_MAX_AGE"
	envSessionDomain      = "SESSION_DOMAIN"
	envSessionSecure      = "SESSION_SECURE"
	envSessionSameSite    = "SESSION_SAME_SITE"
	envSessionPartitioned = "SESSION_PARTITIONED"
)

// SessionOptionsBuilder returns a function that builds SessionOptions from the command line arguments and environment variables
func SessionOptionsBuilder() SessionOptionsBuilderFn {

	sessionKeyArg := flag.String("session-key", "", "session key. Environment: "+envSessionKey)
	sessionNameArg := flag.String("session-name", "gofe_sid", "session name. Environment: "+envSessionName)
	sessionMaxAgeArg := flag.Int("session-max-age", 0, "session max age. Environment: "+envSessionMaxAge)
	sessionDomainArg := flag.String("session-domain", "", "session domain. Environment: "+envSessionDomain)
	sessionSecureArg := flag.Bool("session-secure", true, "session secure. Environment: "+envSessionSecure)
	sessionSameSiteArg := flag.String("session-same-site", "Lax", "session same site: Default, Lax, Strict or None. Environment: "+envSessionSameSite)
	sessionPartitionedArg := flag.Bool("session-partitioned", false, "session partitioned. Environment: "+envSessionPartitioned)

	return func() (*options.SessionOptions, error) {
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
			useSessionName = "gofe_sid"
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
		switch strSessionSameSite {
		case "Lax":
			useSessionSameSite = http.SameSiteLaxMode
		case "Strict":
			useSessionSameSite = http.SameSiteStrictMode
		case "None":
			useSessionSameSite = http.SameSiteNoneMode
		case "Default":
			useSessionSameSite = http.SameSiteDefaultMode
		default:
			return nil, ErrInvalidSessionSameSite
		}

		var useSessionPartitioned bool
		strSessionPartitioned, found := os.LookupEnv(envSessionPartitioned)
		if !found {
			useSessionPartitioned = *sessionPartitionedArg
		} else {
			useSessionPartitioned = strSessionPartitioned == "true"
		}

		return &options.SessionOptions{
			Key:           useSessionKey,
			Name:          useSessionName,
			MaxAge:        useSessionMaxAge,
			Domain:        useSessionDomain,
			SecureCookies: useSessionSecure,
			SameSite:      useSessionSameSite,
			Partitioned:   useSessionPartitioned,
		}, nil
	}
}
