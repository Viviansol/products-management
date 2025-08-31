package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService handles Redis caching operations
type CacheService struct {
	Client *redis.Client
}

// NewCacheService creates a new cache service
func NewCacheService(client *redis.Client) *CacheService {
	return &CacheService{
		Client: client,
	}
}

// Set stores a key-value pair in Redis with expiration
func (s *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return s.Client.Set(ctx, key, jsonValue, expiration).Err()
}

// Get retrieves a value from Redis by key
func (s *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	value, err := s.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("failed to get value: %w", err)
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	return json.Unmarshal([]byte(value), dest)
}

// Delete removes a key from Redis
func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.Client.Del(ctx, key).Err()
}

// DeletePattern removes keys matching a pattern
func (s *CacheService) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := s.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}
	
	if len(keys) > 0 {
		return s.Client.Del(ctx, keys...).Err()
	}

	return nil
}

// Exists checks if a key exists in Redis
func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := s.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return result > 0, nil
}

// SetNX sets a key only if it doesn't exist (for distributed locks)
func (s *CacheService) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return s.Client.SetNX(ctx, key, jsonValue, expiration).Result()
}

// Incr increments a counter in Redis
func (s *CacheService) Incr(ctx context.Context, key string) (int64, error) {
	return s.Client.Incr(ctx, key).Result()
}

// Expire sets expiration for a key
func (s *CacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return s.Client.Expire(ctx, key, expiration).Err()
}
