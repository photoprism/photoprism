package limiter

import (
	"time"

	"golang.org/x/time/rate"
)

const (
	DefaultAuthInterval = time.Second * 10 // average authentication errors per second
	DefaultAuthLimit    = 60               // authentication failure burst rate limit (for access tokens)
)

// Auth limits the number of authentication errors from a single IP per time interval (every 15 seconds by default).
var Auth = NewLimit(rate.Every(DefaultAuthInterval), DefaultAuthLimit)
