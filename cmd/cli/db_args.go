package cli

import (
	"errors"
	"flag"
	"os"

	"github.com/morphy76/g-fe-server/cmd/options"
)

// DBOptionsBuidlerFn is a function that builds a DbOptions object from the command line arguments
type DBOptionsBuidlerFn func() (*options.MongoDBOptions, error)

// ErrRequiredMongoDbUrl is a required mongo db url error
var ErrRequiredMongoDbUrl = errors.New("mongo db url is required")

// IsRequiredMongoDbUrl returns true if the error is a required mongo db url error
func IsRequiredMongoDbUrl(err error) bool {
	return err == ErrRequiredMongoDbUrl
}

const (
	envDBMongoURL      = "DB_MONGO_URL"
	envDBMongoUser     = "DB_MONGO_USER"
	envDBMongoPassword = "DB_MONGO_PASSWORD"
)

// DBOptionsBuilder returns a function that builds a DbOptions object from the command line arguments
func DBOptionsBuilder() DBOptionsBuidlerFn {

	dbMongoUrlArg := flag.String("db-mongo-url", "", "mongo database URL in the form of mongodb://<user>:<pass>@<host>:<port>/<db>?<args>. Environment: "+envDBMongoURL)
	dbMongoUserArg := flag.String("db-mongo-user", "", "mongo database username. Environment: "+envDBMongoUser)
	dbMongoPasswordArg := flag.String("db-mongo-password", "", "mongo database password. Environment: "+envDBMongoPassword)

	rv := func() (*options.MongoDBOptions, error) {

		url, found := os.LookupEnv(envDBMongoURL)
		if !found {
			url = *dbMongoUrlArg
		}
		if url == "" {
			return nil, ErrRequiredMongoDbUrl
		}

		user, found := os.LookupEnv(envDBMongoUser)
		if !found {
			user = *dbMongoUserArg
		}

		password, found := os.LookupEnv(envDBMongoPassword)
		if !found {
			password = *dbMongoPasswordArg
		}

		return &options.MongoDBOptions{
			URL:      url,
			User:     user,
			Password: password,
		}, nil
	}

	return rv
}
