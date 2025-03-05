package example

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Unleash/unleash-client-go/v4"
	featContext "github.com/Unleash/unleash-client-go/v4/context"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/morphy76/g-fe-server/internal/features"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/server"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	moduleName = "module.example"
	modulePath = "/example"
)

func Handler(
	apiRouter *mux.Router,
	feServer *server.FEServer,
	routerLog zerolog.Logger,
) {
	exampleRouter := apiRouter.PathPrefix(modulePath).Subrouter()
	featCtx := featContext.Context{
		Properties: map[string]string{
			"role": "api",
		},
	}
	exampleRouter.Use(features.ModuleVerifier(moduleName, unleash.WithContext(featCtx)))
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Str("module", moduleName).
			Msg("Router registered")
	}

	// test APIs & functions
	exampleRouter.HandleFunc("/up", func(w http.ResponseWriter, r *http.Request) {
		// this API has its own span started by the OTEL SDK integration with the mux router
		useLogger := logger.GetLogger(r.Context(), "example")
		useFEServer := server.ExtractFEServer(r.Context())

		isTestFeatOn := useFEServer.IsFeatureEnabled("test")

		// this is an inner span, started by the application, representing upstream logic
		_, span := trace.SpanFromContext(r.Context()).TracerProvider().Tracer("mboh").Start(r.Context(), "upBiz")
		useLogger.Debug().Msg("start up biz")
		<-time.After(1 * time.Second)
		span.AddEvent("testEventUp")
		// session := session.ExtractSession(r.Context())
		// session.Put("test", uuid.New().String())
		useLogger.Info().Msg("end up biz")
		span.End()

		if isTestFeatOn {
			// this other span represents remote downstream logic
			_, span = trace.SpanFromContext(r.Context()).TracerProvider().Tracer("mboh").Start(r.Context(), "downBiz")
			useLogger.Debug().Msg("start down biz")
			<-time.After(1 * time.Second)

			// a generic request to the same server, but to a different endpoint
			newUrl := fmt.Sprintf("http://%s%s",
				r.Host,
				strings.Replace(r.URL.Path, "up", "down", 1),
			)
			newReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, newUrl, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// mediated by the AIW facade to inject the OTEL context
			newRes, err := useFEServer.GetAIWFacade().Call(r.Context(), newReq)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// the request span, the HTTP call namely, is closed when closing the response body
			transferBodyTest(w, newRes)
			// the downstream span can be closed at the end
			defer span.End()
			<-time.After(1 * time.Second)
			span.AddEvent("testEventDown")
			useLogger.Info().Msg("end down biz")
		} else {
			w.Write([]byte("{\"message\": \"Hello, World!\"}"))
		}
	}).Name("GET " + feServer.ServeOpts.ContextRoot + "/api/" + moduleName + "/up")

	exampleRouter.HandleFunc("/down", func(w http.ResponseWriter, r *http.Request) {
		// this API has its own span started by the OTEL SDK integration with the mux router
		// but it seems not to be bound to the client parent span (TODO)
		useLogger := logger.GetLogger(r.Context(), "example")
		useFEServer := server.ExtractFEServer(r.Context())

		span := trace.SpanFromContext(r.Context())

		useLogger.Debug().Msg("start downer biz")
		// session := session.ExtractSession(r.Context())
		// test := session.Get("test").(string)
		test := uuid.NewString()
		// TODO mongo client should be bound to the OTEL context WIP
		useFEServer.MongoClient.Database("fe_db").Collection("test_collection").InsertOne(r.Context(), map[string]string{"test": test})
		span.AddEvent("testEventDowner")
		w.Write([]byte("{\"message\": \"Hello, World, " + test + "!\"}"))
		<-time.After(1 * time.Second)
		span.RecordError(errors.New("testError"))
		span.SetStatus(codes.Error, "testError")
		useLogger.Info().Msg("end downer biz")
	}).Name("GET " + feServer.ServeOpts.ContextRoot + "/api/" + moduleName + "/down")

	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Str("module", moduleName).
			Msg("Handler registered")
	}
}

// this method grants to close the http request span accordingly to the completion of managing the response body
// not really interested in handling error so far
func transferBodyTest(w http.ResponseWriter, newRes *http.Response) bool {
	w.WriteHeader(newRes.StatusCode)
	body, err := io.ReadAll(newRes.Body)
	// example of business metrics
	addBusinessMetrics(len(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	w.Write(body)
	defer newRes.Body.Close()
	return false
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
