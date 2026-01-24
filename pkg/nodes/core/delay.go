package core

import (
	"context"
	"fmt"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// DelayNode delays message processing for a specified duration
type DelayNode struct {
	duration time.Duration
	timeout  time.Duration
}

// NewDelayNode creates a new delay node
func NewDelayNode() *DelayNode {
	return &DelayNode{
		duration: 1 * time.Second,
		timeout:  1 * time.Minute,
	}
}

// Init initializes the delay node with configuration
func (n *DelayNode) Init(config map[string]interface{}) error {
	// Parse duration
	if durationStr, ok := config["duration"].(string); ok {
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}
		n.duration = duration
	} else if durationMs, ok := config["duration"].(float64); ok {
		n.duration = time.Duration(durationMs) * time.Millisecond
	}

	// Parse timeout
	if timeoutStr, ok := config["timeout"].(string); ok {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %w", err)
		}
		n.timeout = timeout
	}

	// Validate duration
	if n.duration < 0 {
		return fmt.Errorf("duration cannot be negative")
	}

	if n.duration > n.timeout {
		return fmt.Errorf("duration cannot exceed timeout")
	}

	return nil
}

// Execute delays the message by the configured duration
func (n *DelayNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Create a timer for the delay
	timer := time.NewTimer(n.duration)
	defer timer.Stop()

	// Wait for either the delay or context cancellation
	select {
	case <-timer.C:
		// Delay completed successfully
		return msg, nil
	case <-ctx.Done():
		// Context was cancelled
		return msg, fmt.Errorf("delay cancelled: %w", ctx.Err())
	}
}

// Cleanup stops the delay node
func (n *DelayNode) Cleanup() error {
	return nil
}
