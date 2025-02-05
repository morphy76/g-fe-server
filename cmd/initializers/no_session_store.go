//go:build !with_http_session

package initializers

import "github.com/morphy76/g-fe-server/cmd/options"

func CreateSessionStore(
	sessionOptions *options.SessionOptions,
	serveOptions *options.ServeOptions,
) (options.SessionStore, error) {
	return nil, nil
}
