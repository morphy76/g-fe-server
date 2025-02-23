package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type SessionOptions struct {
	SessionKey           string
	SessionName          string
	SessionMaxAge        int
	SessionHttpOnly      bool
	SessionDomain        string
	SessionSecureCookies bool
	SessionSameSite      http.SameSite
}

type Session interface {
	Put(key string, value interface{})
	Get(key string) interface{}
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

func (s *SessionWrapper) Put(key string, value interface{}) {
	prev, found := s.session.Values[key]
	if found && prev == value {
		return
	}
	s.session.Values[key] = value
	s.dirty = true
}

func (s *SessionWrapper) Get(key string) interface{} {
	return s.session.Values[key]
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
