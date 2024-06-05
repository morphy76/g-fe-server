package options

type OidcOptions struct {
	Issuer       string
	ClientId     string
	ClientSecret string
	Scopes       []string
}
