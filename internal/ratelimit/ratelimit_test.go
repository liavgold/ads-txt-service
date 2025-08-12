package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestTokenBucket_Allow(t *testing.T) {
	t.Run("AllowRequestWithTokens", func(t *testing.T) {
		tb := NewTokenBucket(3, time.Second)
		if !tb.Allow() {
			t.Error("Expected Allow to return true for initial token")
		}
		if !tb.Allow() {
			t.Error("Expected Allow to return true for second token")
		}
		if !tb.Allow() {
			t.Error("Expected Allow to return true for third token")
		}
		if tb.Allow() {
			t.Error("Expected Allow to return false when tokens are depleted")
		}
	})

	t.Run("RefillTokens", func(t *testing.T) {
		tb := NewTokenBucket(2, 100*time.Millisecond)
		if !tb.Allow() {
			t.Error("Expected Allow to return true for initial token")
		}
		if !tb.Allow() {
			t.Error("Expected Allow to return true for second token")
		}
		if tb.Allow() {
			t.Error("Expected Allow to return false when tokens are depleted")
		}

		time.Sleep(150 * time.Millisecond)
		if !tb.Allow() {
			t.Error("Expected Allow to return true after refill")
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		tb := NewTokenBucket(5, time.Second)
		var wg sync.WaitGroup
		attempts := 10
		results := make(chan bool, attempts)

		for i := 0; i < attempts; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				results <- tb.Allow()
			}()
		}
		wg.Wait()
		close(results)

		allowed := 0
		for result := range results {
			if result {
				allowed++
			}
		}
		if allowed != 5 {
			t.Errorf("Expected exactly 5 allowed requests, got %d", allowed)
		}
	})

	t.Run("NoOverfillCapacity", func(t *testing.T) {
		tb := NewTokenBucket(2, 100*time.Millisecond)
		time.Sleep(500 * time.Millisecond)
		tb.Allow()                         
		tb.Allow()                        
		if tb.Allow() {
			t.Error("Expected Allow to return false when tokens are depleted after refill")
		}
	})
}
