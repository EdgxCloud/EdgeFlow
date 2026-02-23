package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// RateLimitNode implements a token bucket rate limiter
type RateLimitNode struct {
	rate       int           // Messages per time window
	timeWindow time.Duration // Time window duration
	strategy   string        // drop, queue, delay
	maxQueue   int           // Maximum queue size when strategy is "queue"

	// Token bucket state
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex

	// Queue for queued messages
	queue     []node.Message
	queueChan chan node.Message
	stopChan  chan struct{}
	running   bool
}

// Init initializes the rate limit node
func (n *RateLimitNode) Init(config map[string]interface{}) error {
	// Rate (messages per window)
	if rate, ok := config["rate"].(float64); ok {
		n.rate = int(rate)
	} else if rate, ok := config["rate"].(int); ok {
		n.rate = rate
	} else {
		n.rate = 10 // Default 10 messages
	}

	// Time window
	if window, ok := config["timeWindow"].(float64); ok {
		n.timeWindow = time.Duration(window) * time.Second
	} else if window, ok := config["timeWindow"].(int); ok {
		n.timeWindow = time.Duration(window) * time.Second
	} else if window, ok := config["timeWindow"].(string); ok {
		d, err := time.ParseDuration(window)
		if err != nil {
			n.timeWindow = time.Second
		} else {
			n.timeWindow = d
		}
	} else {
		n.timeWindow = time.Second // Default 1 second
	}

	// Strategy: drop, queue, delay
	if strategy, ok := config["strategy"].(string); ok {
		n.strategy = strategy
	} else {
		n.strategy = "drop" // Default drop excess
	}

	// Max queue size
	if maxQueue, ok := config["maxQueue"].(float64); ok {
		n.maxQueue = int(maxQueue)
	} else if maxQueue, ok := config["maxQueue"].(int); ok {
		n.maxQueue = maxQueue
	} else {
		n.maxQueue = 100 // Default 100 messages
	}

	// Initialize token bucket
	n.tokens = float64(n.rate)
	n.lastUpdate = time.Now()

	// Initialize queue if using queue strategy
	if n.strategy == "queue" {
		n.queue = make([]node.Message, 0, n.maxQueue)
		n.queueChan = make(chan node.Message, n.maxQueue)
		n.stopChan = make(chan struct{})
	}

	return nil
}

// Execute processes a message through the rate limiter
func (n *RateLimitNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Refill tokens based on elapsed time
	n.refillTokens()

	// Check if we have tokens available
	if n.tokens >= 1.0 {
		// Consume a token and pass the message
		n.tokens -= 1.0
		msg.Payload["_rateLimit"] = map[string]interface{}{
			"allowed":   true,
			"tokens":    n.tokens,
			"rate":      n.rate,
			"window":    n.timeWindow.String(),
			"strategy":  n.strategy,
			"timestamp": time.Now().Format(time.RFC3339),
		}
		return msg, nil
	}

	// No tokens available - handle based on strategy
	switch n.strategy {
	case "drop":
		// Drop the message by returning with dropped flag
		msg.Payload["_rateLimit"] = map[string]interface{}{
			"allowed":   false,
			"dropped":   true,
			"reason":    "rate_limit_exceeded",
			"rate":      n.rate,
			"window":    n.timeWindow.String(),
			"timestamp": time.Now().Format(time.RFC3339),
		}
		return msg, fmt.Errorf("rate limit exceeded, message dropped")

	case "queue":
		// Queue the message for later processing
		if len(n.queue) < n.maxQueue {
			n.queue = append(n.queue, msg)
			msg.Payload["_rateLimit"] = map[string]interface{}{
				"allowed":     false,
				"queued":      true,
				"queueLength": len(n.queue),
				"rate":        n.rate,
				"window":      n.timeWindow.String(),
				"timestamp":   time.Now().Format(time.RFC3339),
			}
			return msg, fmt.Errorf("rate limit exceeded, message queued (position %d)", len(n.queue))
		}
		// Queue full, drop message
		msg.Payload["_rateLimit"] = map[string]interface{}{
			"allowed":   false,
			"dropped":   true,
			"reason":    "queue_full",
			"rate":      n.rate,
			"window":    n.timeWindow.String(),
			"timestamp": time.Now().Format(time.RFC3339),
		}
		return msg, fmt.Errorf("rate limit exceeded and queue full, message dropped")

	case "delay":
		// Calculate delay needed to get a token
		tokensNeeded := 1.0 - n.tokens
		refillRate := float64(n.rate) / n.timeWindow.Seconds()
		delay := time.Duration(tokensNeeded/refillRate) * time.Second

		// Wait for token (with context timeout)
		select {
		case <-time.After(delay):
			// Refill and consume token
			n.refillTokens()
			if n.tokens >= 1.0 {
				n.tokens -= 1.0
			}
			msg.Payload["_rateLimit"] = map[string]interface{}{
				"allowed":   true,
				"delayed":   true,
				"delayMs":   delay.Milliseconds(),
				"rate":      n.rate,
				"window":    n.timeWindow.String(),
				"timestamp": time.Now().Format(time.RFC3339),
			}
			return msg, nil
		case <-ctx.Done():
			msg.Payload["_rateLimit"] = map[string]interface{}{
				"allowed":   false,
				"dropped":   true,
				"reason":    "context_cancelled",
				"rate":      n.rate,
				"window":    n.timeWindow.String(),
				"timestamp": time.Now().Format(time.RFC3339),
			}
			return msg, ctx.Err()
		}
	}

	return msg, nil
}

// refillTokens adds tokens based on elapsed time since last update
func (n *RateLimitNode) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(n.lastUpdate)
	n.lastUpdate = now

	// Calculate tokens to add (rate per second * elapsed seconds)
	tokensToAdd := float64(n.rate) * elapsed.Seconds() / n.timeWindow.Seconds()
	n.tokens += tokensToAdd

	// Cap at max tokens (rate)
	if n.tokens > float64(n.rate) {
		n.tokens = float64(n.rate)
	}
}

// ProcessQueue processes queued messages (for queue strategy)
func (n *RateLimitNode) ProcessQueue(ctx context.Context, output func(node.Message) error) error {
	n.mu.Lock()
	if n.running {
		n.mu.Unlock()
		return nil
	}
	n.running = true
	n.mu.Unlock()

	ticker := time.NewTicker(n.timeWindow / time.Duration(n.rate))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-n.stopChan:
			return nil
		case <-ticker.C:
			n.mu.Lock()
			if len(n.queue) > 0 {
				// Dequeue and process first message
				msg := n.queue[0]
				n.queue = n.queue[1:]
				n.mu.Unlock()

				msg.Payload["_rateLimit"] = map[string]interface{}{
					"allowed":        true,
					"fromQueue":      true,
					"remainingQueue": len(n.queue),
					"timestamp":      time.Now().Format(time.RFC3339),
				}
				if err := output(msg); err != nil {
					return err
				}
			} else {
				n.mu.Unlock()
			}
		}
	}
}

// GetStats returns current rate limiter statistics
func (n *RateLimitNode) GetStats() map[string]interface{} {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.refillTokens()

	stats := map[string]interface{}{
		"tokens":     n.tokens,
		"maxTokens":  n.rate,
		"rate":       n.rate,
		"timeWindow": n.timeWindow.String(),
		"strategy":   n.strategy,
	}

	if n.strategy == "queue" {
		stats["queueLength"] = len(n.queue)
		stats["maxQueue"] = n.maxQueue
	}

	return stats
}

// Cleanup releases resources
func (n *RateLimitNode) Cleanup() error {
	if n.stopChan != nil {
		close(n.stopChan)
	}
	return nil
}

// NewRateLimitExecutor creates a new rate limit node executor
func NewRateLimitExecutor() node.Executor {
	return &RateLimitNode{}
}

// init registers the rate limit node
func init() {
	registry := node.GetGlobalRegistry()
	registry.Register(&node.NodeInfo{
		Type:        "rate-limit",
		Name:        "Rate Limit",
		Category:    node.NodeTypeFunction,
		Description: "Rate limit messages using token bucket algorithm",
		Icon:        "gauge",
		Color:       "#7B68EE",
		Properties: []node.PropertySchema{
			{
				Name:        "rate",
				Label:       "Rate",
				Type:        "number",
				Default:     10,
				Required:    true,
				Description: "Maximum number of messages per time window",
			},
			{
				Name:        "timeWindow",
				Label:       "Time Window",
				Type:        "string",
				Default:     "1s",
				Required:    true,
				Description: "Time window duration (e.g., 1s, 5s, 1m)",
			},
			{
				Name:        "strategy",
				Label:       "Strategy",
				Type:        "select",
				Default:     "drop",
				Required:    true,
				Description: "How to handle excess messages",
				Options:     []string{"drop", "queue", "delay"},
			},
			{
				Name:        "maxQueue",
				Label:       "Max Queue Size",
				Type:        "number",
				Default:     100,
				Required:    false,
				Description: "Maximum queue size (for queue strategy)",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Messages to rate limit"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Rate-limited messages"},
		},
		Factory: NewRateLimitExecutor,
	})
}
