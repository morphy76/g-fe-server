package db

import (
	"context"
	"errors"
	"net/url"

	"github.com/morphy76/g-fe-server/internal/options"
	"go.mongodb.org/mongo-driver/mongo"
	mongo_opts "go.mongodb.org/mongo-driver/mongo/options"
)

type DbClient interface{}

var ErrNilDbOptions = errors.New("unknown repository type")

func IsNilDbOptions(err error) bool {
	return err == ErrNilDbOptions
}

func NewClient(dbOptions *options.DbOptions) (DbClient, error) {
	if dbOptions == nil {
		return nil, ErrNilDbOptions
	} else if dbOptions.Type == options.RepositoryTypeMemoryDB {
		return nil, nil
	} else if dbOptions.Type == options.RepositoryTypeMongoDB {
		var clientOpts *mongo_opts.ClientOptions
		serverAPI := mongo_opts.ServerAPI(mongo_opts.ServerAPIVersion1)

		useUrl, err := url.Parse(dbOptions.Url)
		if err != nil {
			return nil, err
		}

		if useUrl.User == nil {
			useCredentials := url.UserPassword(dbOptions.User, dbOptions.Password)
			useUrl.User = useCredentials
		}

		clientOpts = mongo_opts.Client().
			ApplyURI(useUrl.String()).
			SetServerAPIOptions(serverAPI)

		mongoClient, err := mongo.Connect(context.Background(), clientOpts)
		if err != nil {
			return nil, err
		}

		return mongoClient, nil
	} else {
		return nil, ErrNilDbOptions
	}
}
