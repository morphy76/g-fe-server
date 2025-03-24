package session

import (
	"context"
)

type ContextSessionKey string

const (
	ctxSessionKey ContextSessionKey = "session"
)

func ExtractSession(ctx context.Context) (Session, bool) {
	session, ok := ctx.Value(ctxSessionKey).(Session)
	return session, ok
}

func InjectSession(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, ctxSessionKey, session)
}
