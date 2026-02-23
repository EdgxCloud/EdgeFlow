package core

import (
	"context"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// InjectNode sends messages at regular intervals
type InjectNode struct {
	interval    time.Duration
	payload     map[string]interface{}
	topic       string
	toggleMode  bool // if true, alternates payload["value"] between true/false
	toggleState bool
	ticker      *time.Ticker
	stopChan    chan struct{}
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

	// Toggle mode: alternates value between true and false each tick
	if toggle, ok := config["toggle"].(bool); ok {
		n.toggleMode = toggle
	}

	return nil
}

// Execute passes through the message from Run() — must NOT call createMessage()
// again, as that would double-consume the toggle state.
func (n *InjectNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	return msg, nil
}

// Run implements SelfTriggering — sends messages at the configured interval
func (n *InjectNode) Run(ctx context.Context, send func(node.Message)) {
	n.ticker = time.NewTicker(n.interval)
	defer n.ticker.Stop()

	// Send an initial message immediately
	send(n.createMessage())

	for {
		select {
		case <-ctx.Done():
			return
		case <-n.stopChan:
			return
		case <-n.ticker.C:
			send(n.createMessage())
		}
	}
}

// Cleanup stops the inject node
func (n *InjectNode) Cleanup() error {
	if n.ticker != nil {
		n.ticker.Stop()
	}
	select {
	case <-n.stopChan:
		// Already closed
	default:
		close(n.stopChan)
	}
	return nil
}

// createMessage creates a new message with the configured payload
func (n *InjectNode) createMessage() node.Message {
	payload := n.payload

	// In toggle mode, alternate value between true and false
	if n.toggleMode {
		payload = map[string]interface{}{
			"value": n.toggleState,
		}
		// Copy any extra fields from configured payload
		for k, v := range n.payload {
			if k != "value" {
				payload[k] = v
			}
		}
		n.toggleState = !n.toggleState
	}

	return node.Message{
		Type:    node.MessageTypeData,
		Payload: payload,
		Topic:   n.topic,
	}
}
