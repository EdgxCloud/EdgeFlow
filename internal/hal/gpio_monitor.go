package hal

import (
	"log"
	"sync"
	"time"
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

	// Read current active pins
	gpio := h.GPIO()
	activePins := gpio.ActivePins()

	m.mu.RLock()
	defer m.mu.RUnlock()

	pins := make(map[int]*PinState, len(activePins))
	for pin := range activePins {
		if state, ok := m.pins[pin]; ok {
			pinCopy := *state
			pins[pin] = &pinCopy
		}
	}

	return GPIOMonitorState{
		Pins:      pins,
		BoardName: m.boardName,
		GPIOChip:  m.gpioChip,
		Available: true,
		Timestamp: time.Now(),
	}
}

// poll reads all active GPIO pins and broadcasts if any changed
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
		if len(m.pins) > 0 {
			m.pins = make(map[int]*PinState)
			m.prevValues = make(map[int]bool)
			m.mu.Unlock()
			// Broadcast empty state
			m.broadcaster(GPIOMonitorState{
				Pins:      make(map[int]*PinState),
				BoardName: m.boardName,
				GPIOChip:  m.gpioChip,
				Available: true,
				Timestamp: time.Now(),
			})
			return
		}
		m.mu.Unlock()
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
			log.Printf("GPIO monitor: failed to read pin %d: %v", pin, err)
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
			// Update existing pin
			state.Mode = modeStr
			if value != m.prevValues[pin] {
				state.Value = value
				state.EdgeCount++
				state.LastChange = now
				m.prevValues[pin] = value
				changed = true
			}
		}
	}

	// Build state snapshot while still holding the lock
	var state GPIOMonitorState
	if changed {
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

	// Broadcast outside the lock
	if changed && m.broadcaster != nil {
		m.broadcaster(state)
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
