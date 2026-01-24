//go:build linux
// +build linux

package gpio

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"farhotech.com/iot-edge/pkg/node"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// VEML7700 I2C registers
const (
	veml7700DefaultAddr = 0x10

	veml7700RegALSConf    = 0x00 // ALS configuration
	veml7700RegALSWH      = 0x01 // ALS high threshold
	veml7700RegALSWL      = 0x02 // ALS low threshold
	veml7700RegPowerSave  = 0x03 // Power saving mode
	veml7700RegALS        = 0x04 // ALS data output
	veml7700RegWhite      = 0x05 // White channel output
	veml7700RegALSInt     = 0x06 // Interrupt status
)

// VEML7700 gain settings
const (
	veml7700Gain1   = 0x00 // 1x gain
	veml7700Gain2   = 0x01 // 2x gain
	veml7700Gain1_8 = 0x02 // 1/8 gain
	veml7700Gain1_4 = 0x03 // 1/4 gain
)

// VEML7700 integration time settings
const (
	veml7700IT25ms  = 0x0C // 25ms
	veml7700IT50ms  = 0x08 // 50ms
	veml7700IT100ms = 0x00 // 100ms (default)
	veml7700IT200ms = 0x01 // 200ms
	veml7700IT400ms = 0x02 // 400ms
	veml7700IT800ms = 0x03 // 800ms
)

// VEML7700Config holds configuration for VEML7700 node
type VEML7700Config struct {
	I2CBus          string  `json:"i2c_bus"`
	Address         uint16  `json:"address"`
	Gain            string  `json:"gain"`             // "1", "2", "1/8", "1/4"
	IntegrationTime string  `json:"integration_time"` // "25ms", "50ms", "100ms", "200ms", "400ms", "800ms"
	AutoGain        bool    `json:"auto_gain"`
	PollInterval    int     `json:"poll_interval_ms"`
}

// VEML7700Executor implements high accuracy ambient light sensor
type VEML7700Executor struct {
	config          VEML7700Config
	bus             i2c.BusCloser
	dev             i2c.Dev
	mu              sync.Mutex
	hostInited      bool
	initialized     bool
	gain            uint8
	integrationTime uint8
	resolution      float64 // lux per count
}

func init() {
	node.RegisterType("veml7700", func() node.Executor {
		return &VEML7700Executor{}
	})
}

func (e *VEML7700Executor) Init(config json.RawMessage) error {
	e.config = VEML7700Config{
		I2CBus:          "/dev/i2c-1",
		Address:         veml7700DefaultAddr,
		Gain:            "1",
		IntegrationTime: "100ms",
		AutoGain:        true,
		PollInterval:    1000,
	}

	if config != nil {
		if err := json.Unmarshal(config, &e.config); err != nil {
			return fmt.Errorf("failed to parse VEML7700 config: %w", err)
		}
	}

	// Parse gain setting
	switch e.config.Gain {
	case "1":
		e.gain = veml7700Gain1
	case "2":
		e.gain = veml7700Gain2
	case "1/8":
		e.gain = veml7700Gain1_8
	case "1/4":
		e.gain = veml7700Gain1_4
	default:
		e.gain = veml7700Gain1
	}

	// Parse integration time
	switch e.config.IntegrationTime {
	case "25ms":
		e.integrationTime = veml7700IT25ms
	case "50ms":
		e.integrationTime = veml7700IT50ms
	case "100ms":
		e.integrationTime = veml7700IT100ms
	case "200ms":
		e.integrationTime = veml7700IT200ms
	case "400ms":
		e.integrationTime = veml7700IT400ms
	case "800ms":
		e.integrationTime = veml7700IT800ms
	default:
		e.integrationTime = veml7700IT100ms
	}

	e.updateResolution()

	return nil
}

func (e *VEML7700Executor) updateResolution() {
	// Resolution in lux per count based on gain and integration time
	// Base resolution at gain=1, IT=100ms is 0.0576 lux/count
	baseResolution := 0.0576

	// Adjust for gain
	switch e.gain {
	case veml7700Gain2:
		baseResolution /= 2
	case veml7700Gain1_8:
		baseResolution *= 8
	case veml7700Gain1_4:
		baseResolution *= 4
	}

	// Adjust for integration time
	switch e.integrationTime {
	case veml7700IT25ms:
		baseResolution *= 4
	case veml7700IT50ms:
		baseResolution *= 2
	case veml7700IT200ms:
		baseResolution /= 2
	case veml7700IT400ms:
		baseResolution /= 4
	case veml7700IT800ms:
		baseResolution /= 8
	}

	e.resolution = baseResolution
}

func (e *VEML7700Executor) initHardware() error {
	if e.initialized {
		return nil
	}

	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return fmt.Errorf("failed to initialize periph host: %w", err)
		}
		e.hostInited = true
	}

	bus, err := i2creg.Open(e.config.I2CBus)
	if err != nil {
		return fmt.Errorf("failed to open I2C bus %s: %w", e.config.I2CBus, err)
	}
	e.bus = bus
	e.dev = i2c.Dev{Bus: bus, Addr: e.config.Address}

	// Configure the sensor
	if err := e.configure(); err != nil {
		e.bus.Close()
		return fmt.Errorf("failed to configure VEML7700: %w", err)
	}

	e.initialized = true
	return nil
}

func (e *VEML7700Executor) configure() error {
	// Build configuration word
	// Bits 12-11: Gain
	// Bits 9-6: Integration time
	// Bit 0: Shutdown (0 = enabled)
	config := uint16(e.gain)<<11 | uint16(e.integrationTime)<<6

	return e.writeRegister(veml7700RegALSConf, config)
}

func (e *VEML7700Executor) writeRegister(reg uint8, value uint16) error {
	cmd := []byte{reg, byte(value & 0xFF), byte(value >> 8)}
	return e.dev.Tx(cmd, nil)
}

func (e *VEML7700Executor) readRegister(reg uint8) (uint16, error) {
	write := []byte{reg}
	read := make([]byte, 2)
	if err := e.dev.Tx(write, read); err != nil {
		return 0, err
	}
	return uint16(read[0]) | uint16(read[1])<<8, nil
}

func (e *VEML7700Executor) readALS() (uint16, error) {
	return e.readRegister(veml7700RegALS)
}

func (e *VEML7700Executor) readWhite() (uint16, error) {
	return e.readRegister(veml7700RegWhite)
}

func (e *VEML7700Executor) calculateLux(rawALS uint16) float64 {
	lux := float64(rawALS) * e.resolution

	// Apply correction for high lux values (> 1000 lux)
	if lux > 1000 {
		lux = 6.0135e-13*lux*lux*lux*lux - 9.3924e-9*lux*lux*lux + 8.1488e-5*lux*lux + 1.0023*lux
	}

	return lux
}

func (e *VEML7700Executor) autoAdjustGain(rawALS uint16) bool {
	// If reading is too high, reduce sensitivity
	if rawALS > 10000 {
		if e.gain == veml7700Gain2 {
			e.gain = veml7700Gain1
			return true
		} else if e.gain == veml7700Gain1 {
			e.gain = veml7700Gain1_4
			return true
		} else if e.gain == veml7700Gain1_4 {
			e.gain = veml7700Gain1_8
			return true
		}
	}
	// If reading is too low, increase sensitivity
	if rawALS < 100 {
		if e.gain == veml7700Gain1_8 {
			e.gain = veml7700Gain1_4
			return true
		} else if e.gain == veml7700Gain1_4 {
			e.gain = veml7700Gain1
			return true
		} else if e.gain == veml7700Gain1 {
			e.gain = veml7700Gain2
			return true
		}
	}
	return false
}

func (e *VEML7700Executor) Execute(msg *node.Message) (*node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.initHardware(); err != nil {
		return nil, err
	}

	action := "read"
	if payload, ok := msg.Payload.(map[string]interface{}); ok {
		if a, ok := payload["action"].(string); ok {
			action = a
		}
	}

	switch action {
	case "read":
		return e.readLight()
	case "configure":
		return e.handleConfigure(msg)
	case "set_threshold":
		return e.handleSetThreshold(msg)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *VEML7700Executor) readLight() (*node.Message, error) {
	rawALS, err := e.readALS()
	if err != nil {
		return nil, fmt.Errorf("failed to read ALS: %w", err)
	}

	rawWhite, err := e.readWhite()
	if err != nil {
		return nil, fmt.Errorf("failed to read white channel: %w", err)
	}

	// Auto-adjust gain if enabled
	if e.config.AutoGain && e.autoAdjustGain(rawALS) {
		e.updateResolution()
		e.configure()
		// Wait for new reading
		time.Sleep(150 * time.Millisecond)
		rawALS, _ = e.readALS()
		rawWhite, _ = e.readWhite()
	}

	lux := e.calculateLux(rawALS)

	gainStr := "1"
	switch e.gain {
	case veml7700Gain2:
		gainStr = "2"
	case veml7700Gain1_8:
		gainStr = "1/8"
	case veml7700Gain1_4:
		gainStr = "1/4"
	}

	return &node.Message{
		Payload: map[string]interface{}{
			"lux":              lux,
			"raw_als":          rawALS,
			"raw_white":        rawWhite,
			"gain":             gainStr,
			"integration_time": e.config.IntegrationTime,
			"resolution":       e.resolution,
			"timestamp":        time.Now().Unix(),
		},
	}, nil
}

func (e *VEML7700Executor) handleConfigure(msg *node.Message) (*node.Message, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid payload type")
	}

	if gain, ok := payload["gain"].(string); ok {
		switch gain {
		case "1":
			e.gain = veml7700Gain1
		case "2":
			e.gain = veml7700Gain2
		case "1/8":
			e.gain = veml7700Gain1_8
		case "1/4":
			e.gain = veml7700Gain1_4
		}
	}

	if it, ok := payload["integration_time"].(string); ok {
		switch it {
		case "25ms":
			e.integrationTime = veml7700IT25ms
		case "50ms":
			e.integrationTime = veml7700IT50ms
		case "100ms":
			e.integrationTime = veml7700IT100ms
		case "200ms":
			e.integrationTime = veml7700IT200ms
		case "400ms":
			e.integrationTime = veml7700IT400ms
		case "800ms":
			e.integrationTime = veml7700IT800ms
		}
		e.config.IntegrationTime = it
	}

	if autoGain, ok := payload["auto_gain"].(bool); ok {
		e.config.AutoGain = autoGain
	}

	e.updateResolution()
	if err := e.configure(); err != nil {
		return nil, err
	}

	return &node.Message{
		Payload: map[string]interface{}{
			"status":  "configured",
			"message": "VEML7700 configuration updated",
		},
	}, nil
}

func (e *VEML7700Executor) handleSetThreshold(msg *node.Message) (*node.Message, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid payload type")
	}

	if high, ok := payload["high"].(float64); ok {
		if err := e.writeRegister(veml7700RegALSWH, uint16(high)); err != nil {
			return nil, fmt.Errorf("failed to set high threshold: %w", err)
		}
	}

	if low, ok := payload["low"].(float64); ok {
		if err := e.writeRegister(veml7700RegALSWL, uint16(low)); err != nil {
			return nil, fmt.Errorf("failed to set low threshold: %w", err)
		}
	}

	return &node.Message{
		Payload: map[string]interface{}{
			"status":  "threshold_set",
			"message": "Threshold values configured",
		},
	}, nil
}

func (e *VEML7700Executor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized && e.bus != nil {
		// Put sensor in shutdown mode
		e.writeRegister(veml7700RegALSConf, 0x0001)
		e.bus.Close()
		e.initialized = false
	}
	return nil
}
