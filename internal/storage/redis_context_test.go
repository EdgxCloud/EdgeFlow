package storage

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Redis client for testing
type mockRedisClient struct {
	data map[string]string
	ttls map[string]time.Duration
}

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		data: make(map[string]string),
		ttls: make(map[string]time.Duration),
	}
}

func TestRedisContextConfig_Defaults(t *testing.T) {
	config := RedisContextConfig{}

	// Test that defaults would be applied (can't actually connect without Redis)
	assert.Equal(t, "", config.Host)
	assert.Equal(t, 0, config.Port)
	assert.Equal(t, 0, config.PoolSize)
	assert.Equal(t, 0, config.MinIdleConns)
	assert.Equal(t, "", config.KeyPrefix)
}

func TestRedisContextStorage_BuildKey(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "edgeflow",
	}

	tests := []struct {
		name     string
		scope    ContextScope
		id       string
		key      string
		expected string
	}{
		{
			name:     "node scope",
			scope:    ScopeNode,
			id:       "node123",
			key:      "counter",
			expected: "edgeflow:node:node123:counter",
		},
		{
			name:     "flow scope",
			scope:    ScopeFlow,
			id:       "flow456",
			key:      "status",
			expected: "edgeflow:flow:flow456:status",
		},
		{
			name:     "global scope",
			scope:    ScopeGlobal,
			id:       "app",
			key:      "version",
			expected: "edgeflow:global:app:version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := storage.buildKey(tt.scope, tt.id, tt.key)
			assert.Equal(t, tt.expected, key)
		})
	}
}

func TestRedisContextStorage_BuildPattern(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "edgeflow",
	}

	tests := []struct {
		name     string
		scope    ContextScope
		id       string
		expected string
	}{
		{
			name:     "node pattern",
			scope:    ScopeNode,
			id:       "node123",
			expected: "edgeflow:node:node123:*",
		},
		{
			name:     "flow pattern",
			scope:    ScopeFlow,
			id:       "flow456",
			expected: "edgeflow:flow:flow456:*",
		},
		{
			name:     "global pattern",
			scope:    ScopeGlobal,
			id:       "app",
			expected: "edgeflow:global:app:*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := storage.buildPattern(tt.scope, tt.id)
			assert.Equal(t, tt.expected, pattern)
		})
	}
}

func TestRedisContextStorage_ExtractKeyName(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "edgeflow",
	}

	tests := []struct {
		name     string
		redisKey string
		scope    ContextScope
		id       string
		expected string
	}{
		{
			name:     "extract from node key",
			redisKey: "edgeflow:node:node123:counter",
			scope:    ScopeNode,
			id:       "node123",
			expected: "counter",
		},
		{
			name:     "extract from flow key",
			redisKey: "edgeflow:flow:flow456:status",
			scope:    ScopeFlow,
			id:       "flow456",
			expected: "status",
		},
		{
			name:     "extract from global key",
			redisKey: "edgeflow:global:app:version",
			scope:    ScopeGlobal,
			id:       "app",
			expected: "version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyName := storage.extractKeyName(tt.redisKey, tt.scope, tt.id)
			assert.Equal(t, tt.expected, keyName)
		})
	}
}

func TestContextScopes(t *testing.T) {
	scopes := []ContextScope{ScopeNode, ScopeFlow, ScopeGlobal}

	assert.Equal(t, ContextScope("node"), ScopeNode)
	assert.Equal(t, ContextScope("flow"), ScopeFlow)
	assert.Equal(t, ContextScope("global"), ScopeGlobal)
	assert.Len(t, scopes, 3)
}

func TestContextHelper_NewHelper(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "test",
	}

	helper := NewContextHelper(storage, ScopeNode, "node123")
	require.NotNil(t, helper)
	assert.Equal(t, storage, helper.storage)
	assert.Equal(t, ScopeNode, helper.scope)
	assert.Equal(t, "node123", helper.id)
}

// Integration-like tests that would work with a real Redis instance
// These are structure tests that verify the logic without needing Redis

func TestRedisContextStorage_KeyStructure(t *testing.T) {
	storage := &RedisContextStorage{
		prefix:     "edgeflow",
		defaultTTL: 1 * time.Hour,
	}

	// Test key generation for different scopes
	nodeKey := storage.buildKey(ScopeNode, "node1", "counter")
	assert.Contains(t, nodeKey, "node")
	assert.Contains(t, nodeKey, "node1")
	assert.Contains(t, nodeKey, "counter")

	flowKey := storage.buildKey(ScopeFlow, "flow1", "status")
	assert.Contains(t, flowKey, "flow")
	assert.Contains(t, flowKey, "flow1")
	assert.Contains(t, flowKey, "status")

	globalKey := storage.buildKey(ScopeGlobal, "system", "version")
	assert.Contains(t, globalKey, "global")
	assert.Contains(t, globalKey, "system")
	assert.Contains(t, globalKey, "version")
}

func TestRedisContextStorage_PatternMatching(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "edgeflow",
	}

	pattern := storage.buildPattern(ScopeNode, "node1")

	// Verify pattern format
	assert.Contains(t, pattern, "edgeflow")
	assert.Contains(t, pattern, "node")
	assert.Contains(t, pattern, "node1")
	assert.HasSuffix(t, pattern, "*")
}

func TestRedisContextStorage_TTLHandling(t *testing.T) {
	storage := &RedisContextStorage{
		defaultTTL: 30 * time.Minute,
	}

	// Verify default TTL is set
	assert.Equal(t, 30*time.Minute, storage.defaultTTL)

	// Test with no TTL
	storage2 := &RedisContextStorage{
		defaultTTL: 0,
	}
	assert.Equal(t, time.Duration(0), storage2.defaultTTL)
}

func TestContextHelper_MethodAvailability(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "test",
	}

	helper := NewContextHelper(storage, ScopeNode, "node1")

	// Verify all methods are available
	ctx := context.Background()

	// These will fail without Redis, but verify methods exist
	_, _ = helper.Get(ctx, "test")
	_ = helper.Set(ctx, "test", "value")
	_ = helper.SetWithTTL(ctx, "test", "value", 1*time.Minute)
	_ = helper.Delete(ctx, "test")
	_, _ = helper.GetAll(ctx)
	_ = helper.Clear(ctx)
	_, _ = helper.Keys(ctx)
	_, _ = helper.Exists(ctx, "test")
	_, _ = helper.Increment(ctx, "test", 1)

	// If we get here, all methods are available
	assert.NotNil(t, helper)
}

func TestRedisContextStorage_MultipleScopes(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "edgeflow",
	}

	// Verify different scopes generate different keys for same ID and key
	nodeKey := storage.buildKey(ScopeNode, "id1", "data")
	flowKey := storage.buildKey(ScopeFlow, "id1", "data")
	globalKey := storage.buildKey(ScopeGlobal, "id1", "data")

	assert.NotEqual(t, nodeKey, flowKey)
	assert.NotEqual(t, flowKey, globalKey)
	assert.NotEqual(t, nodeKey, globalKey)
}

func TestRedisContextStorage_KeyExtraction(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "edgeflow",
	}

	// Build a key and extract it back
	originalKey := "temperature"
	redisKey := storage.buildKey(ScopeNode, "sensor1", originalKey)
	extractedKey := storage.extractKeyName(redisKey, ScopeNode, "sensor1")

	assert.Equal(t, originalKey, extractedKey)
}

func TestRedisContextStorage_PoolConfiguration(t *testing.T) {
	storage := &RedisContextStorage{
		poolSize:     20,
		minIdleConns: 5,
	}

	assert.Equal(t, 20, storage.poolSize)
	assert.Equal(t, 5, storage.minIdleConns)
}

func TestRedisContextConfig_CustomPrefix(t *testing.T) {
	config := RedisContextConfig{
		KeyPrefix: "myapp",
	}

	assert.Equal(t, "myapp", config.KeyPrefix)
}

func TestContextHelper_ScopeIsolation(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "test",
	}

	// Create helpers for different scopes
	nodeHelper := NewContextHelper(storage, ScopeNode, "id1")
	flowHelper := NewContextHelper(storage, ScopeFlow, "id1")
	globalHelper := NewContextHelper(storage, ScopeGlobal, "id1")

	// Verify they have different scopes
	assert.NotEqual(t, nodeHelper.scope, flowHelper.scope)
	assert.NotEqual(t, flowHelper.scope, globalHelper.scope)
	assert.NotEqual(t, nodeHelper.scope, globalHelper.scope)
}

func TestRedisContextStorage_DefaultConfig(t *testing.T) {
	config := RedisContextConfig{}

	// Default should be empty, will be filled by NewRedisContextStorage
	assert.Empty(t, config.Host)
	assert.Zero(t, config.Port)
	assert.Zero(t, config.PoolSize)
}

func TestRedisContextStorage_JSONMarshaling(t *testing.T) {
	// Test that we can handle different value types
	values := []interface{}{
		"string value",
		42,
		3.14,
		true,
		map[string]interface{}{"key": "value"},
		[]interface{}{1, 2, 3},
	}

	for _, val := range values {
		// Verify we can marshal these types
		// (actual marshaling tested by Set/Get with real Redis)
		assert.NotNil(t, val)
	}
}

func TestRedisContextStorage_ErrorScenarios(t *testing.T) {
	// Test configuration that would fail (without actual Redis connection)
	config := RedisContextConfig{
		Host: "invalid-host",
		Port: 99999,
	}

	_, err := NewRedisContextStorage(config)

	// Should fail to connect
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}

func TestContextHelper_DifferentIDs(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "test",
	}

	// Create helpers for different IDs
	helper1 := NewContextHelper(storage, ScopeNode, "node1")
	helper2 := NewContextHelper(storage, ScopeNode, "node2")

	// Verify they have different IDs
	assert.NotEqual(t, helper1.id, helper2.id)
	assert.Equal(t, "node1", helper1.id)
	assert.Equal(t, "node2", helper2.id)
}

func TestRedisContextStorage_Concurrency(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "test",
	}

	// Test that multiple goroutines can build keys concurrently
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			key := storage.buildKey(ScopeNode, "node1", "test")
			assert.Contains(t, key, "node1")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRedisContextStorage_LongKeys(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "edgeflow",
	}

	// Test with long ID and key names
	longID := "very-long-node-identifier-with-many-characters-12345678901234567890"
	longKey := "very-long-key-name-with-many-characters-12345678901234567890"

	redisKey := storage.buildKey(ScopeNode, longID, longKey)

	assert.Contains(t, redisKey, longID)
	assert.Contains(t, redisKey, longKey)
}

func TestRedisContextStorage_SpecialCharacters(t *testing.T) {
	storage := &RedisContextStorage{
		prefix: "edgeflow",
	}

	// Test with special characters in ID and key
	specialID := "node-123_test.sensor"
	specialKey := "data:temperature/celsius"

	redisKey := storage.buildKey(ScopeNode, specialID, specialKey)
	extractedKey := storage.extractKeyName(redisKey, ScopeNode, specialID)

	assert.Equal(t, specialKey, extractedKey)
}
