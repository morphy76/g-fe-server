package http

type ContextModelKey string
type ContextRepositoryKey string
type ContextSessionKey string

const (
	CTX_CONTEXT_ROOT_KEY ContextModelKey      = "contextModel"
	CTX_REPOSITORY_KEY   ContextRepositoryKey = "repository"
	CTX_SESSION_KEY      ContextSessionKey    = "session"
)
