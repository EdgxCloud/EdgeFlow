package database

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/edgeflow/edgeflow/internal/node"
)

// RedisConfig configuration for the Redis node
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// RedisExecutor executor for the Redis node
type RedisExecutor struct {
	config RedisConfig
	client *redis.Client
	mu     sync.RWMutex
}

// NewRedisExecutor creates a new RedisExecutor
func NewRedisExecutor() node.Executor {
	return &RedisExecutor{}
}

// Init initializes the Redis node with configuration
func (e *RedisExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var redisConfig RedisConfig
	if err := json.Unmarshal(configJSON, &redisConfig); err != nil {
		return fmt.Errorf("invalid redis config: %w", err)
	}

	// Default values
	if redisConfig.Host == "" {
		redisConfig.Host = "localhost"
	}
	if redisConfig.Port == 0 {
		redisConfig.Port = 6379
	}

	e.config = redisConfig
	return nil
}

// Execute executes the node
func (e *RedisExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Connect if not connected
	if e.client == nil {
		if err := e.connect(); err != nil {
			return node.Message{}, fmt.Errorf("failed to connect: %w", err)
		}
	}

	var command string
	var key string
	var value interface{}
	var ttl int

	// Get command from message
	if cmd, ok := msg.Payload["command"].(string); ok {
		command = cmd
	}
	if k, ok := msg.Payload["key"].(string); ok {
		key = k
	}
	if v, ok := msg.Payload["value"]; ok {
		value = v
	}
	if t, ok := msg.Payload["ttl"].(float64); ok {
		ttl = int(t)
	}

	if command == "" {
		return node.Message{}, fmt.Errorf("command is required")
	}
	if key == "" {
		return node.Message{}, fmt.Errorf("key is required")
	}

	switch command {
	case "get":
		val, err := e.client.Get(ctx, key).Result()
		if err == redis.Nil {
			return node.Message{
				Payload: map[string]interface{}{
					"key":    key,
					"exists": false,
				},
			}, nil
		} else if err != nil {
			return node.Message{}, err
		}

		return node.Message{
			Payload: map[string]interface{}{
				"key":    key,
				"value":  val,
				"exists": true,
			},
		}, nil

	case "set":
		if value == nil {
			return node.Message{}, fmt.Errorf("value is required")
		}

		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = v
		default:
			jsonData, _ := json.Marshal(v)
			valueStr = string(jsonData)
		}

		var expiration time.Duration
		if ttl > 0 {
			expiration = time.Duration(ttl) * time.Second
		}

		err := e.client.Set(ctx, key, valueStr, expiration).Err()
		if err != nil {
			return node.Message{}, err
		}

		return node.Message{
			Payload: map[string]interface{}{
				"key":   key,
				"value": valueStr,
				"set":   true,
			},
		}, nil

	case "delete", "del":
		err := e.client.Del(ctx, key).Err()
		if err != nil {
			return node.Message{}, err
		}

		return node.Message{
			Payload: map[string]interface{}{
				"key":     key,
				"deleted": true,
			},
		}, nil

	case "exists":
		count, err := e.client.Exists(ctx, key).Result()
		if err != nil {
			return node.Message{}, err
		}

		return node.Message{
			Payload: map[string]interface{}{
				"key":    key,
				"exists": count > 0,
			},
		}, nil

	case "incr":
		val, err := e.client.Incr(ctx, key).Result()
		if err != nil {
			return node.Message{}, err
		}

		return node.Message{
			Payload: map[string]interface{}{
				"key":   key,
				"value": val,
			},
		}, nil

	case "decr":
		val, err := e.client.Decr(ctx, key).Result()
		if err != nil {
			return node.Message{}, err
		}

		return node.Message{
			Payload: map[string]interface{}{
				"key":   key,
				"value": val,
			},
		}, nil

	default:
		return node.Message{}, fmt.Errorf("unknown command: %s", command)
	}
}

// connect connects to Redis
func (e *RedisExecutor) connect() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.client != nil {
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", e.config.Host, e.config.Port),
		Password: e.config.Password,
		DB:       e.config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return err
	}

	e.client = client
	return nil
}

// Cleanup releases resources
func (e *RedisExecutor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.client != nil {
		e.client.Close()
		e.client = nil
	}
	return nil
}
