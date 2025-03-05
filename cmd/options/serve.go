package options

type ServeOptions struct {
	StaticPath string
	PathOptions
	URLOptions
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
