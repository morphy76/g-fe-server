package middleware

import (
	"context"
	"net/http"
)

type CTX_OWNERSHIP string

const CTX_OWNERSHIP_KEY CTX_OWNERSHIP = "ownership"

type Ownership struct {
	Tenant       string
	Subscription string
}

func TenantResolver(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		newContext := context.WithValue(r.Context(), CTX_OWNERSHIP_KEY, Ownership{
			Tenant:       r.Header.Get("X-Tenant"),
			Subscription: r.Header.Get("X-Subscription"),
		})
		useRequestLogger := r.WithContext(newContext)
		next.ServeHTTP(w, useRequestLogger)
	})
}
