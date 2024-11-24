package db

import (
	"context"
	"errors"
	"net/url"

	"github.com/morphy76/g-fe-server/internal/options"
	"go.mongodb.org/mongo-driver/mongo"
	mongo_opts "go.mongodb.org/mongo-driver/mongo/options"
)

// ErrMissingDBOptions is returned when the db options are missing
var ErrMissingDBOptions = errors.New("missing db options")

// IsMissingDBOptions returns true if the error is ErrMissingDBOptions
func IsMissingDBOptions(err error) bool {
	return err == ErrMissingDBOptions
}

// NewClient creates a new db client based on the db options
func NewClient(dbOptions *options.MongoDBOptions) (*mongo.Client, error) {
	if dbOptions == nil {
		return nil, ErrMissingDBOptions
	} else {
		var clientOpts *mongo_opts.ClientOptions
		serverAPI := mongo_opts.ServerAPI(mongo_opts.ServerAPIVersion1)

		useURL, err := url.Parse(dbOptions.URL)
		if err != nil {
			return nil, err
		}

		if useURL.User == nil {
			useCredentials := url.UserPassword(dbOptions.User, dbOptions.Password)
			useURL.User = useCredentials
		}

		clientOpts = mongo_opts.Client().
			ApplyURI(useURL.String()).
			SetServerAPIOptions(serverAPI)

		mongoClient, err := mongo.Connect(context.Background(), clientOpts)
		if err != nil {
			return nil, err
		}

		return mongoClient, nil
	}
}
