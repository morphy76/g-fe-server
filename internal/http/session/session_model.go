package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type SessionOptions struct {
	Key           string
	Name          string
	Path          string
	MaxAge        int
	HttpOnly      bool
	Domain        string
	SecureCookies bool
	SameSite      http.SameSite
}

type Session interface {
	Put(key string, value any)
	Get(key string) (any, bool)
	GetOrElse(key string, alt any) any
	Delete(key string)
	IsDirty() bool
}

type SessionWrapper struct {
	session *sessions.Session

	dirty bool
}

func NewSessionWrapper(session *sessions.Session) *SessionWrapper {
	return &SessionWrapper{session: session}
}

func (s *SessionWrapper) Put(key string, value any) {
	prev, found := s.session.Values[key]
	if found && prev == value {
		return
	}
	s.session.Values[key] = value
	s.dirty = true
}

func (s *SessionWrapper) Get(key string) (any, bool) {
	rv, found := s.session.Values[key]
	if !found {
		var zero any
		return zero, false
	}
	val, ok := rv.(any)
	if !ok {
		var zero any
		return zero, false
	}
	return val, true
}

func (s *SessionWrapper) GetOrElse(key string, alt any) any {
	rv, found := s.Get(key)
	if !found {
		return alt
	}
	return rv
}

func (s *SessionWrapper) Delete(key string) {
	_, found := s.session.Values[key]
	if !found {
		return
	}
	delete(s.session.Values, key)
	s.dirty = true
}

func (s *SessionWrapper) IsDirty() bool {
	return s.dirty
}
