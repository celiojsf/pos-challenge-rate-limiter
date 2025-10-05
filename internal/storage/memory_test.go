package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_Increment(t *testing.T) {
	store := NewMemoryStorage()
	defer store.Close()

	ctx := context.Background()
	key := "test-key"

	// First increment should return 1
	count, err := store.Increment(ctx, key, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Second increment should return 2
	count, err = store.Increment(ctx, key, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestMemoryStorage_Get(t *testing.T) {
	store := NewMemoryStorage()
	defer store.Close()

	ctx := context.Background()
	key := "test-key"

	// Get non-existent key should return 0
	count, err := store.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Increment and get
	_, err = store.Increment(ctx, key, 5*time.Second)
	require.NoError(t, err)

	count, err = store.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestMemoryStorage_Block(t *testing.T) {
	store := NewMemoryStorage()
	defer store.Close()

	ctx := context.Background()
	key := "test-key"

	// Initially not blocked
	blocked, err := store.IsBlocked(ctx, key)
	require.NoError(t, err)
	assert.False(t, blocked)

	// Set block
	err = store.SetBlock(ctx, key, 2*time.Second)
	require.NoError(t, err)

	// Should be blocked now
	blocked, err = store.IsBlocked(ctx, key)
	require.NoError(t, err)
	assert.True(t, blocked)

	// Wait for expiration
	time.Sleep(3 * time.Second)

	// Should not be blocked anymore
	blocked, err = store.IsBlocked(ctx, key)
	require.NoError(t, err)
	assert.False(t, blocked)
}

func TestMemoryStorage_Expiration(t *testing.T) {
	store := NewMemoryStorage()
	defer store.Close()

	ctx := context.Background()
	key := "test-key"

	// Increment with short expiration
	count, err := store.Increment(ctx, key, 1*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Get should return 0 after expiration
	count, err = store.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// New increment should start from 1 again
	count, err = store.Increment(ctx, key, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
