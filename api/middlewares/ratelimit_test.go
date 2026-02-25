package middlewares

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit_FreshAllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "192.168.99.1:12345" // unique IP so global limiter is effectively fresh for this IP

	nextCalled := false
	next := func(c *gin.Context) {
		nextCalled = true
		c.Status(http.StatusOK)
	}

	handler := RateLimitMiddleware()
	handler(c)
	next(c)

	assert.True(t, nextCalled, "next handler should be called")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimit_OverLimitRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testIP := "10.0.0.42"
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    2,
		window:   time.Minute,
	}
	now := time.Now()
	rl.requests[testIP] = []time.Time{now, now}

	oldGet := getLimiter
	getLimiter = func() *rateLimiter { return rl }
	defer func() { getLimiter = oldGet }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = testIP + ":0"

	nextCalled := false
	next := func(c *gin.Context) {
		nextCalled = true
		c.Status(http.StatusOK)
	}

	handler := RateLimitMiddleware()
	handler(c)
	next(c)

	assert.False(t, nextCalled, "next handler should not be called when over limit")
	assert.Equal(t, 429, w.Code)
}

func TestRateLimit_CleanupExpiredEntries(t *testing.T) {
	window := time.Minute
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    10,
		window:   window,
	}
	old := time.Now().Add(-2 * window)
	rl.requests["1.2.3.4"] = []time.Time{old, old.Add(-time.Hour)}
	rl.requests["5.6.7.8"] = []time.Time{time.Now()} // still valid

	rl.cleanup()

	assert.NotContains(t, rl.requests, "1.2.3.4")
	assert.Contains(t, rl.requests, "5.6.7.8")
	assert.Len(t, rl.requests["5.6.7.8"], 1)
}

func TestRateLimit_ConcurrentAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testIP := "172.16.0.1"
	limit := 10
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   time.Minute,
	}

	oldGet := getLimiter
	getLimiter = func() *rateLimiter { return rl }
	defer func() { getLimiter = oldGet }()

	handler := RateLimitMiddleware()
	numGoroutines := 50
	var allowed, rejected int32
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Request.RemoteAddr = testIP + ":0"
			handler(c)
			if c.IsAborted() {
				atomic.AddInt32(&rejected, 1)
			} else {
				atomic.AddInt32(&allowed, 1)
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, int32(limit), allowed, "exactly limit requests should be allowed")
	assert.Equal(t, int32(numGoroutines-limit), rejected, "rest should be rejected")
}
