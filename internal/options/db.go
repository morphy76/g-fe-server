package options

type RepositoryType int8

const (
	RepositoryTypeMemoryDB RepositoryType = iota
	RepositoryTypeMongoDB  RepositoryType = 1
)

type MongoDbOptions struct {
	Url      string
	User     string
	Password string
}

type DbOptions struct {
	MongoDbOptions
	Type RepositoryType
}
