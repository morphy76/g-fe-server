package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/event"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type startedSignature func(context.Context, *event.CommandStartedEvent)
type succeededSignature func(context.Context, *event.CommandSucceededEvent)
type failedSignature func(context.Context, *event.CommandFailedEvent)

// TODO
func NewMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started:   newCommandStartedEvent(),
		Succeeded: newCommandSucceededEvent(),
		Failed:    newCommandFailedEvent(),
	}
}

func NewPoolMonitor() *event.PoolMonitor {
	return &event.PoolMonitor{
		Event: func(event *event.PoolEvent) {
			meter := otel.GetMeterProvider().Meter("mongo")
			useContext := context.Background()

			histogram, err := meter.Float64Histogram(
				"mongo_pool_event_duration",
				metric.WithDescription("Mongo event duration"),
			)
			if err == nil {
				histogram.Record(useContext, float64(event.Duration.Milliseconds()))
			}

			waitHistogram, err := meter.Float64Histogram(
				"mongo_pool_event_wait_duration",
				metric.WithDescription("Mongo event wait duration"),
			)
			if err == nil && event.PoolOptions != nil {
				waitHistogram.Record(useContext, float64(event.PoolOptions.WaitQueueTimeoutMS))
			}

			if event.Error != nil {
				errorHistogram, err := meter.Float64Histogram(
					"mongo_pool_event_errors",
					metric.WithDescription("Mongo error events"),
				)
				if err == nil {
					errorHistogram.Record(useContext, 1)
				}
			}
		},
	}
}

func newCommandStartedEvent() startedSignature {
	return func(ctx context.Context, newEvent *event.CommandStartedEvent) {
		fmt.Printf("sssssssssssstaaaaaart Event: %+v\n", newEvent)
		trace.SpanFromContext(ctx).TracerProvider().Tracer("mongo").Start(ctx, newEvent.CommandName)
	}
}

func newCommandSucceededEvent() succeededSignature {
	return func(ctx context.Context, newEvent *event.CommandSucceededEvent) {
		fmt.Printf("sssssssuuuuucccccceeeeeesssss Event: %+v\n", newEvent)
		_, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("mongo").Start(ctx, newEvent.CommandName)
		span.End()
	}
}

func newCommandFailedEvent() failedSignature {
	return func(ctx context.Context, newEvent *event.CommandFailedEvent) {
		fmt.Printf("ssssssfaaaaaiiiillllll Event: %+v\n", newEvent)
		_, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("mongo").Start(ctx, newEvent.CommandName)
		span.RecordError(newEvent.Failure)
		span.End()
	}
}
