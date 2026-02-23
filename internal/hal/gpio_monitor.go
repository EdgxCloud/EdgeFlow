package hal

import (
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/logger"
	"go.uber.org/zap"
)

// PinState represents the live state of a single GPIO pin
type PinState struct {
	BCMPin    int       `json:"bcm_pin"`
	Value     bool      `json:"value"`
	Mode      string    `json:"mode"`
	EdgeCount uint64    `json:"edge_count"`
	LastChange time.Time `json:"last_change"`
}

// GPIOMonitorState is the complete GPIO state for broadcasting
type GPIOMonitorState struct {
	Pins      map[int]*PinState `json:"pins"`
	BoardName string            `json:"board_name"`
	GPIOChip  string            `json:"gpio_chip"`
	Available bool              `json:"available"`
	Timestamp time.Time         `json:"timestamp"`
}

// GPIOMonitor tracks active GPIO pins and broadcasts state changes
type GPIOMonitor struct {
	mu          sync.RWMutex
	pins        map[int]*PinState
	prevValues  map[int]bool
	broadcaster func(GPIOMonitorState)
	stopChan    chan struct{}
	pollMs      int
	boardName   string
	gpioChip    string
	pollCount   int // tracks polls for periodic forced broadcast
}

// NewGPIOMonitor creates a new GPIO monitor
func NewGPIOMonitor(pollMs int, broadcaster func(GPIOMonitorState)) *GPIOMonitor {
	boardName := "Unknown"
	gpioChip := ""

	h, err := GetGlobalHAL()
	if err == nil {
		info := h.Info()
		boardName = info.Name
		gpioChip = info.GPIOChip
	}

	return &GPIOMonitor{
		pins:        make(map[int]*PinState),
		prevValues:  make(map[int]bool),
		broadcaster: broadcaster,
		stopChan:    make(chan struct{}),
		pollMs:      pollMs,
		boardName:   boardName,
		gpioChip:    gpioChip,
	}
}

// Start begins the background polling loop
func (m *GPIOMonitor) Start() {
	ticker := time.NewTicker(time.Duration(m.pollMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.poll()
		}
	}
}

// Stop stops the background polling loop
func (m *GPIOMonitor) Stop() {
	close(m.stopChan)
}

// GetState returns the current GPIO state snapshot
func (m *GPIOMonitor) GetState() GPIOMonitorState {
	h, err := GetGlobalHAL()
	if err != nil {
		return GPIOMonitorState{
			Pins:      make(map[int]*PinState),
			BoardName: m.boardName,
			GPIOChip:  m.gpioChip,
			Available: false,
			Timestamp: time.Now(),
		}
	}

	// Read current active pins directly from HAL
	gpio := h.GPIO()
	activePins := gpio.ActivePins()
	now := time.Now()

	m.mu.RLock()
	defer m.mu.RUnlock()

	pins := make(map[int]*PinState, len(activePins))
	for pin, mode := range activePins {
		if state, ok := m.pins[pin]; ok {
			// Use cached state (has edge count, last change, etc.)
			pinCopy := *state
			// But read the current value live
			if val, readErr := gpio.DigitalRead(pin); readErr == nil {
				pinCopy.Value = val
			}
			pins[pin] = &pinCopy
		} else {
			// Pin exists in HAL but not yet in monitor cache — read live
			modeStr := "input"
			switch mode {
			case Output:
				modeStr = "output"
			case PWM:
				modeStr = "pwm"
			}
			val, _ := gpio.DigitalRead(pin)
			pins[pin] = &PinState{
				BCMPin:     pin,
				Value:      val,
				Mode:       modeStr,
				EdgeCount:  0,
				LastChange: now,
			}
		}
	}

	return GPIOMonitorState{
		Pins:      pins,
		BoardName: m.boardName,
		GPIOChip:  m.gpioChip,
		Available: true,
		Timestamp: now,
	}
}

// poll reads all active GPIO pins and broadcasts state
func (m *GPIOMonitor) poll() {
	h, err := GetGlobalHAL()
	if err != nil {
		return
	}

	gpio := h.GPIO()
	if gpio == nil {
		return
	}

	activePins := gpio.ActivePins()
	if len(activePins) == 0 {
		// Clean up old pins if no active pins
		m.mu.Lock()
		hadPins := len(m.pins) > 0
		if hadPins {
			m.pins = make(map[int]*PinState)
			m.prevValues = make(map[int]bool)
		}
		m.pollCount++
		m.mu.Unlock()

		if hadPins {
			// Broadcast empty state when pins go away
			m.broadcaster(GPIOMonitorState{
				Pins:      make(map[int]*PinState),
				BoardName: m.boardName,
				GPIOChip:  m.gpioChip,
				Available: true,
				Timestamp: time.Now(),
			})
		}
		return
	}

	changed := false
	now := time.Now()

	m.mu.Lock()

	// Remove pins that are no longer active
	for pin := range m.pins {
		if _, exists := activePins[pin]; !exists {
			delete(m.pins, pin)
			delete(m.prevValues, pin)
			changed = true
		}
	}

	// Read current values for all active pins
	for pin, mode := range activePins {
		value, err := gpio.DigitalRead(pin)
		if err != nil {
			logger.Warn("GPIO monitor: failed to read pin", zap.Int("pin", pin), zap.Error(err))
			continue
		}

		modeStr := "input"
		switch mode {
		case Output:
			modeStr = "output"
		case PWM:
			modeStr = "pwm"
		}

		state, exists := m.pins[pin]
		if !exists {
			// New pin discovered
			m.pins[pin] = &PinState{
				BCMPin:     pin,
				Value:      value,
				Mode:       modeStr,
				EdgeCount:  0,
				LastChange: now,
			}
			m.prevValues[pin] = value
			changed = true
		} else {
			// Update existing pin — always update current value
			state.Mode = modeStr
			state.Value = value
			if value != m.prevValues[pin] {
				state.EdgeCount++
				state.LastChange = now
				m.prevValues[pin] = value
				changed = true
			}
		}
	}

	m.pollCount++

	// Broadcast on state change OR every 5 polls (~1s) as a heartbeat
	shouldBroadcast := changed || m.pollCount%5 == 0

	var state GPIOMonitorState
	if shouldBroadcast {
		pins := make(map[int]*PinState, len(m.pins))
		for pin, s := range m.pins {
			pinCopy := *s
			pins[pin] = &pinCopy
		}
		state = GPIOMonitorState{
			Pins:      pins,
			BoardName: m.boardName,
			GPIOChip:  m.gpioChip,
			Available: true,
			Timestamp: now,
		}
	}

	m.mu.Unlock()

	if shouldBroadcast && m.broadcaster != nil {
		m.broadcaster(state)
	}

	if changed {
		logger.Debug("GPIO monitor: state changed", zap.Int("active_pins", len(m.pins)))
	}
}

// Global GPIO monitor singleton
var (
	globalGPIOMonitor *GPIOMonitor
	gpioMonitorMu     sync.RWMutex
)

// SetGlobalGPIOMonitor sets the global GPIO monitor
func SetGlobalGPIOMonitor(m *GPIOMonitor) {
	gpioMonitorMu.Lock()
	defer gpioMonitorMu.Unlock()
	globalGPIOMonitor = m
}

// GetGlobalGPIOMonitor returns the global GPIO monitor
func GetGlobalGPIOMonitor() *GPIOMonitor {
	gpioMonitorMu.RLock()
	defer gpioMonitorMu.RUnlock()
	return globalGPIOMonitor
}
