package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func NewCache(redisURL string) (*Cache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return &Cache{client: redis.NewClient(opts)}, nil
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return b, err
}

func (c *Cache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, val, ttl).Err()
}

// InvalidateGame deletes all cache keys matching "{gameID}:*".
func (c *Cache) InvalidateGame(ctx context.Context, gameID string) error {
	pattern := gameID + ":*"
	var cursor uint64
	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("del: %w", err)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}

// CacheKey builds a canonical cache key from parts joined by ":".
func CacheKey(gameID, resource string, params ...string) string {
	parts := append([]string{gameID, resource}, params...)
	return strings.Join(parts, ":")
}
