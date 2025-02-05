//go:build !with_oidc

package cli

import (
	"github.com/morphy76/g-fe-server/internal/auth"
)

// OIDCOptionsBuidlerFn is a function that returns OIDC options
type OIDCOptionsBuidlerFn func() (*auth.OIDCOptions, error)

// OIDCOptionsBuilder returns a function that can be used to build OIDC options
func OIDCOptionsBuilder() OIDCOptionsBuidlerFn {
	return func() (*auth.OIDCOptions, error) {
		return &auth.OIDCOptions{}, nil
	}
}
