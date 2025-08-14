package ratelimiter

import (
	"golang.org/x/time/rate"
	"sync"
)

type DomainRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	limit    rate.Limit
	burst    int
}

func NewDomainRateLimiter(limit rate.Limit, burst int) *DomainRateLimiter {
	return &DomainRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    limit,
		burst:    burst,
	}
}

func (d *DomainRateLimiter) GetLimiter(domain string) *rate.Limiter {
	d.mu.Lock()
	defer d.mu.Unlock()

	limiter, exists := d.limiters[domain]
	if !exists {
		limiter = rate.NewLimiter(d.limit, d.burst)
		d.limiters[domain] = limiter
	}

	return limiter
}
