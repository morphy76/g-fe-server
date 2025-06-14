package cli

import (
	"errors"
	"flag"
	"os"

	"github.com/morphy76/g-fe-server/cmd/options"
)

// AIWOptionsBuilderFn is a function that builds AIW options from the command line arguments
type AIWOptionsBuilderFn func() (*options.AIWOptions, error)

// ErrInvalidFQDN is an invalid FQDN error
var ErrInvalidFQDN = errors.New("invalid AIW FQDN")

const (
	envFQDN = "AIW_FQDN"
)

// AIWOptionsBuilder builds AIW options from command line flags and environment variables.
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
