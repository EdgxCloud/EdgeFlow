package core

import (
	"context"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// InjectNode sends messages at regular intervals
type InjectNode struct {
	interval time.Duration
	payload  map[string]interface{}
	topic    string
	ticker   *time.Ticker
	stopChan chan struct{}
}

// NewInjectNode creates a new inject node
func NewInjectNode() *InjectNode {
	return &InjectNode{
		interval: 1 * time.Second,
		payload:  make(map[string]interface{}),
		stopChan: make(chan struct{}),
	}
}

// Init initializes the inject node with configuration
func (n *InjectNode) Init(config map[string]interface{}) error {
	// Parse interval from new format (intervalType + intervalValue)
	intervalType := "seconds"
	if it, ok := config["intervalType"].(string); ok {
		intervalType = it
	}

	intervalValue := 5
	if iv, ok := config["intervalValue"].(float64); ok {
		intervalValue = int(iv)
	} else if iv, ok := config["intervalValue"].(int); ok {
		intervalValue = iv
	}

	// Convert to duration
	switch intervalType {
	case "seconds":
		n.interval = time.Duration(intervalValue) * time.Second
	case "minutes":
		n.interval = time.Duration(intervalValue) * time.Minute
	case "hours":
		n.interval = time.Duration(intervalValue) * time.Hour
	case "days":
		n.interval = time.Duration(intervalValue) * 24 * time.Hour
	case "months":
		// Approximate 1 month as 30 days
		n.interval = time.Duration(intervalValue) * 24 * 30 * time.Hour
	default:
		n.interval = time.Duration(intervalValue) * time.Second
	}

	// Backward compatibility: parse old "interval" string format
	if intervalStr, ok := config["interval"].(string); ok {
		duration, err := time.ParseDuration(intervalStr)
		if err == nil {
			n.interval = duration
		}
	}

	// Parse payload
	if payload, ok := config["payload"].(map[string]interface{}); ok {
		n.payload = payload
	}

	// Parse topic
	if topic, ok := config["topic"].(string); ok {
		n.topic = topic
	}

	return nil
}

// Execute processes incoming messages and sends periodic injections
func (n *InjectNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Inject node can be triggered manually or run on timer
	if msg.Type == node.MessageTypeEvent && msg.Payload["trigger"] == true {
		return n.createMessage(), nil
	}

	// For timer-based injection, this will be called periodically
	return n.createMessage(), nil
}

// Cleanup stops the inject node
func (n *InjectNode) Cleanup() error {
	if n.ticker != nil {
		n.ticker.Stop()
	}
	close(n.stopChan)
	return nil
}

// createMessage creates a new message with the configured payload
func (n *InjectNode) createMessage() node.Message {
	return node.Message{
		Type:    node.MessageTypeData,
		Payload: n.payload,
		Topic:   n.topic,
	}
}
