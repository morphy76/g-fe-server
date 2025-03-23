package auth

import (
	"context"
	"net/http"

	"github.com/morphy76/g-fe-server/internal/common/health"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
)

// CreateHealthCheck creates a health check for the IAM
func CreateHealthCheck(relyingParty rp.RelyingParty) health.AdditionalCheckFn {
	return func(requestContext context.Context) (health.HealthCheckFn, health.Probe) {
		return func(requestContext context.Context) (string, health.Status) {
			feLogger := logger.GetLogger(requestContext, "feServer")

			iamStatus := health.Inactive
			label := "IAM - OIDC"

			req, err := http.NewRequestWithContext(requestContext, http.MethodGet, relyingParty.Issuer(), nil)
			if err != nil {
				feLogger.Error().Err(err).Msg("Error creating OIDC request")
				return label, health.Inactive
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				feLogger.Error().Err(err).Msg("Error sending OIDC request")
				return label, health.Inactive
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				iamStatus = health.Active
			} else {
				iamStatus = health.Inactive
			}

			return label, iamStatus
		}, health.Live
	}
}
