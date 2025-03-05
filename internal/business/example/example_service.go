package example

import (
	"bytes"
	"context"
	"io"

	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/server"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName   = "exampleService"
	methodNameKey = "methodNameKey"
)

type ExampleService interface {
	DoUp() error
	CallDown() (*ExampleResponse, error)
	DoDown() (string, error)
}

func NewExampleService(requestContext context.Context) ExampleService {
	useLogger := logger.GetLogger(requestContext, serviceName)
	span := trace.SpanFromContext(requestContext)
	feServer := server.ExtractFEServer(requestContext)

	return &exampleService{
		span: span,
		delegate: &exampleServiceDelegate{
			log:            useLogger,
			requestContext: requestContext,
			feServer:       feServer,
		},
	}
}

type exampleService struct {
	span     trace.Span
	delegate *exampleServiceDelegate
}

func (s *exampleService) beforeInvocation(methodName string) (context.Context, error) {
	spanContext, _ := s.span.TracerProvider().Tracer(serviceName).Start(s.delegate.requestContext, methodName)
	return context.WithValue(spanContext, methodNameKey, methodName), nil
}

func (s *exampleService) afterInvocation(serviceContext context.Context) error {
	span := trace.SpanFromContext(serviceContext)
	span.End()
	return nil
}

func (s *exampleService) DoUp() error {
	serviceContext, err := s.beforeInvocation("DoUp")
	if err != nil {
		return err
	}
	defer s.afterInvocation(serviceContext)

	err = s.delegate.doUp(serviceContext)
	if err != nil {
		return err
	}

	return nil
}

func (s *exampleService) CallDown() (*ExampleResponse, error) {
	serviceContext, err := s.beforeInvocation("CallDown")
	if err != nil {
		return nil, err
	}
	defer s.afterInvocation(serviceContext)

	answer, err := s.delegate.callDown(serviceContext)
	if err != nil {
		return nil, err
	}
	if answer == nil {
		return NewExampleResponse("AIW is disabled"), nil
	}

	resBody := <-answer
	reader := io.NopCloser(bytes.NewReader(resBody))

	rv, err := FromJSON(reader)
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (s *exampleService) DoDown() (string, error) {
	serviceContext, err := s.beforeInvocation("DoDown")
	if err != nil {
		return "", err
	}
	defer s.afterInvocation(serviceContext)

	answer, err := s.delegate.doDown(serviceContext)
	if err != nil {
		return "", err
	}

	return answer, nil
}
