package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Redis represents Redis connection
type Redis struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedis creates a new Redis connection
func NewRedis(addr, password string, db int, logger *zap.Logger) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis", 
		zap.String("addr", addr),
		zap.Int("db", db))

	return &Redis{
		client: rdb,
		logger: logger,
	}, nil
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	if err := r.client.Close(); err != nil {
		r.logger.Error("Failed to close Redis connection", zap.Error(err))
		return err
	}

	r.logger.Info("Redis connection closed")
	return nil
}

// Set stores a value in Redis with expiration
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set value in Redis: %w", err)
	}

	return nil
}

// Get retrieves a value from Redis
func (r *Redis) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get value from Redis: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a key from Redis
func (r *Redis) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete key from Redis: %w", err)
	}

	return nil
}

// Exists checks if a key exists in Redis
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return count > 0, nil
}

// SetExpiration sets expiration for a key
func (r *Redis) SetExpiration(ctx context.Context, key string, expiration time.Duration) error {
	if err := r.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	return nil
}

// Increment increments a counter
func (r *Redis) Increment(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment counter: %w", err)
	}

	return val, nil
}

// IncrementWithExpiration increments a counter and sets expiration
func (r *Redis) IncrementWithExpiration(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	pipe := r.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("failed to increment with expiration: %w", err)
	}

	return incrCmd.Val(), nil
}

// GetClient returns the underlying Redis client
func (r *Redis) GetClient() *redis.Client {
	return r.client
}

// Cache interface defines caching operations
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetExpiration(ctx context.Context, key string, expiration time.Duration) error
	Increment(ctx context.Context, key string) (int64, error)
	IncrementWithExpiration(ctx context.Context, key string, expiration time.Duration) (int64, error)
}
