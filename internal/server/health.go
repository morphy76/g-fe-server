package server

// // CreateHealthCheck creates a health check for the IAM
// func CreateHealthCheck(relayingParty rp.RelyingParty) app_http.AdditionalCheckFn {
// 	return testIAMStatus(relayingParty)
// }

// func testIAMStatus(relyingParty rp.RelyingParty) app_http.AdditionalCheckFn {
// 	return func(requestContext context.Context) (app_http.HealthCheckFn, app_http.Probe) {
// 		return func(requestContext context.Context) (string, app_http.Status) {
// 			feLogger := logger.GetLogger(requestContext, "feServer")

// 			iamStatus := app_http.Inactive
// 			label := "IAM - OIDC"

// 			req, err := http.NewRequestWithContext(requestContext, http.MethodGet, relyingParty.Issuer(), nil)
// 			if err != nil {
// 				feLogger.Error().Err(err).Msg("Error creating OIDC request")
// 				return label, app_http.Inactive
// 			}

// 			client := &http.Client{}
// 			resp, err := client.Do(req)
// 			if err != nil {
// 				feLogger.Error().Err(err).Msg("Error sending OIDC request")
// 				return label, app_http.Inactive
// 			}
// 			defer resp.Body.Close()

// 			if resp.StatusCode == http.StatusOK {
// 				iamStatus = app_http.Active
// 			} else {
// 				iamStatus = app_http.Inactive
// 			}

// 			return label, iamStatus
// 		}, app_http.Live
// 	}
// }
