package options

// IntegrationOptions holds the options for the server integrations
type IntegrationOptions struct {
	// MongoDBOptions holds the options for the MongoDB integration
	DBOptions *MongoDBOptions
	// OTelOptions holds the options for the OpenTelemetry integration
	OTelOptions *OTelOptions
	// UnleashOptions holds the options for the Unleash integration
	UnleashOptions *UnleashOptions
	// AIWOptions holds the options for the AIW integration
	AIWOptions *AIWOptions
}
