package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/node"
)

// RelayConfig نود Relay
type RelayConfig struct {
	Pin            int  `json:"pin"`            // شماره پین
	InitialState   bool `json:"initialState"`   // وضعیت اولیه
	ActiveLow      bool `json:"activeLow"`      // فعال در سطح پایین
	PulseMode      bool `json:"pulseMode"`      // حالت پالس
	PulseDuration  int  `json:"pulseDuration"`  // مدت پالس (ms)
}

// RelayExecutor اجراکننده نود Relay
type RelayExecutor struct {
	config RelayConfig
	hal    hal.HAL
	state  bool
}

// NewRelayExecutor ایجاد RelayExecutor
func NewRelayExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var relayConfig RelayConfig
	if err := json.Unmarshal(configJSON, &relayConfig); err != nil {
		return nil, fmt.Errorf("invalid relay config: %w", err)
	}

	// Validate
	if relayConfig.Pin < 0 {
		return nil, fmt.Errorf("invalid pin number")
	}

	// Default pulse duration
	if relayConfig.PulseDuration == 0 {
		relayConfig.PulseDuration = 100
	}

	return &RelayExecutor{
		config: relayConfig,
		state:  relayConfig.InitialState,
	}, nil
}

// Init initializes the Relay executor with config
func (e *RelayExecutor) Init(config map[string]interface{}) error {
	// Config is already parsed in NewRelayExecutor
	return nil
}

// Execute اجرای نود
func (e *RelayExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get HAL if not initialized
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h

		// Setup relay
		if err := e.setup(); err != nil {
			return node.Message{}, fmt.Errorf("failed to setup relay: %w", err)
		}
	}

	// Get command from message
	var command string

	if msg.Payload != nil {
		if cmd, ok := msg.Payload["command"].(string); ok {
			command = cmd
		} else if s, ok := msg.Payload["state"].(bool); ok {
			if s {
				command = "on"
			} else {
				command = "off"
			}
		} else if v, ok := msg.Payload["value"].(bool); ok {
			if v {
				command = "on"
			} else {
				command = "off"
			}
		}
	}

	gpio := e.hal.GPIO()

	switch command {
	case "on":
		e.state = true
		actualValue := e.state
		if e.config.ActiveLow {
			actualValue = !actualValue
		}
		if err := gpio.DigitalWrite(e.config.Pin, actualValue); err != nil {
			return node.Message{}, err
		}

	case "off":
		e.state = false
		actualValue := e.state
		if e.config.ActiveLow {
			actualValue = !actualValue
		}
		if err := gpio.DigitalWrite(e.config.Pin, actualValue); err != nil {
			return node.Message{}, err
		}

	case "toggle":
		e.state = !e.state
		actualValue := e.state
		if e.config.ActiveLow {
			actualValue = !actualValue
		}
		if err := gpio.DigitalWrite(e.config.Pin, actualValue); err != nil {
			return node.Message{}, err
		}

	case "pulse":
		// Turn on
		e.state = true
		actualValue := e.state
		if e.config.ActiveLow {
			actualValue = !actualValue
		}
		if err := gpio.DigitalWrite(e.config.Pin, actualValue); err != nil {
			return node.Message{}, err
		}

		// Wait
		time.Sleep(time.Duration(e.config.PulseDuration) * time.Millisecond)

		// Turn off
		e.state = false
		actualValue = e.state
		if e.config.ActiveLow {
			actualValue = !actualValue
		}
		if err := gpio.DigitalWrite(e.config.Pin, actualValue); err != nil {
			return node.Message{}, err
		}

	default:
		return node.Message{}, fmt.Errorf("unknown command: %s", command)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"pin":     e.config.Pin,
			"state":   e.state,
			"command": command,
		},
	}, nil
}

// setup راه‌اندازی relay
func (e *RelayExecutor) setup() error {
	gpio := e.hal.GPIO()

	// Set mode to output
	if err := gpio.SetMode(e.config.Pin, hal.Output); err != nil {
		return err
	}

	// Set initial state
	initialValue := e.config.InitialState
	if e.config.ActiveLow {
		initialValue = !initialValue
	}
	if err := gpio.DigitalWrite(e.config.Pin, initialValue); err != nil {
		return err
	}

	return nil
}

// Cleanup پاکسازی منابع
func (e *RelayExecutor) Cleanup() error {
	// Turn off relay
	if e.hal != nil {
		gpio := e.hal.GPIO()
		offValue := false
		if e.config.ActiveLow {
			offValue = true
		}
		gpio.DigitalWrite(e.config.Pin, offValue)
	}
	return nil
}
