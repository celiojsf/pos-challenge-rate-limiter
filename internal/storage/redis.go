package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{
		client: client,
	}, nil
}

func (r *RedisStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	pipe := r.client.Pipeline()

	incr := pipe.Incr(ctx, key)
	pipe.ExpireNX(ctx, key, expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}

	return incr.Val(), nil
}

func (r *RedisStorage) Get(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

func (r *RedisStorage) SetBlock(ctx context.Context, key string, expiration time.Duration) error {
	blockKey := fmt.Sprintf("block:%s", key)
	err := r.client.Set(ctx, blockKey, "1", expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set block for key %s: %w", key, err)
	}
	return nil
}

func (r *RedisStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	blockKey := fmt.Sprintf("block:%s", key)
	val, err := r.client.Get(ctx, blockKey).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check block status for key %s: %w", key, err)
	}
	return val == "1", nil
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}
