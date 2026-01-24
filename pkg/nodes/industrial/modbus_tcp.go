// Package industrial provides nodes for industrial protocols
// Supports Modbus TCP/RTU, OPC-UA, and other automation protocols
package industrial

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// ModbusTCPNode implements Modbus TCP client for industrial communication
type ModbusTCPNode struct {
	host          string
	port          int
	unitID        byte
	timeout       time.Duration
	operation     string // read_coils, read_discrete, read_holding, read_input, write_coil, write_register, write_coils, write_registers
	address       uint16
	quantity      uint16
	conn          net.Conn
	mu            sync.Mutex
	transactionID uint16
}

// Modbus function codes
const (
	FuncReadCoils          = 0x01
	FuncReadDiscreteInputs = 0x02
	FuncReadHoldingRegs    = 0x03
	FuncReadInputRegs      = 0x04
	FuncWriteSingleCoil    = 0x05
	FuncWriteSingleReg     = 0x06
	FuncWriteMultiCoils    = 0x0F
	FuncWriteMultiRegs     = 0x10
)

// NewModbusTCPNode creates a new Modbus TCP node
func NewModbusTCPNode() *ModbusTCPNode {
	return &ModbusTCPNode{
		host:      "127.0.0.1",
		port:      502,
		unitID:    1,
		timeout:   5 * time.Second,
		operation: "read_holding",
		address:   0,
		quantity:  1,
	}
}

// Init initializes the Modbus TCP node
func (n *ModbusTCPNode) Init(config map[string]interface{}) error {
	if host, ok := config["host"].(string); ok {
		n.host = host
	}
	if port, ok := config["port"].(float64); ok {
		n.port = int(port)
	}
	if unitID, ok := config["unitId"].(float64); ok {
		n.unitID = byte(unitID)
	}
	if timeout, ok := config["timeout"].(float64); ok {
		n.timeout = time.Duration(timeout) * time.Millisecond
	}
	if op, ok := config["operation"].(string); ok {
		n.operation = op
	}
	if addr, ok := config["address"].(float64); ok {
		n.address = uint16(addr)
	}
	if qty, ok := config["quantity"].(float64); ok {
		n.quantity = uint16(qty)
	}

	return nil
}

// Execute performs Modbus TCP operation
func (n *ModbusTCPNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Get operation parameters from message or use defaults
	operation := n.operation
	address := n.address
	quantity := n.quantity
	var values []uint16

	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}
	if addr, ok := msg.Payload["address"].(float64); ok {
		address = uint16(addr)
	}
	if qty, ok := msg.Payload["quantity"].(float64); ok {
		quantity = uint16(qty)
	}
	if vals, ok := msg.Payload["values"].([]interface{}); ok {
		for _, v := range vals {
			if fv, ok := v.(float64); ok {
				values = append(values, uint16(fv))
			}
		}
	}
	if val, ok := msg.Payload["value"].(float64); ok {
		values = []uint16{uint16(val)}
	}

	// Connect if not connected
	if n.conn == nil {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", n.host, n.port), n.timeout)
		if err != nil {
			return msg, fmt.Errorf("modbus connection failed: %w", err)
		}
		n.conn = conn
	}

	// Set deadline
	n.conn.SetDeadline(time.Now().Add(n.timeout))

	var result interface{}
	var err error

	switch operation {
	case "read_coils":
		result, err = n.readCoils(address, quantity)
	case "read_discrete":
		result, err = n.readDiscreteInputs(address, quantity)
	case "read_holding":
		result, err = n.readHoldingRegisters(address, quantity)
	case "read_input":
		result, err = n.readInputRegisters(address, quantity)
	case "write_coil":
		var coilValue bool
		if len(values) > 0 {
			coilValue = values[0] != 0
		} else if v, ok := msg.Payload["value"].(bool); ok {
			coilValue = v
		}
		err = n.writeSingleCoil(address, coilValue)
		result = map[string]interface{}{"success": err == nil, "address": address}
	case "write_register":
		if len(values) > 0 {
			err = n.writeSingleRegister(address, values[0])
			result = map[string]interface{}{"success": err == nil, "address": address, "value": values[0]}
		} else {
			err = fmt.Errorf("no value provided for write_register")
		}
	case "write_coils":
		var coilValues []bool
		for _, v := range values {
			coilValues = append(coilValues, v != 0)
		}
		err = n.writeMultipleCoils(address, coilValues)
		result = map[string]interface{}{"success": err == nil, "address": address, "quantity": len(coilValues)}
	case "write_registers":
		err = n.writeMultipleRegisters(address, values)
		result = map[string]interface{}{"success": err == nil, "address": address, "quantity": len(values)}
	default:
		err = fmt.Errorf("unknown operation: %s", operation)
	}

	if err != nil {
		// Close connection on error
		n.conn.Close()
		n.conn = nil
		return msg, err
	}

	// Build response
	msg.Payload["result"] = result
	msg.Payload["operation"] = operation
	msg.Payload["address"] = address
	msg.Payload["unitId"] = n.unitID

	return msg, nil
}

// readCoils reads coils (function code 0x01)
func (n *ModbusTCPNode) readCoils(address, quantity uint16) ([]bool, error) {
	request := n.buildRequest(FuncReadCoils, address, quantity, nil)
	response, err := n.sendRequest(request)
	if err != nil {
		return nil, err
	}

	// Parse response
	if len(response) < 9 {
		return nil, fmt.Errorf("invalid response length")
	}

	byteCount := int(response[8])
	if len(response) < 9+byteCount {
		return nil, fmt.Errorf("incomplete response")
	}

	coils := make([]bool, quantity)
	for i := uint16(0); i < quantity; i++ {
		byteIdx := i / 8
		bitIdx := i % 8
		coils[i] = (response[9+byteIdx] & (1 << bitIdx)) != 0
	}

	return coils, nil
}

// readDiscreteInputs reads discrete inputs (function code 0x02)
func (n *ModbusTCPNode) readDiscreteInputs(address, quantity uint16) ([]bool, error) {
	request := n.buildRequest(FuncReadDiscreteInputs, address, quantity, nil)
	response, err := n.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if len(response) < 9 {
		return nil, fmt.Errorf("invalid response length")
	}

	byteCount := int(response[8])
	if len(response) < 9+byteCount {
		return nil, fmt.Errorf("incomplete response")
	}

	inputs := make([]bool, quantity)
	for i := uint16(0); i < quantity; i++ {
		byteIdx := i / 8
		bitIdx := i % 8
		inputs[i] = (response[9+byteIdx] & (1 << bitIdx)) != 0
	}

	return inputs, nil
}

// readHoldingRegisters reads holding registers (function code 0x03)
func (n *ModbusTCPNode) readHoldingRegisters(address, quantity uint16) ([]uint16, error) {
	request := n.buildRequest(FuncReadHoldingRegs, address, quantity, nil)
	response, err := n.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if len(response) < 9 {
		return nil, fmt.Errorf("invalid response length")
	}

	byteCount := int(response[8])
	if len(response) < 9+byteCount {
		return nil, fmt.Errorf("incomplete response")
	}

	registers := make([]uint16, quantity)
	for i := uint16(0); i < quantity; i++ {
		registers[i] = binary.BigEndian.Uint16(response[9+i*2:])
	}

	return registers, nil
}

// readInputRegisters reads input registers (function code 0x04)
func (n *ModbusTCPNode) readInputRegisters(address, quantity uint16) ([]uint16, error) {
	request := n.buildRequest(FuncReadInputRegs, address, quantity, nil)
	response, err := n.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if len(response) < 9 {
		return nil, fmt.Errorf("invalid response length")
	}

	byteCount := int(response[8])
	if len(response) < 9+byteCount {
		return nil, fmt.Errorf("incomplete response")
	}

	registers := make([]uint16, quantity)
	for i := uint16(0); i < quantity; i++ {
		registers[i] = binary.BigEndian.Uint16(response[9+i*2:])
	}

	return registers, nil
}

// writeSingleCoil writes a single coil (function code 0x05)
func (n *ModbusTCPNode) writeSingleCoil(address uint16, value bool) error {
	var coilValue uint16
	if value {
		coilValue = 0xFF00
	}
	request := n.buildRequest(FuncWriteSingleCoil, address, coilValue, nil)
	_, err := n.sendRequest(request)
	return err
}

// writeSingleRegister writes a single register (function code 0x06)
func (n *ModbusTCPNode) writeSingleRegister(address, value uint16) error {
	request := n.buildRequest(FuncWriteSingleReg, address, value, nil)
	_, err := n.sendRequest(request)
	return err
}

// writeMultipleCoils writes multiple coils (function code 0x0F)
func (n *ModbusTCPNode) writeMultipleCoils(address uint16, values []bool) error {
	// Convert bools to bytes
	byteCount := (len(values) + 7) / 8
	data := make([]byte, byteCount)
	for i, v := range values {
		if v {
			data[i/8] |= 1 << (i % 8)
		}
	}
	request := n.buildWriteMultiRequest(FuncWriteMultiCoils, address, uint16(len(values)), data)
	_, err := n.sendRequest(request)
	return err
}

// writeMultipleRegisters writes multiple registers (function code 0x10)
func (n *ModbusTCPNode) writeMultipleRegisters(address uint16, values []uint16) error {
	data := make([]byte, len(values)*2)
	for i, v := range values {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	request := n.buildWriteMultiRequest(FuncWriteMultiRegs, address, uint16(len(values)), data)
	_, err := n.sendRequest(request)
	return err
}

// buildRequest builds a Modbus TCP request
func (n *ModbusTCPNode) buildRequest(funcCode byte, address, value uint16, data []byte) []byte {
	n.transactionID++

	// MBAP Header (7 bytes) + PDU
	pduLen := 6 // Unit ID (1) + Function (1) + Address (2) + Value (2)
	request := make([]byte, 7+pduLen)

	// Transaction ID
	binary.BigEndian.PutUint16(request[0:], n.transactionID)
	// Protocol ID (0 for Modbus)
	binary.BigEndian.PutUint16(request[2:], 0)
	// Length
	binary.BigEndian.PutUint16(request[4:], uint16(pduLen))
	// Unit ID
	request[6] = n.unitID
	// Function code
	request[7] = funcCode
	// Address
	binary.BigEndian.PutUint16(request[8:], address)
	// Value/Quantity
	binary.BigEndian.PutUint16(request[10:], value)

	return request
}

// buildWriteMultiRequest builds a write multiple request
func (n *ModbusTCPNode) buildWriteMultiRequest(funcCode byte, address, quantity uint16, data []byte) []byte {
	n.transactionID++

	pduLen := 7 + len(data) // Unit ID + Function + Address + Quantity + ByteCount + Data
	request := make([]byte, 7+pduLen)

	// MBAP Header
	binary.BigEndian.PutUint16(request[0:], n.transactionID)
	binary.BigEndian.PutUint16(request[2:], 0)
	binary.BigEndian.PutUint16(request[4:], uint16(pduLen))
	request[6] = n.unitID

	// PDU
	request[7] = funcCode
	binary.BigEndian.PutUint16(request[8:], address)
	binary.BigEndian.PutUint16(request[10:], quantity)
	request[12] = byte(len(data))
	copy(request[13:], data)

	return request
}

// sendRequest sends request and receives response
func (n *ModbusTCPNode) sendRequest(request []byte) ([]byte, error) {
	// Send request
	_, err := n.conn.Write(request)
	if err != nil {
		return nil, fmt.Errorf("send failed: %w", err)
	}

	// Read response header (MBAP header = 7 bytes)
	header := make([]byte, 7)
	_, err = n.conn.Read(header)
	if err != nil {
		return nil, fmt.Errorf("read header failed: %w", err)
	}

	// Get PDU length from header
	pduLen := binary.BigEndian.Uint16(header[4:])

	// Read PDU
	pdu := make([]byte, pduLen)
	_, err = n.conn.Read(pdu)
	if err != nil {
		return nil, fmt.Errorf("read pdu failed: %w", err)
	}

	// Check for exception
	if len(pdu) >= 2 && pdu[0]&0x80 != 0 {
		return nil, fmt.Errorf("modbus exception: %d", pdu[1])
	}

	// Combine header and PDU
	response := append(header, pdu...)
	return response, nil
}

// Cleanup closes the Modbus connection
func (n *ModbusTCPNode) Cleanup() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.conn != nil {
		n.conn.Close()
		n.conn = nil
	}
	return nil
}

// NewModbusTCPExecutor creates a new Modbus TCP executor for registry
func NewModbusTCPExecutor() node.Executor {
	return NewModbusTCPNode()
}
