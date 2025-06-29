package session

import (
	"context"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/morphy76/g-fe-server/internal/db"
	"github.com/morphy76/g-fe-server/internal/http/session/mongostore"
)

// CreateSessionStore initializes a MongoDB session store with the provided options and returns it along with a shutdown function.
func CreateSessionStore(
	sessionOptions *options.SessionOptions,
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
	client, dbName, err := db.NewClient(dbOptions, false)
	if err != nil {
		return nil, nil, err
	}

	useOptions := &sessions.Options{
		Path:        sessionOptions.Path,
		MaxAge:      sessionOptions.MaxAge,
		HttpOnly:    true,
		Domain:      sessionOptions.Domain,
		Secure:      sessionOptions.SecureCookies,
		SameSite:    sessionOptions.SameSite,
		Partitioned: sessionOptions.Partitioned,
	}

	store := mongostore.NewMongoStore(
		client.Database(dbName).Collection("http_sessions"),
		useOptions,
		[]byte(sessionOptions.Key),
	)

	shutdownFunc := func() error {
		return client.Disconnect(context.Background())
	}

	return store, shutdownFunc, nil
}
