package repository

import (
	"context"
	"errors"

	"github.com/morphy76/g-fe-server/internal/db"
	impl "github.com/morphy76/g-fe-server/internal/example/repository/impl"
	"github.com/morphy76/g-fe-server/internal/options"
	model "github.com/morphy76/g-fe-server/pkg/example"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRepository(requestContext context.Context) (model.Repository, error) {

	dbOptions := db.ExtractDbOptions(requestContext)
	dbClient := db.ExtractDb(requestContext)

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
