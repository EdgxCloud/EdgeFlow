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
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// PCF8591 control byte bits
const (
	pcf8591DefaultAddr = 0x48

	// Control byte structure:
	// Bit 7: reserved (0)
	// Bit 6: DAC output enable
	// Bits 5-4: Analog input mode
	// Bit 3: Auto-increment flag
	// Bits 1-0: A/D channel number

	pcf8591DACEnable     = 0x40
	pcf8591AutoIncrement = 0x04

	// Input modes
	pcf8591ModeFourSingle = 0x00 // Four single-ended inputs
	pcf8591ModeThreeDiff  = 0x10 // Three differential inputs
	pcf8591ModeMixed      = 0x20 // Single-ended and differential mixed
	pcf8591ModeTwoDiff    = 0x30 // Two differential inputs
)

// PCF8591Config holds configuration for PCF8591 ADC/DAC
type PCF8591Config struct {
	I2CBus       string  `json:"i2c_bus"`
	Address      uint16  `json:"address"`
	InputMode    string  `json:"input_mode"` // "single", "diff3", "mixed", "diff2"
	DACEnabled   bool    `json:"dac_enabled"`
	VRef         float64 `json:"vref"`
	PollInterval int     `json:"poll_interval_ms"`
}

// PCF8591Executor implements 8-bit ADC/DAC
type PCF8591Executor struct {
	config      PCF8591Config
	bus         i2c.BusCloser
	dev         i2c.Dev
	mu          sync.Mutex
	hostInited  bool
	initialized bool
	inputMode   byte
	dacValue    byte
}

func (e *PCF8591Executor) Init(config map[string]interface{}) error {
	e.config = PCF8591Config{
		I2CBus:       "/dev/i2c-1",
		Address:      pcf8591DefaultAddr,
		InputMode:    "single",
		DACEnabled:   false,
		VRef:         3.3,
		PollInterval: 100,
	}

	if config != nil {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := json.Unmarshal(configJSON, &e.config); err != nil {
			return fmt.Errorf("failed to parse PCF8591 config: %w", err)
		}
	}

	// Parse input mode
	switch e.config.InputMode {
	case "single":
		e.inputMode = pcf8591ModeFourSingle
	case "diff3":
		e.inputMode = pcf8591ModeThreeDiff
	case "mixed":
		e.inputMode = pcf8591ModeMixed
	case "diff2":
		e.inputMode = pcf8591ModeTwoDiff
	default:
		e.inputMode = pcf8591ModeFourSingle
	}

	return nil
}

func (e *PCF8591Executor) initHardware() error {
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

	e.initialized = true
	return nil
}

func (e *PCF8591Executor) readChannel(channel int) (byte, float64, error) {
	if channel < 0 || channel > 3 {
		return 0, 0, fmt.Errorf("invalid channel: %d (must be 0-3)", channel)
	}

	// Build control byte
	controlByte := e.inputMode | byte(channel)
	if e.config.DACEnabled {
		controlByte |= pcf8591DACEnable
	}

	// Write control byte and DAC value, read ADC result
	write := []byte{controlByte}
	if e.config.DACEnabled {
		write = append(write, e.dacValue)
	}

	// Read two bytes - first is previous conversion, second is current
	read := make([]byte, 2)
	if err := e.dev.Tx(write, read); err != nil {
		return 0, 0, fmt.Errorf("failed to read channel %d: %w", channel, err)
	}

	// Second byte is the actual reading
	raw := read[1]
	voltage := (float64(raw) / 255.0) * e.config.VRef

	return raw, voltage, nil
}

func (e *PCF8591Executor) readAllChannels() ([]byte, []float64, error) {
	// Enable auto-increment to read all channels in one transaction
	controlByte := e.inputMode | pcf8591AutoIncrement
	if e.config.DACEnabled {
		controlByte |= pcf8591DACEnable
	}

	write := []byte{controlByte}
	if e.config.DACEnabled {
		write = append(write, e.dacValue)
	}

	// Read 5 bytes: 1 previous + 4 channels
	read := make([]byte, 5)
	if err := e.dev.Tx(write, read); err != nil {
		return nil, nil, fmt.Errorf("failed to read all channels: %w", err)
	}

	raw := read[1:5] // Skip first byte (previous conversion)
	voltages := make([]float64, 4)
	for i, r := range raw {
		voltages[i] = (float64(r) / 255.0) * e.config.VRef
	}

	return raw, voltages, nil
}

func (e *PCF8591Executor) writeDAC(value byte) error {
	if !e.config.DACEnabled {
		return fmt.Errorf("DAC is not enabled")
	}

	controlByte := e.inputMode | pcf8591DACEnable
	write := []byte{controlByte, value}

	if err := e.dev.Tx(write, nil); err != nil {
		return fmt.Errorf("failed to write DAC: %w", err)
	}

	e.dacValue = value
	return nil
}

func (e *PCF8591Executor) writeDACVoltage(voltage float64) error {
	if voltage < 0 || voltage > e.config.VRef {
		return fmt.Errorf("voltage out of range: %.3f (must be 0-%.3f)", voltage, e.config.VRef)
	}

	value := byte((voltage / e.config.VRef) * 255)
	return e.writeDAC(value)
}

func (e *PCF8591Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.initHardware(); err != nil {
		return node.Message{}, err
	}

	action := "read"
	channel := 0
	if payload := msg.Payload; payload != nil {
		if a, ok := payload["action"].(string); ok {
			action = a
		}
		if c, ok := payload["channel"].(float64); ok {
			channel = int(c)
		}
	}

	switch action {
	case "read":
		return e.handleReadChannel(channel)
	case "read_all":
		return e.handleReadAll()
	case "write_dac":
		return e.handleWriteDAC(msg)
	case "configure":
		return e.handleConfigure(msg)
	case "status":
		return e.handleStatus()
	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *PCF8591Executor) handleReadChannel(channel int) (node.Message, error) {
	raw, voltage, err := e.readChannel(channel)
	if err != nil {
		return node.Message{}, err
	}

	percentage := (voltage / e.config.VRef) * 100

	return node.Message{
		Payload: map[string]interface{}{
			"channel":    channel,
			"raw":        raw,
			"voltage":    voltage,
			"percentage": percentage,
			"vref":       e.config.VRef,
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

func (e *PCF8591Executor) handleReadAll() (node.Message, error) {
	raw, voltages, err := e.readAllChannels()
	if err != nil {
		return node.Message{}, err
	}

	channels := make([]map[string]interface{}, 4)
	for i := 0; i < 4; i++ {
		channels[i] = map[string]interface{}{
			"channel":    i,
			"raw":        raw[i],
			"voltage":    voltages[i],
			"percentage": (voltages[i] / e.config.VRef) * 100,
		}
	}

	return node.Message{
		Payload: map[string]interface{}{
			"channels":   channels,
			"vref":       e.config.VRef,
			"input_mode": e.config.InputMode,
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

func (e *PCF8591Executor) handleWriteDAC(msg node.Message) (node.Message, error) {
	if !e.config.DACEnabled {
		return node.Message{}, fmt.Errorf("DAC is not enabled in configuration")
	}

	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	var err error
	var dacValue byte
	var dacVoltage float64

	if v, ok := payload["value"].(float64); ok {
		// Raw byte value (0-255)
		dacValue = byte(v)
		err = e.writeDAC(dacValue)
		dacVoltage = (float64(dacValue) / 255.0) * e.config.VRef
	} else if v, ok := payload["voltage"].(float64); ok {
		// Voltage value
		dacVoltage = v
		dacValue = byte((v / e.config.VRef) * 255)
		err = e.writeDACVoltage(v)
	} else if p, ok := payload["percentage"].(float64); ok {
		// Percentage value
		dacVoltage = (p / 100.0) * e.config.VRef
		dacValue = byte((p / 100.0) * 255)
		err = e.writeDACVoltage(dacVoltage)
	} else {
		return node.Message{}, fmt.Errorf("value, voltage, or percentage required")
	}

	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":     "dac_set",
			"raw":        dacValue,
			"voltage":    dacVoltage,
			"percentage": (float64(dacValue) / 255.0) * 100,
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

func (e *PCF8591Executor) handleConfigure(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	if mode, ok := payload["input_mode"].(string); ok {
		e.config.InputMode = mode
		switch mode {
		case "single":
			e.inputMode = pcf8591ModeFourSingle
		case "diff3":
			e.inputMode = pcf8591ModeThreeDiff
		case "mixed":
			e.inputMode = pcf8591ModeMixed
		case "diff2":
			e.inputMode = pcf8591ModeTwoDiff
		}
	}

	if dacEnabled, ok := payload["dac_enabled"].(bool); ok {
		e.config.DACEnabled = dacEnabled
	}

	if vref, ok := payload["vref"].(float64); ok {
		e.config.VRef = vref
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":      "configured",
			"input_mode":  e.config.InputMode,
			"dac_enabled": e.config.DACEnabled,
			"vref":        e.config.VRef,
		},
	}, nil
}

func (e *PCF8591Executor) handleStatus() (node.Message, error) {
	return node.Message{
		Payload: map[string]interface{}{
			"address":     fmt.Sprintf("0x%02X", e.config.Address),
			"input_mode":  e.config.InputMode,
			"dac_enabled": e.config.DACEnabled,
			"dac_value":   e.dacValue,
			"dac_voltage": (float64(e.dacValue) / 255.0) * e.config.VRef,
			"vref":        e.config.VRef,
			"initialized": e.initialized,
		},
	}, nil
}

func (e *PCF8591Executor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized && e.bus != nil {
		// Set DAC to 0 before closing
		if e.config.DACEnabled {
			e.writeDAC(0)
		}
		e.bus.Close()
		e.initialized = false
	}
	return nil
}
