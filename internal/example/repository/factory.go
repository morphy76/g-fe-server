package repository

import (
	"context"
	"errors"

	impl "github.com/morphy76/g-fe-server/internal/example/impl"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/options"
	model "github.com/morphy76/g-fe-server/pkg/example"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRepository(requestContext context.Context) (model.Repository, error) {

	dbOptions := requestContext.Value(app_http.CTX_DB_OPTIONS_KEY).(*options.DbOptions)
	dbClient := requestContext.Value(app_http.CTX_DB_KEY)

	switch dbOptions.Type {
	case options.RepositoryTypeMemoryDB:
		return impl.NewMemoryRepository(), nil
	case options.RepositoryTypeMongoDB:
		if dbClient == nil {
			return nil, errors.New("MongoDB client not found in request context")
		}

		mongoClient := dbClient.(*mongo.Client)

		var rv model.Repository = &impl.MongoRepository{
			DbOptions:  dbOptions,
			Client:     mongoClient,
			UseContext: requestContext,
		}

		return rv, nil
	default:
		return nil, model.ErrUnknownRepositoryType
	}
}
