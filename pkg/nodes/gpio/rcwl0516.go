//go:build linux
// +build linux

package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

// RCWL0516Config holds configuration for RCWL-0516 microwave motion sensor
type RCWL0516Config struct {
	Pin            string `json:"pin"`
	ActiveHigh     bool   `json:"active_high"`
	DebounceMs     int    `json:"debounce_ms"`
	HoldTimeMs     int    `json:"hold_time_ms"`
	EnableCallback bool   `json:"enable_callback"`
	PollInterval   int    `json:"poll_interval_ms"`
}

// RCWL0516Executor implements microwave motion detection sensor
type RCWL0516Executor struct {
	config      RCWL0516Config
	pin         gpio.PinIO
	mu          sync.Mutex
	hostInited  bool
	initialized bool
	lastState   bool
	lastChange  time.Time
	motionCount uint64
}

func (e *RCWL0516Executor) Init(config map[string]interface{}) error {
	e.config = RCWL0516Config{
		Pin:            "GPIO17",
		ActiveHigh:     true,
		DebounceMs:     50,
		HoldTimeMs:     2000,
		EnableCallback: false,
		PollInterval:   100,
	}

	if config != nil {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := json.Unmarshal(configJSON, &e.config); err != nil {
			return fmt.Errorf("failed to parse RCWL0516 config: %w", err)
		}
	}

	return nil
}

func (e *RCWL0516Executor) initHardware() error {
	if e.initialized {
		return nil
	}

	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return fmt.Errorf("failed to initialize periph host: %w", err)
		}
		e.hostInited = true
	}

	pin := gpioreg.ByName(e.config.Pin)
	if pin == nil {
		return fmt.Errorf("failed to find pin %s", e.config.Pin)
	}

	if err := pin.In(gpio.PullDown, gpio.BothEdges); err != nil {
		return fmt.Errorf("failed to configure pin as input: %w", err)
	}

	e.pin = pin
	e.lastChange = time.Now()
	e.initialized = true
	return nil
}

func (e *RCWL0516Executor) readMotion() bool {
	level := e.pin.Read()
	if e.config.ActiveHigh {
		return level == gpio.High
	}
	return level == gpio.Low
}

func (e *RCWL0516Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.initHardware(); err != nil {
		return node.Message{}, err
	}

	action := "read"
	if payload := msg.Payload; payload != nil {
		if a, ok := payload["action"].(string); ok {
			action = a
		}
	}

	switch action {
	case "read":
		return e.readState()
	case "reset_count":
		return e.resetCount()
	case "configure":
		return e.handleConfigure(msg)
	case "status":
		return e.getStatus()
	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *RCWL0516Executor) readState() (node.Message, error) {
	currentState := e.readMotion()
	now := time.Now()

	// Debounce
	if now.Sub(e.lastChange) < time.Duration(e.config.DebounceMs)*time.Millisecond {
		return node.Message{
			Payload: map[string]interface{}{
				"motion":       e.lastState,
				"triggered":    false,
				"motion_count": e.motionCount,
				"timestamp":    now.Unix(),
			},
		}, nil
	}

	triggered := false
	if currentState && !e.lastState {
		// Rising edge - motion detected
		triggered = true
		e.motionCount++
		e.lastChange = now
	} else if !currentState && e.lastState {
		// Falling edge - motion ended
		e.lastChange = now
	}

	e.lastState = currentState

	// Calculate hold time remaining if in motion state
	holdRemaining := 0
	if currentState && e.config.HoldTimeMs > 0 {
		elapsed := now.Sub(e.lastChange).Milliseconds()
		remaining := int64(e.config.HoldTimeMs) - elapsed
		if remaining > 0 {
			holdRemaining = int(remaining)
		}
	}

	return node.Message{
		Payload: map[string]interface{}{
			"motion":             currentState,
			"triggered":          triggered,
			"motion_count":       e.motionCount,
			"hold_remaining_ms":  holdRemaining,
			"last_motion":        e.lastChange.Unix(),
			"timestamp":          now.Unix(),
		},
	}, nil
}

func (e *RCWL0516Executor) resetCount() (node.Message, error) {
	oldCount := e.motionCount
	e.motionCount = 0

	return node.Message{
		Payload: map[string]interface{}{
			"status":     "count_reset",
			"old_count":  oldCount,
			"new_count":  0,
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

func (e *RCWL0516Executor) handleConfigure(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	if activeHigh, ok := payload["active_high"].(bool); ok {
		e.config.ActiveHigh = activeHigh
	}
	if debounce, ok := payload["debounce_ms"].(float64); ok {
		e.config.DebounceMs = int(debounce)
	}
	if holdTime, ok := payload["hold_time_ms"].(float64); ok {
		e.config.HoldTimeMs = int(holdTime)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":       "configured",
			"active_high":  e.config.ActiveHigh,
			"debounce_ms":  e.config.DebounceMs,
			"hold_time_ms": e.config.HoldTimeMs,
		},
	}, nil
}

func (e *RCWL0516Executor) getStatus() (node.Message, error) {
	return node.Message{
		Payload: map[string]interface{}{
			"pin":          e.config.Pin,
			"active_high":  e.config.ActiveHigh,
			"debounce_ms":  e.config.DebounceMs,
			"hold_time_ms": e.config.HoldTimeMs,
			"motion_count": e.motionCount,
			"current_state": e.lastState,
			"last_change":   e.lastChange.Unix(),
			"initialized":   e.initialized,
		},
	}, nil
}

func (e *RCWL0516Executor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized && e.pin != nil {
		e.pin.Halt()
		e.initialized = false
	}
	return nil
}
