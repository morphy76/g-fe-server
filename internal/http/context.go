package http

type ContextModelKey string
type ContextSessionKey string
type ContextSessionStoreKey string
type ContextDbKey string
type ContextDbOptionsKey string

const (
	CTX_CONTEXT_SERVE_KEY ContextModelKey        = "contextModel"
	CTX_SESSION_STORE_KEY ContextSessionStoreKey = "sessionStore"
	CTX_SESSION_KEY       ContextSessionKey      = "session"
	CTX_DB_KEY            ContextDbKey           = "db"
	CTX_DB_OPTIONS_KEY    ContextDbOptionsKey    = "dbOptions"
)
