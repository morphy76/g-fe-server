package middleware

import (
	"context"
	"net/http"

	app_http "github.com/morphy76/g-fe-server/internal/http"
)

type Ownership struct {
	Tenant       string
	Subscription string
}

func TenantResolver(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newContext := context.WithValue(r.Context(), app_http.CTX_OWNERSHIP_KEY, Ownership{
			Tenant:       r.Header.Get("X-Tenant"),
			Subscription: r.Header.Get("X-Subscription"),
		})
		tenantResolverContext := r.WithContext(newContext)
		next.ServeHTTP(w, tenantResolverContext)
	})
}
