package multitenancy

import (
	"context"

	"github.com/morphy76/g-fe-server/internal/common"
)

const tenantCtxKey common.CtxKey = "Tenant"

// Tenant is a simple struct that represents a tenant
type Tenant struct {
	TenantID       string
	SubscriptionID string
	GroupID        string
}

// ExtractTenant returns the Tenant from the context
func ExtractTenant(ctx context.Context) (*Tenant, bool) {
	rv := ctx.Value(tenantCtxKey)
	if rv == nil {
		return nil, false
	} else {
		return rv.(*Tenant), true
	}
}

// InitTenant creates a Context with a new Tenant
func InitTenant(ctx context.Context, tenantID, subscriptionID, groupID string) context.Context {
	tenant := &Tenant{
		TenantID:       tenantID,
		SubscriptionID: subscriptionID,
		GroupID:        groupID,
	}

	return context.WithValue(ctx, tenantCtxKey, tenant)
}
