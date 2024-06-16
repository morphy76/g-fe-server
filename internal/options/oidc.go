package options

type OidcOptions struct {
	Disabled     bool
	Issuer       string
	ClientId     string
	ClientSecret string
	Scopes       []string
}
