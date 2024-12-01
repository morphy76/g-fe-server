package options

// OTelOptions holds the configuration for OpenTelemetry
type OTelOptions struct {
	// Enabled is a flag to enable OpenTelemetry
	Enabled bool
	// ServiceName is the name of the service
	ServiceName string
	// URL is the URL of the OpenTelemetry collector
	URL string
}
