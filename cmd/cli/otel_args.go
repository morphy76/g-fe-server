package cli

import (
	"errors"
	"flag"
	"os"

	"github.com/morphy76/g-fe-server/cmd/options"
)

// OTelOptionsBuidlerFn is a function that returns the OTLP options
type OTelOptionsBuidlerFn func() (*options.OTelOptions, error)

// ErrRequiredOTLPURL is an error that indicates that the OTLP URL is required
var ErrRequiredOTLPURL = errors.New("OTLP export enabled but no URL has been specified")

// IsRequiredOTLPURL returns true if the error is ErrRequiredOTLPURL
func IsRequiredOTLPURL(err error) bool {
	return err == ErrRequiredOTLPURL
}

const (
	envEnableOTelExport = "ENABLE_OTEL_EXPORT"
	envServiceName      = "OTLP_SERVICE_NAME"
	envOTLPURL          = "OTLP_URL"
)

// OTelOptionsBuilder returns a function that returns the OTLP options
func OTelOptionsBuilder() OTelOptionsBuidlerFn {

	otlpEnabledArg := flag.Bool("otel-enabled", false, "Enable to export onto OTLP. Environment: "+envEnableOTelExport)
	otlpServiceNameArg := flag.String("otel-service-name", "fe-server", "OTLP service name. Environment: "+envServiceName)
	otlpUrlArg := flag.String("otlp-url", "", "OTLP collector. Environment: "+envOTLPURL)

	rv := func() (*options.OTelOptions, error) {

		otlpEnabled := *otlpEnabledArg
		otlpEnabledStr, found := os.LookupEnv(envEnableOTelExport)
		if found {
			otlpEnabled = otlpEnabledStr == "true"
		}

		serviceName, found := os.LookupEnv(envServiceName)
		if !found {
			serviceName = *otlpServiceNameArg
		}

		url, found := os.LookupEnv(envOTLPURL)
		if !found {
			url = *otlpUrlArg
		}
		if url == "" && otlpEnabled {
			return nil, ErrRequiredOTLPURL
		}

		return &options.OTelOptions{
			Enabled:     otlpEnabled,
			ServiceName: serviceName,
			URL:         url,
		}, nil
	}

	return rv
}
