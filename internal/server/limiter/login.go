package limiter

import (
	"time"

	"golang.org/x/time/rate"
)

const (
	DefaultLoginInterval = time.Minute // average failed logins per second
	DefaultLoginLimit    = 10          // login failure burst rate limit (for passwords and 2FA)
)

// Login limits the number of failed login attempts from a single IP per time interval (one per minute by default).
var Login = NewLimit(rate.Every(DefaultLoginInterval), DefaultLoginLimit)
