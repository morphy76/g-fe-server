package http

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
)

var mockProvider *consumer.V4HTTPMockProvider
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)

func TestMain(m *testing.M) {
	useMockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: "g-fe-server",
		Provider: "g-be-service",
		LogDir:   logDir,
		PactDir:  pactDir,
	})
	if err != nil {
		fmt.Println("Error creating mock provider: ", err)
		os.Exit(1)
	} else {
		mockProvider = useMockProvider
	}
	code := m.Run()
	os.Exit(code)
}

func listExamples() error {
	url := fmt.Sprintf("http://localhost:%d/example", mockProvider.Server.Port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if _, err := http.DefaultClient.Do(req); err != nil {
		return err
	}
	return nil
}

// func TestTheWholeBody_GET(t *testing.T) {
// 	pact.AddInteraction().
// 		Given("Match the whole response body").
// 		UponReceiving("A a GET request").
// 		WithRequest(dsl.Request{
// 			Method: "GET",
// 			Path:   dsl.Term("/example", "/example"),
// 			Headers: dsl.MapMatcher{
// 				"Content-Type": dsl.Term("application/json; charset=utf-8", `application\/json`),
// 			},
// 		}).
// 		WillRespondWith(dsl.Response{
// 			Status: 200,
// 			Body:   []example.Example{{Age: 10, Name: "Alice"}, {Age: 20, Name: "Bob"}},
// 			Headers: dsl.MapMatcher{
// 				"Content-Type": dsl.Term("application/json; charset=utf-8", `application\/json`),
// 			},
// 		})

// 	if err := pact.Verify(listExamples); err != nil {
// 		t.Fatalf("Error on verifying pact: %v", err)
// 	}
// }
