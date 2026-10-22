package limiter

import (
	"golang.org/x/time/rate"
)

// Default authorization rate-limit parameters.
const (
	DefaultAuthorizeInterval = DefaultAuthInterval // average interval between counted OAuth authorization requests
	DefaultAuthorizeLimit    = DefaultAuthLimit    // burst limit of counted OAuth authorization requests
)

// Authorize limits the OAuth authorization requests, failed client checks and client metadata fetches per client IP.
// It is separate from Auth, so these requests cannot affect the authentication of existing sessions.
var Authorize = NewLimit(rate.Every(DefaultAuthorizeInterval), DefaultAuthorizeLimit)
