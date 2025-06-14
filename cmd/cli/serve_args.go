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
	envProtocol   = "SERVE_PROTOCOL"
	envCertFile   = "SERVE_CERT_FILE"
	envKeyFile    = "SERVE_KEY_FILE"
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
	protocolArg := flag.String("protocol", "http", "protocol of the presentation server. Environment: "+envProtocol)
	certFileArg := flag.String("tls-cert", "", "path to the TLS certificate file. Environment: "+envCertFile)
	keyFileArg := flag.String("tls-key", "", "path to the TLS key file. Environment: "+envKeyFile)

	return func() (*options.URLOptions, error) {

		usePort, found := os.LookupEnv(envPort)
		if !found {
			usePort = *portArg
		}

		useHost, found := os.LookupEnv(envHost)
		if !found {
			useHost = *hostArg
		}

		useProtocol, found := os.LookupEnv(envProtocol)
		if !found {
			useProtocol = *protocolArg
		}

		useCertFile, found := os.LookupEnv(envCertFile)
		if !found {
			useCertFile = *certFileArg
		}

		useKeyFile, found := os.LookupEnv(envKeyFile)
		if !found {
			useKeyFile = *keyFileArg
		}

		if useProtocol != "http" && useProtocol != "https" {
			return nil, errors.New("invalid protocol: must be 'http' or 'https'")
		}

		if useProtocol == "https" && (len(useCertFile) == 0 || len(useKeyFile) == 0) {
			return nil, errors.New("cert and key files must be provided for https protocol")
		}

		return &options.URLOptions{
			Protocol: useProtocol,
			Port:     usePort,
			Host:     useHost,
			CertFile: useCertFile,
			KeyFile:  useKeyFile,
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
		if len(staticPath) == 0 {
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
