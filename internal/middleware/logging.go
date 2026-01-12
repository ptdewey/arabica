package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"arabica/internal/atproto"

	"github.com/rs/zerolog"
)

// getClientIP extracts the real client IP address from the request,
// checking X-Forwarded-For and X-Real-IP headers for reverse proxy setups.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (can contain multiple IPs: client, proxy1, proxy2)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (the original client)
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header (single IP set by some proxies)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr (strip port if present)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr might not have a port
		return r.RemoteAddr
	}
	return ip
}

// LoggingMiddleware returns a middleware that logs HTTP request details with structured logging
func LoggingMiddleware(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code and bytes written
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				bytesWritten:   0,
			}

			// Call the next handler
			next.ServeHTTP(rw, r)

			// Calculate duration
			duration := time.Since(start)

			// Select log level based on status code
			var logEvent *zerolog.Event
			if rw.statusCode >= 500 {
				logEvent = logger.Error()
			} else if rw.statusCode >= 400 {
				logEvent = logger.Warn()
			} else {
				logEvent = logger.Info()
			}

			// Add core fields
			logEvent.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("query", r.URL.RawQuery).
				Int("status", rw.statusCode).
				Dur("duration", duration).
				Str("client_ip", getClientIP(r)).
				Str("user_agent", r.UserAgent()).
				Int64("bytes_written", rw.bytesWritten).
				Str("proto", r.Proto)

			// Add optional fields only if present
			if referer := r.Referer(); referer != "" {
				logEvent.Str("referer", referer)
			}
			if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
				logEvent.Str("request_id", reqID)
			}
			if contentType := r.Header.Get("Content-Type"); contentType != "" {
				logEvent.Str("content_type", contentType)
			}
			if did, err := atproto.GetAuthenticatedDID(r.Context()); err == nil && did != "" {
				logEvent.Str("user_did", did)
			}

			// Log all request headers for debugging malicious traffic
			headers := make(map[string]string)
			for name, values := range r.Header {
				headers[name] = strings.Join(values, ", ")
			}
			logEvent.Interface("headers", headers)

			logEvent.Msg("HTTP request")
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	wroteHeader  bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}
