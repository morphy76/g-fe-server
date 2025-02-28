package server

import (
	"context"
	"errors"

	"github.com/morphy76/g-fe-server/cmd/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
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

	tracerProvider, err := newTraceProvider(ctx, otelOptions.Enabled, otelOptions.ServiceName, otelOptions.URL)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := newMeterProvider(ctx, otelOptions.Enabled, otelOptions.ServiceName, otelOptions.URL)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, enabled bool, serviceName string, url string) (*trace.TracerProvider, error) {
	if enabled {
		expTraces, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpointURL(url),
		)
		if err != nil {
			return nil, err
		}

		return trace.NewTracerProvider(
			trace.WithBatcher(expTraces),
			trace.WithResource(resource.NewSchemaless(
				attribute.String("service.name", serviceName),
			)),
		), nil
	}

	return nil, nil
}

func newMeterProvider(ctx context.Context, enabled bool, serviceName string, url string) (*metric.MeterProvider, error) {
	if enabled {
		expMetrics, err := otlpmetricgrpc.New(
			ctx,
			otlpmetricgrpc.WithInsecure(),
			otlpmetricgrpc.WithEndpointURL(url),
		)
		if err != nil {
			return nil, err
		}

		return metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(expMetrics)),
			metric.WithResource(resource.NewSchemaless(
				attribute.String("service.name", serviceName),
			)),
		), nil
	}

	return nil, nil
}
