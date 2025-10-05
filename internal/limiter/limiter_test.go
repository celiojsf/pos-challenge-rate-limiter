package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/celiojsf/pos-challenge-rate-limiter/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_IPLimit(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := NewRateLimiter(Config{
		Storage:       store,
		IPLimit:       5,
		TokenLimit:    10,
		BlockDuration: 5 * time.Second,
		TokenLimits:   make(map[string]int),
	})

	ctx := context.Background()
	ip := "192.168.1.1"

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, ip, "")
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 6th request should be blocked
	allowed, err := limiter.Allow(ctx, ip, "")
	require.NoError(t, err)
	assert.False(t, allowed, "6th request should be blocked")

	// Further requests should remain blocked
	allowed, err = limiter.Allow(ctx, ip, "")
	require.NoError(t, err)
	assert.False(t, allowed, "Request should remain blocked")
}

func TestRateLimiter_TokenLimit(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := NewRateLimiter(Config{
		Storage:       store,
		IPLimit:       5,
		TokenLimit:    10,
		BlockDuration: 5 * time.Second,
		TokenLimits:   make(map[string]int),
	})

	ctx := context.Background()
	ip := "192.168.1.1"
	token := "test-token"

	// First 10 requests with token should be allowed
	for i := 0; i < 10; i++ {
		allowed, err := limiter.Allow(ctx, ip, token)
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 11th request should be blocked
	allowed, err := limiter.Allow(ctx, ip, token)
	require.NoError(t, err)
	assert.False(t, allowed, "11th request should be blocked")
}

func TestRateLimiter_CustomTokenLimit(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	customToken := "custom-token"
	limiter := NewRateLimiter(Config{
		Storage:       store,
		IPLimit:       5,
		TokenLimit:    10,
		BlockDuration: 5 * time.Second,
		TokenLimits: map[string]int{
			customToken: 3,
		},
	})

	ctx := context.Background()
	ip := "192.168.1.1"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, ip, customToken)
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 4th request should be blocked
	allowed, err := limiter.Allow(ctx, ip, customToken)
	require.NoError(t, err)
	assert.False(t, allowed, "4th request should be blocked")
}

func TestRateLimiter_TokenOverridesIP(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := NewRateLimiter(Config{
		Storage:       store,
		IPLimit:       3,
		TokenLimit:    10,
		BlockDuration: 5 * time.Second,
		TokenLimits:   make(map[string]int),
	})

	ctx := context.Background()
	ip := "192.168.1.1"
	token := "test-token"

	// Make 5 requests with token (should be allowed even though IP limit is 3)
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, ip, token)
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d with token should be allowed", i+1)
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := NewRateLimiter(Config{
		Storage:       store,
		IPLimit:       3,
		TokenLimit:    10,
		BlockDuration: 5 * time.Second,
		TokenLimits:   make(map[string]int),
	})

	ctx := context.Background()

	// Each IP should have its own limit
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}

	for _, ip := range ips {
		for i := 0; i < 3; i++ {
			allowed, err := limiter.Allow(ctx, ip, "")
			require.NoError(t, err)
			assert.True(t, allowed, "Request %d for IP %s should be allowed", i+1, ip)
		}

		// 4th request should be blocked for this IP
		allowed, err := limiter.Allow(ctx, ip, "")
		require.NoError(t, err)
		assert.False(t, allowed, "4th request for IP %s should be blocked", ip)
	}
}

func TestRateLimiter_ResetAfterExpiration(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := NewRateLimiter(Config{
		Storage:       store,
		IPLimit:       2,
		TokenLimit:    10,
		BlockDuration: 1 * time.Second,
		TokenLimits:   make(map[string]int),
	})

	ctx := context.Background()
	ip := "192.168.1.1"

	// Use up the limit
	for i := 0; i < 2; i++ {
		allowed, err := limiter.Allow(ctx, ip, "")
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// Next request should be blocked
	allowed, err := limiter.Allow(ctx, ip, "")
	require.NoError(t, err)
	assert.False(t, allowed)

	// Wait for counter to reset (1 second) plus block duration
	time.Sleep(2 * time.Second)

	// Should be allowed again
	allowed, err = limiter.Allow(ctx, ip, "")
	require.NoError(t, err)
	assert.True(t, allowed, "Request should be allowed after expiration")
}
