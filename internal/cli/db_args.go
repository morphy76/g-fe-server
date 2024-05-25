package cli

import (
	"errors"
	"flag"
	"os"
	"strconv"

	app_context "github.com/morphy76/g-fe-server/internal/http/context"
	model "github.com/morphy76/g-fe-server/pkg/example"
)

type dbOptionsBuidler func() (app_context.DbOptions, error)

func DbOptionsBuilder() dbOptionsBuidler {

	dbTypeArg := flag.String("db", "0", "type of the database: 0: memory - 1: mongo")
	dbMongoUrlArg := flag.String("db-mongo-url", "", "mongo database URL in the form of mongodb://<user>:<pass>@<host>:<port>/<db>?<args>")
	dbMongoUserArg := flag.String("db-mongo-user", "", "mongo database username")
	dbMongoPasswordArg := flag.String("db-mongo-password", "", "mongo database password")
	dbMongoCollectionArg := flag.String("db-mongo-collection", "examples", "mongo collection to use")

	rv := func() (app_context.DbOptions, error) {

		dbType, found := os.LookupEnv("DB_TYPE")
		if !found {
			dbType = *dbTypeArg
		}
		if dbType != "0" && dbType != "1" {
			return app_context.DbOptions{}, errors.New("invalid db type")
		}

		url, found := os.LookupEnv("DB_MONGO_URL")
		if !found {
			url = *dbMongoUrlArg
		}
		if url == "" && dbType == "1" {
			return app_context.DbOptions{}, errors.New("mongo db url is required")
		}

		user, found := os.LookupEnv("DB_MONGO_USER")
		if !found {
			user = *dbMongoUserArg
		}

		password, found := os.LookupEnv("DB_MONGO_PASSWORD")
		if !found {
			password = *dbMongoPasswordArg
		}

		collection, found := os.LookupEnv("DB_MONGO_COLLECTION")
		if !found {
			collection = *dbMongoCollectionArg
		}

		dbTypeAsInt, err := strconv.Atoi(dbType)
		if err != nil {
			return app_context.DbOptions{}, err
		}
		useDbType := model.RepositoryType(dbTypeAsInt)

		return app_context.DbOptions{
			// TODO change to use mongodb
			Type: useDbType,
			MongoDbOptions: app_context.MongoDbOptions{
				Url:        url,
				User:       user,
				Password:   password,
				Collection: collection,
			},
		}, nil
	}

	return rv
}
