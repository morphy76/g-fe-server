package api

import (
	"fmt"

	"github.com/morphy76/g-fe-server/internal/options"
)

const routeName = "/example"

func RegisteredRouteUri(serveOptions *options.ServeOptions) string {
	return fmt.Sprintf("route:%s:%s://%s:%s%s/api/example", routeName, serveOptions.Protocol, serveOptions.Host, serveOptions.Port, serveOptions.ContextRoot)
}

func UnRegisteredRouteUri() string {
	return fmt.Sprintf("unroute:%s", routeName)
}
