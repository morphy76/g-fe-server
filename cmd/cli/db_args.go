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

// ErrRequiredMongoDbName is a required mongo db name error
var ErrRequiredMongoDbName = errors.New("mongo db name is required")

const (
	envDBMongoURL      = "DB_MONGO_URL"
	envDBMongoDatabase = "DB_MONGO_DATABASE"
	envDBMongoUser     = "DB_MONGO_USER"
	envDBMongoPassword = "DB_MONGO_PASSWORD"
)

// DBOptionsBuilder returns a function that builds a DbOptions object from the command line arguments
func DBOptionsBuilder() DBOptionsBuidlerFn {

	dbMongoUrlArg := flag.String("db-mongo-url", "", "mongo database URL in the form of mongodb://<user>:<pass>@<host>:<port>/<db>?<args>. Environment: "+envDBMongoURL)
	dbMongoUserArg := flag.String("db-mongo-user", "", "mongo database username. Environment: "+envDBMongoUser)
	dbMongoPasswordArg := flag.String("db-mongo-password", "", "mongo database password. Environment: "+envDBMongoPassword)
	dbMongoDatabaseArg := flag.String("db-mongo-database", "", "mongo database name. Environment: "+envDBMongoDatabase)

	rv := func() (*options.MongoDBOptions, error) {

		url, found := os.LookupEnv(envDBMongoURL)
		if !found {
			url = *dbMongoUrlArg
		}
		if url == "" {
			return nil, ErrRequiredMongoDbUrl
		}

		db, found := os.LookupEnv(envDBMongoDatabase)
		if !found {
			db = *dbMongoDatabaseArg
		}
		if db == "" {
			return nil, ErrRequiredMongoDbName
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
			Database: db,
		}, nil
	}

	return rv
}
