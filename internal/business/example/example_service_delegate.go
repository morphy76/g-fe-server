package example

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/morphy76/g-fe-server/internal/server"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type exampleServiceDelegate struct {
	log            zerolog.Logger
	requestContext context.Context
	feServer       *server.FEServer
}

func (d *exampleServiceDelegate) doUp(serviceContext context.Context) error {
	d.log.Debug().Msg("start up biz")

	span := trace.SpanFromContext(serviceContext)
	<-time.After(1 * time.Second)
	span.AddEvent("testEventUp")

	d.log.Info().Msg("end up biz")

	return nil
}

func (d *exampleServiceDelegate) callDown(ctx context.Context) (chan []byte, error) {
	d.log.Debug().Msg("start call down biz")

	isTestFeatOn := d.feServer.IsFeatureEnabled("call.down")

	if !isTestFeatOn {
		d.log.Info().Msg("feature is off")
		return nil, nil
	}

	span := trace.SpanFromContext(ctx)
	rvCh := make(chan []byte)
	go func() {
		defer close(rvCh)
		downAnswer, err := d.feServer.AIWfacade.CallDown(ctx)
		if err != nil {
			d.log.Error().Err(err).Msg("error in call down biz")
			return
		}
		rvCh <- downAnswer
		span.AddEvent("testEventCallDown")
	}()

	d.log.Info().Msg("end call down biz")

	return rvCh, nil
}

func (d *exampleServiceDelegate) doDown(ctx context.Context) (string, error) {
	d.log.Debug().Msg("start down biz")

	span := trace.SpanFromContext(ctx)
	<-time.After(1 * time.Second)
	test := uuid.NewString()
	addBusinessMetrics(len(test))
	span.AddEvent("testEventDown")

	d.feServer.MongoClient.Database("fe_db").Collection("test_collection").InsertOne(ctx, map[string]string{"test": test})

	span.RecordError(errors.New("testError"))
	span.SetStatus(codes.Error, "testError")

	d.log.Info().Msg("end down biz")

	return test, nil
}

func addBusinessMetrics(tokens int) {
	meter := otel.GetMeterProvider().Meter("fe_server.metrics")

	var totalTokens int64
	tokensCounter, err := meter.Int64ObservableCounter(
		"fe_server.tokens",
		metric.WithDescription("The number of tokens processed"),
		metric.WithUnit("10"),
	)
	if err == nil {
		meter.RegisterCallback(
			func(ctx context.Context, observer metric.Observer) error {
				totalTokens += int64(tokens)
				observer.ObserveInt64(
					tokensCounter,
					totalTokens,
					metric.WithAttributes(
						attribute.Bool("billable", true),
						attribute.String("tenant", "todo"),
						attribute.String("subscription", "todo"),
					),
				)
				return nil
			},
			tokensCounter,
		)
	}
}
