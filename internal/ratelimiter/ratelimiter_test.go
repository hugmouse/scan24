package ratelimiter

import (
	"golang.org/x/time/rate"
	"testing"
	"time"
)

func TestDomainRateLimiter(t *testing.T) {
	limiter := NewDomainRateLimiter(rate.Limit(1), 1)

	domainLimiter := limiter.GetLimiter("example.com")

	if !domainLimiter.Allow() {
		t.Error("Expected first request to be allowed")
	}

	if domainLimiter.Allow() {
		t.Error("Expected second request to be denied")
	}

	time.Sleep(time.Second)

	if !domainLimiter.Allow() {
		t.Error("Expected third request to be allowed after waiting")
	}
}

func TestDomainRateLimiter_DifferentDomains(t *testing.T) {
	limiter := NewDomainRateLimiter(rate.Limit(1), 1)

	domainLimiter1 := limiter.GetLimiter("example.com")
	domainLimiter2 := limiter.GetLimiter("test.com")

	if !domainLimiter1.Allow() {
		t.Error("Expected first request to example.com to be allowed")
	}
	if !domainLimiter2.Allow() {
		t.Error("Expected first request to test.com to be allowed")
	}

	if domainLimiter1.Allow() {
		t.Error("Expected second request to example.com to be denied")
	}
	if domainLimiter2.Allow() {
		t.Error("Expected second request to test.com to be denied")
	}
}
