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

type AIWOptionsBuilderFn func() (*options.AIWOptions, error)

// ErrInvalidContextRoot is an invalid context root error
var ErrInvalidContextRoot = errors.New("invalid context root")

// ErrInvalidStaticPath is an invalid static path error
var ErrInvalidStaticPath = errors.New("invalid static path")

// ErrInvalidFQDN is an invalid FQDN error
var ErrInvalidFQDN = errors.New("invalid AIW FQDN")

const (
	envNonFnRoot  = "NON_FUNCTIONAL_ROOT"
	envCtxRoot    = "CONTEXT_ROOT"
	envStaticPath = "STATIC_PATH"
	envPort       = "SERVE_PORT"
	envHost       = "SERVE_HOST"
	envFQDN       = "AIW_FQDN"
)

func PathOptionsBuilder() PathOptionsBuilderFn {
	nonFnRootArg := flag.String("non-fn", "/g", "presentation server non functional root. Environment: "+envNonFnRoot)
	ctxRootArg := flag.String("ctx", "", "presentation server context root. Environment: "+envCtxRoot)

	return func() (*options.PathOptions, error) {
		nonFnPath, found := os.LookupEnv(envNonFnRoot)
		if !found {
			nonFnPath = *nonFnRootArg
		}
		if len(nonFnPath) == 0 {
			return nil, ErrInvalidStaticPath
		}

		ctxRoot, found := os.LookupEnv(envCtxRoot)
		if !found {
			ctxRoot = *ctxRootArg
		}
		if len(ctxRoot) == 0 || !strings.HasPrefix(ctxRoot, "/") {
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

func AIWOptionsBuilder() AIWOptionsBuilderFn {
	fqdnArg := flag.String("aiw-fqdn", "", "fully qualified domain name of the application. Environment: "+envFQDN)

	return func() (*options.AIWOptions, error) {
		fqdn, found := os.LookupEnv(envFQDN)
		if !found {
			fqdn = *fqdnArg
		}
		if len(fqdn) == 0 {
			return nil, ErrInvalidFQDN
		}

		return &options.AIWOptions{
			FQDN: fqdn,
		}, nil
	}
}

// ServeOptionsBuilder returns a function that builds ServeOptions from the command line arguments and environment variables
func ServeOptionsBuilder() ServeOptionsBuilderFn {

	staticPathArg := flag.String("static", "/static", "static path of the served application. Environment: "+envStaticPath)
	pathOptionsBuilder := PathOptionsBuilder()
	urlOptionsBuilder := URLOptionsBuilder()
	aiwOptionsBuilder := AIWOptionsBuilder()

	return func() (*options.ServeOptions, error) {

		pathOptions, err := pathOptionsBuilder()
		if err != nil {
			return nil, err
		}

		staticPath, found := os.LookupEnv(envStaticPath)
		if !found {
			staticPath = *staticPathArg
		}
		if len(staticPath) == 0 {
			return nil, ErrInvalidStaticPath
		}

		urlOptions, err := urlOptionsBuilder()
		if err != nil {
			return nil, err
		}

		aiwOptions, err := aiwOptionsBuilder()
		if err != nil {
			return nil, err
		}

		return &options.ServeOptions{
			StaticPath:  staticPath,
			PathOptions: *pathOptions,
			URLOptions:  *urlOptions,
			AIWOptions:  *aiwOptions,
		}, nil
	}
}
