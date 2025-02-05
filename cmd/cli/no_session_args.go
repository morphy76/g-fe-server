//go:build !with_http_session

package cli

import (
	"github.com/morphy76/g-fe-server/internal/http/session"
)

// SessionOptionsBuilderFn is a function that returns SessionOptions
type SessionOptionsBuilderFn func() (*session.SessionOptions, error)

// SessionOptionsBuilder returns a function that builds SessionOptions from the command line arguments and environment variables
func SessionOptionsBuilder() SessionOptionsBuilderFn {
	return func() (*session.SessionOptions, error) {
		return &session.SessionOptions{}, nil
	}
}
