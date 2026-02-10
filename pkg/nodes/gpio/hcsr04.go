//go:build linux
// +build linux

package gpio

import (
	"context"
	"fmt"
	"time"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/node"
)

type HCSR04Node struct {
	triggerPin  int
	echoPin     int
	unit        string // "cm" or "inch"
	halInstance hal.HAL
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

	// Get HAL and configure GPIO pins
	h, err := hal.GetGlobalHAL()
	if err != nil {
		return fmt.Errorf("failed to get HAL: %w", err)
	}
	n.halInstance = h

	gpio := h.GPIO()

	if err := gpio.SetMode(n.triggerPin, hal.Output); err != nil {
		return fmt.Errorf("failed to set trigger pin %d as output: %w", n.triggerPin, err)
	}
	if err := gpio.SetMode(n.echoPin, hal.Input); err != nil {
		return fmt.Errorf("failed to set echo pin %d as input: %w", n.echoPin, err)
	}

	return nil
}

func (n *HCSR04Node) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	gpio := n.halInstance.GPIO()

	// Send trigger pulse
	gpio.DigitalWrite(n.triggerPin, false)
	time.Sleep(2 * time.Microsecond)
	gpio.DigitalWrite(n.triggerPin, true)
	time.Sleep(10 * time.Microsecond)
	gpio.DigitalWrite(n.triggerPin, false)

	// Wait for echo to go high
	timeout := time.Now().Add(100 * time.Millisecond)
	for {
		val, err := gpio.DigitalRead(n.echoPin)
		if err != nil {
			return msg, fmt.Errorf("failed to read echo pin: %w", err)
		}
		if val {
			break
		}
		if time.Now().After(timeout) {
			return msg, fmt.Errorf("timeout waiting for echo")
		}
	}
	startTime := time.Now()

	// Wait for echo to go low
	timeout = time.Now().Add(100 * time.Millisecond)
	for {
		val, err := gpio.DigitalRead(n.echoPin)
		if err != nil {
			return msg, fmt.Errorf("failed to read echo pin: %w", err)
		}
		if !val {
			break
		}
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
	return nil
}

func NewHCSR04Executor(config map[string]interface{}) (node.Executor, error) {
	executor := NewHCSR04Node()
	if err := executor.Init(config); err != nil {
		return nil, err
	}
	return executor, nil
}
