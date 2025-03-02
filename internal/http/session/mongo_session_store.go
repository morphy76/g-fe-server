package session

import (
	"context"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/morphy76/g-fe-server/internal/db"
	"github.com/morphy76/g-fe-server/internal/http/session/mongostore"
)

func CreateSessionStore(
	sessionOptions *SessionOptions,
	dbOptions *options.MongoDBOptions,
	contextRoot string,
) (sessions.Store, func() error, error) {

	useURL, err := url.Parse(dbOptions.URL)
	if err != nil {
		return nil, nil, err
	}

	if useURL.User == nil {
		useCredentials := url.UserPassword(dbOptions.User, dbOptions.Password)
		useURL.User = useCredentials
	}
	client, err := db.NewClient(dbOptions, false)
	if err != nil {
		return nil, nil, err
	}

	store := mongostore.NewMongoStore(
		client.Database(useURL.Path).Collection(sessionOptions.SessionName),
		sessionOptions.SessionMaxAge,
		true,
		[]byte(sessionOptions.SessionKey),
	)

	store.Options = &sessions.Options{
		Path:     contextRoot,
		MaxAge:   sessionOptions.SessionMaxAge,
		HttpOnly: sessionOptions.SessionHttpOnly,
		Domain:   sessionOptions.SessionDomain,
		Secure:   sessionOptions.SessionSecureCookies,
		SameSite: sessionOptions.SessionSameSite,
	}

	shutdownFunc := func() error {
		return client.Disconnect(context.Background())
	}

	return store, shutdownFunc, nil
}
