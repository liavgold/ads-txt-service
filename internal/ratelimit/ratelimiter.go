package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu         sync.Mutex
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
}

func NewTokenBucket(capacity int, refillPeriod time.Duration) *TokenBucket {
	sec := refillPeriod.Seconds()
	if sec <= 0 {
		sec = 1.0
	}
	return &TokenBucket{
		capacity:   float64(capacity),
		tokens:     float64(capacity),
		refillRate: float64(capacity) / sec,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) refillLocked() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	if elapsed <= 0 {
		return
	}
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refillLocked()
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

func (tb *TokenBucket) AllowWithRemaining() (bool, float64) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refillLocked()
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true, tb.tokens
	}
	return false, tb.tokens
}

func (tb *TokenBucket) Remaining() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refillLocked()
	return tb.tokens
}
