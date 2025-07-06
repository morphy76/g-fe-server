package features

import (
	"net/http"

	"github.com/Unleash/unleash-client-go/v4"
	"github.com/gorilla/mux"
	"github.com/morphy76/g-fe-server/internal/server"
)

// ModuleVerifier is a middleware that checks if a specific module is enabled
func ModuleVerifier(moduleName string, opts ...unleash.FeatureOption) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			feServer, err := server.ExtractFEServer(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			moduleEnabled := feServer.IsFeatureEnabled(moduleName, opts...)
			if !moduleEnabled {
				http.Error(w, "", http.StatusNotFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
