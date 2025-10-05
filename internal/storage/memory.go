package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type entry struct {
	count      int64
	expiration time.Time
}

type MemoryStorage struct {
	mu      sync.RWMutex
	data    map[string]*entry
	blocked map[string]time.Time
}

func NewMemoryStorage() *MemoryStorage {
	m := &MemoryStorage{
		data:    make(map[string]*entry),
		blocked: make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go m.cleanup()

	return m
}

func (m *MemoryStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	if e, exists := m.data[key]; exists {
		if now.Before(e.expiration) {
			e.count++
			return e.count, nil
		}
	}

	m.data[key] = &entry{
		count:      1,
		expiration: now.Add(expiration),
	}

	return 1, nil
}

func (m *MemoryStorage) Get(ctx context.Context, key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if e, exists := m.data[key]; exists {
		if time.Now().Before(e.expiration) {
			return e.count, nil
		}
	}

	return 0, nil
}

func (m *MemoryStorage) SetBlock(ctx context.Context, key string, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	blockKey := fmt.Sprintf("block:%s", key)
	m.blocked[blockKey] = time.Now().Add(expiration)

	return nil
}

func (m *MemoryStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	blockKey := fmt.Sprintf("block:%s", key)
	if expiration, exists := m.blocked[blockKey]; exists {
		if time.Now().Before(expiration) {
			return true, nil
		}
	}

	return false, nil
}

func (m *MemoryStorage) Close() error {
	return nil
}

func (m *MemoryStorage) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()

		// Clean up expired entries
		for key, e := range m.data {
			if now.After(e.expiration) {
				delete(m.data, key)
			}
		}

		// Clean up expired blocks
		for key, expiration := range m.blocked {
			if now.After(expiration) {
				delete(m.blocked, key)
			}
		}

		m.mu.Unlock()
	}
}
