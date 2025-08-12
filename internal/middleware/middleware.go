package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"ads-txt-service/internal/logger"
	"ads-txt-service/internal/ratelimit"
)

type clientLimiter struct {
	limiter  *ratelimit.TokenBucket
	lastSeen time.Time
}

type RateLimiter struct {
	clientLimiters map[string]*clientLimiter
	mu             sync.Mutex
	capacity       int
	refillPeriod   time.Duration
	cleanupPeriod  time.Duration
	log            *logger.Logger
}

func NewRateLimiter(capacity int, refillPeriod time.Duration, lg *logger.Logger) *RateLimiter {
	rl := &RateLimiter{
		clientLimiters: make(map[string]*clientLimiter),
		capacity:       capacity,
		refillPeriod:   refillPeriod,
		cleanupPeriod:  1 * time.Minute,
		log:            lg,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupPeriod)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, cl := range rl.clientLimiters {
			if time.Since(cl.lastSeen) > rl.cleanupPeriod {
				delete(rl.clientLimiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (rl *RateLimiter) RateLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			rl.log.Debug("[ratelimit] MIDDLEWARE called, ip=%s path=%s method=%s\n", clientIP, r.URL.Path, r.Method)

			rl.mu.Lock()
			cl, exists := rl.clientLimiters[clientIP]
			if !exists {
				rl.log.Debug("[ratelimit] creating limiter for ip=%s\n", clientIP)
				cl = &clientLimiter{
					limiter:  ratelimit.NewTokenBucket(rl.capacity, rl.refillPeriod),
					lastSeen: time.Now(),
				}
				rl.clientLimiters[clientIP] = cl
			} else {
				cl.lastSeen = time.Now()
			}
			limiter := cl.limiter
			rl.mu.Unlock()

			allowed, remaining := limiter.AllowWithRemaining()
			if !allowed {
				rl.log.Info("[ratelimit] BLOCK ip=%s remaining=%.2f\n", clientIP, remaining)
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			rl.log.Info("[ratelimit] ALLOW ip=%s remaining=%.2f\n", clientIP, remaining)
			next.ServeHTTP(w, r)
		})
	}
}
