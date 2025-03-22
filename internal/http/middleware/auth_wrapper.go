package middleware

import (
	"net/http"

	"github.com/morphy76/g-fe-server/internal/auth"
	"github.com/morphy76/g-fe-server/internal/server"
)

func IsAuthenticated(feServer *server.FEServer) func(handler http.Handler) http.Handler {
	return auth.IsAuthenticated(
		feServer.RelayingParty,
		feServer.ResourceServer,
		feServer.SessionName,
		*feServer.CookieStore,
		feServer.SessionStore,
	)
}
