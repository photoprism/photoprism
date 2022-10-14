package limiter

import (
	"time"

	"golang.org/x/time/rate"
)

const DefaultAuthLimit = 10
const DefaultAuthInterval = time.Minute

// Auth limits failed authentication requests (one per minute).
var Auth = NewLimit(rate.Every(DefaultAuthInterval), DefaultAuthLimit)
