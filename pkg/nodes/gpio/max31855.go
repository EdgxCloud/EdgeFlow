package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/hal"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// MAX31855 Fault bits
const (
	max31855FaultOC    = 0x01 // Open circuit
	max31855FaultSCG   = 0x02 // Short to GND
	max31855FaultSCV   = 0x04 // Short to VCC
	max31855FaultAny   = 0x10000
)

// MAX31855Config configuration for MAX31855 thermocouple interface
type MAX31855Config struct {
	SPIBus     int     `json:"spi_bus"`     // SPI bus number (default: 0)
	SPIDevice  int     `json:"spi_device"`  // SPI device number (default: 0)
	Speed      int     `json:"speed"`       // SPI speed in Hz (default: 5MHz)
	CSPin      int     `json:"cs_pin"`      // Chip select GPIO pin (optional, for software CS)
	TempOffset float64 `json:"temp_offset"` // Temperature calibration offset
}

// MAX31855Executor executes MAX31855 thermocouple readings
type MAX31855Executor struct {
	config      MAX31855Config
	hal         hal.HAL
	mu          sync.Mutex
	initialized bool
}

// NewMAX31855Executor creates a new MAX31855 executor
func NewMAX31855Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var tcConfig MAX31855Config
	if err := json.Unmarshal(configJSON, &tcConfig); err != nil {
		return nil, fmt.Errorf("invalid MAX31855 config: %w", err)
	}

	// Defaults
	if tcConfig.Speed == 0 {
		tcConfig.Speed = 5000000 // 5MHz max
	}

	return &MAX31855Executor{
		config: tcConfig,
	}, nil
}

// Init initializes the MAX31855 executor
func (e *MAX31855Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute reads the MAX31855 thermocouple
func (e *MAX31855Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get HAL
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h
	}

	// Initialize CS pin if specified
	if !e.initialized && e.config.CSPin > 0 {
		gpio := e.hal.GPIO()
		gpio.SetMode(e.config.CSPin, hal.Output)
		gpio.DigitalWrite(e.config.CSPin, true) // CS high (deselected)
		e.initialized = true
	}

	// Parse command
	payload := msg.Payload
	if payload == nil {
		// Default: read temperature
		return e.readTemperature()
	}

	action, _ := payload["action"].(string)

	switch action {
	case "read", "":
		return e.readTemperature()

	case "raw":
		return e.readRaw()

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// readTemperature reads the thermocouple temperature
func (e *MAX31855Executor) readTemperature() (node.Message, error) {
	raw, err := e.readRawData()
	if err != nil {
		return node.Message{}, err
	}

	// Check for fault
	fault := raw&max31855FaultAny != 0
	var faultType string
	if fault {
		if raw&max31855FaultOC != 0 {
			faultType = "open_circuit"
		} else if raw&max31855FaultSCG != 0 {
			faultType = "short_to_gnd"
		} else if raw&max31855FaultSCV != 0 {
			faultType = "short_to_vcc"
		}
	}

	// Extract thermocouple temperature (bits 31-18)
	// 14-bit signed value, 0.25°C resolution
	tcRaw := int32(raw >> 18)
	if tcRaw&0x2000 != 0 { // Sign extend if negative
		tcRaw |= ^int32(0x3FFF)
	}
	tcTempC := float64(tcRaw) * 0.25
	tcTempC += e.config.TempOffset

	// Extract internal (cold junction) temperature (bits 15-4)
	// 12-bit signed value, 0.0625°C resolution
	intRaw := int32((raw >> 4) & 0x0FFF)
	if intRaw&0x0800 != 0 { // Sign extend if negative
		intRaw |= ^int32(0x0FFF)
	}
	intTempC := float64(intRaw) * 0.0625

	// Convert to Fahrenheit
	tcTempF := tcTempC*9/5 + 32
	intTempF := intTempC*9/5 + 32

	return node.Message{
		Payload: map[string]interface{}{
			"thermocouple_c":   tcTempC,
			"thermocouple_f":   tcTempF,
			"internal_c":       intTempC,
			"internal_f":       intTempF,
			"fault":            fault,
			"fault_type":       faultType,
			"valid":            !fault,
			"sensor":           "MAX31855",
			"timestamp":        time.Now().Unix(),
		},
	}, nil
}

// readRaw returns the raw 32-bit value
func (e *MAX31855Executor) readRaw() (node.Message, error) {
	raw, err := e.readRawData()
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"raw":       raw,
			"raw_hex":   fmt.Sprintf("0x%08X", raw),
			"sensor":    "MAX31855",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// readRawData reads the 32-bit raw data from MAX31855
func (e *MAX31855Executor) readRawData() (uint32, error) {
	spi := e.hal.SPI()
	gpio := e.hal.GPIO()

	// Open SPI device
	if err := spi.Open(e.config.SPIBus, e.config.SPIDevice); err != nil {
		return 0, fmt.Errorf("failed to open SPI device: %w", err)
	}
	if err := spi.SetSpeed(e.config.Speed); err != nil {
		return 0, fmt.Errorf("failed to set SPI speed: %w", err)
	}

	// Assert CS if using software control
	if e.config.CSPin > 0 {
		gpio.DigitalWrite(e.config.CSPin, false)
		time.Sleep(1 * time.Microsecond)
	}

	// Read 4 bytes (send zeros to clock out data)
	writeData := make([]byte, 4)
	data, err := spi.Transfer(writeData)
	if err != nil {
		if e.config.CSPin > 0 {
			gpio.DigitalWrite(e.config.CSPin, true)
		}
		return 0, fmt.Errorf("SPI transfer failed: %w", err)
	}

	// Deassert CS
	if e.config.CSPin > 0 {
		gpio.DigitalWrite(e.config.CSPin, true)
	}

	// Combine bytes (MSB first)
	raw := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])

	return raw, nil
}

// Cleanup releases resources
func (e *MAX31855Executor) Cleanup() error {
	return nil
}
