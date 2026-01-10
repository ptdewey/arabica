package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func TestResponseWriter_WriteHeader(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: recorder,
			statusCode:     http.StatusOK,
		}

		rw.WriteHeader(http.StatusNotFound)

		if rw.statusCode != http.StatusNotFound {
			t.Errorf("statusCode = %d, want %d", rw.statusCode, http.StatusNotFound)
		}
		if !rw.wroteHeader {
			t.Error("wroteHeader should be true after WriteHeader")
		}
	})

	t.Run("only writes header once", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: recorder,
			statusCode:     http.StatusOK,
		}

		rw.WriteHeader(http.StatusNotFound)
		rw.WriteHeader(http.StatusInternalServerError) // Should be ignored

		if rw.statusCode != http.StatusNotFound {
			t.Errorf("statusCode = %d, want %d (second WriteHeader should be ignored)", rw.statusCode, http.StatusNotFound)
		}
	})
}

func TestResponseWriter_Write(t *testing.T) {
	t.Run("counts bytes written", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: recorder,
			statusCode:     http.StatusOK,
		}

		n, err := rw.Write([]byte("hello"))
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		if n != 5 {
			t.Errorf("Write() returned %d, want 5", n)
		}
		if rw.bytesWritten != 5 {
			t.Errorf("bytesWritten = %d, want 5", rw.bytesWritten)
		}
	})

	t.Run("accumulates bytes from multiple writes", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: recorder,
			statusCode:     http.StatusOK,
		}

		rw.Write([]byte("hello"))
		rw.Write([]byte(" world"))

		if rw.bytesWritten != 11 {
			t.Errorf("bytesWritten = %d, want 11", rw.bytesWritten)
		}
	})

	t.Run("Write sets header if not already written", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: recorder,
			statusCode:     http.StatusOK,
		}

		rw.Write([]byte("test"))

		if !rw.wroteHeader {
			t.Error("wroteHeader should be true after Write")
		}
		if rw.statusCode != http.StatusOK {
			t.Errorf("statusCode = %d, want %d", rw.statusCode, http.StatusOK)
		}
	})
}

func TestLoggingMiddleware(t *testing.T) {
	t.Run("logs request details", func(t *testing.T) {
		// Create a buffer to capture log output
		var buf bytes.Buffer
		logger := zerolog.New(&buf)

		// Create a simple handler that returns 200 OK
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Wrap with logging middleware
		middleware := LoggingMiddleware(logger)
		wrapped := middleware(handler)

		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/test?q=1", nil)
		req.Header.Set("User-Agent", "test-agent")
		recorder := httptest.NewRecorder()

		// Execute
		wrapped.ServeHTTP(recorder, req)

		// Verify response
		if recorder.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusOK)
		}

		// Verify log contains expected fields
		logOutput := buf.String()
		expectedFields := []string{
			`"method":"GET"`,
			`"path":"/test"`,
			`"query":"q=1"`,
			`"status":200`,
			`"user_agent":"test-agent"`,
		}

		for _, field := range expectedFields {
			if !bytes.Contains([]byte(logOutput), []byte(field)) {
				t.Errorf("log output missing %s, got: %s", field, logOutput)
			}
		}
	})

	t.Run("handles 404 status", func(t *testing.T) {
		var buf bytes.Buffer
		logger := zerolog.New(&buf)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		middleware := LoggingMiddleware(logger)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
		recorder := httptest.NewRecorder()

		wrapped.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusNotFound)
		}

		logOutput := buf.String()
		if !bytes.Contains([]byte(logOutput), []byte(`"status":404`)) {
			t.Errorf("log should contain status 404, got: %s", logOutput)
		}
		if !bytes.Contains([]byte(logOutput), []byte(`"level":"warn"`)) {
			t.Errorf("log should be warn level for 4xx status, got: %s", logOutput)
		}
	})

	t.Run("handles 500 status", func(t *testing.T) {
		var buf bytes.Buffer
		logger := zerolog.New(&buf)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		middleware := LoggingMiddleware(logger)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		recorder := httptest.NewRecorder()

		wrapped.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusInternalServerError)
		}

		logOutput := buf.String()
		if !bytes.Contains([]byte(logOutput), []byte(`"status":500`)) {
			t.Errorf("log should contain status 500, got: %s", logOutput)
		}
		if !bytes.Contains([]byte(logOutput), []byte(`"level":"error"`)) {
			t.Errorf("log should be error level for 5xx status, got: %s", logOutput)
		}
	})

	t.Run("includes referer when present", func(t *testing.T) {
		var buf bytes.Buffer
		logger := zerolog.New(&buf)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := LoggingMiddleware(logger)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Referer", "https://example.com/previous")
		recorder := httptest.NewRecorder()

		wrapped.ServeHTTP(recorder, req)

		logOutput := buf.String()
		if !bytes.Contains([]byte(logOutput), []byte(`"referer":"https://example.com/previous"`)) {
			t.Errorf("log should contain referer, got: %s", logOutput)
		}
	})

	t.Run("includes request ID when present", func(t *testing.T) {
		var buf bytes.Buffer
		logger := zerolog.New(&buf)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := LoggingMiddleware(logger)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "abc-123-xyz")
		recorder := httptest.NewRecorder()

		wrapped.ServeHTTP(recorder, req)

		logOutput := buf.String()
		if !bytes.Contains([]byte(logOutput), []byte(`"request_id":"abc-123-xyz"`)) {
			t.Errorf("log should contain request_id, got: %s", logOutput)
		}
	})

	t.Run("tracks bytes written", func(t *testing.T) {
		var buf bytes.Buffer
		logger := zerolog.New(&buf)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello world"))
		})

		middleware := LoggingMiddleware(logger)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		recorder := httptest.NewRecorder()

		wrapped.ServeHTTP(recorder, req)

		logOutput := buf.String()
		if !bytes.Contains([]byte(logOutput), []byte(`"bytes_written":11`)) {
			t.Errorf("log should contain bytes_written:11, got: %s", logOutput)
		}
	})
}
