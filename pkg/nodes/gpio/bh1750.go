package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/hal"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// BH1750 I2C commands
const (
	bh1750PowerDown          = 0x00
	bh1750PowerOn            = 0x01
	bh1750Reset              = 0x07
	bh1750ContHighRes        = 0x10 // Continuous high resolution mode (1 lux, 120ms)
	bh1750ContHighRes2       = 0x11 // Continuous high resolution mode 2 (0.5 lux, 120ms)
	bh1750ContLowRes         = 0x13 // Continuous low resolution mode (4 lux, 16ms)
	bh1750OneTimeHighRes     = 0x20 // One-time high resolution mode
	bh1750OneTimeHighRes2    = 0x21 // One-time high resolution mode 2
	bh1750OneTimeLowRes      = 0x23 // One-time low resolution mode
)

// BH1750Config configuration for BH1750 light sensor
type BH1750Config struct {
	Bus        string  `json:"bus"`        // I2C bus (default: "")
	Address    int     `json:"address"`    // I2C address (default: 0x23)
	Mode       string  `json:"mode"`       // Mode: "high", "high2", "low" (default: "high")
	Continuous bool    `json:"continuous"` // Continuous mode (default: true)
	MTreg      int     `json:"mtreg"`      // Measurement time register (default: 69)
	Offset     float64 `json:"offset"`     // Calibration offset
	Scale      float64 `json:"scale"`      // Calibration scale factor (default: 1.0)
}

// BH1750Executor executes BH1750 light sensor readings
type BH1750Executor struct {
	config     BH1750Config
	hal        hal.HAL
	bus        i2c.BusCloser
	dev        i2c.Dev
	lastRead   time.Time
	lastLux    float64
	mu         sync.Mutex
	hostInited bool
	inited     bool
}

// NewBH1750Executor creates a new BH1750 executor
func NewBH1750Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var bhConfig BH1750Config
	if err := json.Unmarshal(configJSON, &bhConfig); err != nil {
		return nil, fmt.Errorf("invalid BH1750 config: %w", err)
	}

	// Default address (ADDR pin LOW = 0x23, ADDR pin HIGH = 0x5C)
	if bhConfig.Address == 0 {
		bhConfig.Address = 0x23
	}

	// Validate address
	if bhConfig.Address != 0x23 && bhConfig.Address != 0x5C {
		return nil, fmt.Errorf("invalid I2C address: 0x%02X (must be 0x23 or 0x5C)", bhConfig.Address)
	}

	// Default mode
	if bhConfig.Mode == "" {
		bhConfig.Mode = "high"
	}

	// Default continuous mode
	if !bhConfig.Continuous {
		bhConfig.Continuous = true
	}

	// Default measurement time register (69 is default in datasheet)
	if bhConfig.MTreg == 0 {
		bhConfig.MTreg = 69
	}

	// Clamp MTreg to valid range (31-254)
	if bhConfig.MTreg < 31 {
		bhConfig.MTreg = 31
	}
	if bhConfig.MTreg > 254 {
		bhConfig.MTreg = 254
	}

	// Default scale
	if bhConfig.Scale == 0 {
		bhConfig.Scale = 1.0
	}

	return &BH1750Executor{
		config: bhConfig,
	}, nil
}

// Init initializes the BH1750 executor
func (e *BH1750Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute reads the BH1750 sensor
func (e *BH1750Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Initialize hardware if needed
	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init periph host: %w", err)
		}
		e.hostInited = true
	}

	// Open I2C bus if not already open
	if e.bus == nil {
		busName := e.config.Bus
		if busName == "" {
			busName = "" // Let periph.io find the default bus
		}

		bus, err := i2creg.Open(busName)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to open I2C bus: %w", err)
		}
		e.bus = bus
		e.dev = i2c.Dev{Bus: e.bus, Addr: uint16(e.config.Address)}
	}

	// Initialize sensor if needed
	if !e.inited {
		if err := e.initSensor(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init BH1750: %w", err)
		}
		e.inited = true
	}

	// Read light level
	lux, err := e.readLux()
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read BH1750: %w", err)
	}

	// Apply calibration
	lux = lux*e.config.Scale + e.config.Offset

	// Clamp to valid range
	if lux < 0 {
		lux = 0
	}

	e.lastRead = time.Now()
	e.lastLux = lux

	// Determine light level description
	lightLevel := getLightLevel(lux)

	return node.Message{
		Payload: map[string]interface{}{
			"lux":         lux,
			"light_level": lightLevel,
			"unit":        "lx",
			"mode":        e.config.Mode,
			"address":     fmt.Sprintf("0x%02X", e.config.Address),
			"sensor":      "BH1750",
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

// initSensor initializes the BH1750 sensor
func (e *BH1750Executor) initSensor() error {
	// Power on
	if _, err := e.dev.Write([]byte{bh1750PowerOn}); err != nil {
		return fmt.Errorf("power on failed: %w", err)
	}

	// Reset
	if _, err := e.dev.Write([]byte{bh1750Reset}); err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}

	// Set measurement time if not default
	if e.config.MTreg != 69 {
		// MTreg is split into high bits (01000_MT[7:5]) and low bits (011_MT[4:0])
		highBits := 0x40 | (e.config.MTreg >> 5)
		lowBits := 0x60 | (e.config.MTreg & 0x1F)

		if _, err := e.dev.Write([]byte{byte(highBits)}); err != nil {
			return fmt.Errorf("set MTreg high failed: %w", err)
		}
		if _, err := e.dev.Write([]byte{byte(lowBits)}); err != nil {
			return fmt.Errorf("set MTreg low failed: %w", err)
		}
	}

	// Set mode
	mode := e.getModeCommand()
	if _, err := e.dev.Write([]byte{mode}); err != nil {
		return fmt.Errorf("set mode failed: %w", err)
	}

	// Wait for first measurement
	time.Sleep(e.getMeasurementTime())

	return nil
}

// getModeCommand returns the I2C command for the configured mode
func (e *BH1750Executor) getModeCommand() byte {
	if e.config.Continuous {
		switch e.config.Mode {
		case "high":
			return bh1750ContHighRes
		case "high2":
			return bh1750ContHighRes2
		case "low":
			return bh1750ContLowRes
		default:
			return bh1750ContHighRes
		}
	} else {
		switch e.config.Mode {
		case "high":
			return bh1750OneTimeHighRes
		case "high2":
			return bh1750OneTimeHighRes2
		case "low":
			return bh1750OneTimeLowRes
		default:
			return bh1750OneTimeHighRes
		}
	}
}

// getMeasurementTime returns the measurement time for the configured mode
func (e *BH1750Executor) getMeasurementTime() time.Duration {
	// Base times (with default MTreg=69)
	baseTime := 120 * time.Millisecond // High resolution
	if e.config.Mode == "low" {
		baseTime = 16 * time.Millisecond
	}

	// Scale by MTreg
	scaledTime := time.Duration(float64(baseTime) * float64(e.config.MTreg) / 69.0)
	return scaledTime + 10*time.Millisecond // Add safety margin
}

// readLux reads the light level in lux
func (e *BH1750Executor) readLux() (float64, error) {
	// In one-time mode, trigger new measurement
	if !e.config.Continuous {
		mode := e.getModeCommand()
		if _, err := e.dev.Write([]byte{mode}); err != nil {
			return 0, fmt.Errorf("trigger measurement failed: %w", err)
		}
		time.Sleep(e.getMeasurementTime())
	}

	// Read 2 bytes
	data := make([]byte, 2)
	if err := e.dev.Tx(nil, data); err != nil {
		return 0, fmt.Errorf("read failed: %w", err)
	}

	// Convert to lux
	// Raw value / 1.2 for standard mode
	// Raw value / 1.2 / 2 for high res mode 2
	raw := float64(uint16(data[0])<<8 | uint16(data[1]))
	lux := raw / 1.2

	// Adjust for MTreg
	lux = lux * (69.0 / float64(e.config.MTreg))

	// High resolution mode 2 has 0.5 lux resolution
	if e.config.Mode == "high2" {
		lux /= 2.0
	}

	return lux, nil
}

// getLightLevel returns a human-readable description of the light level
func getLightLevel(lux float64) string {
	switch {
	case lux < 1:
		return "pitch_black"
	case lux < 10:
		return "very_dark"
	case lux < 50:
		return "dark"
	case lux < 200:
		return "dim"
	case lux < 400:
		return "normal_indoor"
	case lux < 1000:
		return "bright_indoor"
	case lux < 10000:
		return "overcast"
	case lux < 25000:
		return "daylight"
	case lux < 50000:
		return "bright_daylight"
	default:
		return "direct_sunlight"
	}
}

// Cleanup releases resources
func (e *BH1750Executor) Cleanup() error {
	if e.bus != nil {
		// Power down the sensor
		e.dev.Write([]byte{bh1750PowerDown})
		e.bus.Close()
		e.bus = nil
	}
	e.inited = false
	return nil
}
