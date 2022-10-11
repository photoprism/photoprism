package limiter

import (
	"time"

	"golang.org/x/time/rate"
)

// Auth limits failed authentication requests (one per minute).
var Auth = NewLimit(rate.Every(time.Minute), 10)
