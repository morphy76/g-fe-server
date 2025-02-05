//go:build with_http_session && with_mongodb

package initializers

import (
	"github.com/morphy76/g-fe-server/cmd/options"
)

func CreateSessionStore(
	sessionOptions *options.SessionOptions,
	serveOptions *options.ServeOptions,
) (options.SessionStore, error) {
	// TODO from memstore to https://github.com/kidstuff/mongostore
	return nil, nil
}
