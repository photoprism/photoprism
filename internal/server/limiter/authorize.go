package limiter

import (
	"golang.org/x/time/rate"
)

// Authorize limits the unauthenticated OAuth authorization requests per client IP. It is separate from Auth,
// so these requests cannot affect the authentication of existing sessions.
var Authorize = NewLimit(rate.Every(DefaultAuthInterval), DefaultAuthLimit)
