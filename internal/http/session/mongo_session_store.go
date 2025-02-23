package session

import (
	"net/url"

	"github.com/globalsign/mgo"
	"github.com/gorilla/sessions"
	"github.com/kidstuff/mongostore"
	"github.com/morphy76/g-fe-server/cmd/options"
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

	dbSession, err := mgo.Dial(useURL.String())
	if err != nil {
		return nil, nil, err
	}

	store := mongostore.NewMongoStore(
		dbSession.DB(dbOptions.Database).C(sessionOptions.SessionName),
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
		dbSession.Close()
		return nil
	}

	return store, shutdownFunc, nil
}
