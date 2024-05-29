package cli

import (
	"errors"
	"flag"
	"os"
	"strconv"

	"github.com/morphy76/g-fe-server/internal/options"
)

type dbOptionsBuidler func() (*options.DbOptions, error)

var errUnknownDbType = errors.New("unknown db type")
var errRequiredMongoDbUrl = errors.New("mongo db url is required")

func IsUnknownDbType(err error) bool {
	return err == errUnknownDbType
}

func IsRequiredMongoDbUrl(err error) bool {
	return err == errRequiredMongoDbUrl
}

const (
	ENV_DB_TYPE                = "DB_TYPE"
	ENV_DB_MONGO_URL           = "DB_MONGO_URL"
	ENV_DB_MONGO_USER          = "DB_MONGO_USER"
	ENV_DB_MONGO_PASS          = "DB_MONGO_PASSWORD"
	ENV_DB_MONG_MAX_POOL_SIZE  = "DB_MONGO_MAX_POOL_SIZE"
	ENV_DB_MONGO_MIN_POOL_SIZE = "DB_MONGO_MIN_POOL_SIZE"
)

func DbOptionsBuilder() dbOptionsBuidler {

	dbTypeArg := flag.String("db", "0", "type of the database: 0: memory - 1: mongo. Environment: "+ENV_DB_TYPE)
	dbMongoUrlArg := flag.String("db-mongo-url", "", "mongo database URL in the form of mongodb://<user>:<pass>@<host>:<port>/<db>?<args>. Environment: "+ENV_DB_MONGO_URL)
	dbMongoUserArg := flag.String("db-mongo-user", "", "mongo database username. Environment: "+ENV_DB_MONGO_USER)
	dbMongoPasswordArg := flag.String("db-mongo-password", "", "mongo database password. Environment: "+ENV_DB_MONGO_PASS)
	dbMongoMaxPoolSizeArg := flag.Uint64("db-mongo-max-pool-size", 100, "mongo database maximum pool size. Environment: "+ENV_DB_MONG_MAX_POOL_SIZE)
	dbMongoMinPoolSizeArg := flag.Uint64("db-mongo-min-pool-size", 1, "mongo database minimum pool size. Environment: "+ENV_DB_MONGO_MIN_POOL_SIZE)

	rv := func() (*options.DbOptions, error) {

		dbType, found := os.LookupEnv(ENV_DB_TYPE)
		if !found {
			dbType = *dbTypeArg
		}
		if dbType != "0" && dbType != "1" {
			return nil, errUnknownDbType
		}

		url, found := os.LookupEnv(ENV_DB_MONGO_URL)
		if !found {
			url = *dbMongoUrlArg
		}
		if url == "" && dbType == "1" {
			return nil, errRequiredMongoDbUrl
		}

		user, found := os.LookupEnv(ENV_DB_MONGO_USER)
		if !found {
			user = *dbMongoUserArg
		}

		password, found := os.LookupEnv(ENV_DB_MONGO_PASS)
		if !found {
			password = *dbMongoPasswordArg
		}

		dbTypeAsInt, err := strconv.Atoi(dbType)
		if err != nil {
			return nil, err
		}
		useDbType := options.RepositoryType(dbTypeAsInt)

		maxPoolSize, found := os.LookupEnv(ENV_DB_MONG_MAX_POOL_SIZE)
		if !found {
			maxPoolSize = strconv.FormatUint(*dbMongoMaxPoolSizeArg, 10)
		}
		maxPoolSizeAsInt, err := strconv.ParseUint(maxPoolSize, 10, 64)
		if err != nil {
			return nil, err
		}
		if maxPoolSizeAsInt <= 0 {
			maxPoolSizeAsInt = 100
		}

		minPoolSize, found := os.LookupEnv(ENV_DB_MONGO_MIN_POOL_SIZE)
		if !found {
			minPoolSize = strconv.FormatUint(*dbMongoMinPoolSizeArg, 10)
		}
		minPoolSizeAsInt, err := strconv.ParseUint(minPoolSize, 10, 64)
		if err != nil {
			return nil, err
		}
		if minPoolSizeAsInt <= 0 {
			minPoolSizeAsInt = 1
		}

		return &options.DbOptions{
			Type: useDbType,
			MongoDbOptions: options.MongoDbOptions{
				Url:         url,
				User:        user,
				Password:    password,
				MaxPoolSize: maxPoolSizeAsInt,
				MinPoolSize: minPoolSizeAsInt,
			},
		}, nil
	}

	return rv
}
