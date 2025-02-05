package cli

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/morphy76/g-fe-server/cmd/options"
)

// PathOptionsBuilderFn is a function that returns PathOptions
type PathOptionsBuilderFn func() (*options.PathOptions, error)

// ServeOptionsBuilderFn is a function that returns ServeOptions
type ServeOptionsBuilderFn func() (*options.ServeOptions, error)

// URLOptionsBuilderFn is a function that returns URLOptions
type URLOptionsBuilderFn func() (*options.URLOptions, error)

// ErrInvalidContextRoot is an invalid context root error
var ErrInvalidContextRoot = errors.New("invalid context root")

// ErrInvalidStaticPath is an invalid static path error
var ErrInvalidStaticPath = errors.New("invalid static path")

const (
	envNonFnRoot  = "NON_FUNCTIONAL_ROOT"
	envCtxRoot    = "CONTEXT_ROOT"
	envStaticPath = "STATIC_PATH"
	envPort       = "SERVE_PORT"
	envHost       = "SERVE_HOST"
)

func PathOptionsBuilder() PathOptionsBuilderFn {
	nonFnRootArg := flag.String("non-fn", "/g", "presentation server non functional root. Environment: "+envNonFnRoot)
	ctxRootArg := flag.String("ctx", "", "presentation server context root. Environment: "+envCtxRoot)

	return func() (*options.PathOptions, error) {
		nonFnPath, found := os.LookupEnv(envNonFnRoot)
		if !found {
			nonFnPath = *nonFnRootArg
		}
		if len(nonFnPath) == 0 || strings.Contains(nonFnPath, " ") {
			return nil, ErrInvalidStaticPath
		}

		ctxRoot, found := os.LookupEnv(envCtxRoot)
		if !found {
			ctxRoot = *ctxRootArg
		}
		if len(ctxRoot) == 0 || strings.Contains(ctxRoot, " ") || !strings.HasPrefix(ctxRoot, "/") {
			return nil, ErrInvalidContextRoot
		}

		return &options.PathOptions{
			NonFunctionalRoot: nonFnPath,
			ContextRoot:       ctxRoot,
		}, nil
	}
}

// URLOptionsBuilder returns a function that builds URLOptions from the command line arguments and environment variables
func URLOptionsBuilder() URLOptionsBuilderFn {
	portArg := flag.String("port", "8080", "binding port of the presentation server. Environment: "+envPort)
	hostArg := flag.String("host", "0.0.0.0", "binding host of the presentation server. Environment: "+envHost)

	return func() (*options.URLOptions, error) {

		usePort, found := os.LookupEnv(envPort)
		if !found {
			usePort = *portArg
		}

		useHost, found := os.LookupEnv(envHost)
		if !found {
			useHost = *hostArg
		}

		return &options.URLOptions{
			Protocol: "http",
			Port:     usePort,
			Host:     useHost,
		}, nil
	}
}

// ServeOptionsBuilder returns a function that builds ServeOptions from the command line arguments and environment variables
func ServeOptionsBuilder() ServeOptionsBuilderFn {

	staticPathArg := flag.String("static", "/static", "static path of the served application. Environment: "+envStaticPath)
	pathOptionsBuilder := PathOptionsBuilder()
	urlOptionsBuilder := URLOptionsBuilder()

	return func() (*options.ServeOptions, error) {

		pathOptions, err := pathOptionsBuilder()
		if err != nil {
			return nil, err
		}

		staticPath, found := os.LookupEnv(envStaticPath)
		if !found {
			staticPath = *staticPathArg
		}
		if len(staticPath) == 0 || strings.Contains(staticPath, " ") {
			return nil, ErrInvalidStaticPath
		}

		urlOptions, err := urlOptionsBuilder()
		if err != nil {
			return nil, err
		}

		return &options.ServeOptions{
			StaticPath:  staticPath,
			PathOptions: *pathOptions,
			URLOptions:  *urlOptions,
		}, nil
	}
}
