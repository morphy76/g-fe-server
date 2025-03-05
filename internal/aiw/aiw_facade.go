package aiw

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/morphy76/g-fe-server/cmd/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type AIWFacade struct {
	AIWOptions *options.AIWOptions
	HttpClient *http.Client
}

// Fake method, it actually calls the presentation server again
func (aiw *AIWFacade) CallDown(ctx context.Context) ([]byte, error) {

	_, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("AIW").Start(ctx, "CallDown")
	defer span.End()

	aiwOptions := aiw.AIWOptions

	newUrl := fmt.Sprintf("%s/api/example/down",
		aiwOptions.FQDN,
	)

	newReq, err := http.NewRequestWithContext(ctx, http.MethodGet, newUrl, nil)
	if err != nil {
		return nil, err
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(newReq.Header))
	newRes, err := aiw.HttpClient.Do(newReq)
	if err != nil {
		return nil, err
	}
	defer newRes.Body.Close()
	rv, err := io.ReadAll(newRes.Body)
	if err != nil {
		return nil, err
	}

	return rv, nil
}
