package limiter

import (
	"time"

	"golang.org/x/time/rate"
)

const (
	DefaultAuthInterval  = time.Second * 10 // average authentication errors per second
	DefaultAuthLimit     = 60               // authentication failure burst rate limit (for access tokens)
	DefaultLoginInterval = time.Minute      // average failed logins per second
	DefaultLoginLimit    = 10               // login failure burst rate limit (for passwords and 2FA)
)

// Auth limits the number of authentication errors from a single IP per time interval (every 15 seconds by default).
var Auth = NewLimit(rate.Every(DefaultAuthInterval), DefaultAuthLimit)

// Login limits the number of failed login attempts from a single IP per time interval (one per minute by default).
var Login = NewLimit(rate.Every(DefaultLoginInterval), DefaultLoginLimit)
