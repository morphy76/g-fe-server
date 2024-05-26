package cli

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/morphy76/g-fe-server/internal/options"
)

type serveOptionsBuilder func() (*options.ServeOptions, error)

var errInvalidContextRoot = errors.New("invalid context root")
var errInvalidStaticPath = errors.New("invalid static path")

func IsInvalidContextRoot(err error) bool {
	return err == errInvalidContextRoot
}

func IsInvalidStaticPath(err error) bool {
	return err == errInvalidStaticPath
}

const (
	ENV_CONTEXT_ROOT = "CONTEXT_ROOT"
	ENV_STATIC_PATH  = "STATIC_PATH"
	ENV_PORT         = "PORT"
	ENV_HOST         = "HOST"
	ENV_SESSION_KEY  = "SESSION_KEY"
)

func ServeOptionsBuilder() serveOptionsBuilder {

	ctxRootArg := flag.String("ctx", "", "presentation server context root. Environment: "+ENV_CONTEXT_ROOT)
	staticPathArg := flag.String("static", "/static", "static path of the served application. Environment: "+ENV_STATIC_PATH)
	portArg := flag.String("port", "8080", "binding port of the presentation server. Environment: "+ENV_PORT)
	hostArg := flag.String("host", "0.0.0.0", "binding host of the presentation server. Environment: "+ENV_HOST)
	sessionKeyArg := flag.String("session-key", "", "session key. Environment: "+ENV_SESSION_KEY)

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
			useSessionKey = uuid.New().String()
		}

		return &options.ServeOptions{
			ContextRoot: ctxRoot,
			StaticPath:  staticPath,
			Port:        usePort,
			Host:        useHost,
			SessionKey:  useSessionKey,
		}, nil
	}

	return rv
}
