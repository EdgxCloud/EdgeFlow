package core

import (
	"context"
	"fmt"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/robfig/cron/v3"
)

// ScheduleNode triggers messages based on cron expressions
type ScheduleNode struct {
	cronExpr  string
	payload   map[string]interface{}
	topic     string
	timezone  string
	cron      *cron.Cron
	outputChan chan node.Message
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewScheduleNode creates a new schedule node
func NewScheduleNode() *ScheduleNode {
	return &ScheduleNode{
		payload:    make(map[string]interface{}),
		timezone:   "Local",
		outputChan: make(chan node.Message, 10),
	}
}

// Init initializes the schedule node with configuration
func (n *ScheduleNode) Init(config map[string]interface{}) error {
	// Parse cron expression
	if cronExpr, ok := config["cron"].(string); ok {
		n.cronExpr = cronExpr
	} else {
		return fmt.Errorf("cron expression is required")
	}

	// Parse payload
	if payload, ok := config["payload"].(map[string]interface{}); ok {
		n.payload = payload
	}

	// Parse topic
	if topic, ok := config["topic"].(string); ok {
		n.topic = topic
	}

	// Parse timezone
	if timezone, ok := config["timezone"].(string); ok {
		n.timezone = timezone
	}

	// Initialize cron scheduler
	var loc *time.Location
	var err error

	if n.timezone == "Local" || n.timezone == "" {
		loc = time.Local
	} else {
		loc, err = time.LoadLocation(n.timezone)
		if err != nil {
			return fmt.Errorf("invalid timezone: %w", err)
		}
	}

	n.cron = cron.New(cron.WithLocation(loc), cron.WithSeconds())

	// Add cron job
	_, err = n.cron.AddFunc(n.cronExpr, func() {
		select {
		case n.outputChan <- n.createMessage():
		default:
			// Channel full, skip this trigger
		}
	})

	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	return nil
}

// Start starts the cron scheduler
func (n *ScheduleNode) Start(ctx context.Context) error {
	n.ctx, n.cancel = context.WithCancel(ctx)
	n.cron.Start()
	return nil
}

// Execute processes incoming messages
func (n *ScheduleNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Schedule node typically doesn't process incoming messages
	// It generates messages based on cron schedule
	// But we can support manual trigger
	if msg.Type == node.MessageTypeEvent && msg.Payload["trigger"] == true {
		return n.createMessage(), nil
	}

	return node.Message{}, nil
}

// Cleanup stops the schedule node
func (n *ScheduleNode) Cleanup() error {
	if n.cron != nil {
		ctx := n.cron.Stop()
		<-ctx.Done()
	}
	if n.cancel != nil {
		n.cancel()
	}
	close(n.outputChan)
	return nil
}

// createMessage creates a new message with the configured payload
func (n *ScheduleNode) createMessage() node.Message {
	payload := make(map[string]interface{})
	for k, v := range n.payload {
		payload[k] = v
	}
	payload["timestamp"] = time.Now().Unix()
	payload["scheduled"] = true

	return node.Message{
		Type:    node.MessageTypeData,
		Payload: payload,
		Topic:   n.topic,
	}
}

// GetOutputChannel returns the channel for scheduled messages
func (n *ScheduleNode) GetOutputChannel() <-chan node.Message {
	return n.outputChan
}
