//go:build !with_http_session

package options

import "github.com/gorilla/sessions"

type SessionStore interface {
	sessions.Store
}

type SessionOptions *interface{}
