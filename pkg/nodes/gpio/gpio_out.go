package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/node"
)

// gpioStateStore stores GPIO state for persistence
var gpioStateStore = struct {
	sync.RWMutex
	states map[int]bool
}{
	states: make(map[int]bool),
}

// GPIOOutConfig GPIO Out node configuration
type GPIOOutConfig struct {
	Pin          int    `json:"pin"`          // pin number
	InitialValue bool   `json:"initialValue"` // initial value
	Invert       bool   `json:"invert"`       // invert output
	PersistState bool   `json:"persistState"` // Persist state across restarts
	StateFile    string `json:"stateFile"`    // File to persist state (default: /var/lib/edgeflow/gpio_states.json)
	FailsafeMode bool   `json:"failsafeMode"` // Enable failsafe mode
	FailsafeValue bool  `json:"failsafeValue"` // Value to set on error/cleanup
	DriveStrength int   `json:"driveStrength"` // Drive strength (mA) for supported pins
}

// GPIOOutExecutor GPIO Out node executor
type GPIOOutExecutor struct {
	config      GPIOOutConfig
	hal         hal.HAL
	currentValue bool
}

// NewGPIOOutExecutor create GPIOOutExecutor
func NewGPIOOutExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var gpioConfig GPIOOutConfig
	if err := json.Unmarshal(configJSON, &gpioConfig); err != nil {
		return nil, fmt.Errorf("invalid gpio config: %w", err)
	}

	// Validate
	if gpioConfig.Pin < 0 {
		return nil, fmt.Errorf("invalid pin number")
	}

	// Default state file
	if gpioConfig.PersistState && gpioConfig.StateFile == "" {
		gpioConfig.StateFile = "/var/lib/edgeflow/gpio_states.json"
	}

	return &GPIOOutExecutor{
		config: gpioConfig,
	}, nil
}

// Init initializes the GPIO Out executor with config
func (e *GPIOOutExecutor) Init(config map[string]interface{}) error {
	// Config is already parsed in NewGPIOOutExecutor
	return nil
}

// Execute execute node
func (e *GPIOOutExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get HAL if not initialized
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h

		// Setup GPIO
		if err := e.setup(); err != nil {
			return node.Message{}, fmt.Errorf("failed to setup GPIO: %w", err)
		}
	}

	// Get value from message
	var value bool
	if msg.Payload != nil {
		// Try different field names
		if v, ok := msg.Payload["value"].(bool); ok {
			value = v
		} else if v, ok := msg.Payload["state"].(bool); ok {
			value = v
		} else if v, ok := msg.Payload["on"].(bool); ok {
			value = v
		} else if v, ok := msg.Payload["payload"].(bool); ok {
			value = v
		} else {
			// Try to convert from number
			if v, ok := msg.Payload["value"].(float64); ok {
				value = v > 0
			} else if v, ok := msg.Payload["payload"].(float64); ok {
				value = v > 0
			}
		}
	}

	// Apply invert
	if e.config.Invert {
		value = !value
	}

	// Write to GPIO
	gpio := e.hal.GPIO()
	if err := gpio.DigitalWrite(e.config.Pin, value); err != nil {
		// Failsafe mode: set to failsafe value on error
		if e.config.FailsafeMode {
			failsafeValue := e.config.FailsafeValue
			if e.config.Invert {
				failsafeValue = !failsafeValue
			}
			gpio.DigitalWrite(e.config.Pin, failsafeValue)
		}
		return node.Message{}, fmt.Errorf("failed to write GPIO: %w", err)
	}

	// Update current value
	e.currentValue = value

	// Persist state if enabled
	if e.config.PersistState {
		e.saveState(value)
	}

	// Return message with actual written value
	return node.Message{
		Payload: map[string]interface{}{
			"pin":   e.config.Pin,
			"value": value,
		},
	}, nil
}

// setup initialize GPIO
func (e *GPIOOutExecutor) setup() error {
	gpio := e.hal.GPIO()

	// Set mode to output
	if err := gpio.SetMode(e.config.Pin, hal.Output); err != nil {
		return err
	}

	// Set initial value
	initialValue := e.config.InitialValue
	if e.config.Invert {
		initialValue = !initialValue
	}
	if err := gpio.DigitalWrite(e.config.Pin, initialValue); err != nil {
		return err
	}

	return nil
}

// Cleanup cleanup resources
func (e *GPIOOutExecutor) Cleanup() error {
	if e.hal != nil {
		gpio := e.hal.GPIO()

		// Use failsafe value if failsafe mode is enabled
		if e.config.FailsafeMode {
			failsafeValue := e.config.FailsafeValue
			if e.config.Invert {
				failsafeValue = !failsafeValue
			}
			gpio.DigitalWrite(e.config.Pin, failsafeValue)
		} else {
			// Reset to initial value
			initialValue := e.config.InitialValue
			if e.config.Invert {
				initialValue = !initialValue
			}
			gpio.DigitalWrite(e.config.Pin, initialValue)
		}
	}
	return nil
}

// saveState saves the current GPIO state to file for persistence
func (e *GPIOOutExecutor) saveState(value bool) {
	gpioStateStore.Lock()
	defer gpioStateStore.Unlock()

	gpioStateStore.states[e.config.Pin] = value

	// Ensure directory exists
	dir := filepath.Dir(e.config.StateFile)
	os.MkdirAll(dir, 0755)

	// Save to file
	data, err := json.MarshalIndent(gpioStateStore.states, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(e.config.StateFile, data, 0644)
}

// loadState loads the persisted GPIO state from file
func (e *GPIOOutExecutor) loadState() (bool, bool) {
	gpioStateStore.RLock()
	defer gpioStateStore.RUnlock()

	// Check in-memory first
	if value, ok := gpioStateStore.states[e.config.Pin]; ok {
		return value, true
	}

	// Try loading from file
	data, err := os.ReadFile(e.config.StateFile)
	if err != nil {
		return false, false
	}

	var states map[int]bool
	if err := json.Unmarshal(data, &states); err != nil {
		return false, false
	}

	if value, ok := states[e.config.Pin]; ok {
		return value, true
	}

	return false, false
}
