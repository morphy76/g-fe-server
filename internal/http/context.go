package http

type ContextModelKey string
type ContextRepositoryKey string
type ContextSessionKey string
type ContextSessionStoreKey string

const (
	CTX_CONTEXT_SERVE_KEY ContextModelKey        = "contextModel"
	CTX_REPOSITORY_KEY    ContextRepositoryKey   = "repository"
	CTX_SESSION_STORE_KEY ContextSessionStoreKey = "sessionStore"
	CTX_SESSION_KEY       ContextSessionKey      = "session"
)
