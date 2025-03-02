package aiw

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type AIWFacade struct {
	HttpClient *http.Client
}

func (aiw *AIWFacade) Call(ctx context.Context, req *http.Request) (*http.Response, error) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	return aiw.HttpClient.Do(req)
}
