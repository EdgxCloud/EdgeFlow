//go:build linux
// +build linux

package gpio

import (
	"context"
	"fmt"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stianeikeland/go-rpio/v4"
)

type HCSR04Node struct {
	triggerPin int
	echoPin    int
	unit       string // "cm" or "inch"
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

	if err := rpio.Open(); err != nil {
		return fmt.Errorf("failed to open GPIO: %w", err)
	}

	trigger := rpio.Pin(n.triggerPin)
	echo := rpio.Pin(n.echoPin)

	trigger.Output()
	echo.Input()

	return nil
}

func (n *HCSR04Node) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	trigger := rpio.Pin(n.triggerPin)
	echo := rpio.Pin(n.echoPin)

	trigger.Low()
	time.Sleep(2 * time.Microsecond)
	trigger.High()
	time.Sleep(10 * time.Microsecond)
	trigger.Low()

	timeout := time.Now().Add(100 * time.Millisecond)
	for echo.Read() == rpio.Low {
		if time.Now().After(timeout) {
			return msg, fmt.Errorf("timeout waiting for echo")
		}
	}
	startTime := time.Now()

	timeout = time.Now().Add(100 * time.Millisecond)
	for echo.Read() == rpio.High {
		if time.Now().After(timeout) {
			return msg, fmt.Errorf("timeout reading echo")
		}
	}
	duration := time.Since(startTime)

	distanceCm := float64(duration.Microseconds()) / 58.0

	var distance float64
	if n.unit == "inch" {
		distance = distanceCm / 2.54
	} else {
		distance = distanceCm
	}

	msg.Payload = map[string]interface{}{
		"distance": distance,
		"unit":     n.unit,
	}

	return msg, nil
}

func (n *HCSR04Node) Cleanup() error {
	return rpio.Close()
}

func NewHCSR04Executor(config map[string]interface{}) (node.Executor, error) {
	executor := NewHCSR04Node()
	if err := executor.Init(config); err != nil {
		return nil, err
	}
	return executor, nil
}
