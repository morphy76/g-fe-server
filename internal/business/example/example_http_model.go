package example

type ExampleResponse struct {
	Message string `json:"message"`
}

func NewExampleResponse(message string) *ExampleResponse {
	return &ExampleResponse{Message: message}
}
