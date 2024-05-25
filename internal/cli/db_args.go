package cli

import (
	app_context "github.com/morphy76/g-fe-server/internal/http/context"
	model "github.com/morphy76/g-fe-server/pkg/example"

	"flag"
	"os"
)

type DbModelBuidler func() app_context.DbModel

func BuildDbModel() (rv DbModelBuidler) {

	dbMongoUrlArg := flag.String("db-mongo-url", "", "MongoDB URL")
	dbMongoUserArg := flag.String("db-mongo-user", "", "MongoDB User")
	dbMongoPasswordArg := flag.String("db-mongo-password", "", "MongoDB Password")

	rv = func() app_context.DbModel {

		url, found := os.LookupEnv("DB_MONGO_URL")
		if !found {
			url = *dbMongoUrlArg
		}

		user, found := os.LookupEnv("DB_MONGO_USER")
		if !found {
			user = *dbMongoUserArg
		}

		password, found := os.LookupEnv("DB_MONGO_PASSWORD")
		if !found {
			password = *dbMongoPasswordArg
		}

		return app_context.DbModel{
			// TODO change to use mongodb
			Type:     model.RepositoryTypeMemoryDB,
			Url:      url,
			User:     user,
			Password: password,
		}
	}

	return rv
}
