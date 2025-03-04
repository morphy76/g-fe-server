package cli

import (
	"flag"
	"os"

	"github.com/morphy76/g-fe-server/cmd/options"
)

type UnleashOptionsBuilderFn func() (*options.UnleashOptions, error)

const (
	unleashEnabled  = "UNLEASH_ENABLED"
	unleashAppName  = "UNLEASH_APP_NAME"
	envUnleashURL   = "UNLEASH_URL"
	envUnleashToken = "UNLEASH_TOKEN"
)

func UnleashOptionsBuilder() UnleashOptionsBuilderFn {
	unleashEnabledArg := flag.Bool("unleash-enabled", false, "Unleash Enabled. Environment: "+unleashEnabled)
	unleashAppNameArg := flag.String("unleash-app-name", "gfe", "Unleash App Name. Environment: "+unleashAppName)
	unleashURLArg := flag.String("unleash-url", "", "Unleash URL. Environment: "+envUnleashURL)
	unleashTokenArg := flag.String("unleash-token", "", "Unleash Token. Environment: "+envUnleashToken)

	return func() (*options.UnleashOptions, error) {

		enabled := *unleashEnabledArg
		enabledStr, found := os.LookupEnv(unleashEnabled)
		if found {
			enabled = enabledStr == "true"
		}

		url, found := os.LookupEnv(envUnleashURL)
		if !found {
			url = *unleashURLArg
		}

		appName, found := os.LookupEnv(unleashAppName)
		if !found {
			appName = *unleashAppNameArg
		}

		token, found := os.LookupEnv(envUnleashToken)
		if !found {
			token = *unleashTokenArg
		}

		return &options.UnleashOptions{
			Enabled: enabled,
			AppName: appName,
			URL:     url,
			Token:   token,
		}, nil
	}
}
