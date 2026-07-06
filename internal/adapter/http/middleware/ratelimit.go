package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type clientBucket struct {
	tokens    float64
	lastCheck time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientBucket
	rate     float64
	capacity float64
}

func NewRateLimiter(rps int) *RateLimiter {
	if rps <= 0 {
		rps = 50
	}
	return &RateLimiter{
		clients:  make(map[string]*clientBucket),
		rate:     float64(rps),
		capacity: float64(rps),
	}
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			rl.mu.Lock()
			now := time.Now()
			bucket, exists := rl.clients[ip]
			if !exists {
				bucket = &clientBucket{tokens: rl.capacity, lastCheck: now}
				rl.clients[ip] = bucket
			}

			elapsed := now.Sub(bucket.lastCheck).Seconds()
			bucket.lastCheck = now
			bucket.tokens += elapsed * rl.rate
			if bucket.tokens > rl.capacity {
				bucket.tokens = rl.capacity
			}

			if bucket.tokens < 1.0 {
				rl.mu.Unlock()
				w.Header().Set("Retry-After", "1")
				http.Error(w, `{"error":"rate_limit_exceeded"}`, http.StatusTooManyRequests)
				return
			}

			bucket.tokens -= 1.0
			rl.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
