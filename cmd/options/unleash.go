package options

// UnleashOptions holds the configuration options for the Unleash feature toggle service.
type UnleashOptions struct {
	// Enabled indicates whether the Unleash service is enabled.
	Enabled bool
	// AppName is the name of the application using Unleash.
	AppName string
	// URL is the endpoint URL for the Unleash service.
	URL string
	// Token is the API token for authenticating with the Unleash service.
	Token string
	// Environment is the deployment environment (e.g., production, staging).
	Environment string
}
