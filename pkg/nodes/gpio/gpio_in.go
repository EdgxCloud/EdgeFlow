package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/node"
)

// GPIOInConfig GPIO In node configuration
type GPIOInConfig struct {
	Pin           int    `json:"pin"`           // pin number
	PullMode      string `json:"pullMode"`      // none, up, down
	EdgeMode      string `json:"edgeMode"`      // none, rising, falling, both
	Debounce      int    `json:"debounce"`      // Debounce time (ms)
	PollInterval  int    `json:"pollInterval"`  // Polling interval (ms) if edge is none
	InterruptMode bool   `json:"interruptMode"` // Use hardware interrupt if available
	ReadInitial   bool   `json:"readInitial"`   // Read and emit initial value on startup
	Glitch        int    `json:"glitch"`        // Glitch filter time in microseconds (hardware debounce)
}

// GPIOInExecutor GPIO In node executor
type GPIOInExecutor struct {
	config     GPIOInConfig
	hal        hal.HAL
	outputChan chan node.Message
	stopChan   chan struct{}
	lastValue  bool
	lastChange time.Time
}

// NewGPIOInExecutor create GPIOInExecutor
func NewGPIOInExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var gpioConfig GPIOInConfig
	if err := json.Unmarshal(configJSON, &gpioConfig); err != nil {
		return nil, fmt.Errorf("invalid gpio config: %w", err)
	}

	// Validate
	if gpioConfig.Pin < 0 {
		return nil, fmt.Errorf("invalid pin number")
	}

	// Default values
	if gpioConfig.PullMode == "" {
		gpioConfig.PullMode = "none"
	}
	if gpioConfig.EdgeMode == "" {
		gpioConfig.EdgeMode = "none"
	}
	if gpioConfig.Debounce == 0 {
		gpioConfig.Debounce = 50
	}
	if gpioConfig.PollInterval == 0 {
		gpioConfig.PollInterval = 100
	}
	// Enable interrupt mode by default when edge detection is used
	if gpioConfig.EdgeMode != "none" && !gpioConfig.InterruptMode {
		gpioConfig.InterruptMode = true
	}

	return &GPIOInExecutor{
		config:     gpioConfig,
		outputChan: make(chan node.Message, 10),
		stopChan:   make(chan struct{}),
	}, nil
}

// Init initializes the GPIO In executor with config
func (e *GPIOInExecutor) Init(config map[string]interface{}) error {
	// Config is already parsed in NewGPIOInExecutor
	return nil
}

// Execute execute node
func (e *GPIOInExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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

		// Emit initial value if configured
		if e.config.ReadInitial {
			msg := node.Message{
				Payload: map[string]interface{}{
					"pin":     e.config.Pin,
					"value":   e.lastValue,
					"initial": true,
				},
			}
			select {
			case e.outputChan <- msg:
			default:
			}
		}

		// Start monitoring
		if e.config.EdgeMode != "none" {
			// Edge detection mode (interrupt-driven)
			go e.edgeLoop()
		} else {
			// Polling mode
			go e.pollLoop()
		}
	}

	// Wait for input or context cancellation
	select {
	case <-ctx.Done():
		return node.Message{}, ctx.Err()
	case gpioMsg := <-e.outputChan:
		return gpioMsg, nil
	}
}

// setup initialize GPIO
func (e *GPIOInExecutor) setup() error {
	gpio := e.hal.GPIO()

	// Set mode to input
	if err := gpio.SetMode(e.config.Pin, hal.Input); err != nil {
		return err
	}

	// Set pull mode
	var pull hal.PullMode
	switch e.config.PullMode {
	case "up":
		pull = hal.PullUp
	case "down":
		pull = hal.PullDown
	default:
		pull = hal.PullNone
	}
	if err := gpio.SetPull(e.config.Pin, pull); err != nil {
		return err
	}

	// Read initial value
	value, err := gpio.DigitalRead(e.config.Pin)
	if err != nil {
		return err
	}
	e.lastValue = value

	return nil
}

// edgeLoop edge detection loop
func (e *GPIOInExecutor) edgeLoop() {
	gpio := e.hal.GPIO()

	// Map edge mode
	var edge hal.EdgeMode
	switch e.config.EdgeMode {
	case "rising":
		edge = hal.EdgeRising
	case "falling":
		edge = hal.EdgeFalling
	case "both":
		edge = hal.EdgeBoth
	default:
		edge = hal.EdgeNone
	}

	// Watch for edge changes
	gpio.WatchEdge(e.config.Pin, edge, func(pin int, value bool) {
		// Debounce
		now := time.Now()
		if now.Sub(e.lastChange) < time.Duration(e.config.Debounce)*time.Millisecond {
			return
		}
		e.lastChange = now
		e.lastValue = value

		// Determine edge type
		edgeType := "unknown"
		if value && !e.lastValue {
			edgeType = "rising"
		} else if !value && e.lastValue {
			edgeType = "falling"
		}

		// Send message
		msg := node.Message{
			Payload: map[string]interface{}{
				"pin":      pin,
				"value":    value,
				"edge":     edgeType,
				"interrupt": e.config.InterruptMode,
			},
		}

		select {
		case e.outputChan <- msg:
		default:
		}
	})
}

// pollLoop polling loop
func (e *GPIOInExecutor) pollLoop() {
	gpio := e.hal.GPIO()
	ticker := time.NewTicker(time.Duration(e.config.PollInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopChan:
			return
		case <-ticker.C:
			value, err := gpio.DigitalRead(e.config.Pin)
			if err != nil {
				continue
			}

			// Check if value changed
			if value != e.lastValue {
				// Debounce
				now := time.Now()
				if now.Sub(e.lastChange) < time.Duration(e.config.Debounce)*time.Millisecond {
					continue
				}
				e.lastChange = now
				e.lastValue = value

				// Send message
				msg := node.Message{
					Payload: map[string]interface{}{
						"pin":   e.config.Pin,
						"value": value,
					},
				}

				select {
				case e.outputChan <- msg:
				default:
				}
			}
		}
	}
}

// Cleanup cleanup resources
func (e *GPIOInExecutor) Cleanup() error {
	close(e.stopChan)
	close(e.outputChan)
	return nil
}
