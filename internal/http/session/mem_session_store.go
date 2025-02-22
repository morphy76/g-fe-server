package session

import (
	"github.com/gorilla/sessions"
	"github.com/quasoft/memstore"
)

func CreateSessionStore(
	sessionOptions *SessionOptions,
	contextRoot string,
) (SessionStore, error) {
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
