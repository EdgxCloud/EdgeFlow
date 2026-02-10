package gpio

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/node"
)

// SerialConfig Serial node configuration
type SerialConfig struct {
	Port       string `json:"port"`       // Serial port path (e.g., /dev/ttyS0, /dev/ttyUSB0)
	BaudRate   int    `json:"baudRate"`   // Baud rate (9600, 115200, etc.)
	DataBits   int    `json:"dataBits"`   // Data bits (7, 8)
	StopBits   int    `json:"stopBits"`   // Stop bits (1, 2)
	Parity     string `json:"parity"`     // Parity: "none", "odd", "even"
	Mode       string `json:"mode"`       // "write", "read", "readwrite"
	Timeout    int    `json:"timeout"`    // Read timeout in milliseconds
	Delimiter  string `json:"delimiter"`  // Message delimiter for read mode
	BufferSize int    `json:"bufferSize"` // Read buffer size
}

// SerialExecutor Serial node executor
type SerialExecutor struct {
	config SerialConfig
	hal    hal.HAL
}

// NewSerialExecutor create SerialExecutor
func NewSerialExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var serialConfig SerialConfig
	if err := json.Unmarshal(configJSON, &serialConfig); err != nil {
		return nil, fmt.Errorf("invalid serial config: %w", err)
	}

	// Defaults
	if serialConfig.BaudRate == 0 {
		serialConfig.BaudRate = 9600
	}
	if serialConfig.DataBits == 0 {
		serialConfig.DataBits = 8
	}
	if serialConfig.StopBits == 0 {
		serialConfig.StopBits = 1
	}
	if serialConfig.Parity == "" {
		serialConfig.Parity = "none"
	}
	if serialConfig.Mode == "" {
		serialConfig.Mode = "readwrite"
	}
	if serialConfig.Timeout == 0 {
		serialConfig.Timeout = 1000 // 1 second
	}
	if serialConfig.BufferSize == 0 {
		serialConfig.BufferSize = 1024
	}

	// Validate
	if serialConfig.Port == "" {
		return nil, fmt.Errorf("serial port is required")
	}
	if serialConfig.BaudRate <= 0 {
		return nil, fmt.Errorf("invalid baud rate")
	}
	if serialConfig.DataBits != 7 && serialConfig.DataBits != 8 {
		return nil, fmt.Errorf("data bits must be 7 or 8")
	}
	if serialConfig.StopBits != 1 && serialConfig.StopBits != 2 {
		return nil, fmt.Errorf("stop bits must be 1 or 2")
	}
	if serialConfig.Parity != "none" && serialConfig.Parity != "odd" && serialConfig.Parity != "even" {
		return nil, fmt.Errorf("parity must be none, odd, or even")
	}
	if serialConfig.Mode != "write" && serialConfig.Mode != "read" && serialConfig.Mode != "readwrite" {
		return nil, fmt.Errorf("mode must be write, read, or readwrite")
	}

	return &SerialExecutor{
		config: serialConfig,
	}, nil
}

// Init initializes the Serial executor with config
func (e *SerialExecutor) Init(config map[string]interface{}) error {
	// Config is already parsed in NewSerialExecutor
	return nil
}

// Execute execute node
func (e *SerialExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get HAL if not initialized
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h

		// Setup Serial
		if err := e.setup(); err != nil {
			return node.Message{}, fmt.Errorf("failed to setup Serial: %w", err)
		}
	}

	serial := e.hal.Serial()

	// Handle based on mode
	switch e.config.Mode {
	case "write":
		return e.handleWrite(serial, msg)
	case "read":
		return e.handleRead(serial, msg)
	case "readwrite":
		return e.handleReadWrite(serial, msg)
	default:
		return node.Message{}, fmt.Errorf("invalid mode: %s", e.config.Mode)
	}
}

// handleWrite handles write-only mode
func (e *SerialExecutor) handleWrite(serial hal.SerialProvider, msg node.Message) (node.Message, error) {
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
		} else if v, ok := msg.Payload["payload"].(string); ok {
			writeData = []byte(v)
		}
	}

	if len(writeData) == 0 {
		return node.Message{}, fmt.Errorf("no data to write")
	}

	// Write to serial
	n, err := serial.Write(writeData)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to write serial: %w", err)
	}

	// Return result
	return node.Message{
		Payload: map[string]interface{}{
			"port":    e.config.Port,
			"written": n,
			"data":    writeData,
			"hex":     hex.EncodeToString(writeData),
		},
	}, nil
}

// handleRead handles read-only mode
func (e *SerialExecutor) handleRead(serial hal.SerialProvider, msg node.Message) (node.Message, error) {
	buffer := make([]byte, e.config.BufferSize)

	// Set read timeout
	timeout := time.Duration(e.config.Timeout) * time.Millisecond
	deadline := time.Now().Add(timeout)

	var readData []byte
	for time.Now().Before(deadline) {
		n, err := serial.Read(buffer)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to read serial: %w", err)
		}

		if n > 0 {
			readData = append(readData, buffer[:n]...)

			// Check for delimiter if specified
			if e.config.Delimiter != "" {
				if contains(readData, []byte(e.config.Delimiter)) {
					break
				}
			} else {
				// No delimiter, return immediately
				break
			}
		}

		// Small delay before next read
		time.Sleep(10 * time.Millisecond)
	}

	if len(readData) == 0 {
		return node.Message{}, fmt.Errorf("no data received (timeout)")
	}

	// Return result
	return node.Message{
		Payload: map[string]interface{}{
			"port":   e.config.Port,
			"data":   readData,
			"hex":    hex.EncodeToString(readData),
			"string": string(readData),
			"length": len(readData),
		},
	}, nil
}

// handleReadWrite handles read-write mode
func (e *SerialExecutor) handleReadWrite(serial hal.SerialProvider, msg node.Message) (node.Message, error) {
	// First write
	writeResult, err := e.handleWrite(serial, msg)
	if err != nil {
		return node.Message{}, err
	}

	// Small delay before reading
	time.Sleep(50 * time.Millisecond)

	// Then read
	buffer := make([]byte, e.config.BufferSize)
	timeout := time.Duration(e.config.Timeout) * time.Millisecond
	deadline := time.Now().Add(timeout)

	var readData []byte
	for time.Now().Before(deadline) {
		n, err := serial.Read(buffer)
		if err != nil {
			// Ignore read errors in readwrite mode
			break
		}

		if n > 0 {
			readData = append(readData, buffer[:n]...)

			// Check for delimiter if specified
			if e.config.Delimiter != "" {
				if contains(readData, []byte(e.config.Delimiter)) {
					break
				}
			} else {
				// No delimiter, return after first read
				break
			}
		}

		time.Sleep(10 * time.Millisecond)
	}

	// Combine write and read results
	result := map[string]interface{}{
		"port":    e.config.Port,
		"written": writeResult.Payload["written"],
		"sent":    writeResult.Payload["data"],
	}

	if len(readData) > 0 {
		result["received"] = readData
		result["hex"] = hex.EncodeToString(readData)
		result["string"] = string(readData)
		result["length"] = len(readData)
	}

	return node.Message{
		Payload: result,
	}, nil
}

// setup initialize Serial
func (e *SerialExecutor) setup() error {
	serial := e.hal.Serial()

	// Open serial port
	if err := serial.Open(e.config.Port); err != nil {
		return err
	}

	// Set baud rate
	if err := serial.SetBaudRate(e.config.BaudRate); err != nil {
		return err
	}

	// Set data bits
	if err := serial.SetDataBits(e.config.DataBits); err != nil {
		return err
	}

	// Set stop bits
	if err := serial.SetStopBits(e.config.StopBits); err != nil {
		return err
	}

	// Set parity
	var parity byte
	switch e.config.Parity {
	case "none":
		parity = 0
	case "odd":
		parity = 1
	case "even":
		parity = 2
	}
	if err := serial.SetParity(parity); err != nil {
		return err
	}

	return nil
}

// Cleanup cleanup resources
func (e *SerialExecutor) Cleanup() error {
	if e.hal != nil {
		e.hal.Serial().Close()
	}
	return nil
}

// contains checks if a byte slice contains a subsequence
func contains(data, substr []byte) bool {
	if len(substr) == 0 {
		return true
	}
	if len(data) < len(substr) {
		return false
	}

	for i := 0; i <= len(data)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if data[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
