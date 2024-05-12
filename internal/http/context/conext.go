package context

import "g-fe-server/pkg/example"

type ContxtModelKey string
type ContextRepositoryKey string

type ContextModel struct {
	ContextRoot string
	StaticPath  string
}

type DbModel struct {
	Type       example.RepositoryType
	Uri        string
	Db         string
	Collection string
}

const (
	CTX_CONTEXT_ROOT_KEY ContxtModelKey       = "contextModel"
	CTX_REPOSITORY_KEY   ContextRepositoryKey = "repository"
)
