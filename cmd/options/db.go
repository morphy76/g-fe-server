package options

// MongoDBOptions holds the options for the MongoDB client.
type MongoDBOptions struct {
	// URL is the connection string for the MongoDB database.
	URL string
	// User is the username for authenticating with the MongoDB database.
	User string
	// Password is the password for authenticating with the MongoDB database.
	Password string
}
