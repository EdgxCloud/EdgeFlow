package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// TriggerNode sends a message, then optionally sends a second message after a delay
// Useful for debouncing, timeout patterns, and pulse generation
type TriggerNode struct {
	// Configuration
	op              string        // "send" or "send-then-send" or "send-then-nothing"
	initialPayload  interface{}   // Payload to send immediately
	secondPayload   interface{}   // Payload to send after delay
	delay           time.Duration // Delay before second message
	duration        time.Duration // Duration to keep initial state (for send-then-send)
	extend          bool          // Extend delay on new message
	reset           bool          // Reset on second message
	resetOnNewMsg   bool          // Reset timer on any new message

	// Runtime state
	timer           *time.Timer
	timerMu         sync.Mutex
	outputChan      chan node.Message
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	lastMessage     node.Message
	messageMu       sync.RWMutex
}

// NewTriggerNode creates a new Trigger node
func NewTriggerNode() *TriggerNode {
	return &TriggerNode{
		op:             "send-then-send",
		initialPayload: nil, // nil means pass through original message
		secondPayload:  nil, // nil means don't send second message
		delay:          250 * time.Millisecond,
		extend:         false,
		reset:          false,
		resetOnNewMsg:  false,
		outputChan:     make(chan node.Message, 10),
	}
}

// Init initializes the Trigger node
func (n *TriggerNode) Init(config map[string]interface{}) error {
	// Parse operation mode
	if op, ok := config["op"].(string); ok {
		if op != "send" && op != "send-then-send" && op != "send-then-nothing" {
			return fmt.Errorf("op must be 'send', 'send-then-send', or 'send-then-nothing'")
		}
		n.op = op
	}

	// Parse initial payload
	if payload, ok := config["initialPayload"]; ok {
		n.initialPayload = payload
	}

	// Parse second payload
	if payload, ok := config["secondPayload"]; ok {
		n.secondPayload = payload
	}

	// Parse delay
	if delayStr, ok := config["delay"].(string); ok {
		duration, err := time.ParseDuration(delayStr)
		if err != nil {
			return fmt.Errorf("invalid delay duration: %w", err)
		}
		n.delay = duration
	} else if delayMs, ok := config["delay"].(float64); ok {
		n.delay = time.Duration(delayMs) * time.Millisecond
	}

	// Parse duration (for send-then-send mode)
	if durationStr, ok := config["duration"].(string); ok {
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		n.duration = duration
	} else if durationMs, ok := config["duration"].(float64); ok {
		n.duration = time.Duration(durationMs) * time.Millisecond
	} else {
		n.duration = n.delay // Default duration = delay
	}

	// Parse extend flag
	if extend, ok := config["extend"].(bool); ok {
		n.extend = extend
	}

	// Parse reset flag
	if reset, ok := config["reset"].(bool); ok {
		n.reset = reset
	}

	// Parse resetOnNewMsg flag
	if resetOnNewMsg, ok := config["resetOnNewMsg"].(bool); ok {
		n.resetOnNewMsg = resetOnNewMsg
	}

	return nil
}

// Start starts the Trigger node
func (n *TriggerNode) Start(ctx context.Context) error {
	n.ctx, n.cancel = context.WithCancel(ctx)
	return nil
}

// Execute processes incoming messages
func (n *TriggerNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	n.timerMu.Lock()
	defer n.timerMu.Unlock()

	// Store last message for potential use
	n.messageMu.Lock()
	n.lastMessage = msg
	n.messageMu.Unlock()

	// Handle based on operation mode
	switch n.op {
	case "send":
		// Simple send mode - just pass through or send initial payload
		return n.createOutputMessage(msg, n.initialPayload), nil

	case "send-then-send":
		// Send initial message immediately
		initialMsg := n.createOutputMessage(msg, n.initialPayload)

		// Send initial message to output channel
		select {
		case n.outputChan <- initialMsg:
		default:
			// Channel full, skip
		}

		// Handle timer for second message
		if n.timer != nil {
			if n.extend {
				// Extend the timer (restart it)
				n.timer.Stop()
				n.timer = time.AfterFunc(n.delay, func() {
					n.sendSecondMessage(msg)
				})
			} else if n.resetOnNewMsg {
				// Reset the timer
				n.timer.Stop()
				n.timer = time.AfterFunc(n.delay, func() {
					n.sendSecondMessage(msg)
				})
			}
			// If extend and resetOnNewMsg are both false, ignore new messages
		} else {
			// Start new timer
			n.timer = time.AfterFunc(n.delay, func() {
				n.sendSecondMessage(msg)
			})
		}

		// Return empty message since we already sent via channel
		return node.Message{}, nil

	case "send-then-nothing":
		// Send message, then block subsequent messages for duration
		if n.timer == nil {
			// Not currently blocking, send the message
			outputMsg := n.createOutputMessage(msg, n.initialPayload)

			// Start blocking timer
			n.timer = time.AfterFunc(n.duration, func() {
				n.timerMu.Lock()
				n.timer = nil
				n.timerMu.Unlock()
			})

			return outputMsg, nil
		}

		// Currently blocking, don't send
		return node.Message{}, nil

	default:
		return node.Message{}, fmt.Errorf("unknown operation mode: %s", n.op)
	}
}

// sendSecondMessage sends the second message after the delay
func (n *TriggerNode) sendSecondMessage(originalMsg node.Message) {
	n.timerMu.Lock()
	n.timer = nil
	n.timerMu.Unlock()

	secondMsg := n.createOutputMessage(originalMsg, n.secondPayload)

	select {
	case n.outputChan <- secondMsg:
	case <-n.ctx.Done():
		// Node is stopping
	default:
		// Channel full, skip
	}
}

// createOutputMessage creates an output message from template
func (n *TriggerNode) createOutputMessage(originalMsg node.Message, payload interface{}) node.Message {
	msg := node.Message{
		Type:  node.MessageTypeData,
		Topic: originalMsg.Topic,
	}

	// Set payload
	if payload == nil {
		// nil means use original message payload
		msg.Payload = originalMsg.Payload
	} else {
		// Convert payload to map[string]interface{}
		switch p := payload.(type) {
		case map[string]interface{}:
			msg.Payload = p
		default:
			msg.Payload = map[string]interface{}{"value": p}
		}
	}

	return msg
}

// GetOutputChannel returns the async output channel
func (n *TriggerNode) GetOutputChannel() <-chan node.Message {
	return n.outputChan
}

// Cleanup stops the node and cleans up resources
func (n *TriggerNode) Cleanup() error {
	if n.cancel != nil {
		n.cancel()
	}

	n.timerMu.Lock()
	if n.timer != nil {
		n.timer.Stop()
		n.timer = nil
	}
	n.timerMu.Unlock()

	n.wg.Wait()

	if n.outputChan != nil {
		close(n.outputChan)
	}

	return nil
}
