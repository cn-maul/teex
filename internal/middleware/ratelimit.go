package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// visitor tracks a single IP's request count within a time window.
type visitor struct {
	mu       sync.Mutex
	count    int
	lastSeen time.Time
}

// RateLimiterConfig holds the rate limit parameters for a group of routes.
type RateLimiterConfig struct {
	MaxRequests int
	Window      time.Duration
}

// allVisitors tracks all rate limiter instances for a shared cleanup goroutine.
var (
	globalMu       sync.Mutex
	allLimiterMaps []map[string]*visitor
	allLimiterCfgs []RateLimiterConfig
	allMapMus      []*sync.Mutex
	cleanupOnce    sync.Once
	cleanupQuit    = make(chan struct{})
)

// StopCleanup signals the background cleanup goroutine to exit.
// Call this during graceful server shutdown.
func StopCleanup() {
	close(cleanupQuit)
}

// startCleanup launches a single background goroutine that sweeps expired
// entries from every registered rate limiter map. It is safe to call multiple
// times -- only the first call actually starts the goroutine.
func startCleanup() {
	cleanupOnce.Do(func() {
		go func() {
			for {
				select {
				case <-time.After(time.Minute):
					globalMu.Lock()
					for i, visitors := range allLimiterMaps {
						mu := allMapMus[i]
						window := allLimiterCfgs[i].Window
						mu.Lock()
						for ip, v := range visitors {
							v.mu.Lock()
							if time.Since(v.lastSeen) > window*2 {
								delete(visitors, ip)
							}
							v.mu.Unlock()
						}
						mu.Unlock()
					}
					globalMu.Unlock()
				case <-cleanupQuit:
					return
				}
			}
		}()
	})
}

// RateLimiter returns a Gin middleware that limits requests per IP using the
// provided configuration.  An in-memory map is used so no external dependency
// is required.  A single shared background goroutine handles cleanup for all
// rate limiter instances.
func RateLimiter(cfg RateLimiterConfig) gin.HandlerFunc {
	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
	)

	// Register this limiter's maps with the global cleanup coordinator.
	globalMu.Lock()
	allLimiterMaps = append(allLimiterMaps, visitors)
	allLimiterCfgs = append(allLimiterCfgs, cfg)
	allMapMus = append(allMapMus, &mu)
	globalMu.Unlock()
	startCleanup()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Get or create the visitor entry.
		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			v = &visitor{}
			visitors[ip] = v
		}
		mu.Unlock()

		v.mu.Lock()

		now := time.Now()
		// Reset the window if it has expired.
		if now.Sub(v.lastSeen) > cfg.Window {
			v.count = 0
			v.lastSeen = now
		}

		v.count++
		count := v.count
		v.lastSeen = now
		v.mu.Unlock()

		if count > cfg.MaxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Pre-built rate limiter configurations for common endpoint groups.
var (
	// LoginRateLimit allows 10 requests per minute per IP.
	LoginRateLimit = RateLimiterConfig{MaxRequests: 10, Window: time.Minute}
	// RegisterRateLimit allows 5 requests per minute per IP.
	RegisterRateLimit = RateLimiterConfig{MaxRequests: 5, Window: time.Minute}
	// GeneralRateLimit allows 120 requests per minute per IP.
	GeneralRateLimit = RateLimiterConfig{MaxRequests: 120, Window: time.Minute}
)
