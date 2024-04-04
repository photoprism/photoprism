package limiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limit represents an IP-based rate limiter.
type Limit struct {
	limiters  map[string]*rate.Limiter
	mu        *sync.RWMutex
	rateLimit rate.Limit // rateLimit defines the maximum frequency of the requests.
	burstSize int        // burstSize is the maximum number of requests that can be performed at once.
}

// NewLimit returns a new Limit with the specified request and burst rate limit per second.
func NewLimit(limit rate.Limit, burst int) *Limit {
	// Check burst to enforce a minimum size of 3, or disable rate limiting if less than 1.
	if burst < 1 {
		// A burst of zero does not allow any requests unless limit == Inf, which disables rate limiting:
		limit = rate.Inf
		burst = 0
	} else if burst < 3 {
		// If rate limiting is not deactivated, the minimum burst must be 3 for 2-Factor Authentication (2FA)
		// to work, as 3 tokens must be available for the authentication check:
		burst = 3
	}

	// Create and return new IP-based rate limiter.
	return &Limit{
		limiters:  make(map[string]*rate.Limiter),
		mu:        &sync.RWMutex{},
		rateLimit: limit,
		burstSize: burst,
	}
}

// AddIP adds a new rate limiter for the specified IP address.
func (i *Limit) AddIP(ip string) *rate.Limiter {
	if ip == "" {
		ip = DefaultIP
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.rateLimit, i.burstSize)

	i.limiters[ip] = limiter

	return limiter
}

// IP returns the rate limiter for the specified IP address.
// TODO: Normalize IPv6 addresses so that hosts with multiple addresses cannot be used for spray attacks.
func (i *Limit) IP(ip string) *rate.Limiter {
	// Default to 0.0.0.0 if no address was provided.
	if ip == "" {
		ip = DefaultIP
	}

	i.mu.RLock()
	limiter, exists := i.limiters[ip]

	if !exists {
		i.mu.RUnlock()
		return i.AddIP(ip)
	}

	i.mu.RUnlock()

	return limiter
}

// Allow checks if a new request is allowed at this time and increments the request counter by 1.
func (i *Limit) Allow(ip string) bool {
	return i.IP(ip).Allow()
}

// AllowN checks if a new request is allowed at this time and increments the request counter by n.
func (i *Limit) AllowN(ip string, n int) bool {
	return i.IP(ip).AllowN(time.Now(), n)
}

// Request tries to increment the request counter and returns the result as new *Request.
func (i *Limit) Request(ip string) *Request {
	return NewRequest(i.IP(ip), 1)
}

// RequestN tries to increment the request counter by n and returns the result as new *Request.
func (i *Limit) RequestN(ip string, n int) *Request {
	return NewRequest(i.IP(ip), n)
}

// Reserve increments the request counter and returns a rate.Reservation.
func (i *Limit) Reserve(ip string) *rate.Reservation {
	return i.IP(ip).Reserve()
}

// ReserveN increments the request counter by n and returns a rate.Reservation.
func (i *Limit) ReserveN(ip string, n int) *rate.Reservation {
	return i.IP(ip).ReserveN(time.Now(), n)
}

// Reject checks if the request rate limit has been exceeded, but does not modify the counter.
func (i *Limit) Reject(ip string) bool {
	return i.IP(ip).Tokens() < 1
}
