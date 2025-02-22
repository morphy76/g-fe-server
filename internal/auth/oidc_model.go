package auth

// OIDCOptions holds the configuration for the OIDC client
type OIDCOptions struct {
	Disabled      bool
	Issuer        string
	ClientID      string
	ClientSecret  string
	Scopes        []string
	ExtraAuthArgs map[string]string
}
