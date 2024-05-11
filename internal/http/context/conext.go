package context

type ContxtModelKey string

type ContextModel struct {
	ContextRoot string
	StaticPath  string
}

const CTX_CONTEXT_ROOT_KEY ContxtModelKey = "contextModel"
