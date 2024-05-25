package example

import (
	impl "github.com/morphy76/g-fe-server/internal/example/impl"
	app_context "github.com/morphy76/g-fe-server/internal/http/context"
	model "github.com/morphy76/g-fe-server/pkg/example"
)

func NewRepository(dbOptions app_context.DbOptions) (model.Repository, error) {
	switch dbOptions.Type {
	case model.RepositoryTypeMemoryDB:
		return impl.NewMemoryRepository(), nil
	case model.RepositoryTypeMongoDB:
		return &impl.MongoRepository{
			Url:        dbOptions.Url,
			Username:   dbOptions.User,
			Password:   dbOptions.Password,
			Collection: dbOptions.Collection,
		}, nil
	default:
		return nil, model.ErrUnknownRepositoryType
	}
}
