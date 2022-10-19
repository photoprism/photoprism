package limiter

import (
	"time"

	"golang.org/x/time/rate"
)

const DefaultLoginLimit = 10
const DefaultLoginInterval = time.Minute

// Login limits failed authentication requests (one per minute).
var Login = NewLimit(rate.Every(DefaultLoginInterval), DefaultLoginLimit)
