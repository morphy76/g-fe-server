package cli

import (
	"context"
	"errors"

	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/internal/serve"
	"github.com/rs/zerolog/log"
)

func SetupOTEL(parentContext context.Context, otelOptions *options.OtelOptions) (func(), error) {
	otelShutdown, err := serve.SetupOTelSDK(parentContext, otelOptions)
	shutdownFn := func() {
		err = errors.Join(err, otelShutdown(parentContext))
	}
	log.Trace().
		Msg("Opentelemetry ready")
	return shutdownFn, err
}
