package limiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limit represents an IP request rate limit.
type Limit struct {
	limiters  map[string]*rate.Limiter
	mu        *sync.RWMutex
	rateLimit rate.Limit
	burstSize int
}

// NewLimit returns a new Limit with the specified request and burst rate limit per second.
func NewLimit(r rate.Limit, b int) *Limit {
	i := &Limit{
		limiters:  make(map[string]*rate.Limiter),
		mu:        &sync.RWMutex{},
		rateLimit: r,
		burstSize: b,
	}

	return i
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
func (i *Limit) IP(ip string) *rate.Limiter {
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
