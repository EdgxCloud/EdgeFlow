package gpio

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"go.bug.st/serial"
)

// Modbus function codes
const (
	FuncReadCoils          = 0x01
	FuncReadDiscreteInputs = 0x02
	FuncReadHoldingRegs    = 0x03
	FuncReadInputRegs      = 0x04
	FuncWriteSingleCoil    = 0x05
	FuncWriteSingleReg     = 0x06
	FuncWriteMultipleCoils = 0x0F
	FuncWriteMultipleRegs  = 0x10
)

// ModbusConfig configuration for Modbus node
type ModbusConfig struct {
	// Connection settings
	Mode     string `json:"mode"`      // "tcp" or "rtu"
	Host     string `json:"host"`      // TCP host (e.g., "192.168.1.100")
	Port     int    `json:"port"`      // TCP port (default: 502)
	Device   string `json:"device"`    // Serial device (e.g., "/dev/ttyUSB0")
	BaudRate int    `json:"baud_rate"` // Serial baud rate (default: 9600)
	DataBits int    `json:"data_bits"` // Serial data bits (default: 8)
	StopBits int    `json:"stop_bits"` // Serial stop bits (default: 1)
	Parity   string `json:"parity"`    // Serial parity: "none", "odd", "even"

	// Modbus settings
	SlaveID  byte   `json:"slave_id"`  // Slave/Unit ID (1-247)
	Function string `json:"function"`  // Function: "read_coils", "read_holding", etc.
	Address  uint16 `json:"address"`   // Starting register address
	Quantity uint16 `json:"quantity"`  // Number of registers/coils to read
	Value    int    `json:"value"`     // Value to write (for write functions)
	Values   []int  `json:"values"`    // Values to write (for write multiple)

	// Timeout
	Timeout int `json:"timeout"` // Timeout in milliseconds (default: 1000)
}

// ModbusExecutor executes Modbus operations
type ModbusExecutor struct {
	config      ModbusConfig
	tcpConn     net.Conn
	serialPort  serial.Port
	mu          sync.Mutex
	transID     uint16
}

// NewModbusExecutor creates a new Modbus executor
func NewModbusExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var mbConfig ModbusConfig
	if err := json.Unmarshal(configJSON, &mbConfig); err != nil {
		return nil, fmt.Errorf("invalid Modbus config: %w", err)
	}

	// Validate mode
	if mbConfig.Mode == "" {
		mbConfig.Mode = "tcp"
	}
	if mbConfig.Mode != "tcp" && mbConfig.Mode != "rtu" {
		return nil, fmt.Errorf("invalid mode: %s (must be 'tcp' or 'rtu')", mbConfig.Mode)
	}

	// TCP defaults
	if mbConfig.Mode == "tcp" {
		if mbConfig.Host == "" {
			return nil, fmt.Errorf("host is required for TCP mode")
		}
		if mbConfig.Port == 0 {
			mbConfig.Port = 502
		}
	}

	// RTU defaults
	if mbConfig.Mode == "rtu" {
		if mbConfig.Device == "" {
			return nil, fmt.Errorf("device is required for RTU mode")
		}
		if mbConfig.BaudRate == 0 {
			mbConfig.BaudRate = 9600
		}
		if mbConfig.DataBits == 0 {
			mbConfig.DataBits = 8
		}
		if mbConfig.StopBits == 0 {
			mbConfig.StopBits = 1
		}
		if mbConfig.Parity == "" {
			mbConfig.Parity = "none"
		}
	}

	// Common defaults
	if mbConfig.SlaveID == 0 {
		mbConfig.SlaveID = 1
	}
	if mbConfig.Timeout == 0 {
		mbConfig.Timeout = 1000
	}
	if mbConfig.Quantity == 0 {
		mbConfig.Quantity = 1
	}

	// Validate function
	validFunctions := map[string]bool{
		"read_coils": true, "read_discrete": true, "read_holding": true, "read_input": true,
		"write_coil": true, "write_register": true, "write_coils": true, "write_registers": true,
	}
	if mbConfig.Function == "" {
		mbConfig.Function = "read_holding"
	}
	if !validFunctions[mbConfig.Function] {
		return nil, fmt.Errorf("invalid function: %s", mbConfig.Function)
	}

	return &ModbusExecutor{
		config: mbConfig,
	}, nil
}

// Init initializes the Modbus executor
func (e *ModbusExecutor) Init(config map[string]interface{}) error {
	return nil
}

// Execute performs the Modbus operation
func (e *ModbusExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Connect if not connected
	if err := e.connect(); err != nil {
		return node.Message{}, fmt.Errorf("connection failed: %w", err)
	}

	// Override config from message payload if present
	config := e.config
	if msg.Payload != nil {
		if addr, ok := msg.Payload["address"].(float64); ok {
			config.Address = uint16(addr)
		}
		if qty, ok := msg.Payload["quantity"].(float64); ok {
			config.Quantity = uint16(qty)
		}
		if val, ok := msg.Payload["value"].(float64); ok {
			config.Value = int(val)
		}
		if fn, ok := msg.Payload["function"].(string); ok {
			config.Function = fn
		}
	}

	// Execute function
	var result interface{}
	var err error

	switch config.Function {
	case "read_coils":
		result, err = e.readCoils(config.Address, config.Quantity)
	case "read_discrete":
		result, err = e.readDiscreteInputs(config.Address, config.Quantity)
	case "read_holding":
		result, err = e.readHoldingRegisters(config.Address, config.Quantity)
	case "read_input":
		result, err = e.readInputRegisters(config.Address, config.Quantity)
	case "write_coil":
		err = e.writeSingleCoil(config.Address, config.Value != 0)
		result = config.Value != 0
	case "write_register":
		err = e.writeSingleRegister(config.Address, uint16(config.Value))
		result = config.Value
	case "write_coils":
		values := make([]bool, len(config.Values))
		for i, v := range config.Values {
			values[i] = v != 0
		}
		err = e.writeMultipleCoils(config.Address, values)
		result = values
	case "write_registers":
		values := make([]uint16, len(config.Values))
		for i, v := range config.Values {
			values[i] = uint16(v)
		}
		err = e.writeMultipleRegisters(config.Address, values)
		result = config.Values
	default:
		return node.Message{}, fmt.Errorf("unsupported function: %s", config.Function)
	}

	if err != nil {
		return node.Message{}, fmt.Errorf("Modbus error: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"function":  config.Function,
			"address":   config.Address,
			"quantity":  config.Quantity,
			"slave_id":  config.SlaveID,
			"result":    result,
			"mode":      config.Mode,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// connect establishes connection to the Modbus device
func (e *ModbusExecutor) connect() error {
	if e.config.Mode == "tcp" {
		if e.tcpConn != nil {
			return nil // Already connected
		}

		addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)
		conn, err := net.DialTimeout("tcp", addr, time.Duration(e.config.Timeout)*time.Millisecond)
		if err != nil {
			return fmt.Errorf("TCP connect failed: %w", err)
		}
		e.tcpConn = conn
	} else {
		if e.serialPort != nil {
			return nil // Already connected
		}

		mode := &serial.Mode{
			BaudRate: e.config.BaudRate,
			DataBits: e.config.DataBits,
			StopBits: serial.StopBits(e.config.StopBits),
		}

		switch e.config.Parity {
		case "odd":
			mode.Parity = serial.OddParity
		case "even":
			mode.Parity = serial.EvenParity
		default:
			mode.Parity = serial.NoParity
		}

		port, err := serial.Open(e.config.Device, mode)
		if err != nil {
			return fmt.Errorf("serial open failed: %w", err)
		}
		port.SetReadTimeout(time.Duration(e.config.Timeout) * time.Millisecond)
		e.serialPort = port
	}

	return nil
}

// sendRequest sends a Modbus request and returns the response
func (e *ModbusExecutor) sendRequest(functionCode byte, data []byte) ([]byte, error) {
	if e.config.Mode == "tcp" {
		return e.sendTCPRequest(functionCode, data)
	}
	return e.sendRTURequest(functionCode, data)
}

// sendTCPRequest sends a Modbus TCP request
func (e *ModbusExecutor) sendTCPRequest(functionCode byte, data []byte) ([]byte, error) {
	e.transID++

	// Build MBAP header + PDU
	// MBAP: Transaction ID (2) + Protocol ID (2) + Length (2) + Unit ID (1)
	// PDU: Function code (1) + Data (N)
	pduLen := 1 + len(data)
	request := make([]byte, 7+pduLen)

	binary.BigEndian.PutUint16(request[0:2], e.transID)   // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)           // Protocol ID (0 = Modbus)
	binary.BigEndian.PutUint16(request[4:6], uint16(pduLen+1)) // Length
	request[6] = e.config.SlaveID                         // Unit ID
	request[7] = functionCode                             // Function code
	copy(request[8:], data)                               // Data

	// Set timeout
	e.tcpConn.SetDeadline(time.Now().Add(time.Duration(e.config.Timeout) * time.Millisecond))

	// Send request
	if _, err := e.tcpConn.Write(request); err != nil {
		e.tcpConn.Close()
		e.tcpConn = nil
		return nil, fmt.Errorf("write failed: %w", err)
	}

	// Read response header
	header := make([]byte, 7)
	if _, err := e.tcpConn.Read(header); err != nil {
		return nil, fmt.Errorf("read header failed: %w", err)
	}

	// Verify transaction ID
	respTransID := binary.BigEndian.Uint16(header[0:2])
	if respTransID != e.transID {
		return nil, fmt.Errorf("transaction ID mismatch: expected %d, got %d", e.transID, respTransID)
	}

	// Get response length
	respLen := binary.BigEndian.Uint16(header[4:6]) - 1 // Subtract unit ID

	// Read response PDU
	pdu := make([]byte, respLen)
	if _, err := e.tcpConn.Read(pdu); err != nil {
		return nil, fmt.Errorf("read PDU failed: %w", err)
	}

	// Check for exception
	if pdu[0]&0x80 != 0 {
		return nil, fmt.Errorf("Modbus exception: code %d", pdu[1])
	}

	return pdu[1:], nil // Return data after function code
}

// sendRTURequest sends a Modbus RTU request
func (e *ModbusExecutor) sendRTURequest(functionCode byte, data []byte) ([]byte, error) {
	// Build RTU frame: Slave ID (1) + Function (1) + Data (N) + CRC (2)
	frame := make([]byte, 2+len(data)+2)
	frame[0] = e.config.SlaveID
	frame[1] = functionCode
	copy(frame[2:], data)

	// Calculate CRC
	crc := crc16(frame[:len(frame)-2])
	binary.LittleEndian.PutUint16(frame[len(frame)-2:], crc)

	// Clear input buffer
	e.serialPort.ResetInputBuffer()

	// Send request
	if _, err := e.serialPort.Write(frame); err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	// Wait for response (3.5 character times minimum)
	time.Sleep(5 * time.Millisecond)

	// Read response
	response := make([]byte, 256)
	n, err := e.serialPort.Read(response)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}
	if n < 4 {
		return nil, fmt.Errorf("response too short: %d bytes", n)
	}

	response = response[:n]

	// Verify CRC
	respCRC := binary.LittleEndian.Uint16(response[n-2:])
	calcCRC := crc16(response[:n-2])
	if respCRC != calcCRC {
		return nil, fmt.Errorf("CRC mismatch: expected %04X, got %04X", calcCRC, respCRC)
	}

	// Check for exception
	if response[1]&0x80 != 0 {
		return nil, fmt.Errorf("Modbus exception: code %d", response[2])
	}

	// Return data (skip slave ID, function code, and CRC)
	return response[2 : n-2], nil
}

// readCoils reads coil status (function 0x01)
func (e *ModbusExecutor) readCoils(address, quantity uint16) ([]bool, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)

	resp, err := e.sendRequest(FuncReadCoils, data)
	if err != nil {
		return nil, err
	}

	// Parse response
	byteCount := int(resp[0])
	coils := make([]bool, quantity)
	for i := uint16(0); i < quantity; i++ {
		byteIdx := int(i / 8)
		bitIdx := uint(i % 8)
		if byteIdx < byteCount {
			coils[i] = (resp[1+byteIdx] & (1 << bitIdx)) != 0
		}
	}

	return coils, nil
}

// readDiscreteInputs reads discrete inputs (function 0x02)
func (e *ModbusExecutor) readDiscreteInputs(address, quantity uint16) ([]bool, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)

	resp, err := e.sendRequest(FuncReadDiscreteInputs, data)
	if err != nil {
		return nil, err
	}

	// Parse response (same format as read coils)
	byteCount := int(resp[0])
	inputs := make([]bool, quantity)
	for i := uint16(0); i < quantity; i++ {
		byteIdx := int(i / 8)
		bitIdx := uint(i % 8)
		if byteIdx < byteCount {
			inputs[i] = (resp[1+byteIdx] & (1 << bitIdx)) != 0
		}
	}

	return inputs, nil
}

// readHoldingRegisters reads holding registers (function 0x03)
func (e *ModbusExecutor) readHoldingRegisters(address, quantity uint16) ([]uint16, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)

	resp, err := e.sendRequest(FuncReadHoldingRegs, data)
	if err != nil {
		return nil, err
	}

	// Parse response
	byteCount := int(resp[0])
	registers := make([]uint16, quantity)
	for i := 0; i < int(quantity) && i*2+2 <= byteCount; i++ {
		registers[i] = binary.BigEndian.Uint16(resp[1+i*2 : 3+i*2])
	}

	return registers, nil
}

// readInputRegisters reads input registers (function 0x04)
func (e *ModbusExecutor) readInputRegisters(address, quantity uint16) ([]uint16, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)

	resp, err := e.sendRequest(FuncReadInputRegs, data)
	if err != nil {
		return nil, err
	}

	// Parse response (same format as holding registers)
	byteCount := int(resp[0])
	registers := make([]uint16, quantity)
	for i := 0; i < int(quantity) && i*2+2 <= byteCount; i++ {
		registers[i] = binary.BigEndian.Uint16(resp[1+i*2 : 3+i*2])
	}

	return registers, nil
}

// writeSingleCoil writes a single coil (function 0x05)
func (e *ModbusExecutor) writeSingleCoil(address uint16, value bool) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	if value {
		data[2] = 0xFF
		data[3] = 0x00
	} else {
		data[2] = 0x00
		data[3] = 0x00
	}

	_, err := e.sendRequest(FuncWriteSingleCoil, data)
	return err
}

// writeSingleRegister writes a single register (function 0x06)
func (e *ModbusExecutor) writeSingleRegister(address, value uint16) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], value)

	_, err := e.sendRequest(FuncWriteSingleReg, data)
	return err
}

// writeMultipleCoils writes multiple coils (function 0x0F)
func (e *ModbusExecutor) writeMultipleCoils(address uint16, values []bool) error {
	quantity := uint16(len(values))
	byteCount := (quantity + 7) / 8

	data := make([]byte, 5+byteCount)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)
	data[4] = byte(byteCount)

	// Pack coils into bytes
	for i, v := range values {
		if v {
			byteIdx := i / 8
			bitIdx := uint(i % 8)
			data[5+byteIdx] |= 1 << bitIdx
		}
	}

	_, err := e.sendRequest(FuncWriteMultipleCoils, data)
	return err
}

// writeMultipleRegisters writes multiple registers (function 0x10)
func (e *ModbusExecutor) writeMultipleRegisters(address uint16, values []uint16) error {
	quantity := uint16(len(values))
	byteCount := quantity * 2

	data := make([]byte, 5+byteCount)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)
	data[4] = byte(byteCount)

	for i, v := range values {
		binary.BigEndian.PutUint16(data[5+i*2:7+i*2], v)
	}

	_, err := e.sendRequest(FuncWriteMultipleRegs, data)
	return err
}

// crc16 calculates Modbus CRC-16
func crc16(data []byte) uint16 {
	crc := uint16(0xFFFF)

	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}

	return crc
}

// Cleanup releases resources
func (e *ModbusExecutor) Cleanup() error {
	if e.tcpConn != nil {
		e.tcpConn.Close()
		e.tcpConn = nil
	}
	if e.serialPort != nil {
		e.serialPort.Close()
		e.serialPort = nil
	}
	return nil
}
