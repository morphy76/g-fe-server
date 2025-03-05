package options

type ServeOptions struct {
	StaticPath string
	PathOptions
	URLOptions
	AIWOptions
}

type PathOptions struct {
	NonFunctionalRoot string
	ContextRoot       string
}

type URLOptions struct {
	Protocol string
	Port     string
	Host     string
}

type AIWOptions struct {
	FQDN string
}
