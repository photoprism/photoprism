package api

import (
	"time"
)

// nowRFC3339 returns a time formatted according to RFC 3339 in UTC.
func nowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }

// HealthResponse is the response type for GET /api/v1/cluster/health.
// swagger:model HealthResponse
type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

// NewHealthResponse returns a standard health response with a status and RFC 3339 UTC timestamp.
func NewHealthResponse(status string) *HealthResponse {
	return &HealthResponse{Status: status, Time: nowRFC3339()}
}
