package example

import (
	impl "github.com/morphy76/g-fe-server/internal/example/impl"
	"github.com/morphy76/g-fe-server/internal/options"
	model "github.com/morphy76/g-fe-server/pkg/example"
)

func NewRepository(dbOptions *options.DbOptions) (model.Repository, error) {
	switch dbOptions.Type {
	case options.RepositoryTypeMemoryDB:
		return impl.NewMemoryRepository(), nil
	case options.RepositoryTypeMongoDB:
		return &impl.MongoRepository{
			Url:      dbOptions.Url,
			Username: dbOptions.User,
			Password: dbOptions.Password,
		}, nil
	default:
		return nil, model.ErrUnknownRepositoryType
	}
}
