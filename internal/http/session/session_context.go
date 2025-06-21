package session

import (
	"context"
)

// ContextSessionKey is a type for the context key used to store the session.
type ContextSessionKey string

const (
	ctxSessionKey ContextSessionKey = "session"
)

// ExtractSession extracts the session from the context.
func ExtractSession(ctx context.Context) (Session, bool) {
	session, ok := ctx.Value(ctxSessionKey).(Session)
	return session, ok
}

// InjectSession injects the session into the context.
func InjectSession(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, ctxSessionKey, session)
}
