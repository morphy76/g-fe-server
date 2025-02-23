package server

import (
	"context"
	"errors"

	"github.com/morphy76/g-fe-server/cmd/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SetupOTelSDK sets up the OTel SDK
func SetupOTelSDK(otelOptions *options.OTelOptions) (shutdown func() error, err error) {

	ctx := context.Background()
	var shutdownFuncs []func(context.Context) error

	shutdown = func() error {
		var err error
		for _, fn := range shutdownFuncs {
			useCtx := ctx
			err = errors.Join(err, fn(useCtx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown())
	}

	propagator := newPropagator()
	otel.SetTextMapPropagator(propagator)

	tracerProvider, err := newTraceProvider(otelOptions.Enabled, otelOptions.ServiceName, otelOptions.URL)
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

func newTraceProvider(enabled bool, serviceName string, url string) (*trace.TracerProvider, error) {
	var traceProvider *trace.TracerProvider

	if enabled {
		traceExporter, err := zipkin.New(url)
		if err != nil {
			return nil, err
		}
		traceProvider = trace.NewTracerProvider(
			trace.WithBatcher(traceExporter),
			trace.WithResource(resource.NewSchemaless(
				attribute.String("service.name", serviceName),
			)),
		)
	} else {
		traceProvider = trace.NewTracerProvider()
	}

	return traceProvider, nil
}
