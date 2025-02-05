//go:build with_http_session

package session

import (
	"context"

	"github.com/gorilla/sessions"
)

type ContextSessionKey string

const (
	ctxSessionKey ContextSessionKey = "session"
)

func ExtractSession(ctx context.Context) *sessions.Session {
	return ctx.Value(ctxSessionKey).(*sessions.Session)
}

func InjectSession(ctx context.Context, session *sessions.Session) context.Context {
	return context.WithValue(ctx, ctxSessionKey, session)
}
