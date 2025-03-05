package otel

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/morphy76/g-fe-server/cmd/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	metrics "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/shirou/gopsutil/v3/cpu"
)

// SetupOTelSDK sets up the OTel SDK
func SetupOTelSDK(otelOptions *options.OTelOptions) (shutdown func() error, err error) {
	if otelOptions.Enabled == false {
		return nil, nil
	}

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

	useResource, err := newResource(otelOptions.ServiceName)
	if err != nil {
		handleErr(err)
		return
	}

	tracerProvider, err := newTraceProvider(ctx, useResource, otelOptions.URL)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := newMeterProvider(ctx, useResource, otelOptions.URL)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return shutdown, err
}

func newResource(serviceName string) (*resource.Resource, error) {
	return resource.New(
		context.Background(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, useResource *resource.Resource, url string) (*trace.TracerProvider, error) {
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
		trace.WithResource(useResource),
	), nil
}

func newMeterProvider(ctx context.Context, useResource *resource.Resource, url string) (*metric.MeterProvider, error) {
	expMetrics, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpointURL(url),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(expMetrics, metric.WithInterval(10*time.Second))),
		metric.WithResource(useResource),
	)

	shouldReturn, result, err := addRuntimeGauges(meterProvider)
	if shouldReturn {
		return result, err
	}

	return meterProvider, nil
}

func addRuntimeGauges(meterProvider *metric.MeterProvider) (bool, *metric.MeterProvider, error) {
	meter := meterProvider.Meter("go.runtime")
	goRoutineCount, err := meter.Int64ObservableGauge(
		"runtime.go.goroutines",
		metrics.WithDescription("The number of goroutines that currently exist."),
	)
	if err != nil {
		return true, nil, err
	}

	currentMemory, err := meter.Int64ObservableGauge(
		"runtime.memory.current",
		metrics.WithDescription("The current memory usage in KiB."),
	)
	if err != nil {
		return true, nil, err
	}

	totalAlloc, err := meter.Int64ObservableGauge(
		"runtime.memory.total_alloc",
		metrics.WithDescription("The total memory allocated in KiB."),
	)
	if err != nil {
		return true, nil, err
	}

	sysMemory, err := meter.Int64ObservableGauge(
		"runtime.memory.sys",
		metrics.WithDescription("The total memory obtained from the OS in KiB."),
	)
	if err != nil {
		return true, nil, err
	}

	heapAlloc, err := meter.Int64ObservableGauge(
		"runtime.memory.heap_alloc",
		metrics.WithDescription("The total memory allocated by the runtime in KiB."),
	)
	if err != nil {
		return true, nil, err
	}

	gcCount, err := meter.Int64ObservableGauge(
		"runtime.gc.count",
		metrics.WithDescription("The number of garbage collections."),
	)
	if err != nil {
		return true, nil, err
	}

	gcPauseTotal, err := meter.Int64ObservableGauge(
		"runtime.gc.pause_total",
		metrics.WithDescription("The total pause time in nanoseconds."),
	)
	if err != nil {
		return true, nil, err
	}

	cpuLoad, err := meter.Float64ObservableGauge(
		"runtime.cpu.load",
		metrics.WithDescription("The current CPU load."),
	)
	if err != nil {
		return true, nil, err
	}

	_, err = meter.RegisterCallback(
		func(ctx context.Context, observer metrics.Observer) error {

			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)

			cpuStat, _ := cpu.Percent(0, false)

			observer.ObserveInt64(currentMemory, int64(mem.Alloc/1024))
			observer.ObserveInt64(totalAlloc, int64(mem.TotalAlloc/1024))
			observer.ObserveInt64(sysMemory, int64(mem.Sys/1024))
			observer.ObserveInt64(heapAlloc, int64(mem.HeapAlloc/1024))

			observer.ObserveInt64(goRoutineCount, int64(runtime.NumGoroutine()))

			observer.ObserveInt64(gcCount, int64(mem.NumGC))
			observer.ObserveInt64(gcPauseTotal, int64(mem.PauseTotalNs))

			observer.ObserveFloat64(cpuLoad, cpuStat[0])

			return nil
		},
		currentMemory,
		totalAlloc,
		sysMemory,
		heapAlloc,
		goRoutineCount,
		gcCount,
		gcPauseTotal,
		cpuLoad,
	)
	if err != nil {
		return true, nil, err
	}
	return false, nil, nil
}
