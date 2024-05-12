package example

import (
	impl "g-fe-server/internal/example/impl"
	app_context "g-fe-server/internal/http/context"
	model "g-fe-server/pkg/example"
)

const (
	RepositoryTypeMemoryDB model.RepositoryType = iota
	RepositoryTypeMongoDB  model.RepositoryType = 1
)

func NewRepository(dbModel app_context.DbModel) (model.Repository, error) {
	switch dbModel.Type {
	case RepositoryTypeMemoryDB:
		return impl.NewMemoryRepository(), nil
	case RepositoryTypeMongoDB:
		return &impl.MongoRepository{
			Uri:  dbModel.Uri,
			Db:   dbModel.Db,
			Coll: dbModel.Collection,
		}, nil
	default:
		return nil, model.ErrUnknownRepositoryType
	}
}
