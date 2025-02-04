// go:build !with-http-session

package cli

import (
	"github.com/morphy76/g-fe-server/cmd/options"
)

// SessionOptionsBuilderFn is a function that returns SessionOptions
type SessionOptionsBuilderFn func() (*options.SessionOptions, error)

// SessionOptionsBuilder returns a function that builds SessionOptions from the command line arguments and environment variables
func SessionOptionsBuilder() SessionOptionsBuilderFn {
	return func() (interface{}, error) {
		return nil, nil
	}
}
