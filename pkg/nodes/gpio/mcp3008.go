package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/hal"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

// MCP3008Config configuration for MCP3008 ADC
type MCP3008Config struct {
	Bus       string    `json:"bus"`        // SPI bus (default: "")
	ChipSelect int      `json:"cs"`         // Chip select (0 or 1, default: 0)
	Speed     int       `json:"speed"`      // SPI speed in Hz (default: 1000000)
	Channel   int       `json:"channel"`    // ADC channel 0-7 (default: 0)
	Mode      string    `json:"mode"`       // "single" or "diff" (default: "single")
	VRef      float64   `json:"vref"`       // Reference voltage (default: 3.3)
	Samples   int       `json:"samples"`    // Number of samples to average (default: 1)
	Scale     float64   `json:"scale"`      // Output scale factor (default: 1.0)
	Offset    float64   `json:"offset"`     // Output offset (default: 0.0)
	ReadAll   bool      `json:"read_all"`   // Read all 8 channels (default: false)
}

// MCP3008Executor executes MCP3008 ADC readings
type MCP3008Executor struct {
	config     MCP3008Config
	hal        hal.HAL
	port       spi.PortCloser
	conn       spi.Conn
	mu         sync.Mutex
	hostInited bool
}

// NewMCP3008Executor creates a new MCP3008 executor
func NewMCP3008Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var mcpConfig MCP3008Config
	if err := json.Unmarshal(configJSON, &mcpConfig); err != nil {
		return nil, fmt.Errorf("invalid MCP3008 config: %w", err)
	}

	// Defaults
	if mcpConfig.Speed == 0 {
		mcpConfig.Speed = 1000000 // 1 MHz (max 3.6 MHz at 5V, 1.35 MHz at 2.7V)
	}

	if mcpConfig.VRef == 0 {
		mcpConfig.VRef = 3.3 // Raspberry Pi default
	}

	if mcpConfig.Samples == 0 {
		mcpConfig.Samples = 1
	}

	if mcpConfig.Scale == 0 {
		mcpConfig.Scale = 1.0
	}

	if mcpConfig.Mode == "" {
		mcpConfig.Mode = "single"
	}

	// Validate channel
	if mcpConfig.Channel < 0 || mcpConfig.Channel > 7 {
		return nil, fmt.Errorf("invalid channel: %d (must be 0-7)", mcpConfig.Channel)
	}

	// Validate mode
	if mcpConfig.Mode != "single" && mcpConfig.Mode != "diff" {
		return nil, fmt.Errorf("invalid mode: %s (must be 'single' or 'diff')", mcpConfig.Mode)
	}

	return &MCP3008Executor{
		config: mcpConfig,
	}, nil
}

// Init initializes the MCP3008 executor
func (e *MCP3008Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute reads the MCP3008 ADC
func (e *MCP3008Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Initialize hardware if needed
	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init periph host: %w", err)
		}
		e.hostInited = true
	}

	// Open SPI port if not already open
	if e.port == nil {
		busName := e.config.Bus
		if busName == "" {
			// Default SPI bus on Raspberry Pi
			busName = fmt.Sprintf("/dev/spidev0.%d", e.config.ChipSelect)
		}

		port, err := spireg.Open(busName)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to open SPI bus: %w", err)
		}
		e.port = port

		// Connect with settings
		conn, err := port.Connect(physic.Frequency(e.config.Speed)*physic.Hertz, spi.Mode0, 8)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to connect SPI: %w", err)
		}
		e.conn = conn
	}

	// Read channel(s)
	if e.config.ReadAll {
		// Read all 8 channels
		channels := make([]map[string]interface{}, 8)
		for ch := 0; ch < 8; ch++ {
			raw, voltage, err := e.readChannel(ch)
			if err != nil {
				return node.Message{}, fmt.Errorf("failed to read channel %d: %w", ch, err)
			}

			// Apply scale and offset
			scaledValue := voltage*e.config.Scale + e.config.Offset

			channels[ch] = map[string]interface{}{
				"channel": ch,
				"raw":     raw,
				"voltage": voltage,
				"value":   scaledValue,
			}
		}

		return node.Message{
			Payload: map[string]interface{}{
				"channels":  channels,
				"vref":      e.config.VRef,
				"mode":      e.config.Mode,
				"sensor":    "MCP3008",
				"timestamp": time.Now().Unix(),
			},
		}, nil
	}

	// Read single channel with averaging
	var totalRaw int
	for i := 0; i < e.config.Samples; i++ {
		raw, _, err := e.readChannel(e.config.Channel)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to read channel: %w", err)
		}
		totalRaw += raw

		if i < e.config.Samples-1 {
			time.Sleep(1 * time.Millisecond)
		}
	}

	avgRaw := totalRaw / e.config.Samples
	voltage := float64(avgRaw) / 1023.0 * e.config.VRef
	scaledValue := voltage*e.config.Scale + e.config.Offset

	// Calculate percentage
	percent := float64(avgRaw) / 1023.0 * 100.0

	return node.Message{
		Payload: map[string]interface{}{
			"channel":   e.config.Channel,
			"raw":       avgRaw,
			"voltage":   voltage,
			"value":     scaledValue,
			"percent":   percent,
			"vref":      e.config.VRef,
			"mode":      e.config.Mode,
			"samples":   e.config.Samples,
			"sensor":    "MCP3008",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// readChannel reads a single ADC channel
func (e *MCP3008Executor) readChannel(channel int) (int, float64, error) {
	// MCP3008 SPI protocol:
	// Byte 1: Start bit (0x01)
	// Byte 2: Single/Diff mode (bit 7) + Channel (bits 6-4)
	//         Single-ended: 1xxx xxxx (0x80 | (channel << 4))
	//         Differential: 0xxx xxxx (channel << 4)
	// Byte 3: Don't care
	//
	// Response:
	// Byte 1: Don't care
	// Byte 2: Null bit + 2 MSBs of result
	// Byte 3: 8 LSBs of result

	var configByte byte
	if e.config.Mode == "single" {
		configByte = 0x80 | byte(channel<<4)
	} else {
		configByte = byte(channel << 4)
	}

	tx := []byte{0x01, configByte, 0x00}
	rx := make([]byte, 3)

	if err := e.conn.Tx(tx, rx); err != nil {
		return 0, 0, fmt.Errorf("SPI transfer failed: %w", err)
	}

	// Extract 10-bit result
	// rx[1] contains 2 MSBs in lower 2 bits
	// rx[2] contains 8 LSBs
	raw := int(rx[1]&0x03)<<8 | int(rx[2])

	// Convert to voltage
	voltage := float64(raw) / 1023.0 * e.config.VRef

	return raw, voltage, nil
}

// Cleanup releases resources
func (e *MCP3008Executor) Cleanup() error {
	if e.port != nil {
		e.port.Close()
		e.port = nil
	}
	return nil
}

// MCP3208Config configuration for MCP3208 (12-bit version)
type MCP3208Config struct {
	MCP3008Config
}

// MCP3208Executor is similar to MCP3008 but with 12-bit resolution
type MCP3208Executor struct {
	*MCP3008Executor
}

// NewMCP3208Executor creates a new MCP3208 executor
func NewMCP3208Executor(config map[string]interface{}) (node.Executor, error) {
	exec, err := NewMCP3008Executor(config)
	if err != nil {
		return nil, err
	}

	// Convert to MCP3208
	mcpExec := exec.(*MCP3008Executor)
	return &MCP3208Executor{
		MCP3008Executor: mcpExec,
	}, nil
}

// Execute reads the MCP3208 ADC (12-bit)
func (e *MCP3208Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Initialize hardware if needed
	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init periph host: %w", err)
		}
		e.hostInited = true
	}

	// Open SPI port if not already open
	if e.port == nil {
		busName := e.config.Bus
		if busName == "" {
			busName = fmt.Sprintf("/dev/spidev0.%d", e.config.ChipSelect)
		}

		port, err := spireg.Open(busName)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to open SPI bus: %w", err)
		}
		e.port = port

		conn, err := port.Connect(physic.Frequency(e.config.Speed)*physic.Hertz, spi.Mode0, 8)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to connect SPI: %w", err)
		}
		e.conn = conn
	}

	// Read single channel with averaging
	var totalRaw int
	for i := 0; i < e.config.Samples; i++ {
		raw, err := e.readChannel12(e.config.Channel)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to read channel: %w", err)
		}
		totalRaw += raw

		if i < e.config.Samples-1 {
			time.Sleep(1 * time.Millisecond)
		}
	}

	avgRaw := totalRaw / e.config.Samples
	voltage := float64(avgRaw) / 4095.0 * e.config.VRef // 12-bit = 4095
	scaledValue := voltage*e.config.Scale + e.config.Offset
	percent := float64(avgRaw) / 4095.0 * 100.0

	return node.Message{
		Payload: map[string]interface{}{
			"channel":    e.config.Channel,
			"raw":        avgRaw,
			"voltage":    voltage,
			"value":      scaledValue,
			"percent":    percent,
			"vref":       e.config.VRef,
			"mode":       e.config.Mode,
			"resolution": 12,
			"samples":    e.config.Samples,
			"sensor":     "MCP3208",
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

// readChannel12 reads a single 12-bit ADC channel
func (e *MCP3208Executor) readChannel12(channel int) (int, error) {
	// MCP3208 uses same protocol as MCP3008 but returns 12 bits
	var configByte byte
	if e.config.Mode == "single" {
		configByte = 0x80 | byte(channel<<4)
	} else {
		configByte = byte(channel << 4)
	}

	tx := []byte{0x01, configByte, 0x00}
	rx := make([]byte, 3)

	if err := e.conn.Tx(tx, rx); err != nil {
		return 0, fmt.Errorf("SPI transfer failed: %w", err)
	}

	// Extract 12-bit result
	// rx[1] contains 4 MSBs in lower 4 bits
	// rx[2] contains 8 LSBs
	raw := int(rx[1]&0x0F)<<8 | int(rx[2])

	return raw, nil
}
