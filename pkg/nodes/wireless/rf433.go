package wireless

import (
	"context"
	"fmt"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// RF433Executor provides 433MHz RF wireless communication
type RF433Executor struct {
	dataPin     int
	protocol    int
	pulseLength int
	repeatCount int
}

// NewRF433Executor creates a new RF433 executor
func NewRF433Executor() node.Executor {
	return &RF433Executor{
		protocol:    1,
		pulseLength: 350,
		repeatCount: 10,
	}
}

func (e *RF433Executor) Init(config map[string]interface{}) error {
	if dp, ok := config["dataPin"].(float64); ok {
		e.dataPin = int(dp)
	}
	if p, ok := config["protocol"].(float64); ok {
		e.protocol = int(p)
	}
	if pl, ok := config["pulseLength"].(float64); ok {
		e.pulseLength = int(pl)
	}
	if rc, ok := config["repeatCount"].(float64); ok {
		e.repeatCount = int(rc)
	}
	return nil
}

func (e *RF433Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	action, _ := msg.Payload["action"].(string)
	switch action {
	case "send", "receive", "listen":
		return msg, fmt.Errorf("RF433 hardware requires Linux with GPIO support (use rf433 GPIO node)")
	default:
		msg.Payload["info"] = "433MHz RF wireless transceiver"
		msg.Payload["protocol"] = e.protocol
		msg.Payload["pulse_length_us"] = e.pulseLength
		msg.Payload["actions"] = []string{"send", "receive", "listen"}
		msg.Payload["note"] = "Configure rf433 GPIO node for hardware access on Linux"
		return msg, nil
	}
}

func (e *RF433Executor) Cleanup() error { return nil }
