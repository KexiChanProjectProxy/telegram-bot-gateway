package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter implements a token bucket rate limiter using Redis
type RateLimiter struct {
	client               *redis.Client
	requestsPerSecond    int
	burst                int
	cleanupInterval      time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(client *redis.Client, requestsPerSecond, burst int, cleanupInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		client:            client,
		requestsPerSecond: requestsPerSecond,
		burst:             burst,
		cleanupInterval:   cleanupInterval,
	}
}

// RateLimitMiddleware creates a middleware for rate limiting
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get identifier for rate limiting
		identifier := getIdentifier(c)

		// Check rate limit
		allowed, remaining, resetAt, err := limiter.Allow(c.Request.Context(), identifier)
		if err != nil {
			// On error, log but allow request to proceed
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.requestsPerSecond))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if !allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(resetAt.Sub(time.Now()).Seconds()), 10))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Limit: %d requests per second. Try again after %s",
					limiter.requestsPerSecond, resetAt.Format(time.RFC3339)),
				"retry_after": resetAt.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Allow checks if a request should be allowed based on rate limit
func (rl *RateLimiter) Allow(ctx context.Context, identifier string) (allowed bool, remaining int, resetAt time.Time, err error) {
	now := time.Now()
	key := fmt.Sprintf("ratelimit:%s", identifier)

	// Use Lua script for atomic rate limiting check
	script := redis.NewScript(`
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		-- Get current count and timestamp
		local current = redis.call('GET', key)

		if current == false then
			-- First request in this window
			redis.call('SET', key, 1, 'EX', window)
			return {1, limit - 1, now + window}
		end

		local count = tonumber(current)

		if count < limit then
			-- Increment count
			redis.call('INCR', key)
			local ttl = redis.call('TTL', key)
			return {1, limit - count - 1, now + ttl}
		else
			-- Rate limit exceeded
			local ttl = redis.call('TTL', key)
			return {0, 0, now + ttl}
		end
	`)

	result, err := script.Run(ctx, rl.client,
		[]string{key},
		rl.requestsPerSecond,
		1, // 1 second window
		now.Unix(),
	).Result()

	if err != nil {
		return false, 0, time.Time{}, err
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) != 3 {
		return false, 0, time.Time{}, fmt.Errorf("unexpected result format")
	}

	allowedInt, _ := resultSlice[0].(int64)
	remainingInt, _ := resultSlice[1].(int64)
	resetAtInt, _ := resultSlice[2].(int64)

	return allowedInt == 1, int(remainingInt), time.Unix(resetAtInt, 0), nil
}

// getIdentifier returns the identifier for rate limiting
func getIdentifier(c *gin.Context) string {
	// Try to get from auth context first
	authCtx, exists := GetAuthContext(c)
	if exists {
		if authCtx.IsAPIKey {
			return fmt.Sprintf("apikey:%d", *authCtx.APIKeyID)
		}
		return fmt.Sprintf("user:%d", authCtx.UserID)
	}

	// Fall back to IP address
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// PerUserRateLimitMiddleware creates a rate limiter specific to authenticated users
func PerUserRateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, exists := GetAuthContext(c)
		if !exists {
			// No auth context, skip rate limiting
			c.Next()
			return
		}

		var identifier string
		if authCtx.IsAPIKey {
			identifier = fmt.Sprintf("apikey:%d", *authCtx.APIKeyID)
		} else {
			identifier = fmt.Sprintf("user:%d", authCtx.UserID)
		}

		allowed, remaining, resetAt, err := limiter.Allow(c.Request.Context(), identifier)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.requestsPerSecond))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if !allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(resetAt.Sub(time.Now()).Seconds()), 10))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "You have exceeded your request quota",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GlobalRateLimitMiddleware creates a global rate limiter for all requests
func GlobalRateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, remaining, resetAt, err := limiter.Allow(c.Request.Context(), "global")
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.requestsPerSecond))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if !allowed {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service temporarily unavailable",
				"message": "The service is experiencing high load. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SlidingWindowRateLimiter implements a sliding window rate limiter
type SlidingWindowRateLimiter struct {
	client    *redis.Client
	limit     int
	window    time.Duration
}

// NewSlidingWindowRateLimiter creates a new sliding window rate limiter
func NewSlidingWindowRateLimiter(client *redis.Client, limit int, window time.Duration) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// Allow checks if a request is allowed using sliding window algorithm
func (sw *SlidingWindowRateLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	now := time.Now()
	key := fmt.Sprintf("ratelimit:sw:%s", identifier)
	windowStart := now.Add(-sw.window).UnixNano()

	// Lua script for atomic sliding window check
	script := redis.NewScript(`
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local ttl = tonumber(ARGV[4])

		-- Remove old entries
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

		-- Count current requests
		local count = redis.call('ZCARD', key)

		if count < limit then
			-- Add current request
			redis.call('ZADD', key, now, now)
			redis.call('EXPIRE', key, ttl)
			return 1
		else
			return 0
		end
	`)

	result, err := script.Run(ctx, sw.client,
		[]string{key},
		sw.limit,
		windowStart,
		now.UnixNano(),
		int(sw.window.Seconds())+1,
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// SlidingWindowMiddleware creates middleware for sliding window rate limiting
func SlidingWindowMiddleware(limiter *SlidingWindowRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := getIdentifier(c)

		allowed, err := limiter.Allow(c.Request.Context(), identifier)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
