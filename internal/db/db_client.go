package db

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/morphy76/g-fe-server/cmd/options"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mongo_opts "go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ErrMissingDBOptions is returned when the db options are missing
var ErrMissingDBOptions = errors.New("missing db options")

// NewClient creates a new db client based on the db options
func NewClient(dbOptions *options.MongoDBOptions, withMonitor bool) (*mongo.Client, error) {
	if dbOptions == nil {
		return nil, ErrMissingDBOptions
	} else {
		serverAPI := mongo_opts.ServerAPI(mongo_opts.ServerAPIVersion1)

		useURL, err := url.Parse(dbOptions.URL)
		if err != nil {
			return nil, err
		}

		if useURL.User == nil {
			useCredentials := url.UserPassword(dbOptions.User, dbOptions.Password)
			useURL.User = useCredentials
		}

		clientOpts := mongo_opts.Client().
			ApplyURI(useURL.String()).
			SetServerAPIOptions(serverAPI).
			SetHTTPClient(instrumentNewHTTPClient())
		if withMonitor {
			clientOpts = clientOpts.SetMonitor(NewMonitor())
			clientOpts.SetPoolMonitor(NewPoolMonitor())
		}

		mongoClient, err := mongo.Connect(clientOpts)
		if err != nil {
			return nil, err
		}

		return mongoClient, nil
	}
}

func instrumentNewHTTPClient() *http.Client {
	transport := otelhttp.NewTransport(http.DefaultTransport)
	client := &http.Client{
		Transport: transport,
	}
	return client
}
