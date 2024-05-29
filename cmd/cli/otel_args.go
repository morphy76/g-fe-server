package cli

import (
	"errors"
	"flag"
	"os"

	"github.com/morphy76/g-fe-server/internal/options"
)

type otelOptionsBuidler func() (*options.OtelOptions, error)

var errRequiredOTLPUrl = errors.New("OTLP export enabled but no URL has been specified")

func IsRequiredOTLPUrl(err error) bool {
	return err == errRequiredOTLPUrl
}

const (
	ENV_ENABLE_OTEL_EXPORT = "ENABLE_OTEL_EXPORT"
	ENV_OTLP_URL           = "OTLP_URL"
)

func OtelOptionsBuilder() otelOptionsBuidler {

	otlpEnabledArg := flag.Bool("otel-enabled", false, "Enable to export onto OTLP. Environment: "+ENV_ENABLE_OTEL_EXPORT)
	otlpUrlArg := flag.String("otlp-url", "", "OTLP collector. Environment: "+ENV_OTLP_URL)

	rv := func() (*options.OtelOptions, error) {

		otlpEnabled := *otlpEnabledArg
		otlpEnabledStr, found := os.LookupEnv(ENV_ENABLE_OTEL_EXPORT)
		if found {
			otlpEnabled = otlpEnabledStr == "true"
		}

		url, found := os.LookupEnv(ENV_OTLP_URL)
		if !found {
			url = *otlpUrlArg
		}
		if url == "" && otlpEnabled {
			return nil, errRequiredOTLPUrl
		}

		return &options.OtelOptions{
			Enabled: otlpEnabled,
			Url:     url,
		}, nil
	}

	return rv
}
