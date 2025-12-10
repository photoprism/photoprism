package places

import (
	"time"
)

// ApiName is the backend API name.
const ApiName = "places"

// LocationServiceUrls specifies S2 cell location query URLs.
var LocationServiceUrls = []string{
	"https://places.photoprism.app/v1/location/%s",
}

// ReverseServiceUrls specifies reverse location query URLs.
var ReverseServiceUrls = []string{
	"https://places.photoprism.app/v1/reverse",
}

// SearchServiceUrls specifies location name query URLs.
var SearchServiceUrls = []string{
	"https://places.photoprism.app/v1/search",
}

// Retries specifies the number of attempts to retry the service request.
var Retries = 2

// RetryDelay specifies the waiting time between retries.
var RetryDelay = 100 * time.Millisecond

// Key is the hub places API key (overridden via environment/config in production).
var Key = "f60f5b25d59c397989e3cd374f81cdd7710a4fca" //nolint:gosec // example/default key

// Secret is the hub places API secret (overridden in production).
var Secret = "photoprism" //nolint:gosec // example/default secret

// UserAgent overrides the default HTTP User-Agent header for hub places calls.
var UserAgent = ""
