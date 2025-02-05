//go:build with_http_session && !with_mongodb

package session

import (
	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/quasoft/memstore"
)

func CreateSessionStore(
	sessionOptions *options.SessionOptions,
	contextRoot string,
) (options.SessionStore, error) {
	sessionStore := memstore.NewMemStore([]byte(sessionOptions.SessionKey))
	sessionStore.Options = &sessions.Options{
		Path:     contextRoot,
		MaxAge:   sessionOptions.SessionMaxAge,
		HttpOnly: sessionOptions.SessionHttpOnly,
		Domain:   sessionOptions.SessionDomain,
		Secure:   sessionOptions.SessionSecureCookies,
		SameSite: sessionOptions.SessionSameSite,
	}
	return sessionStore, nil
}
