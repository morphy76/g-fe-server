package http

import "context"

// Status represents the status of a subsystem
type Status string

// Probe represents the type of probe
type Probe int

const (
	// OK is the status of a subsystem that is active
	Active Status = "OK"
	// KO is the status of a subsystem that is inactive
	Inactive Status = "KO"
)

const (
	// Live is the probe type for live checks
	Live Probe = iota + 1
	// Ready is the probe type for ready checks
	Ready
)

// HealthResponse represents the response of a health check
type HealthResponse struct {
	Status     Status                    `json:"status"`
	Message    string                    `json:"message,omitempty"`
	SubSystems map[string]HealthResponse `json:"subsystems,omitempty"`
}

// Error represents an error response
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// AdditionalCheckFn is a function that returns a health check function and a probe
type AdditionalCheckFn func(requestContext context.Context) (HealthCheckFn, Probe)

// HealthCheckFn is a function that performs a health check
type HealthCheckFn func(requestContext context.Context) (string, Status)
