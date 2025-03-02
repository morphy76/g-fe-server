package db

import (
	"context"
	"errors"
	"fmt"

	v2_bson "go.mongodb.org/mongo-driver/v2/bson"

	old_event "go.mongodb.org/mongo-driver/event"
	v2_event "go.mongodb.org/mongo-driver/v2/event"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type oldStartedSignarture func(ctx context.Context, oldEvent *old_event.CommandStartedEvent)
type newStartedSignarture func(ctx context.Context, newEvent *v2_event.CommandStartedEvent)
type oldSucceededSignarture func(ctx context.Context, oldEvent *old_event.CommandSucceededEvent)
type newSucceededSignarture func(ctx context.Context, newEvent *v2_event.CommandSucceededEvent)
type oldFailedSignarture func(ctx context.Context, oldEvent *old_event.CommandFailedEvent)
type newFailedSignarture func(ctx context.Context, newEvent *v2_event.CommandFailedEvent)

func NewMonitor() *v2_event.CommandMonitor {
	oldMonitor := otelmongo.NewMonitor(
		otelmongo.WithTracerProvider(otel.GetTracerProvider()),
	)
	return &v2_event.CommandMonitor{
		Started:   newCommandStartedEvent(oldMonitor.Started),
		Succeeded: newCommandSucceededEvent(oldMonitor.Succeeded),
		Failed:    newCommandFailedEvent(oldMonitor.Failed),
	}
}

func NewPoolMonitor() *v2_event.PoolMonitor {
	return &v2_event.PoolMonitor{
		Event: func(event *v2_event.PoolEvent) {
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

func newCommandStartedEvent(old oldStartedSignarture) newStartedSignarture {
	return func(ctx context.Context, newEvent *v2_event.CommandStartedEvent) {
		oldCmdStartedEvent := new(old_event.CommandStartedEvent)
		old(ctx, oldCmdStartedEvent)
		newEvent.Command = v2_bson.Raw(oldCmdStartedEvent.Command)
		newEvent.DatabaseName = oldCmdStartedEvent.DatabaseName
		newEvent.RequestID = oldCmdStartedEvent.RequestID
		newEvent.ConnectionID = oldCmdStartedEvent.ConnectionID
		newEvent.CommandName = oldCmdStartedEvent.CommandName
		newEvent.ServerConnectionID = oldCmdStartedEvent.ServerConnectionID64
		if oldCmdStartedEvent.ServiceID != nil {
			useServiceID, _ := v2_bson.ObjectIDFromHex(oldCmdStartedEvent.ServiceID.Hex())
			newEvent.ServiceID = &useServiceID
		}

		fmt.Printf("sssssssssssstaaaaaart Event: %+v\n", newEvent)
	}
}

func newCommandSucceededEvent(old oldSucceededSignarture) newSucceededSignarture {
	return func(ctx context.Context, newEvent *v2_event.CommandSucceededEvent) {
		oldCmdSucceededEvent := new(old_event.CommandSucceededEvent)
		old(ctx, oldCmdSucceededEvent)
		newEvent.Reply = v2_bson.Raw(oldCmdSucceededEvent.Reply)
		newEvent.CommandName = oldCmdSucceededEvent.CommandName
		newEvent.RequestID = oldCmdSucceededEvent.RequestID
		newEvent.ConnectionID = oldCmdSucceededEvent.ConnectionID
		newEvent.ServerConnectionID = oldCmdSucceededEvent.ServerConnectionID64
		if oldCmdSucceededEvent.ServiceID != nil {
			useServiceID, _ := v2_bson.ObjectIDFromHex(oldCmdSucceededEvent.ServiceID.Hex())
			newEvent.ServiceID = &useServiceID
		}
		newEvent.DatabaseName = oldCmdSucceededEvent.DatabaseName
		newEvent.Duration = oldCmdSucceededEvent.Duration
		newEvent.CommandFinishedEvent = v2_event.CommandFinishedEvent{
			CommandName:        oldCmdSucceededEvent.CommandName,
			RequestID:          oldCmdSucceededEvent.RequestID,
			ConnectionID:       oldCmdSucceededEvent.ConnectionID,
			ServerConnectionID: oldCmdSucceededEvent.ServerConnectionID64,
		}
		if oldCmdSucceededEvent.CommandFinishedEvent.ServiceID != nil {
			finishedServiceID, _ := v2_bson.ObjectIDFromHex(oldCmdSucceededEvent.CommandFinishedEvent.ServiceID.Hex())
			newEvent.CommandFinishedEvent.ServiceID = &finishedServiceID
		}

		fmt.Printf("sssssssuuuuucccccceeeeeesssss Event: %+v\n", newEvent)
	}
}

func newCommandFailedEvent(old oldFailedSignarture) newFailedSignarture {
	return func(ctx context.Context, newEvent *v2_event.CommandFailedEvent) {
		oldCmdFailedEvent := new(old_event.CommandFailedEvent)
		old(ctx, oldCmdFailedEvent)
		newEvent.CommandName = oldCmdFailedEvent.CommandName
		newEvent.RequestID = oldCmdFailedEvent.RequestID
		newEvent.ConnectionID = oldCmdFailedEvent.ConnectionID
		newEvent.ServerConnectionID = oldCmdFailedEvent.ServerConnectionID64
		if oldCmdFailedEvent.ServiceID != nil {
			useServiceID, _ := v2_bson.ObjectIDFromHex(oldCmdFailedEvent.ServiceID.Hex())
			newEvent.ServiceID = &useServiceID
		}
		newEvent.DatabaseName = oldCmdFailedEvent.DatabaseName
		newEvent.CommandFinishedEvent = v2_event.CommandFinishedEvent{
			CommandName:        oldCmdFailedEvent.CommandName,
			RequestID:          oldCmdFailedEvent.RequestID,
			ConnectionID:       oldCmdFailedEvent.ConnectionID,
			ServerConnectionID: oldCmdFailedEvent.ServerConnectionID64,
		}
		if oldCmdFailedEvent.CommandFinishedEvent.ServiceID != nil {
			finishedServiceID, _ := v2_bson.ObjectIDFromHex(oldCmdFailedEvent.CommandFinishedEvent.ServiceID.Hex())
			newEvent.CommandFinishedEvent.ServiceID = &finishedServiceID
		}
		newEvent.Failure = errors.New(oldCmdFailedEvent.Failure)
		newEvent.Duration = oldCmdFailedEvent.Duration

		fmt.Printf("ssssssfaaaaaiiiillllll Event: %+v\n", newEvent)
	}
}
