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
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.rateLimit, i.burstSize)

	i.limiters[ip] = limiter

	return limiter
}

// IP returns the rate limiter for the specified IP address.
func (i *Limit) IP(ip string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.limiters[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}

	i.mu.Unlock()

	return limiter
}

// Allow reports whether the request is allowed at this time and increments the request counter.
func (i *Limit) Allow(ip string) bool {
	return i.IP(ip).Allow()
}

// Reserve increments the request counter and returns a rate.Reservation.
func (i *Limit) Reserve(ip string) *rate.Reservation {
	return i.IP(ip).Reserve()
}

// ReserveN increments the request counter by n and returns a rate.Reservation.
func (i *Limit) ReserveN(ip string, n int) *rate.Reservation {
	return i.IP(ip).ReserveN(time.Now(), n)
}

// Reject reports whether the request limit has been exceeded, but does not change the request counter.
func (i *Limit) Reject(ip string) bool {
	return i.IP(ip).Tokens() < 1
}
