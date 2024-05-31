package middleware

import (
	"net/http"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/serve"
)

func TenantResolver(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newContext := app_http.InjectOwnership(r.Context(), serve.Ownership{
			Tenant:       r.Header.Get("X-Tenant"),
			Subscription: r.Header.Get("X-Subscription"),
		})
		tenantResolverContext := r.WithContext(newContext)
		next.ServeHTTP(w, tenantResolverContext)
	})
}
