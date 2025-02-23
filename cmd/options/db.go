package options

// MongoDBOptions holds the options for the MongoDB client
type MongoDBOptions struct {
	URL      string
	Database string
	User     string
	Password string
}
