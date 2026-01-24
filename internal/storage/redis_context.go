package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// ContextScope defines the scope of context storage
type ContextScope string

const (
	// ScopeNode is node-level context
	ScopeNode ContextScope = "node"

	// ScopeFlow is flow-level context
	ScopeFlow ContextScope = "flow"

	// ScopeGlobal is global context
	ScopeGlobal ContextScope = "global"
)

// RedisContextStorage implements context storage using Redis
type RedisContextStorage struct {
	client *redis.Client
	mu     sync.RWMutex

	// Key prefix for namespacing
	prefix string

	// Default TTL for keys (0 = no expiry)
	defaultTTL time.Duration

	// Connection pool stats
	poolSize     int
	minIdleConns int
}

// RedisContextConfig holds Redis configuration
type RedisContextConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DefaultTTL   time.Duration
	KeyPrefix    string
}

// NewRedisContextStorage creates a new Redis context storage
func NewRedisContextStorage(config RedisContextConfig) (*RedisContextStorage, error) {
	// Set defaults
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 6379
	}
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.MinIdleConns == 0 {
		config.MinIdleConns = 2
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "edgeflow"
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	storage := &RedisContextStorage{
		client:       client,
		prefix:       config.KeyPrefix,
		defaultTTL:   config.DefaultTTL,
		poolSize:     config.PoolSize,
		minIdleConns: config.MinIdleConns,
	}

	return storage, nil
}

// Get retrieves a value from context storage
func (r *RedisContextStorage) Get(ctx context.Context, scope ContextScope, id, key string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	redisKey := r.buildKey(scope, id, key)

	val, err := r.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return nil, nil // Key not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key %s: %w", redisKey, err)
	}

	// Try to unmarshal as JSON
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		// Return as string if not JSON
		return val, nil
	}

	return result, nil
}

// Set stores a value in context storage
func (r *RedisContextStorage) Set(ctx context.Context, scope ContextScope, id, key string, value interface{}) error {
	return r.SetWithTTL(ctx, scope, id, key, value, r.defaultTTL)
}

// SetWithTTL stores a value with a custom TTL
func (r *RedisContextStorage) SetWithTTL(ctx context.Context, scope ContextScope, id, key string, value interface{}, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	redisKey := r.buildKey(scope, id, key)

	// Marshal value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Set with TTL
	if ttl > 0 {
		err = r.client.Set(ctx, redisKey, data, ttl).Err()
	} else {
		err = r.client.Set(ctx, redisKey, data, 0).Err()
	}

	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", redisKey, err)
	}

	return nil
}

// Delete removes a value from context storage
func (r *RedisContextStorage) Delete(ctx context.Context, scope ContextScope, id, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	redisKey := r.buildKey(scope, id, key)

	err := r.client.Del(ctx, redisKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", redisKey, err)
	}

	return nil
}

// GetAll retrieves all keys for a given scope and ID
func (r *RedisContextStorage) GetAll(ctx context.Context, scope ContextScope, id string) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pattern := r.buildPattern(scope, id)

	// Scan for matching keys
	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error
		batch, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}

		keys = append(keys, batch...)

		if cursor == 0 {
			break
		}
	}

	// Retrieve all values
	result := make(map[string]interface{})
	for _, redisKey := range keys {
		val, err := r.client.Get(ctx, redisKey).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get key %s: %w", redisKey, err)
		}

		// Extract the key name (remove prefix and scope)
		keyName := r.extractKeyName(redisKey, scope, id)

		// Try to unmarshal as JSON
		var value interface{}
		if err := json.Unmarshal([]byte(val), &value); err != nil {
			// Store as string if not JSON
			value = val
		}

		result[keyName] = value
	}

	return result, nil
}

// Clear removes all keys for a given scope and ID
func (r *RedisContextStorage) Clear(ctx context.Context, scope ContextScope, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	pattern := r.buildPattern(scope, id)

	// Scan and delete matching keys
	var cursor uint64

	for {
		var keys []string
		var err error
		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys: %w", err)
		}

		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("failed to delete keys: %w", err)
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

// Keys returns all keys for a given scope and ID
func (r *RedisContextStorage) Keys(ctx context.Context, scope ContextScope, id string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pattern := r.buildPattern(scope, id)

	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error
		batch, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}

		// Extract key names
		for _, redisKey := range batch {
			keyName := r.extractKeyName(redisKey, scope, id)
			keys = append(keys, keyName)
		}

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// Exists checks if a key exists
func (r *RedisContextStorage) Exists(ctx context.Context, scope ContextScope, id, key string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	redisKey := r.buildKey(scope, id, key)

	count, err := r.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return count > 0, nil
}

// Increment increments a numeric value
func (r *RedisContextStorage) Increment(ctx context.Context, scope ContextScope, id, key string, delta int64) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	redisKey := r.buildKey(scope, id, key)

	val, err := r.client.IncrBy(ctx, redisKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", redisKey, err)
	}

	return val, nil
}

// Expire sets a TTL on an existing key
func (r *RedisContextStorage) Expire(ctx context.Context, scope ContextScope, id, key string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	redisKey := r.buildKey(scope, id, key)

	err := r.client.Expire(ctx, redisKey, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiry on key %s: %w", redisKey, err)
	}

	return nil
}

// TTL returns the remaining TTL of a key
func (r *RedisContextStorage) TTL(ctx context.Context, scope ContextScope, id, key string) (time.Duration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	redisKey := r.buildKey(scope, id, key)

	ttl, err := r.client.TTL(ctx, redisKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", redisKey, err)
	}

	return ttl, nil
}

// Close closes the Redis connection
func (r *RedisContextStorage) Close() error {
	return r.client.Close()
}

// Ping tests the Redis connection
func (r *RedisContextStorage) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// PoolStats returns connection pool statistics
func (r *RedisContextStorage) PoolStats() *redis.PoolStats {
	return r.client.PoolStats()
}

// buildKey constructs a Redis key from scope, ID, and key
func (r *RedisContextStorage) buildKey(scope ContextScope, id, key string) string {
	return fmt.Sprintf("%s:%s:%s:%s", r.prefix, scope, id, key)
}

// buildPattern constructs a Redis pattern for scanning
func (r *RedisContextStorage) buildPattern(scope ContextScope, id string) string {
	return fmt.Sprintf("%s:%s:%s:*", r.prefix, scope, id)
}

// extractKeyName extracts the key name from a full Redis key
func (r *RedisContextStorage) extractKeyName(redisKey string, scope ContextScope, id string) string {
	prefix := fmt.Sprintf("%s:%s:%s:", r.prefix, scope, id)
	if len(redisKey) > len(prefix) {
		return redisKey[len(prefix):]
	}
	return redisKey
}

// ContextHelper provides convenience methods for working with context storage
type ContextHelper struct {
	storage *RedisContextStorage
	scope   ContextScope
	id      string
}

// NewContextHelper creates a new context helper
func NewContextHelper(storage *RedisContextStorage, scope ContextScope, id string) *ContextHelper {
	return &ContextHelper{
		storage: storage,
		scope:   scope,
		id:      id,
	}
}

// Get retrieves a value
func (h *ContextHelper) Get(ctx context.Context, key string) (interface{}, error) {
	return h.storage.Get(ctx, h.scope, h.id, key)
}

// Set stores a value
func (h *ContextHelper) Set(ctx context.Context, key string, value interface{}) error {
	return h.storage.Set(ctx, h.scope, h.id, key, value)
}

// SetWithTTL stores a value with TTL
func (h *ContextHelper) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return h.storage.SetWithTTL(ctx, h.scope, h.id, key, value, ttl)
}

// Delete removes a value
func (h *ContextHelper) Delete(ctx context.Context, key string) error {
	return h.storage.Delete(ctx, h.scope, h.id, key)
}

// GetAll retrieves all values
func (h *ContextHelper) GetAll(ctx context.Context) (map[string]interface{}, error) {
	return h.storage.GetAll(ctx, h.scope, h.id)
}

// Clear removes all values
func (h *ContextHelper) Clear(ctx context.Context) error {
	return h.storage.Clear(ctx, h.scope, h.id)
}

// Keys returns all keys
func (h *ContextHelper) Keys(ctx context.Context) ([]string, error) {
	return h.storage.Keys(ctx, h.scope, h.id)
}

// Exists checks if a key exists
func (h *ContextHelper) Exists(ctx context.Context, key string) (bool, error) {
	return h.storage.Exists(ctx, h.scope, h.id, key)
}

// Increment increments a numeric value
func (h *ContextHelper) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return h.storage.Increment(ctx, h.scope, h.id, key, delta)
}
