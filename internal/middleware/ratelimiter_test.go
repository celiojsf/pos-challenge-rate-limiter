package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/celiojsf/pos-challenge-rate-limiter/internal/limiter"
	"github.com/celiojsf/pos-challenge-rate-limiter/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiterMiddleware_IPRateLimit(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	rl := limiter.NewRateLimiter(limiter.Config{
		Storage:       store,
		IPLimit:       3,
		TokenLimit:    10,
		BlockDuration: 5 * time.Second,
		TokenLimits:   make(map[string]int),
	})

	middleware := NewRateLimiterMiddleware(rl)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "you have reached the maximum number of requests")
}

func TestRateLimiterMiddleware_TokenRateLimit(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	rl := limiter.NewRateLimiter(limiter.Config{
		Storage:       store,
		IPLimit:       3,
		TokenLimit:    5,
		BlockDuration: 5 * time.Second,
		TokenLimits:   make(map[string]int),
	})

	middleware := NewRateLimiterMiddleware(rl)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// First 5 requests with token should succeed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("API_KEY", "test-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("API_KEY", "test-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimiterMiddleware_DifferentIPs(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	rl := limiter.NewRateLimiter(limiter.Config{
		Storage:       store,
		IPLimit:       2,
		TokenLimit:    10,
		BlockDuration: 5 * time.Second,
		TokenLimits:   make(map[string]int),
	})

	middleware := NewRateLimiterMiddleware(rl)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	ips := []string{"192.168.1.1:12345", "192.168.1.2:12345", "192.168.1.3:12345"}

	// Each IP should be able to make 2 requests
	for _, ip := range ips {
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code, "Request %d for IP %s should succeed", i+1, ip)
		}

		// 3rd request should be blocked
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code, "3rd request for IP %s should be blocked", ip)
	}
}

func TestRateLimiterMiddleware_XForwardedFor(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	rl := limiter.NewRateLimiter(limiter.Config{
		Storage:       store,
		IPLimit:       2,
		TokenLimit:    10,
		BlockDuration: 5 * time.Second,
		TokenLimits:   make(map[string]int),
	})

	middleware := NewRateLimiterMiddleware(rl)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test with X-Forwarded-For header
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.1:12345" // Different from X-Forwarded-For
		req.Header.Set("X-Forwarded-For", "192.168.1.100")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 3rd request should be blocked (same X-Forwarded-For IP)
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
