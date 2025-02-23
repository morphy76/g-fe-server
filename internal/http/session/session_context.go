package session

import (
	"context"
)

type ContextSessionKey string

const (
	ctxSessionKey ContextSessionKey = "session"
)

func ExtractSession(ctx context.Context) Session {
	return ctx.Value(ctxSessionKey).(Session)
}

func InjectSession(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, ctxSessionKey, session)
}
