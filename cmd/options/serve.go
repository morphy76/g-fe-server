package options

// ServeOptions holds the configuration for serving the application.
type ServeOptions struct {
	// StaticPath is the path to the static files to be served.
	StaticPath string
	// PathOptions holds the options for the paths used in the application.
	PathOptions
	// URLOptions holds the options for the serving URL configuration.
	URLOptions
	// TODO CORS Options
	// TODO CSPOptions
}

// PathOptions holds the options for the paths used in the application.
type PathOptions struct {
	// NonFunctionalRoot is the root path for non-functional resources like health probes.
	NonFunctionalRoot string
	// ContextRoot is the root path for the application context.
	ContextRoot string
}

// URLOptions holds the options for the serving URL configuration.
type URLOptions struct {
	// Protocol is the protocol used for serving the application (e.g., http, https).
	Protocol string
	// Port is the port on which the application will be served.
	Port string
	// Host is the host name or IP address on which the application will be served.
	Host string
	// Path to the TLS certificate file
	CertFile string
	// Path to the TLS key file
	KeyFile string // Path to the TLS key file
}
