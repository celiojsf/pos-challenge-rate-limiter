package storage

import (
	"context"
	"time"
)

// Storage is the interface for rate limiter storage
type Storage interface {
	// Increment increments the counter for the given key and returns the new value
	Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)

	// Get returns the current value for the given key
	Get(ctx context.Context, key string) (int64, error)

	// SetBlock sets a block for the given key with expiration
	SetBlock(ctx context.Context, key string, expiration time.Duration) error

	// IsBlocked checks if the given key is blocked
	IsBlocked(ctx context.Context, key string) (bool, error)

	// Close closes the storage connection
	Close() error
}
