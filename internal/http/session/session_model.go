package session

import (
	"github.com/gorilla/sessions"
)

var staticSessionAssert Session = (*Wrapper)(nil)

// Session interface defines the methods for session management.
type Session interface {
	Put(key string, value any)
	Get(key string) (any, bool)
	GetOrElse(key string, alt any) any
	Delete(key string)
	IsDirty() bool
	Flashes() []interface{}
}

// Wrapper is a wrapper around the Gorilla sessions.Session that implements the Session interface.
type Wrapper struct {
	session *sessions.Session

	dirty bool
}

// NewWrapper creates a new Wrapper from a Gorilla sessions.Session.
func NewWrapper(session *sessions.Session) *Wrapper {
	return &Wrapper{session: session}
}

// Put adds a key-value pair to the session. If the value is the same as the previous value, it does nothing.
func (s *Wrapper) Put(key string, value any) {
	prev, found := s.session.Values[key]
	if found && prev == value {
		return
	}
	s.session.Values[key] = value
	s.dirty = true
}

// Get retrieves a value from the session by key. It returns the value and a boolean indicating if the key was found.
func (s *Wrapper) Get(key string) (any, bool) {
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

// GetOrElse retrieves a value from the session by key, returning an alternative value if the key is not found.
func (s *Wrapper) GetOrElse(key string, alt any) any {
	rv, found := s.Get(key)
	if !found {
		return alt
	}
	return rv
}

// Delete removes a key-value pair from the session. If the key does not exist, it does nothing.
func (s *Wrapper) Delete(key string) {
	_, found := s.session.Values[key]
	if !found {
		return
	}
	delete(s.session.Values, key)
	s.dirty = true
}

// Flashes retrieves and clears the flash messages from the session.
func (s *Wrapper) Flashes() []interface{} {
	return s.session.Flashes()
}

// IsDirty checks if the session has been modified since it was last saved. It returns true if the session is dirty (modified), otherwise false.
func (s *Wrapper) IsDirty() bool {
	return s.dirty
}
