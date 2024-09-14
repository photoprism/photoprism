package limiter

import (
	"time"

	"golang.org/x/time/rate"
)

// Request represents a request for the specified number of limiter tokens.
type Request struct {
	allow   bool
	limiter *rate.Limiter
	Tokens  int
}

// NewRequest checks if a request is allowed, reserves the required tokens,
// and returns a new Request to revert the reservation if successful.
func NewRequest(l *rate.Limiter, n int) *Request {
	if l.AllowN(time.Now(), n) {
		return &Request{
			allow:   true,
			limiter: l,
			Tokens:  n,
		}
	} else {
		return &Request{
			allow:   false,
			limiter: l,
			Tokens:  0,
		}
	}
}

// Allow checks if the request is allowed.
func (r *Request) Allow() bool {
	return r.allow
}

// Reject returns true if the request should be rejected.
func (r *Request) Reject() bool {
	return !r.allow
}

// Success returns the rate limit tokens that have been reserved for this request, if any.
func (r *Request) Success() {
	if r.Tokens != 0 && r.limiter != nil {
		r.limiter.ReserveN(time.Now(), -1*r.Tokens)
		r.Tokens = 0
	}
}
