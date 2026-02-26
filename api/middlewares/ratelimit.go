package middlewares

import (
	"sync"
	"time"

	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

var limiter = &rateLimiter{
	requests: make(map[string][]time.Time),
	limit:    2000, // 每分钟最多 2000 次请求
	window:   time.Minute,
}

// getLimiter is used by the middleware; tests may override to inject a local rateLimiter.
var getLimiter = func() *rateLimiter { return limiter }

func init() {
	go limiter.cleanupLoop()
}

// cleanup runs one pass of removing expired entries; used by cleanupLoop and by tests.
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for ip, requests := range rl.requests {
		var valid []time.Time
		for _, t := range requests {
			if now.Sub(t) < rl.window {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = valid
		}
	}
}

func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.cleanup()
	}
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rl := getLimiter()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		requests := rl.requests[ip]

		var validRequests []time.Time
		for _, t := range requests {
			if now.Sub(t) < rl.window {
				validRequests = append(validRequests, t)
			}
		}

		if len(validRequests) >= rl.limit {
			response.Error(c, 429, "RATE_LIMIT_EXCEEDED", "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		validRequests = append(validRequests, now)
		rl.requests[ip] = validRequests

		c.Next()
	}
}
