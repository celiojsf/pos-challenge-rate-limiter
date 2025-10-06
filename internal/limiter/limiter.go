package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/celiojsf/pos-challenge-rate-limiter/internal/storage"
)

type Config struct {
	Storage       storage.Storage
	IPLimit       int
	BlockDuration time.Duration
	TokenLimits   map[string]int
}

type RateLimiter struct {
	config Config
}

func NewRateLimiter(cfg Config) *RateLimiter {
	return &RateLimiter{
		config: cfg,
	}
}

// Allow checks if a request should be allowed based on IP or token
func (rl *RateLimiter) Allow(ctx context.Context, ip string, token string) (bool, error) {
	// Token takes precedence over IP
	if token != "" {
		return rl.checkToken(ctx, token)
	}

	return rl.checkIP(ctx, ip)
}

// IsTokenRegistered checks if a token is registered in the configuration
func (rl *RateLimiter) IsTokenRegistered(token string) bool {
	_, exists := rl.config.TokenLimits[token]
	return exists
}

func (rl *RateLimiter) checkIP(ctx context.Context, ip string) (bool, error) {
	key := fmt.Sprintf("ratelimit:ip:%s", ip)

	// Check if IP is blocked
	blocked, err := rl.config.Storage.IsBlocked(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check if IP is blocked: %w", err)
	}
	if blocked {
		return false, nil
	}

	// Increment counter
	count, err := rl.config.Storage.Increment(ctx, key, 1*time.Second)
	if err != nil {
		return false, fmt.Errorf("failed to increment IP counter: %w", err)
	}

	// Check if limit exceeded
	if count > int64(rl.config.IPLimit) {
		// Block the IP
		if err := rl.config.Storage.SetBlock(ctx, key, rl.config.BlockDuration); err != nil {
			return false, fmt.Errorf("failed to block IP: %w", err)
		}
		return false, nil
	}

	return true, nil
}

func (rl *RateLimiter) checkToken(ctx context.Context, token string) (bool, error) {
	// Check if token exists in the configured tokens
	limit, exists := rl.config.TokenLimits[token]
	if !exists {
		// Token not registered, deny access
		return false, nil
	}

	key := fmt.Sprintf("ratelimit:token:%s", token)

	// Check if token is blocked
	blocked, err := rl.config.Storage.IsBlocked(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check if token is blocked: %w", err)
	}
	if blocked {
		return false, nil
	}

	// Increment counter
	count, err := rl.config.Storage.Increment(ctx, key, 1*time.Second)
	if err != nil {
		return false, fmt.Errorf("failed to increment token counter: %w", err)
	}

	// Check if limit exceeded
	if count > int64(limit) {
		// Block the token
		if err := rl.config.Storage.SetBlock(ctx, key, rl.config.BlockDuration); err != nil {
			return false, fmt.Errorf("failed to block token: %w", err)
		}
		return false, nil
	}

	return true, nil
}
