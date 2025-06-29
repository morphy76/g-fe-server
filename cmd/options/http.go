package options

// HTTPOptions holds the options for the HTTP server.
type HTTPOptions struct {
	// ServeOptions holds the options for serving the HTTP server.
	ServeOptions *ServeOptions
	// SessionOptions holds the options for session management.
	SessionOptions *SessionOptions
}
