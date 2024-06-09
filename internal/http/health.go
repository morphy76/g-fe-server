package http

import "context"

type Status string

const (
	Active   Status = "OK"
	Inactive Status = "KO"
)

type HealthResponse struct {
	Status     Status                    `json:"status"`
	Message    string                    `json:"message,omitempty"`
	SubSystems map[string]HealthResponse `json:"subsystems,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type HealthCheckFn func(requestContext context.Context) (string, Status)
