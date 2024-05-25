package context

import "github.com/morphy76/g-fe-server/pkg/example"

type ContextModelKey string
type ContextRepositoryKey string

type ServeOptions struct {
	ContextRoot string
	StaticPath  string
	Port        string
	Host        string
}

type MongoDbOptions struct {
	Url        string
	User       string
	Password   string
	Collection string
}

type DbOptions struct {
	MongoDbOptions
	Type example.RepositoryType
}

const (
	CTX_CONTEXT_ROOT_KEY ContextModelKey      = "contextModel"
	CTX_REPOSITORY_KEY   ContextRepositoryKey = "repository"
)
