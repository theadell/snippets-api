package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"snippets.adelh.dev/app/internal/config"
)

type RedisCache struct {
	client  *redis.Client
	ttl     time.Duration
	logger  *slog.Logger
	enabled bool
}

func NewRedisCache(cfg config.RedisConfig) *RedisCache {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("redis connection failed", "error", err)
	}

	return &RedisCache{
		client:  client,
		ttl:     cfg.TTL,
		logger:  logger,
		enabled: cfg.Enabled,
	}
}

// Get retrieves a value from the cache into dst.
// Returns true if on cache hit
// Returns false on cache miss or error
func (c *RedisCache) Get(ctx context.Context, key string, dst any) bool {
	if !c.enabled {
		return false
	}
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			c.logger.Warn("redis get failed", "key", key, "error", err)
		}
		return false
	}

	if err := json.Unmarshal([]byte(val), dst); err != nil {
		c.logger.Warn("failed to unmarshal cached value", "key", key, "error", err)
		c.client.Del(ctx, key)
		return false
	}
	return true
}

// Set stores a value in the cache.
// This operation is fire-and-forget. Errors are logged but not returned.
func (c *RedisCache) Set(ctx context.Context, key string, value any) {
	if !c.enabled {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Warn("failed to marshal value for cache", "key", key, "error", err)
		return
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		c.logger.Warn("failed to set cache key", "key", key, "error", err)
	}
}

// Delete removes a key from the cache.
func (c *RedisCache) Delete(ctx context.Context, key string) {
	if !c.enabled {
		return
	}
	if err := c.client.Del(ctx, key).Err(); err != nil {
		c.logger.Warn("failed to delete cache key", "key", key, "error", err)
	}
}

// Close closes the Redis client.
func (c *RedisCache) Close() error {
	return c.client.Close()
}
