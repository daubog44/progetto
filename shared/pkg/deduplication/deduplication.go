package deduplication

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Deduplicator defines the interface for checking and enforcing uniqueness of items.
type Deduplicator interface {
	// IsUnique checks if the given key is unique.
	// It returns true if the key has NOT been seen before (and effectively "claims" it).
	// It returns false if the key has already been seen (duplicate).
	// The key is stored with the given TTL.
	IsUnique(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

// RedisDeduplicator implements Deduplicator using Redis.
type RedisDeduplicator struct {
	client *redis.Client
	prefix string
}

// NewRedisDeduplicator creates a new RedisDeduplicator.
func NewRedisDeduplicator(client *redis.Client, prefix string) *RedisDeduplicator {
	return &RedisDeduplicator{
		client: client,
		prefix: prefix,
	}
}

// IsUnique implements Deduplicator.
// It uses the SET NX command to atomically set the key if it doesn't exist.
func (d *RedisDeduplicator) IsUnique(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	fullKey := fmt.Sprintf("%s:%s", d.prefix, key)

	// SET key value NX EX ttl
	// NX: Only set the key if it does not already exist.
	// EX: Set the specified expire time, in seconds.
	success, err := d.client.SetNX(ctx, fullKey, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check uniqueness in redis: %w", err)
	}

	return success, nil
}
