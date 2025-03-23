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

	dbName, err := extractDBNameFromURL(useURL)
	if err != nil {
		return nil, nil, err
	}
	store := mongostore.NewMongoStore(
		client.Database(dbName).Collection("http_sessions"),
		sessionOptions.MaxAge,
		true,
		[]byte(sessionOptions.Key),
	)

	store.Options = &sessions.Options{
		Path:     sessionOptions.Path,
		MaxAge:   sessionOptions.MaxAge,
		HttpOnly: sessionOptions.HttpOnly,
		Domain:   sessionOptions.Domain,
		Secure:   sessionOptions.SecureCookies,
		SameSite: sessionOptions.SameSite,
	}

	shutdownFunc := func() error {
		return client.Disconnect(context.Background())
	}

	return store, shutdownFunc, nil
}

func extractDBNameFromURL(useURL *url.URL) (string, error) {
	dbName := useURL.Path
	if len(dbName) > 1 {
		dbName = dbName[1:]
	}
	return dbName, nil
}
