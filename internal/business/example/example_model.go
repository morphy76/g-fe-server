package example

import (
	"encoding/json"
	"io"
)

type ExampleResponse struct {
	Message string `json:"message"`
}

func NewExampleResponse(message string) *ExampleResponse {
	return &ExampleResponse{Message: message}
}

func FromJSON(data io.Reader) (*ExampleResponse, error) {
	var v ExampleResponse
	err := json.NewDecoder(data).Decode(&v)
	return &v, err
}
