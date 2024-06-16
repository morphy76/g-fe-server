package api

import (
	"fmt"

	"github.com/morphy76/g-fe-server/internal/options"
)

const routeName = "/example"

func RegisteredRouteUri(serveOptions *options.ServeOptions) string {
	return fmt.Sprintf("route:%s:%s", routeName, callbackUrl(serveOptions))
}

func callbackUrl(serveOptions *options.ServeOptions) string {
	return fmt.Sprintf("%s%s/api%s", serveOptions.CallbackUrl, serveOptions.ContextRoot, routeName)
}

func UnRegisteredRouteUri() string {
	return fmt.Sprintf("unroute:%s", routeName)
}
