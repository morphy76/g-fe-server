//go:build !with_http_session

package session

import "github.com/gorilla/sessions"

type SessionStore interface {
	sessions.Store
}

type SessionOptions struct {
	SessionName string
}
