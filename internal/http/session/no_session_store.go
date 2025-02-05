//go:build !with_http_session

package session

func CreateSessionStore(
	sessionOptions *SessionOptions,
	contextRoot string,
) (SessionStore, error) {
	return nil, nil
}
