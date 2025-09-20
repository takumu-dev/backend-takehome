package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"blog-platform/internal/infrastructure/config"
	"blog-platform/internal/infrastructure/http/errors"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	// RequestsPerSecond defines the rate limit (requests per second)
	RequestsPerSecond float64
	// BurstSize defines the maximum burst size
	BurstSize int
	// SkipSuccessful when true, only failed requests count towards the limit
	SkipSuccessful bool
	// KeyGenerator generates the key for rate limiting (default: IP address)
	KeyGenerator func(c echo.Context) string
}

// DefaultRateLimiterConfig returns default configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 10,  // 10 requests per second
		BurstSize:         20,  // Allow burst of 20 requests
		SkipSuccessful:    false,
		KeyGenerator: func(c echo.Context) string {
			return c.RealIP()
		},
	}
}

// AuthRateLimiterConfig returns configuration for auth endpoints
func AuthRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 2,   // 2 requests per second for auth
		BurstSize:         5,   // Allow burst of 5 requests
		SkipSuccessful:    false,
		KeyGenerator: func(c echo.Context) string {
			return c.RealIP()
		},
	}
}

// rateLimiterEntry represents a rate limiter instance
type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// cleanupExpiredLimiters removes expired rate limiters
func cleanupExpiredLimiters(limiters map[string]*rateLimiterEntry, mu *sync.RWMutex) {
	mu.Lock()
	defer mu.Unlock()
	
	cutoff := time.Now().Add(-time.Hour)
	for key, entry := range limiters {
		if entry.lastSeen.Before(cutoff) {
			delete(limiters, key)
		}
	}
}

// RateLimiterMiddleware creates a rate limiting middleware with configuration
func RateLimiterMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	return RateLimiterWithConfig(RateLimiterConfig{
		RequestsPerSecond: cfg.RateLimit.DefaultRequestsPerSecond,
		BurstSize:         cfg.RateLimit.DefaultBurstSize,
		KeyGenerator: func(c echo.Context) string {
			return c.RealIP()
		},
	})
}

// AuthRateLimiterMiddleware creates a rate limiting middleware for auth endpoints
func AuthRateLimiterMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	return RateLimiterWithConfig(RateLimiterConfig{
		RequestsPerSecond: cfg.RateLimit.AuthRequestsPerSecond,
		BurstSize:         cfg.RateLimit.AuthBurstSize,
		KeyGenerator: func(c echo.Context) string {
			return c.RealIP()
		},
	})
}

// RateLimiterWithConfig creates a rate limiting middleware with custom config
func RateLimiterWithConfig(config RateLimiterConfig) echo.MiddlewareFunc {
	limiterMap := make(map[string]*rateLimiterEntry)
	mu := sync.RWMutex{}

	// Cleanup goroutine to remove expired limiters
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cleanupExpiredLimiters(limiterMap, &mu)
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientIP := config.KeyGenerator(c)

			mu.Lock()
			entry, exists := limiterMap[clientIP]
			if !exists || time.Since(entry.lastSeen) > time.Hour {
				entry = &rateLimiterEntry{
					limiter:  rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.BurstSize),
					lastSeen: time.Now(),
				}
				limiterMap[clientIP] = entry
			} else {
				entry.lastSeen = time.Now()
			}
			mu.Unlock()

			if !entry.limiter.Allow() {
				// Set rate limit headers
				c.Response().Header().Set("X-RateLimit-Limit", strconv.FormatFloat(config.RequestsPerSecond, 'f', -1, 64))
				c.Response().Header().Set("X-RateLimit-Remaining", "0")
				c.Response().Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Second).Unix(), 10))

				return errors.HandleError(c, errors.ErrTooManyRequests)
			}

			// Set rate limit headers for successful requests
			c.Response().Header().Set("X-RateLimit-Limit", strconv.FormatFloat(config.RequestsPerSecond, 'f', -1, 64))
			// Note: Getting exact remaining tokens from rate.Limiter is not straightforward
			// This is an approximation
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(config.BurstSize-1))

			return next(c)
		}
	}
}

// RateLimitDefault returns rate limiting middleware with default config
func RateLimitDefault() echo.MiddlewareFunc {
	return RateLimiterWithConfig(DefaultRateLimiterConfig())
}

// RateLimitAuth returns rate limiting middleware for auth endpoints
func RateLimitAuth() echo.MiddlewareFunc {
	return RateLimiterWithConfig(AuthRateLimiterConfig())
}
