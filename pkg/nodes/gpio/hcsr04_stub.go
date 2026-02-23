//go:build !linux
// +build !linux

package gpio

import (
	"context"
	"fmt"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

type HCSR04Node struct {
	triggerPin int
	echoPin    int
	unit       string
}

func NewHCSR04Node() *HCSR04Node {
	return &HCSR04Node{
		triggerPin: 23,
		echoPin:    24,
		unit:       "cm",
	}
}

func (n *HCSR04Node) Init(config map[string]interface{}) error {
	if trigger, ok := config["triggerPin"].(float64); ok {
		n.triggerPin = int(trigger)
	} else if trigger, ok := config["triggerPin"].(int); ok {
		n.triggerPin = trigger
	}
	if echo, ok := config["echoPin"].(float64); ok {
		n.echoPin = int(echo)
	} else if echo, ok := config["echoPin"].(int); ok {
		n.echoPin = echo
	}
	if unit, ok := config["unit"].(string); ok {
		n.unit = unit
	}
	return nil
}

func (n *HCSR04Node) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	return msg, fmt.Errorf("HC-SR04 sensor is only supported on Linux platforms")
}

func (n *HCSR04Node) Cleanup() error {
	return nil
}

func NewHCSR04Executor(config map[string]interface{}) (node.Executor, error) {
	executor := NewHCSR04Node()
	if err := executor.Init(config); err != nil {
		return nil, err
	}
	return executor, nil
}
