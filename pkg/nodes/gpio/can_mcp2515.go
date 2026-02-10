//go:build linux
// +build linux

package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/node"
)

// MCP2515 Register addresses
const (
	mcp2515RegRXF0SIDH  = 0x00
	mcp2515RegRXF0SIDL  = 0x01
	mcp2515RegRXF0EID8  = 0x02
	mcp2515RegRXF0EID0  = 0x03
	mcp2515RegBFPCTRL   = 0x0C
	mcp2515RegTXRTSCTRL = 0x0D
	mcp2515RegCANSTAT   = 0x0E
	mcp2515RegCANCTRL   = 0x0F
	mcp2515RegTEC       = 0x1C
	mcp2515RegREC       = 0x1D
	mcp2515RegCNF3      = 0x28
	mcp2515RegCNF2      = 0x29
	mcp2515RegCNF1      = 0x2A
	mcp2515RegCANINTE   = 0x2B
	mcp2515RegCANINTF   = 0x2C
	mcp2515RegEFLG      = 0x2D
	mcp2515RegTXB0CTRL  = 0x30
	mcp2515RegTXB0SIDH  = 0x31
	mcp2515RegTXB1CTRL  = 0x40
	mcp2515RegTXB1SIDH  = 0x41
	mcp2515RegTXB2CTRL  = 0x50
	mcp2515RegTXB2SIDH  = 0x51
	mcp2515RegRXB0CTRL  = 0x60
	mcp2515RegRXB0SIDH  = 0x61
	mcp2515RegRXB1CTRL  = 0x70
)

// MCP2515 SPI instructions
const (
	mcp2515CmdReset       = 0xC0
	mcp2515CmdRead        = 0x03
	mcp2515CmdWrite       = 0x02
	mcp2515CmdReadRXB0    = 0x90
	mcp2515CmdReadRXB1    = 0x94
	mcp2515CmdLoadTXB0    = 0x40
	mcp2515CmdLoadTXB1    = 0x42
	mcp2515CmdLoadTXB2    = 0x44
	mcp2515CmdRTSTXB0     = 0x81
	mcp2515CmdRTSTXB1     = 0x82
	mcp2515CmdRTSTXB2     = 0x84
	mcp2515CmdReadStatus  = 0xA0
	mcp2515CmdRXStatus    = 0xB0
	mcp2515CmdBitModify   = 0x05
)

// MCP2515 Modes
const (
	mcp2515ModeNormal     = 0x00
	mcp2515ModeSleep      = 0x20
	mcp2515ModeLoopback   = 0x40
	mcp2515ModeListenOnly = 0x60
	mcp2515ModeConfig     = 0x80
)

// CANMessage represents a CAN frame
type CANMessage struct {
	ID       uint32 `json:"id"`
	Extended bool   `json:"extended"`
	RTR      bool   `json:"rtr"`
	DLC      uint8  `json:"dlc"`
	Data     []byte `json:"data"`
}

// MCP2515Config configuration for MCP2515 CAN controller
type MCP2515Config struct {
	SPIBus    int `json:"spi_bus"`    // SPI bus number (default: 0)
	SPIDevice int `json:"spi_device"` // SPI device number (default: 0)
	Speed     int `json:"speed"`      // SPI speed in Hz (default: 10MHz)
	CSPin     int `json:"cs_pin"`     // Chip select GPIO pin
	IntPin    int `json:"int_pin"`    // Interrupt GPIO pin (optional)
	Bitrate   int `json:"bitrate"`    // CAN bitrate: 125000, 250000, 500000, 1000000
	Crystal   int `json:"crystal"`    // Crystal frequency: 8000000 or 16000000 (default: 16MHz)
}

// MCP2515Executor executes MCP2515 CAN operations
type MCP2515Executor struct {
	config      MCP2515Config
	hal         hal.HAL
	spi         hal.SPIProvider
	mu          sync.Mutex
	initialized bool
	running     bool
	stopChan    chan struct{}
	rxChan      chan CANMessage
}

// NewMCP2515Executor creates a new MCP2515 executor
func NewMCP2515Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var canConfig MCP2515Config
	if err := json.Unmarshal(configJSON, &canConfig); err != nil {
		return nil, fmt.Errorf("invalid MCP2515 config: %w", err)
	}

	// Defaults
	if canConfig.Speed == 0 {
		canConfig.Speed = 10000000 // 10MHz
	}
	if canConfig.Bitrate == 0 {
		canConfig.Bitrate = 500000 // 500kbps
	}
	if canConfig.Crystal == 0 {
		canConfig.Crystal = 16000000 // 16MHz crystal
	}

	return &MCP2515Executor{
		config:   canConfig,
		stopChan: make(chan struct{}),
		rxChan:   make(chan CANMessage, 32),
	}, nil
}

// Init initializes the MCP2515 executor
func (e *MCP2515Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles CAN operations
func (e *MCP2515Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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

	// Initialize MCP2515
	if !e.initialized {
		if err := e.initMCP2515(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init MCP2515: %w", err)
		}
		e.initialized = true
	}

	// Parse command
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("payload is nil")
	}

	action, _ := payload["action"].(string)

	switch action {
	case "send", "transmit":
		// Send CAN message
		id := uint32(getFloat(payload, "id", 0))
		extended, _ := payload["extended"].(bool)
		rtr, _ := payload["rtr"].(bool)

		var data []byte
		if dataIface, ok := payload["data"].([]interface{}); ok {
			for _, v := range dataIface {
				if b, ok := v.(float64); ok {
					data = append(data, byte(b))
				}
			}
		}

		canMsg := CANMessage{
			ID:       id,
			Extended: extended,
			RTR:      rtr,
			DLC:      uint8(len(data)),
			Data:     data,
		}

		if err := e.sendMessage(canMsg); err != nil {
			return node.Message{}, err
		}

		return node.Message{
			Payload: map[string]interface{}{
				"action":    "send",
				"id":        id,
				"extended":  extended,
				"data":      data,
				"timestamp": time.Now().Unix(),
			},
		}, nil

	case "receive", "read":
		// Read pending CAN message
		canMsg, err := e.receiveMessage()
		if err != nil {
			return node.Message{}, err
		}
		if canMsg == nil {
			return node.Message{
				Payload: map[string]interface{}{
					"action":    "receive",
					"available": false,
					"timestamp": time.Now().Unix(),
				},
			}, nil
		}

		return node.Message{
			Payload: map[string]interface{}{
				"action":    "receive",
				"available": true,
				"id":        canMsg.ID,
				"extended":  canMsg.Extended,
				"rtr":       canMsg.RTR,
				"dlc":       canMsg.DLC,
				"data":      canMsg.Data,
				"timestamp": time.Now().Unix(),
			},
		}, nil

	case "status":
		return e.getStatus()

	case "set_filter":
		// Set acceptance filter
		filterNum := int(getFloat(payload, "filter", 0))
		filterID := uint32(getFloat(payload, "id", 0))
		mask := uint32(getFloat(payload, "mask", 0x7FF))
		extended, _ := payload["extended"].(bool)

		if err := e.setFilter(filterNum, filterID, mask, extended); err != nil {
			return node.Message{}, err
		}

		return node.Message{
			Payload: map[string]interface{}{
				"action":    "set_filter",
				"filter":    filterNum,
				"id":        filterID,
				"mask":      mask,
				"timestamp": time.Now().Unix(),
			},
		}, nil

	case "set_mode":
		// Set operating mode
		mode, _ := payload["mode"].(string)
		if err := e.setMode(mode); err != nil {
			return node.Message{}, err
		}

		return node.Message{
			Payload: map[string]interface{}{
				"action":    "set_mode",
				"mode":      mode,
				"timestamp": time.Now().Unix(),
			},
		}, nil

	case "reset":
		if err := e.reset(); err != nil {
			return node.Message{}, err
		}
		e.initialized = false

		return node.Message{
			Payload: map[string]interface{}{
				"action":    "reset",
				"timestamp": time.Now().Unix(),
			},
		}, nil

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// initMCP2515 initializes the MCP2515 CAN controller
func (e *MCP2515Executor) initMCP2515() error {
	// Open SPI device via HAL
	spi := e.hal.SPI()
	if err := spi.Open(e.config.SPIBus, e.config.SPIDevice); err != nil {
		return fmt.Errorf("failed to open SPI device: %w", err)
	}
	e.spi = spi

	// Reset MCP2515
	if _, err := e.spiTransfer([]byte{mcp2515CmdReset}); err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}
	time.Sleep(10 * time.Millisecond)

	// Enter configuration mode
	if err := e.writeRegister(mcp2515RegCANCTRL, mcp2515ModeConfig); err != nil {
		return err
	}

	// Wait for config mode
	for i := 0; i < 10; i++ {
		stat, err := e.readRegister(mcp2515RegCANSTAT)
		if err != nil {
			return err
		}
		if stat&0xE0 == mcp2515ModeConfig {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Configure bit timing
	cnf1, cnf2, cnf3 := e.calculateBitTiming()
	if err := e.writeRegister(mcp2515RegCNF1, cnf1); err != nil {
		return err
	}
	if err := e.writeRegister(mcp2515RegCNF2, cnf2); err != nil {
		return err
	}
	if err := e.writeRegister(mcp2515RegCNF3, cnf3); err != nil {
		return err
	}

	// Configure RX buffers - receive all messages
	if err := e.writeRegister(mcp2515RegRXB0CTRL, 0x60); err != nil { // Turn mask/filters off
		return err
	}
	if err := e.writeRegister(mcp2515RegRXB1CTRL, 0x60); err != nil {
		return err
	}

	// Enable interrupts (RX buffer full)
	if err := e.writeRegister(mcp2515RegCANINTE, 0x03); err != nil {
		return err
	}

	// Enter normal mode
	if err := e.writeRegister(mcp2515RegCANCTRL, mcp2515ModeNormal); err != nil {
		return err
	}

	return nil
}

// spiTransfer performs an SPI transfer using the HAL provider
func (e *MCP2515Executor) spiTransfer(data []byte) ([]byte, error) {
	return e.spi.Transfer(data)
}

// calculateBitTiming calculates CNF1, CNF2, CNF3 for desired bitrate
func (e *MCP2515Executor) calculateBitTiming() (byte, byte, byte) {
	// Timing configurations for 16MHz crystal
	// Format: CNF1, CNF2, CNF3
	if e.config.Crystal == 16000000 {
		switch e.config.Bitrate {
		case 1000000: // 1Mbps
			return 0x00, 0x80, 0x80
		case 500000: // 500kbps
			return 0x00, 0x90, 0x82
		case 250000: // 250kbps
			return 0x00, 0xB1, 0x85
		case 125000: // 125kbps
			return 0x01, 0xB1, 0x85
		case 100000: // 100kbps
			return 0x01, 0xB4, 0x86
		case 50000: // 50kbps
			return 0x03, 0xB4, 0x86
		default: // Default 500kbps
			return 0x00, 0x90, 0x82
		}
	}

	// 8MHz crystal
	switch e.config.Bitrate {
	case 1000000:
		return 0x00, 0x80, 0x00
	case 500000:
		return 0x00, 0x90, 0x02
	case 250000:
		return 0x00, 0xB1, 0x05
	case 125000:
		return 0x01, 0xB1, 0x05
	default:
		return 0x00, 0x90, 0x02
	}
}

// sendMessage sends a CAN message
func (e *MCP2515Executor) sendMessage(msg CANMessage) error {
	// Check TX buffer 0 is available
	status, err := e.readStatus()
	if err != nil {
		return err
	}

	// Find free TX buffer
	var txCtrl, txSidh byte
	var loadCmd, rtsCmd byte

	if status&0x04 == 0 { // TXB0 free
		txCtrl = mcp2515RegTXB0CTRL
		txSidh = mcp2515RegTXB0SIDH
		loadCmd = mcp2515CmdLoadTXB0
		rtsCmd = mcp2515CmdRTSTXB0
	} else if status&0x10 == 0 { // TXB1 free
		txCtrl = mcp2515RegTXB1CTRL
		txSidh = mcp2515RegTXB1SIDH
		loadCmd = mcp2515CmdLoadTXB1
		rtsCmd = mcp2515CmdRTSTXB1
	} else if status&0x40 == 0 { // TXB2 free
		txCtrl = mcp2515RegTXB2CTRL
		txSidh = mcp2515RegTXB2SIDH
		loadCmd = mcp2515CmdLoadTXB2
		rtsCmd = mcp2515CmdRTSTXB2
	} else {
		return fmt.Errorf("no free TX buffer")
	}

	_ = txCtrl // Used for priority setting if needed

	// Build TX buffer data
	var sidh, sidl, eid8, eid0 byte
	if msg.Extended {
		// Extended ID (29-bit)
		sidh = byte((msg.ID >> 21) & 0xFF)
		sidl = byte(((msg.ID >> 13) & 0xE0) | 0x08 | ((msg.ID >> 16) & 0x03))
		eid8 = byte((msg.ID >> 8) & 0xFF)
		eid0 = byte(msg.ID & 0xFF)
	} else {
		// Standard ID (11-bit)
		sidh = byte((msg.ID >> 3) & 0xFF)
		sidl = byte((msg.ID << 5) & 0xE0)
		eid8 = 0
		eid0 = 0
	}

	dlc := msg.DLC
	if dlc > 8 {
		dlc = 8
	}
	if msg.RTR {
		dlc |= 0x40
	}

	// Load TX buffer using load instruction
	txData := make([]byte, 14)
	txData[0] = loadCmd
	txData[1] = sidh
	txData[2] = sidl
	txData[3] = eid8
	txData[4] = eid0
	txData[5] = dlc
	copy(txData[6:], msg.Data)

	if _, err := e.spiTransfer(txData); err != nil {
		return fmt.Errorf("failed to load TX buffer: %w", err)
	}

	// Request to send
	if _, err := e.spiTransfer([]byte{rtsCmd}); err != nil {
		return fmt.Errorf("failed to send RTS: %w", err)
	}

	// Wait for transmission complete
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("transmission timeout")
		default:
			ctrl, _ := e.readRegister(txSidh - 1) // TX control register is before SIDH
			if ctrl&0x08 == 0 {                   // TXREQ cleared
				if ctrl&0x70 != 0 { // Check for errors
					return fmt.Errorf("transmission error: 0x%02X", ctrl)
				}
				return nil
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// receiveMessage receives a CAN message if available
func (e *MCP2515Executor) receiveMessage() (*CANMessage, error) {
	// Check RX status
	rxStatus, err := e.rxStatus()
	if err != nil {
		return nil, err
	}

	var readCmd byte
	var rxCtrl byte

	if rxStatus&0x40 != 0 { // Message in RXB0
		readCmd = mcp2515CmdReadRXB0
		rxCtrl = mcp2515RegRXB0CTRL
	} else if rxStatus&0x80 != 0 { // Message in RXB1
		readCmd = mcp2515CmdReadRXB1
		rxCtrl = mcp2515RegRXB1CTRL
	} else {
		return nil, nil // No message available
	}

	// Read RX buffer
	txData := make([]byte, 14)
	txData[0] = readCmd

	rxData, err := e.spiTransfer(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to read RX buffer: %w", err)
	}

	// Parse message
	sidh := rxData[1]
	sidl := rxData[2]
	eid8 := rxData[3]
	eid0 := rxData[4]
	dlc := rxData[5]

	canMsg := &CANMessage{}

	if sidl&0x08 != 0 { // Extended frame
		canMsg.Extended = true
		canMsg.ID = uint32(sidh)<<21 | uint32(sidl&0xE0)<<13 | uint32(sidl&0x03)<<16 | uint32(eid8)<<8 | uint32(eid0)
	} else { // Standard frame
		canMsg.Extended = false
		canMsg.ID = uint32(sidh)<<3 | uint32(sidl>>5)
	}

	canMsg.RTR = dlc&0x40 != 0
	canMsg.DLC = dlc & 0x0F
	if canMsg.DLC > 8 {
		canMsg.DLC = 8
	}

	canMsg.Data = make([]byte, canMsg.DLC)
	copy(canMsg.Data, rxData[6:6+canMsg.DLC])

	// Clear interrupt flag
	if rxCtrl == mcp2515RegRXB0CTRL {
		e.bitModify(mcp2515RegCANINTF, 0x01, 0x00)
	} else {
		e.bitModify(mcp2515RegCANINTF, 0x02, 0x00)
	}

	return canMsg, nil
}

// setFilter sets a CAN acceptance filter
func (e *MCP2515Executor) setFilter(filterNum int, id uint32, mask uint32, extended bool) error {
	// Enter config mode
	if err := e.writeRegister(mcp2515RegCANCTRL, mcp2515ModeConfig); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)

	// Filter register addresses
	filterRegs := []byte{0x00, 0x04, 0x08, 0x10, 0x14, 0x18} // RXF0-RXF5
	maskRegs := []byte{0x20, 0x24}                           // RXM0, RXM1

	if filterNum < 0 || filterNum > 5 {
		filterNum = 0
	}

	filterBase := filterRegs[filterNum]
	maskBase := maskRegs[0]
	if filterNum >= 3 {
		maskBase = maskRegs[1]
	}

	// Write filter
	var sidh, sidl, eid8, eid0 byte
	if extended {
		sidh = byte((id >> 21) & 0xFF)
		sidl = byte(((id >> 13) & 0xE0) | 0x08 | ((id >> 16) & 0x03))
		eid8 = byte((id >> 8) & 0xFF)
		eid0 = byte(id & 0xFF)
	} else {
		sidh = byte((id >> 3) & 0xFF)
		sidl = byte((id << 5) & 0xE0)
	}

	e.writeRegister(filterBase, sidh)
	e.writeRegister(filterBase+1, sidl)
	e.writeRegister(filterBase+2, eid8)
	e.writeRegister(filterBase+3, eid0)

	// Write mask
	if extended {
		sidh = byte((mask >> 21) & 0xFF)
		sidl = byte(((mask >> 13) & 0xE0) | ((mask >> 16) & 0x03))
		eid8 = byte((mask >> 8) & 0xFF)
		eid0 = byte(mask & 0xFF)
	} else {
		sidh = byte((mask >> 3) & 0xFF)
		sidl = byte((mask << 5) & 0xE0)
	}

	e.writeRegister(maskBase, sidh)
	e.writeRegister(maskBase+1, sidl)
	e.writeRegister(maskBase+2, eid8)
	e.writeRegister(maskBase+3, eid0)

	// Enable filters on RX buffer
	e.writeRegister(mcp2515RegRXB0CTRL, 0x00)
	e.writeRegister(mcp2515RegRXB1CTRL, 0x00)

	// Return to normal mode
	return e.writeRegister(mcp2515RegCANCTRL, mcp2515ModeNormal)
}

// setMode sets the operating mode
func (e *MCP2515Executor) setMode(mode string) error {
	var modeVal byte
	switch mode {
	case "normal":
		modeVal = mcp2515ModeNormal
	case "sleep":
		modeVal = mcp2515ModeSleep
	case "loopback":
		modeVal = mcp2515ModeLoopback
	case "listen":
		modeVal = mcp2515ModeListenOnly
	case "config":
		modeVal = mcp2515ModeConfig
	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}

	return e.writeRegister(mcp2515RegCANCTRL, modeVal)
}

// getStatus returns CAN bus status
func (e *MCP2515Executor) getStatus() (node.Message, error) {
	canStat, _ := e.readRegister(mcp2515RegCANSTAT)
	canCtrl, _ := e.readRegister(mcp2515RegCANCTRL)
	errFlag, _ := e.readRegister(mcp2515RegEFLG)
	tec, _ := e.readRegister(mcp2515RegTEC)
	rec, _ := e.readRegister(mcp2515RegREC)
	intFlag, _ := e.readRegister(mcp2515RegCANINTF)

	_ = canCtrl // Available for future use

	// Decode mode
	var mode string
	switch canStat & 0xE0 {
	case mcp2515ModeNormal:
		mode = "normal"
	case mcp2515ModeSleep:
		mode = "sleep"
	case mcp2515ModeLoopback:
		mode = "loopback"
	case mcp2515ModeListenOnly:
		mode = "listen"
	case mcp2515ModeConfig:
		mode = "config"
	}

	return node.Message{
		Payload: map[string]interface{}{
			"mode":              mode,
			"tx_error_count":    tec,
			"rx_error_count":    rec,
			"error_warning":     errFlag&0x01 != 0,
			"rx_error_warning":  errFlag&0x02 != 0,
			"tx_error_warning":  errFlag&0x04 != 0,
			"rx_error_passive":  errFlag&0x08 != 0,
			"tx_error_passive":  errFlag&0x10 != 0,
			"bus_off":           errFlag&0x20 != 0,
			"rx0_overflow":      errFlag&0x40 != 0,
			"rx1_overflow":      errFlag&0x80 != 0,
			"interrupt_flags":   intFlag,
			"bitrate":           e.config.Bitrate,
			"sensor":            "MCP2515",
			"timestamp":         time.Now().Unix(),
		},
	}, nil
}

// reset resets the MCP2515
func (e *MCP2515Executor) reset() error {
	_, err := e.spiTransfer([]byte{mcp2515CmdReset})
	return err
}

// readRegister reads a register
func (e *MCP2515Executor) readRegister(reg byte) (byte, error) {
	data := []byte{mcp2515CmdRead, reg, 0x00}
	rxData, err := e.spiTransfer(data)
	if err != nil {
		return 0, err
	}
	return rxData[2], nil
}

// writeRegister writes a register
func (e *MCP2515Executor) writeRegister(reg, value byte) error {
	_, err := e.spiTransfer([]byte{mcp2515CmdWrite, reg, value})
	return err
}

// bitModify modifies bits in a register
func (e *MCP2515Executor) bitModify(reg, mask, data byte) error {
	_, err := e.spiTransfer([]byte{mcp2515CmdBitModify, reg, mask, data})
	return err
}

// readStatus reads the status register
func (e *MCP2515Executor) readStatus() (byte, error) {
	data := []byte{mcp2515CmdReadStatus, 0x00}
	rxData, err := e.spiTransfer(data)
	if err != nil {
		return 0, err
	}
	return rxData[1], nil
}

// rxStatus reads the RX status
func (e *MCP2515Executor) rxStatus() (byte, error) {
	data := []byte{mcp2515CmdRXStatus, 0x00}
	rxData, err := e.spiTransfer(data)
	if err != nil {
		return 0, err
	}
	return rxData[1], nil
}

// Cleanup releases resources
func (e *MCP2515Executor) Cleanup() error {
	if e.running {
		close(e.stopChan)
	}
	if e.spi != nil {
		e.spi.Close()
	}
	return nil
}
