package industrial

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"go.bug.st/serial"
)

// ModbusRTUNode implements Modbus RTU client over serial
type ModbusRTUNode struct {
	port      string
	baudRate  int
	dataBits  int
	stopBits  int
	parity    string // none, odd, even
	unitID    byte
	timeout   time.Duration
	operation string
	address   uint16
	quantity  uint16
	serialPort serial.Port
	mu        sync.Mutex
}

// NewModbusRTUNode creates a new Modbus RTU node
func NewModbusRTUNode() *ModbusRTUNode {
	return &ModbusRTUNode{
		port:      "/dev/ttyUSB0",
		baudRate:  9600,
		dataBits:  8,
		stopBits:  1,
		parity:    "none",
		unitID:    1,
		timeout:   1 * time.Second,
		operation: "read_holding",
		address:   0,
		quantity:  1,
	}
}

// Init initializes the Modbus RTU node
func (n *ModbusRTUNode) Init(config map[string]interface{}) error {
	if port, ok := config["port"].(string); ok {
		n.port = port
	}
	if baud, ok := config["baudRate"].(float64); ok {
		n.baudRate = int(baud)
	}
	if dataBits, ok := config["dataBits"].(float64); ok {
		n.dataBits = int(dataBits)
	}
	if stopBits, ok := config["stopBits"].(float64); ok {
		n.stopBits = int(stopBits)
	}
	if parity, ok := config["parity"].(string); ok {
		n.parity = parity
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

// Execute performs Modbus RTU operation
func (n *ModbusRTUNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Get operation parameters from message
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

	// Open serial port if not open
	if n.serialPort == nil {
		err := n.openPort()
		if err != nil {
			return msg, fmt.Errorf("failed to open serial port: %w", err)
		}
	}

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
		n.closePort()
		return msg, err
	}

	msg.Payload["result"] = result
	msg.Payload["operation"] = operation
	msg.Payload["address"] = address
	msg.Payload["unitId"] = n.unitID

	return msg, nil
}

// openPort opens the serial port
func (n *ModbusRTUNode) openPort() error {
	mode := &serial.Mode{
		BaudRate: n.baudRate,
		DataBits: n.dataBits,
	}

	switch n.stopBits {
	case 1:
		mode.StopBits = serial.OneStopBit
	case 2:
		mode.StopBits = serial.TwoStopBits
	}

	switch n.parity {
	case "none":
		mode.Parity = serial.NoParity
	case "odd":
		mode.Parity = serial.OddParity
	case "even":
		mode.Parity = serial.EvenParity
	}

	port, err := serial.Open(n.port, mode)
	if err != nil {
		return err
	}

	port.SetReadTimeout(n.timeout)
	n.serialPort = port
	return nil
}

// closePort closes the serial port
func (n *ModbusRTUNode) closePort() {
	if n.serialPort != nil {
		n.serialPort.Close()
		n.serialPort = nil
	}
}

// readCoils reads coils via RTU
func (n *ModbusRTUNode) readCoils(address, quantity uint16) ([]bool, error) {
	request := n.buildRequest(FuncReadCoils, address, quantity)
	response, err := n.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if len(response) < 3 {
		return nil, fmt.Errorf("invalid response")
	}

	byteCount := int(response[2])
	if len(response) < 3+byteCount {
		return nil, fmt.Errorf("incomplete response")
	}

	coils := make([]bool, quantity)
	for i := uint16(0); i < quantity; i++ {
		byteIdx := i / 8
		bitIdx := i % 8
		coils[i] = (response[3+byteIdx] & (1 << bitIdx)) != 0
	}

	return coils, nil
}

// readDiscreteInputs reads discrete inputs via RTU
func (n *ModbusRTUNode) readDiscreteInputs(address, quantity uint16) ([]bool, error) {
	request := n.buildRequest(FuncReadDiscreteInputs, address, quantity)
	response, err := n.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if len(response) < 3 {
		return nil, fmt.Errorf("invalid response")
	}

	inputs := make([]bool, quantity)
	for i := uint16(0); i < quantity; i++ {
		byteIdx := i / 8
		bitIdx := i % 8
		inputs[i] = (response[3+byteIdx] & (1 << bitIdx)) != 0
	}

	return inputs, nil
}

// readHoldingRegisters reads holding registers via RTU
func (n *ModbusRTUNode) readHoldingRegisters(address, quantity uint16) ([]uint16, error) {
	request := n.buildRequest(FuncReadHoldingRegs, address, quantity)
	response, err := n.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if len(response) < 3 {
		return nil, fmt.Errorf("invalid response")
	}

	byteCount := int(response[2])
	if len(response) < 3+byteCount {
		return nil, fmt.Errorf("incomplete response")
	}

	registers := make([]uint16, quantity)
	for i := uint16(0); i < quantity; i++ {
		registers[i] = binary.BigEndian.Uint16(response[3+i*2:])
	}

	return registers, nil
}

// readInputRegisters reads input registers via RTU
func (n *ModbusRTUNode) readInputRegisters(address, quantity uint16) ([]uint16, error) {
	request := n.buildRequest(FuncReadInputRegs, address, quantity)
	response, err := n.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if len(response) < 3 {
		return nil, fmt.Errorf("invalid response")
	}

	registers := make([]uint16, quantity)
	for i := uint16(0); i < quantity; i++ {
		registers[i] = binary.BigEndian.Uint16(response[3+i*2:])
	}

	return registers, nil
}

// writeSingleCoil writes a single coil via RTU
func (n *ModbusRTUNode) writeSingleCoil(address uint16, value bool) error {
	var coilValue uint16
	if value {
		coilValue = 0xFF00
	}
	request := n.buildRequest(FuncWriteSingleCoil, address, coilValue)
	_, err := n.sendRequest(request)
	return err
}

// writeSingleRegister writes a single register via RTU
func (n *ModbusRTUNode) writeSingleRegister(address, value uint16) error {
	request := n.buildRequest(FuncWriteSingleReg, address, value)
	_, err := n.sendRequest(request)
	return err
}

// writeMultipleCoils writes multiple coils via RTU
func (n *ModbusRTUNode) writeMultipleCoils(address uint16, values []bool) error {
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

// writeMultipleRegisters writes multiple registers via RTU
func (n *ModbusRTUNode) writeMultipleRegisters(address uint16, values []uint16) error {
	data := make([]byte, len(values)*2)
	for i, v := range values {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	request := n.buildWriteMultiRequest(FuncWriteMultiRegs, address, uint16(len(values)), data)
	_, err := n.sendRequest(request)
	return err
}

// buildRequest builds a Modbus RTU request (without CRC - will be added)
func (n *ModbusRTUNode) buildRequest(funcCode byte, address, value uint16) []byte {
	request := make([]byte, 6)
	request[0] = n.unitID
	request[1] = funcCode
	binary.BigEndian.PutUint16(request[2:], address)
	binary.BigEndian.PutUint16(request[4:], value)
	return n.addCRC(request)
}

// buildWriteMultiRequest builds a write multiple request
func (n *ModbusRTUNode) buildWriteMultiRequest(funcCode byte, address, quantity uint16, data []byte) []byte {
	request := make([]byte, 7+len(data))
	request[0] = n.unitID
	request[1] = funcCode
	binary.BigEndian.PutUint16(request[2:], address)
	binary.BigEndian.PutUint16(request[4:], quantity)
	request[6] = byte(len(data))
	copy(request[7:], data)
	return n.addCRC(request)
}

// sendRequest sends RTU request and receives response
func (n *ModbusRTUNode) sendRequest(request []byte) ([]byte, error) {
	// Clear input buffer
	n.serialPort.ResetInputBuffer()

	// Send request
	_, err := n.serialPort.Write(request)
	if err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	// Wait for response (inter-frame delay)
	time.Sleep(50 * time.Millisecond)

	// Read response
	response := make([]byte, 256)
	totalRead := 0

	for {
		n, err := n.serialPort.Read(response[totalRead:])
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
		totalRead += n
		if totalRead >= 5 { // Minimum response size
			break
		}
	}

	if totalRead < 5 {
		return nil, fmt.Errorf("incomplete response: got %d bytes", totalRead)
	}

	response = response[:totalRead]

	// Verify CRC
	if !n.verifyCRC(response) {
		return nil, fmt.Errorf("CRC error")
	}

	// Check for exception
	if response[1]&0x80 != 0 {
		return nil, fmt.Errorf("modbus exception: %d", response[2])
	}

	// Return without CRC
	return response[:len(response)-2], nil
}

// addCRC adds CRC16 to request
func (n *ModbusRTUNode) addCRC(data []byte) []byte {
	crc := n.calculateCRC(data)
	return append(data, byte(crc&0xFF), byte(crc>>8))
}

// verifyCRC verifies CRC16 in response
func (n *ModbusRTUNode) verifyCRC(data []byte) bool {
	if len(data) < 3 {
		return false
	}
	receivedCRC := uint16(data[len(data)-1])<<8 | uint16(data[len(data)-2])
	calculatedCRC := n.calculateCRC(data[:len(data)-2])
	return receivedCRC == calculatedCRC
}

// calculateCRC calculates CRC16 (Modbus)
func (n *ModbusRTUNode) calculateCRC(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// Cleanup closes the serial port
func (n *ModbusRTUNode) Cleanup() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.closePort()
	return nil
}

// NewModbusRTUExecutor creates a new Modbus RTU executor
func NewModbusRTUExecutor() node.Executor {
	return NewModbusRTUNode()
}
