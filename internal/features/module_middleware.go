package features

import (
	"net/http"

	"github.com/Unleash/unleash-client-go/v4"
	"github.com/gorilla/mux"
	"github.com/morphy76/g-fe-server/internal/server"
)

func ModuleVerifier(moduleName string, opts ...unleash.FeatureOption) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			feServer := server.ExtractFEServer(r.Context())
			if feServer == nil {
				http.Error(w, "FEServer not found in context", http.StatusInternalServerError)
				return
			}
			moduleEnabled := feServer.IsFeatureEnabled(moduleName, opts...)
			if !moduleEnabled {
				http.Error(w, "", http.StatusNotImplemented)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
