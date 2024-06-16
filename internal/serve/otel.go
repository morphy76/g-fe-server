package serve

import (
	"context"
	"errors"

	"github.com/morphy76/g-fe-server/internal/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

const (
	OTEL_APP_NAME     = "g-fe-server"
	OTEL_GW_NAME      = "gateway"
	OTEL_EXAMPLE_NAME = "g-be-server"
)

func SetupOTelSDK(ctx context.Context, otelOptions *options.OtelOptions) (shutdown func(context.Context) error, err error) {

	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	propagator := newPropagator()
	otel.SetTextMapPropagator(propagator)

	tracerProvider, err := newTraceProvider(otelOptions.Enabled, otelOptions.Url)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	return shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(enabled bool, url string) (*trace.TracerProvider, error) {

	var traceProvider *trace.TracerProvider

	if enabled {
		traceExporter, err := zipkin.New(url)
		if err != nil {
			return nil, err
		}
		traceProvider = trace.NewTracerProvider(
			trace.WithBatcher(traceExporter),
			trace.WithResource(resource.NewSchemaless(
				attribute.String("service.name", OTEL_APP_NAME),
			)),
		)
	} else {
		traceProvider = trace.NewTracerProvider()
	}

	return traceProvider, nil
}
