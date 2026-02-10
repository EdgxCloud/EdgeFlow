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

// RC522 Register addresses
const (
	rc522RegCommand       = 0x01
	rc522RegComIEn        = 0x02
	rc522RegDivIEn        = 0x03
	rc522RegComIrq        = 0x04
	rc522RegDivIrq        = 0x05
	rc522RegError         = 0x06
	rc522RegStatus1       = 0x07
	rc522RegStatus2       = 0x08
	rc522RegFIFOData      = 0x09
	rc522RegFIFOLevel     = 0x0A
	rc522RegWaterLevel    = 0x0B
	rc522RegControl       = 0x0C
	rc522RegBitFraming    = 0x0D
	rc522RegColl          = 0x0E
	rc522RegMode          = 0x11
	rc522RegTxMode        = 0x12
	rc522RegRxMode        = 0x13
	rc522RegTxControl     = 0x14
	rc522RegTxASK         = 0x15
	rc522RegTxSel         = 0x16
	rc522RegRxSel         = 0x17
	rc522RegRxThreshold   = 0x18
	rc522RegDemod         = 0x19
	rc522RegMfTx          = 0x1C
	rc522RegMfRx          = 0x1D
	rc522RegSerialSpeed   = 0x1F
	rc522RegCRCResultH    = 0x21
	rc522RegCRCResultL    = 0x22
	rc522RegModWidth      = 0x24
	rc522RegRFCfg         = 0x26
	rc522RegGsN           = 0x27
	rc522RegCWGsP         = 0x28
	rc522RegModGsP        = 0x29
	rc522RegTMode         = 0x2A
	rc522RegTPrescaler    = 0x2B
	rc522RegTReloadH      = 0x2C
	rc522RegTReloadL      = 0x2D
	rc522RegTCounterValH  = 0x2E
	rc522RegTCounterValL  = 0x2F
	rc522RegTestSel1      = 0x31
	rc522RegTestSel2      = 0x32
	rc522RegTestPinEn     = 0x33
	rc522RegTestPinValue  = 0x34
	rc522RegTestBus       = 0x35
	rc522RegAutoTest      = 0x36
	rc522RegVersion       = 0x37
	rc522RegAnalogTest    = 0x38
	rc522RegTestDAC1      = 0x39
	rc522RegTestDAC2      = 0x3A
	rc522RegTestADC       = 0x3B
)

// RC522 Commands
const (
	rc522CmdIdle           = 0x00
	rc522CmdMem            = 0x01
	rc522CmdGenerateRandID = 0x02
	rc522CmdCalcCRC        = 0x03
	rc522CmdTransmit       = 0x04
	rc522CmdNoCmdChange    = 0x07
	rc522CmdReceive        = 0x08
	rc522CmdTransceive     = 0x0C
	rc522CmdMFAuthent      = 0x0E
	rc522CmdSoftReset      = 0x0F
)

// PICC Commands
const (
	piccCmdREQA       = 0x26
	piccCmdWUPA       = 0x52
	piccCmdCT         = 0x88 // Cascade tag
	piccCmdSELCL1     = 0x93
	piccCmdSELCL2     = 0x95
	piccCmdSELCL3     = 0x97
	piccCmdHLTA       = 0x50
	piccCmdMFAuthKeyA = 0x60
	piccCmdMFAuthKeyB = 0x61
	piccCmdMFRead     = 0x30
	piccCmdMFWrite    = 0xA0
	piccCmdMFDecr     = 0xC0
	piccCmdMFIncr     = 0xC1
	piccCmdMFRestore  = 0xC2
	piccCmdMFTransfer = 0xB0
	piccCmdULWrite    = 0xA2
)

// RC522Config configuration for RC522 RFID reader
type RC522Config struct {
	SPIBus    int `json:"spi_bus"`    // SPI bus number (default: 0)
	SPIDevice int `json:"spi_device"` // SPI device number (default: 0)
	Speed     int `json:"speed"`      // SPI speed in Hz (default: 1MHz)
	ResetPin  int `json:"reset_pin"`  // Reset GPIO pin (optional)
	IRQPin    int `json:"irq_pin"`    // IRQ GPIO pin (optional)
}

// RC522Executor executes RC522 RFID operations
type RC522Executor struct {
	config      RC522Config
	hal         hal.HAL
	spi         hal.SPIProvider
	mu          sync.Mutex
	initialized bool
}

// NewRC522Executor creates a new RC522 executor
func NewRC522Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var rfidConfig RC522Config
	if err := json.Unmarshal(configJSON, &rfidConfig); err != nil {
		return nil, fmt.Errorf("invalid RC522 config: %w", err)
	}

	// Defaults
	if rfidConfig.Speed == 0 {
		rfidConfig.Speed = 1000000 // 1MHz
	}

	return &RC522Executor{
		config: rfidConfig,
	}, nil
}

// Init initializes the RC522 executor
func (e *RC522Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles RFID operations
func (e *RC522Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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

	// Initialize RC522
	if !e.initialized {
		if err := e.initRC522(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init RC522: %w", err)
		}
		e.initialized = true
	}

	// Parse command
	payload := msg.Payload
	if payload == nil {
		// Default: scan for card
		return e.scanCard()
	}

	action, _ := payload["action"].(string)

	switch action {
	case "scan", "read", "":
		return e.scanCard()

	case "read_uid":
		return e.readUID()

	case "read_block":
		block := int(getFloat(payload, "block", 0))
		keyType, _ := payload["key_type"].(string)
		var key []byte
		if keyIface, ok := payload["key"].([]interface{}); ok {
			for _, v := range keyIface {
				if b, ok := v.(float64); ok {
					key = append(key, byte(b))
				}
			}
		}
		return e.readBlock(block, keyType, key)

	case "write_block":
		block := int(getFloat(payload, "block", 0))
		keyType, _ := payload["key_type"].(string)
		var key []byte
		if keyIface, ok := payload["key"].([]interface{}); ok {
			for _, v := range keyIface {
				if b, ok := v.(float64); ok {
					key = append(key, byte(b))
				}
			}
		}
		var data []byte
		if dataIface, ok := payload["data"].([]interface{}); ok {
			for _, v := range dataIface {
				if b, ok := v.(float64); ok {
					data = append(data, byte(b))
				}
			}
		}
		return e.writeBlock(block, keyType, key, data)

	case "halt":
		return e.haltCard()

	case "antenna":
		on, _ := payload["on"].(bool)
		return e.setAntenna(on)

	case "version":
		return e.getVersion()

	case "reset":
		return e.reset()

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// initRC522 initializes the RC522
func (e *RC522Executor) initRC522() error {
	// Open SPI device
	e.spi = e.hal.SPI()
	if err := e.spi.Open(e.config.SPIBus, e.config.SPIDevice); err != nil {
		return fmt.Errorf("failed to open SPI bus %d device %d: %w", e.config.SPIBus, e.config.SPIDevice, err)
	}
	if err := e.spi.SetSpeed(e.config.Speed); err != nil {
		return fmt.Errorf("failed to set SPI speed: %w", err)
	}

	gpio := e.hal.GPIO()

	// Setup reset pin if specified
	if e.config.ResetPin > 0 {
		gpio.SetMode(e.config.ResetPin, hal.Output)
		gpio.DigitalWrite(e.config.ResetPin, false)
		time.Sleep(50 * time.Millisecond)
		gpio.DigitalWrite(e.config.ResetPin, true)
		time.Sleep(50 * time.Millisecond)
	}

	// Soft reset
	e.writeRegister(rc522RegCommand, rc522CmdSoftReset)
	time.Sleep(50 * time.Millisecond)

	// Check version
	version, err := e.readRegister(rc522RegVersion)
	if err != nil {
		return err
	}
	if version != 0x91 && version != 0x92 {
		return fmt.Errorf("unexpected RC522 version: 0x%02X", version)
	}

	// Timer: TPrescaler*TReloadVal/6.78MHz = 24ms
	e.writeRegister(rc522RegTMode, 0x8D)       // TAuto=1, TGated=0, TAutoRestart=0, TPrescaler_Hi=0x0D
	e.writeRegister(rc522RegTPrescaler, 0x3E)  // TPrescaler_Lo
	e.writeRegister(rc522RegTReloadH, 0x00)    // TReloadVal = 30
	e.writeRegister(rc522RegTReloadL, 0x1E)

	// Force 100% ASK modulation
	e.writeRegister(rc522RegTxASK, 0x40)

	// Set CRC preset to 0x6363
	e.writeRegister(rc522RegMode, 0x3D)

	// Enable antenna
	e.setAntennaOn()

	return nil
}

// scanCard scans for RFID card and returns UID
func (e *RC522Executor) scanCard() (node.Message, error) {
	// Request card
	status, atqa := e.request(piccCmdREQA)
	if status != 0 {
		return node.Message{
			Payload: map[string]interface{}{
				"card_present": false,
				"sensor":       "RC522",
				"timestamp":    time.Now().Unix(),
			},
		}, nil
	}

	// Anti-collision and select
	status, uid := e.antiCollision()
	if status != 0 {
		return node.Message{
			Payload: map[string]interface{}{
				"card_present": false,
				"error":        "anti-collision failed",
				"sensor":       "RC522",
				"timestamp":    time.Now().Unix(),
			},
		}, nil
	}

	// Determine card type from ATQA
	cardType := e.getCardType(atqa)

	return node.Message{
		Payload: map[string]interface{}{
			"card_present": true,
			"uid":          uid,
			"uid_hex":      fmt.Sprintf("%X", uid),
			"uid_decimal":  e.uidToDecimal(uid),
			"atqa":         atqa,
			"card_type":    cardType,
			"sensor":       "RC522",
			"timestamp":    time.Now().Unix(),
		},
	}, nil
}

// readUID reads just the UID
func (e *RC522Executor) readUID() (node.Message, error) {
	return e.scanCard()
}

// readBlock reads a block from the card
func (e *RC522Executor) readBlock(block int, keyType string, key []byte) (node.Message, error) {
	// Default key
	if len(key) == 0 {
		key = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	}

	// Request and select card
	status, _ := e.request(piccCmdREQA)
	if status != 0 {
		return node.Message{}, fmt.Errorf("no card present")
	}

	status, uid := e.antiCollision()
	if status != 0 {
		return node.Message{}, fmt.Errorf("anti-collision failed")
	}

	// Select card
	e.selectCard(uid)

	// Authenticate
	authCmd := byte(piccCmdMFAuthKeyA)
	if keyType == "B" {
		authCmd = byte(piccCmdMFAuthKeyB)
	}

	if err := e.authenticate(authCmd, block, key, uid); err != nil {
		return node.Message{}, fmt.Errorf("authentication failed: %w", err)
	}

	// Read block
	data, err := e.mifareRead(block)
	if err != nil {
		return node.Message{}, fmt.Errorf("read failed: %w", err)
	}

	// Stop crypto
	e.stopCrypto()

	return node.Message{
		Payload: map[string]interface{}{
			"block":     block,
			"data":      data,
			"data_hex":  fmt.Sprintf("%X", data),
			"uid":       uid,
			"sensor":    "RC522",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// writeBlock writes data to a block
func (e *RC522Executor) writeBlock(block int, keyType string, key, data []byte) (node.Message, error) {
	// Default key
	if len(key) == 0 {
		key = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	}

	// Pad data to 16 bytes
	if len(data) < 16 {
		padded := make([]byte, 16)
		copy(padded, data)
		data = padded
	}

	// Request and select card
	status, _ := e.request(piccCmdREQA)
	if status != 0 {
		return node.Message{}, fmt.Errorf("no card present")
	}

	status, uid := e.antiCollision()
	if status != 0 {
		return node.Message{}, fmt.Errorf("anti-collision failed")
	}

	e.selectCard(uid)

	// Authenticate
	authCmd := byte(piccCmdMFAuthKeyA)
	if keyType == "B" {
		authCmd = byte(piccCmdMFAuthKeyB)
	}

	if err := e.authenticate(authCmd, block, key, uid); err != nil {
		return node.Message{}, fmt.Errorf("authentication failed: %w", err)
	}

	// Write block
	if err := e.mifareWrite(block, data[:16]); err != nil {
		return node.Message{}, fmt.Errorf("write failed: %w", err)
	}

	e.stopCrypto()

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "write_block",
			"block":     block,
			"uid":       uid,
			"sensor":    "RC522",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// request sends REQA or WUPA command
func (e *RC522Executor) request(cmd byte) (int, []byte) {
	e.writeRegister(rc522RegBitFraming, 0x07) // 7 bits

	result, backData := e.communicate(rc522CmdTransceive, []byte{cmd})
	if result != 0 || len(backData) != 2 {
		return 1, nil
	}

	return 0, backData
}

// antiCollision performs anti-collision
func (e *RC522Executor) antiCollision() (int, []byte) {
	e.writeRegister(rc522RegBitFraming, 0x00)

	cmd := []byte{piccCmdSELCL1, 0x20}
	result, backData := e.communicate(rc522CmdTransceive, cmd)

	if result != 0 || len(backData) != 5 {
		return 1, nil
	}

	// Check BCC
	bcc := byte(0)
	for i := 0; i < 4; i++ {
		bcc ^= backData[i]
	}
	if bcc != backData[4] {
		return 2, nil
	}

	return 0, backData[:4]
}

// selectCard selects a card
func (e *RC522Executor) selectCard(uid []byte) {
	cmd := make([]byte, 9)
	cmd[0] = piccCmdSELCL1
	cmd[1] = 0x70
	copy(cmd[2:6], uid)

	// Calculate BCC
	bcc := byte(0)
	for i := 0; i < 4; i++ {
		bcc ^= uid[i]
	}
	cmd[6] = bcc

	// Calculate CRC
	crc := e.calculateCRC(cmd[:7])
	cmd[7] = crc[0]
	cmd[8] = crc[1]

	e.communicate(rc522CmdTransceive, cmd)
}

// authenticate authenticates with card
func (e *RC522Executor) authenticate(authMode byte, block int, key, uid []byte) error {
	cmd := make([]byte, 12)
	cmd[0] = authMode
	cmd[1] = byte(block)
	copy(cmd[2:8], key)
	copy(cmd[8:12], uid)

	result, _ := e.communicate(rc522CmdMFAuthent, cmd)
	if result != 0 {
		return fmt.Errorf("auth failed")
	}

	// Check MFCrypto1On
	status, _ := e.readRegister(rc522RegStatus2)
	if status&0x08 == 0 {
		return fmt.Errorf("crypto not enabled")
	}

	return nil
}

// mifareRead reads a block
func (e *RC522Executor) mifareRead(block int) ([]byte, error) {
	cmd := []byte{piccCmdMFRead, byte(block)}
	crc := e.calculateCRC(cmd)
	cmd = append(cmd, crc...)

	result, backData := e.communicate(rc522CmdTransceive, cmd)
	if result != 0 || len(backData) != 16 {
		return nil, fmt.Errorf("read failed")
	}

	return backData, nil
}

// mifareWrite writes a block
func (e *RC522Executor) mifareWrite(block int, data []byte) error {
	cmd := []byte{piccCmdMFWrite, byte(block)}
	crc := e.calculateCRC(cmd)
	cmd = append(cmd, crc...)

	result, backData := e.communicate(rc522CmdTransceive, cmd)
	if result != 0 || len(backData) != 1 || backData[0]&0x0F != 0x0A {
		return fmt.Errorf("write cmd failed")
	}

	writeData := make([]byte, 18)
	copy(writeData, data)
	crc = e.calculateCRC(writeData[:16])
	writeData[16] = crc[0]
	writeData[17] = crc[1]

	result, backData = e.communicate(rc522CmdTransceive, writeData)
	if result != 0 || len(backData) != 1 || backData[0]&0x0F != 0x0A {
		return fmt.Errorf("write data failed")
	}

	return nil
}

// stopCrypto stops crypto operations
func (e *RC522Executor) stopCrypto() {
	status, _ := e.readRegister(rc522RegStatus2)
	e.writeRegister(rc522RegStatus2, status&^0x08)
}

// haltCard sends HALT command
func (e *RC522Executor) haltCard() (node.Message, error) {
	cmd := []byte{piccCmdHLTA, 0x00}
	crc := e.calculateCRC(cmd)
	cmd = append(cmd, crc...)

	e.communicate(rc522CmdTransceive, cmd)

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "halt",
			"sensor":    "RC522",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// communicate performs command communication
func (e *RC522Executor) communicate(cmd byte, data []byte) (int, []byte) {
	irqEn := byte(0x00)
	waitIrq := byte(0x00)

	if cmd == rc522CmdMFAuthent {
		irqEn = 0x12
		waitIrq = 0x10
	} else if cmd == rc522CmdTransceive {
		irqEn = 0x77
		waitIrq = 0x30
	}

	e.writeRegister(rc522RegComIEn, irqEn|0x80)
	e.clearBitMask(rc522RegComIrq, 0x80)
	e.setBitMask(rc522RegFIFOLevel, 0x80) // Flush FIFO

	e.writeRegister(rc522RegCommand, rc522CmdIdle)

	// Write data to FIFO
	for _, b := range data {
		e.writeRegister(rc522RegFIFOData, b)
	}

	e.writeRegister(rc522RegCommand, cmd)

	if cmd == rc522CmdTransceive {
		e.setBitMask(rc522RegBitFraming, 0x80)
	}

	// Wait for completion
	timeout := 2000
	for i := 0; i < timeout; i++ {
		irq, _ := e.readRegister(rc522RegComIrq)
		if irq&waitIrq != 0 {
			break
		}
		if irq&0x01 != 0 {
			return 1, nil // Timer expired
		}
		time.Sleep(1 * time.Millisecond)
	}

	e.clearBitMask(rc522RegBitFraming, 0x80)

	// Check for errors
	errReg, _ := e.readRegister(rc522RegError)
	if errReg&0x1B != 0 {
		return 2, nil
	}

	// Read data from FIFO
	fifoLen, _ := e.readRegister(rc522RegFIFOLevel)
	backData := make([]byte, fifoLen)
	for i := byte(0); i < fifoLen; i++ {
		backData[i], _ = e.readRegister(rc522RegFIFOData)
	}

	return 0, backData
}

// calculateCRC calculates CRC-A
func (e *RC522Executor) calculateCRC(data []byte) []byte {
	e.clearBitMask(rc522RegDivIrq, 0x04)
	e.setBitMask(rc522RegFIFOLevel, 0x80)

	for _, b := range data {
		e.writeRegister(rc522RegFIFOData, b)
	}

	e.writeRegister(rc522RegCommand, rc522CmdCalcCRC)

	for i := 0; i < 255; i++ {
		irq, _ := e.readRegister(rc522RegDivIrq)
		if irq&0x04 != 0 {
			break
		}
	}

	crcL, _ := e.readRegister(rc522RegCRCResultL)
	crcH, _ := e.readRegister(rc522RegCRCResultH)

	return []byte{crcL, crcH}
}

// setAntenna enables or disables the antenna
func (e *RC522Executor) setAntenna(on bool) (node.Message, error) {
	if on {
		e.setAntennaOn()
	} else {
		e.setAntennaOff()
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "antenna",
			"antenna":   on,
			"sensor":    "RC522",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// setAntennaOn enables the antenna
func (e *RC522Executor) setAntennaOn() {
	val, _ := e.readRegister(rc522RegTxControl)
	if val&0x03 == 0 {
		e.setBitMask(rc522RegTxControl, 0x03)
	}
}

// setAntennaOff disables the antenna
func (e *RC522Executor) setAntennaOff() {
	e.clearBitMask(rc522RegTxControl, 0x03)
}

// getVersion returns the RC522 version
func (e *RC522Executor) getVersion() (node.Message, error) {
	version, _ := e.readRegister(rc522RegVersion)

	var versionStr string
	switch version {
	case 0x88:
		versionStr = "clone"
	case 0x90:
		versionStr = "v0.0"
	case 0x91:
		versionStr = "v1.0"
	case 0x92:
		versionStr = "v2.0"
	default:
		versionStr = "unknown"
	}

	return node.Message{
		Payload: map[string]interface{}{
			"version":     version,
			"version_str": versionStr,
			"sensor":      "RC522",
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

// reset performs a soft reset
func (e *RC522Executor) reset() (node.Message, error) {
	e.writeRegister(rc522RegCommand, rc522CmdSoftReset)
	time.Sleep(50 * time.Millisecond)
	e.initialized = false

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "reset",
			"sensor":    "RC522",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// getCardType determines card type from ATQA
func (e *RC522Executor) getCardType(atqa []byte) string {
	if len(atqa) < 2 {
		return "unknown"
	}

	switch atqa[0] {
	case 0x04:
		if atqa[1] == 0x00 {
			return "MIFARE_Classic_1K"
		}
	case 0x02:
		if atqa[1] == 0x00 {
			return "MIFARE_Classic_4K"
		}
	case 0x44:
		if atqa[1] == 0x00 {
			return "MIFARE_Ultralight"
		}
	case 0x08:
		if atqa[1] == 0x00 {
			return "MIFARE_Plus"
		}
	}

	return "unknown"
}

// uidToDecimal converts UID bytes to decimal
func (e *RC522Executor) uidToDecimal(uid []byte) uint64 {
	var result uint64
	for i, b := range uid {
		result |= uint64(b) << (8 * (len(uid) - 1 - i))
	}
	return result
}

// readRegister reads a register via SPI
func (e *RC522Executor) readRegister(reg byte) (byte, error) {
	data := []byte{((reg << 1) & 0x7E) | 0x80, 0x00}
	resp, err := e.spi.Transfer(data)
	if err != nil {
		return 0, err
	}
	if len(resp) < 2 {
		return 0, fmt.Errorf("SPI read: short response (%d bytes)", len(resp))
	}
	return resp[1], nil
}

// writeRegister writes a register via SPI
func (e *RC522Executor) writeRegister(reg, value byte) error {
	data := []byte{(reg << 1) & 0x7E, value}
	_, err := e.spi.Transfer(data)
	return err
}

// setBitMask sets bits in a register
func (e *RC522Executor) setBitMask(reg, mask byte) {
	val, _ := e.readRegister(reg)
	e.writeRegister(reg, val|mask)
}

// clearBitMask clears bits in a register
func (e *RC522Executor) clearBitMask(reg, mask byte) {
	val, _ := e.readRegister(reg)
	e.writeRegister(reg, val&^mask)
}

// Cleanup releases resources
func (e *RC522Executor) Cleanup() error {
	e.setAntennaOff()
	if e.spi != nil {
		e.spi.Close()
	}
	return nil
}
