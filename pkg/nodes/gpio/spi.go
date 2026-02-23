package gpio

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/EdgxCloud/EdgeFlow/internal/hal"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// SPIConfig SPI node configuration
type SPIConfig struct {
	Bus      int    `json:"bus"`      // Bus number (0 or 1)
	Device   int    `json:"device"`   // Device/CS number (0, 1)
	Speed    int    `json:"speed"`    // SPI speed in Hz
	Mode     int    `json:"mode"`     // SPI mode (0-3)
	BitsWord int    `json:"bitsWord"` // Bits per word (8, 16)
	Mode3Wire bool  `json:"mode3Wire"` // 3-wire mode
}

// SPIExecutor SPI node executor
type SPIExecutor struct {
	config SPIConfig
	hal    hal.HAL
}

// NewSPIExecutor create SPIExecutor
func NewSPIExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var spiConfig SPIConfig
	if err := json.Unmarshal(configJSON, &spiConfig); err != nil {
		return nil, fmt.Errorf("invalid spi config: %w", err)
	}

	// Defaults
	if spiConfig.Speed == 0 {
		spiConfig.Speed = 1000000 // 1 MHz
	}
	if spiConfig.BitsWord == 0 {
		spiConfig.BitsWord = 8
	}

	// Validate
	if spiConfig.Bus < 0 || spiConfig.Bus > 1 {
		return nil, fmt.Errorf("invalid bus number (0 or 1)")
	}
	if spiConfig.Device < 0 || spiConfig.Device > 1 {
		return nil, fmt.Errorf("invalid device number (0 or 1)")
	}
	if spiConfig.Mode < 0 || spiConfig.Mode > 3 {
		return nil, fmt.Errorf("invalid SPI mode (0-3)")
	}

	return &SPIExecutor{
		config: spiConfig,
	}, nil
}

// Init initializes the SPI executor with config
func (e *SPIExecutor) Init(config map[string]interface{}) error {
	// Config is already parsed in NewSPIExecutor
	return nil
}

// Execute execute node
func (e *SPIExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get HAL if not initialized
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h

		// Setup SPI
		if err := e.setup(); err != nil {
			return node.Message{}, fmt.Errorf("failed to setup SPI: %w", err)
		}
	}

	spi := e.hal.SPI()

	// Get data to write from message
	var writeData []byte

	if msg.Payload != nil {
		if d, ok := msg.Payload["data"].([]interface{}); ok {
			writeData = make([]byte, len(d))
			for i, v := range d {
				if num, ok := v.(float64); ok {
					writeData[i] = byte(num)
				}
			}
		} else if d, ok := msg.Payload["data"].(string); ok {
			// Try to decode hex string
			decoded, err := hex.DecodeString(d)
			if err == nil {
				writeData = decoded
			} else {
				writeData = []byte(d)
			}
		} else if v, ok := msg.Payload["value"].(float64); ok {
			writeData = []byte{byte(v)}
		}
	}

	if len(writeData) == 0 {
		return node.Message{}, fmt.Errorf("no data to write")
	}

	// Transfer (write and read simultaneously)
	readData, err := spi.Transfer(writeData)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to transfer SPI: %w", err)
	}

	// Return data
	return node.Message{
		Payload: map[string]interface{}{
			"bus":      e.config.Bus,
			"device":   e.config.Device,
			"written":  writeData,
			"read":     readData,
			"hex":      hex.EncodeToString(readData),
			"length":   len(readData),
		},
	}, nil
}

// setup initialize SPI
func (e *SPIExecutor) setup() error {
	spi := e.hal.SPI()

	// Open SPI
	if err := spi.Open(e.config.Bus, e.config.Device); err != nil {
		return err
	}

	// Set speed
	if err := spi.SetSpeed(e.config.Speed); err != nil {
		return err
	}

	// Set mode
	if err := spi.SetMode(byte(e.config.Mode)); err != nil {
		return err
	}

	// Set bits per word
	if err := spi.SetBitsPerWord(byte(e.config.BitsWord)); err != nil {
		return err
	}

	return nil
}

// Cleanup cleanup resources
func (e *SPIExecutor) Cleanup() error {
	if e.hal != nil {
		e.hal.SPI().Close()
	}
	return nil
}
