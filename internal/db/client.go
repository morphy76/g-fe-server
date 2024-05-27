package db

import (
	"context"
	"errors"
	"net/url"

	"github.com/morphy76/g-fe-server/internal/options"
	"go.mongodb.org/mongo-driver/mongo"
	mongo_opts "go.mongodb.org/mongo-driver/mongo/options"
)

type DbClient any
type MemoryDbClient struct{}

var ErrMissingDbOptions = errors.New("missing db options")

var ErrUnknownDbType = errors.New("unknown database type")

func IsMissingDbOptions(err error) bool {
	return err == ErrMissingDbOptions
}

func IsUnknownDbType(err error) bool {
	return err == ErrUnknownDbType
}

func NewClient(dbOptions *options.DbOptions) (DbClient, error) {
	if dbOptions == nil {
		return nil, ErrMissingDbOptions
	} else if dbOptions.Type == options.RepositoryTypeMemoryDB {
		var rv *MemoryDbClient = new(MemoryDbClient)
		return rv, nil
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
		return nil, ErrUnknownDbType
	}
}
