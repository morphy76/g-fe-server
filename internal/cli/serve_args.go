package cli

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/google/uuid"
	app_context "github.com/morphy76/g-fe-server/internal/http/context"
)

type serveOptionsBuilder func() (app_context.ServeOptions, error)

func ServeOptionsBuilder() serveOptionsBuilder {

	ctxRootArg := flag.String("ctx", "", "presentation server context root")
	staticPathArg := flag.String("static", "/static", "static path of the served application")
	portArg := flag.String("port", "8080", "binding port of the presentation server")
	hostArg := flag.String("host", "0.0.0.0", "binding host of the presentation server")
	sessionKeyArg := flag.String("session-key", "", "session key")

	rv := func() (app_context.ServeOptions, error) {

		ctxRoot, found := os.LookupEnv("CONTEXT_ROOT")
		if !found {
			ctxRoot = *ctxRootArg
		}
		if len(ctxRoot) == 0 || strings.Contains(ctxRoot, " ") || !strings.HasPrefix(ctxRoot, "/") {
			return app_context.ServeOptions{}, errors.New("invalid context root")
		}

		staticPath, found := os.LookupEnv("STATIC_PATH")
		if !found {
			staticPath = *staticPathArg
		}
		if len(staticPath) == 0 || strings.Contains(staticPath, " ") {
			return app_context.ServeOptions{}, errors.New("invalid static path")
		}

		usePort, found := os.LookupEnv("PORT")
		if !found {
			usePort = *portArg
		}

		useHost, found := os.LookupEnv("HOST")
		if !found {
			useHost = *hostArg
		}

		useSessionKey, found := os.LookupEnv("SESSION_KEY")
		if !found {
			useSessionKey = *sessionKeyArg
		}
		if len(useSessionKey) == 0 {
			useSessionKey = uuid.New().String()
		}

		return app_context.ServeOptions{
			ContextRoot: ctxRoot,
			StaticPath:  staticPath,
			Port:        usePort,
			Host:        useHost,
			SessionKey:  useSessionKey,
		}, nil
	}

	return rv
}
