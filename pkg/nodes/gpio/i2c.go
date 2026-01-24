package gpio

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/node"
)

// I2CConfig نود I2C
type I2CConfig struct {
	Address  int    `json:"address"`  // آدرس I2C (0x00-0x7F)
	Register int    `json:"register"` // شماره رجیستر
	Length   int    `json:"length"`   // طول داده برای خواندن
	Mode     string `json:"mode"`     // read یا write
}

// I2CExecutor اجراکننده نود I2C
type I2CExecutor struct {
	config I2CConfig
	hal    hal.HAL
}

// NewI2CExecutor ایجاد I2CExecutor
func NewI2CExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var i2cConfig I2CConfig
	if err := json.Unmarshal(configJSON, &i2cConfig); err != nil {
		return nil, fmt.Errorf("invalid i2c config: %w", err)
	}

	// Validate
	if i2cConfig.Address < 0 || i2cConfig.Address > 0x7F {
		return nil, fmt.Errorf("invalid I2C address (must be 0x00-0x7F)")
	}

	// Default values
	if i2cConfig.Mode == "" {
		i2cConfig.Mode = "read"
	}
	if i2cConfig.Length == 0 {
		i2cConfig.Length = 1
	}

	return &I2CExecutor{
		config: i2cConfig,
	}, nil
}

// Init initializes the I2C executor with config
func (e *I2CExecutor) Init(config map[string]interface{}) error {
	// Config is already parsed in NewI2CExecutor
	return nil
}

// Execute اجرای نود
func (e *I2CExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get HAL if not initialized
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h
	}

	i2c := e.hal.I2C()

	// Open I2C with address
	if err := i2c.Open(byte(e.config.Address)); err != nil {
		return node.Message{}, fmt.Errorf("failed to open I2C: %w", err)
	}

	// Get mode from message or config
	mode := e.config.Mode
	register := e.config.Register
	length := e.config.Length

	if msg.Payload != nil {
		if m, ok := msg.Payload["mode"].(string); ok {
			mode = m
		}
		if r, ok := msg.Payload["register"].(float64); ok {
			register = int(r)
		}
		if l, ok := msg.Payload["length"].(float64); ok {
			length = int(l)
		}
	}

	if mode == "read" {
		// Read from I2C
		var data []byte
		var err error

		if register >= 0 {
			data, err = i2c.ReadRegister(byte(register), length)
		} else {
			data, err = i2c.Read(length)
		}

		if err != nil {
			return node.Message{}, fmt.Errorf("failed to read I2C: %w", err)
		}

		// Return data
		return node.Message{
			Payload: map[string]interface{}{
				"address":  e.config.Address,
				"register": register,
				"data":     data,
				"hex":      hex.EncodeToString(data),
				"length":   len(data),
			},
		}, nil

	} else {
		// Write to I2C
		var data []byte

		// Get data from message
		if msg.Payload != nil {
			if d, ok := msg.Payload["data"].([]interface{}); ok {
				data = make([]byte, len(d))
				for i, v := range d {
					if num, ok := v.(float64); ok {
						data[i] = byte(num)
					}
				}
			} else if d, ok := msg.Payload["data"].(string); ok {
				// Try to decode hex string
				decoded, err := hex.DecodeString(d)
				if err == nil {
					data = decoded
				} else {
					data = []byte(d)
				}
			} else if v, ok := msg.Payload["value"].(float64); ok {
				data = []byte{byte(v)}
			}
		}

		if len(data) == 0 {
			return node.Message{}, fmt.Errorf("no data to write")
		}

		// Write
		var err error
		if register >= 0 {
			err = i2c.WriteRegister(byte(register), data)
		} else {
			err = i2c.Write(data)
		}

		if err != nil {
			return node.Message{}, fmt.Errorf("failed to write I2C: %w", err)
		}

		return node.Message{
			Payload: map[string]interface{}{
				"address":  e.config.Address,
				"register": register,
				"written":  len(data),
			},
		}, nil
	}
}

// Cleanup پاکسازی منابع
func (e *I2CExecutor) Cleanup() error {
	if e.hal != nil {
		e.hal.I2C().Close()
	}
	return nil
}
