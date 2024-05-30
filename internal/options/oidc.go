package options

type OidcOptions struct {
	Issuer       string
	ClientId     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}
