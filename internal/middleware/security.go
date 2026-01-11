package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// XSS protection (legacy but still useful for older browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy - disable unnecessary features
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy
		// Allows: self for scripts/styles, inline styles (for Tailwind), unpkg for HTMX/Alpine
		csp := strings.Join([]string{
			"default-src 'self'",
			"script-src 'self' https://unpkg.com",
			"style-src 'self' 'unsafe-inline'", // unsafe-inline needed for Tailwind
			"img-src 'self' https: data:",      // Allow external images (avatars) and data URIs
			"font-src 'self'",
			"connect-src 'self'",
			"frame-ancestors 'none'",
			"base-uri 'self'",
			"form-action 'self'",
		}, "; ")
		w.Header().Set("Content-Security-Policy", csp)

		next.ServeHTTP(w, r)
	})
}

// RateLimiter implements a simple per-IP rate limiter using token bucket algorithm
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // requests per window
	window   time.Duration // time window
	cleanup  time.Duration // cleanup interval for old entries
}

type visitor struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per window
// window: time window for rate limiting
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
		cleanup:  window * 2,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.Sub(v.lastReset) > rl.cleanup {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:    rl.rate - 1, // Use one token
			lastReset: now,
		}
		return true
	}

	// Reset tokens if window has passed
	if now.Sub(v.lastReset) >= rl.window {
		v.tokens = rl.rate - 1
		v.lastReset = now
		return true
	}

	// Check if tokens available
	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

// RateLimitConfig holds configuration for rate limiting different endpoint types
type RateLimitConfig struct {
	// AuthLimiter for login/auth endpoints (stricter)
	AuthLimiter *RateLimiter
	// APILimiter for general API endpoints
	APILimiter *RateLimiter
	// GlobalLimiter for all other requests
	GlobalLimiter *RateLimiter
}

// NewDefaultRateLimitConfig creates rate limiters with sensible defaults
func NewDefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		AuthLimiter:   NewRateLimiter(5, time.Minute),   // 5 auth attempts per minute
		APILimiter:    NewRateLimiter(60, time.Minute),  // 60 API calls per minute
		GlobalLimiter: NewRateLimiter(120, time.Minute), // 120 requests per minute
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(config *RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			path := r.URL.Path

			var limiter *RateLimiter

			// Select appropriate limiter based on path
			switch {
			case strings.HasPrefix(path, "/auth/") || path == "/login" || path == "/oauth/callback":
				limiter = config.AuthLimiter
			case strings.HasPrefix(path, "/api/"):
				limiter = config.APILimiter
			default:
				limiter = config.GlobalLimiter
			}

			if !limiter.Allow(ip) {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP is defined in logging.go

// MaxBodySize limits the size of request bodies
const (
	MaxJSONBodySize = 1 << 20 // 1 MB for JSON requests
	MaxFormBodySize = 1 << 20 // 1 MB for form submissions
)

// LimitBodyMiddleware limits request body size to prevent DoS
func LimitBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			contentType := r.Header.Get("Content-Type")
			var maxSize int64

			switch {
			case strings.HasPrefix(contentType, "application/json"):
				maxSize = MaxJSONBodySize
			case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"),
				strings.HasPrefix(contentType, "multipart/form-data"):
				maxSize = MaxFormBodySize
			default:
				maxSize = MaxJSONBodySize // Default limit
			}

			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
		}

		next.ServeHTTP(w, r)
	})
}
