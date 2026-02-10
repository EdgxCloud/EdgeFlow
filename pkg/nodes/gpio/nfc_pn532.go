// +build linux

package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// PN532 Commands
const (
	pn532CmdGetFirmwareVersion = 0x02
	pn532CmdSAMConfiguration   = 0x14
	pn532CmdRFConfiguration    = 0x32
	pn532CmdInListPassiveTarget = 0x4A
	pn532CmdInDataExchange     = 0x40
	pn532CmdInCommunicateThru  = 0x42
	pn532CmdInDeselect         = 0x44
	pn532CmdInRelease          = 0x52
	pn532CmdInSelect           = 0x54
	pn532CmdInAutoPoll         = 0x60

	// MIFARE Commands
	pn532MifareRead    = 0x30
	pn532MifareWrite   = 0xA0
	pn532MifareAuthA   = 0x60
	pn532MifareAuthB   = 0x61

	// Frame markers
	pn532Preamble  = 0x00
	pn532StartCode1 = 0x00
	pn532StartCode2 = 0xFF
	pn532HostToPn532 = 0xD4
	pn532Pn532ToHost = 0xD5
	pn532Ack        = 0x00
	pn532Postamble  = 0x00
)

// PN532Config configuration for PN532 NFC module
type PN532Config struct {
	Bus     string `json:"bus"`     // I2C bus (default: "")
	Address int    `json:"address"` // I2C address (default: 0x24)
	IRQPin  int    `json:"irq_pin"` // IRQ GPIO pin (optional)
}

// PN532Executor executes PN532 NFC operations
type PN532Executor struct {
	config      PN532Config
	bus         i2c.BusCloser
	dev         i2c.Dev
	mu          sync.Mutex
	hostInited  bool
	initialized bool
}

// NewPN532Executor creates a new PN532 executor
func NewPN532Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var nfcConfig PN532Config
	if err := json.Unmarshal(configJSON, &nfcConfig); err != nil {
		return nil, fmt.Errorf("invalid PN532 config: %w", err)
	}

	if nfcConfig.Address == 0 {
		nfcConfig.Address = 0x24
	}

	return &PN532Executor{
		config: nfcConfig,
	}, nil
}

// Init initializes the PN532 executor
func (e *PN532Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles NFC operations
func (e *PN532Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init periph host: %w", err)
		}
		e.hostInited = true
	}

	if e.bus == nil {
		bus, err := i2creg.Open(e.config.Bus)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to open I2C bus: %w", err)
		}
		e.bus = bus
		e.dev = i2c.Dev{Bus: e.bus, Addr: uint16(e.config.Address)}
	}

	if !e.initialized {
		if err := e.initPN532(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init PN532: %w", err)
		}
		e.initialized = true
	}

	payload := msg.Payload
	if payload == nil {
		return e.scanTag()
	}

	action, _ := payload["action"].(string)

	switch action {
	case "scan", "read_uid", "":
		return e.scanTag()

	case "read_block":
		block := int(getFloat(payload, "block", 4))
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
		block := int(getFloat(payload, "block", 4))
		keyType, _ := payload["key_type"].(string)
		var key, data []byte
		if keyIface, ok := payload["key"].([]interface{}); ok {
			for _, v := range keyIface {
				if b, ok := v.(float64); ok {
					key = append(key, byte(b))
				}
			}
		}
		if dataIface, ok := payload["data"].([]interface{}); ok {
			for _, v := range dataIface {
				if b, ok := v.(float64); ok {
					data = append(data, byte(b))
				}
			}
		}
		return e.writeBlock(block, keyType, key, data)

	case "read_ndef":
		return e.readNDEF()

	case "write_ndef":
		text, _ := payload["text"].(string)
		return e.writeNDEF(text)

	case "firmware":
		return e.getFirmwareVersion()

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// initPN532 initializes the PN532
func (e *PN532Executor) initPN532() error {
	// Get firmware version to verify communication
	if _, err := e.sendCommand(pn532CmdGetFirmwareVersion, nil); err != nil {
		return fmt.Errorf("failed to get firmware version: %w", err)
	}

	// Configure SAM (Security Access Module)
	// Mode 0x01 = Normal mode, timeout 0x14 = 1 second
	if _, err := e.sendCommand(pn532CmdSAMConfiguration, []byte{0x01, 0x14, 0x01}); err != nil {
		return fmt.Errorf("failed to configure SAM: %w", err)
	}

	return nil
}

// scanTag scans for NFC tags
func (e *PN532Executor) scanTag() (node.Message, error) {
	// InListPassiveTarget: max 1 target, 106 kbps type A (ISO14443A)
	response, err := e.sendCommand(pn532CmdInListPassiveTarget, []byte{0x01, 0x00})
	if err != nil {
		return node.Message{
			Payload: map[string]interface{}{
				"tag_present": false,
				"sensor":      "PN532",
				"timestamp":   time.Now().Unix(),
			},
		}, nil
	}

	if len(response) < 1 || response[0] == 0 {
		return node.Message{
			Payload: map[string]interface{}{
				"tag_present": false,
				"sensor":      "PN532",
				"timestamp":   time.Now().Unix(),
			},
		}, nil
	}

	// Parse response
	numTargets := response[0]
	if numTargets == 0 || len(response) < 6 {
		return node.Message{
			Payload: map[string]interface{}{
				"tag_present": false,
				"sensor":      "PN532",
				"timestamp":   time.Now().Unix(),
			},
		}, nil
	}

	targetNum := response[1]
	sensRes := (uint16(response[2]) << 8) | uint16(response[3])
	selRes := response[4]
	uidLen := response[5]

	if len(response) < 6+int(uidLen) {
		return node.Message{}, fmt.Errorf("invalid response length")
	}

	uid := response[6 : 6+uidLen]
	tagType := e.getTagType(sensRes, selRes)

	return node.Message{
		Payload: map[string]interface{}{
			"tag_present": true,
			"uid":         uid,
			"uid_hex":     fmt.Sprintf("%X", uid),
			"uid_decimal": e.uidToDecimal(uid),
			"target_num":  targetNum,
			"sens_res":    sensRes,
			"sel_res":     selRes,
			"tag_type":    tagType,
			"sensor":      "PN532",
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

// readBlock reads a block from MIFARE tag
func (e *PN532Executor) readBlock(block int, keyType string, key []byte) (node.Message, error) {
	// First scan for tag
	scanResult, err := e.scanTag()
	if err != nil {
		return node.Message{}, err
	}
	payload := scanResult.Payload
	if !payload["tag_present"].(bool) {
		return node.Message{}, fmt.Errorf("no tag present")
	}
	uid := payload["uid"].([]byte)

	// Default key
	if len(key) == 0 {
		key = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	}

	// Authenticate
	authCmd := byte(pn532MifareAuthA)
	if keyType == "B" {
		authCmd = pn532MifareAuthB
	}

	authData := make([]byte, 2+6+len(uid))
	authData[0] = authCmd
	authData[1] = byte(block)
	copy(authData[2:8], key)
	copy(authData[8:], uid)

	if _, err := e.sendCommand(pn532CmdInDataExchange, append([]byte{0x01}, authData...)); err != nil {
		return node.Message{}, fmt.Errorf("authentication failed: %w", err)
	}

	// Read block
	readData := []byte{0x01, pn532MifareRead, byte(block)}
	response, err := e.sendCommand(pn532CmdInDataExchange, readData)
	if err != nil {
		return node.Message{}, fmt.Errorf("read failed: %w", err)
	}

	if len(response) < 17 {
		return node.Message{}, fmt.Errorf("invalid read response")
	}

	blockData := response[1:17]

	return node.Message{
		Payload: map[string]interface{}{
			"block":     block,
			"data":      blockData,
			"data_hex":  fmt.Sprintf("%X", blockData),
			"uid":       uid,
			"sensor":    "PN532",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// writeBlock writes data to a block
func (e *PN532Executor) writeBlock(block int, keyType string, key, data []byte) (node.Message, error) {
	scanResult, err := e.scanTag()
	if err != nil {
		return node.Message{}, err
	}
	payload := scanResult.Payload
	if !payload["tag_present"].(bool) {
		return node.Message{}, fmt.Errorf("no tag present")
	}
	uid := payload["uid"].([]byte)

	if len(key) == 0 {
		key = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	}

	// Pad data to 16 bytes
	if len(data) < 16 {
		padded := make([]byte, 16)
		copy(padded, data)
		data = padded
	}

	// Authenticate
	authCmd := byte(pn532MifareAuthA)
	if keyType == "B" {
		authCmd = pn532MifareAuthB
	}

	authData := make([]byte, 2+6+len(uid))
	authData[0] = authCmd
	authData[1] = byte(block)
	copy(authData[2:8], key)
	copy(authData[8:], uid)

	if _, err := e.sendCommand(pn532CmdInDataExchange, append([]byte{0x01}, authData...)); err != nil {
		return node.Message{}, fmt.Errorf("authentication failed: %w", err)
	}

	// Write block
	writeData := make([]byte, 19)
	writeData[0] = 0x01
	writeData[1] = pn532MifareWrite
	writeData[2] = byte(block)
	copy(writeData[3:], data[:16])

	if _, err := e.sendCommand(pn532CmdInDataExchange, writeData); err != nil {
		return node.Message{}, fmt.Errorf("write failed: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "write_block",
			"block":     block,
			"uid":       uid,
			"sensor":    "PN532",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// readNDEF reads NDEF message from tag
func (e *PN532Executor) readNDEF() (node.Message, error) {
	// Read block 4 (first data block after manufacturer block)
	result, err := e.readBlock(4, "A", nil)
	if err != nil {
		return node.Message{}, err
	}

	payload := result.Payload
	data := payload["data"].([]byte)

	// Parse simple NDEF text record
	var text string
	if len(data) > 7 && data[0] == 0x03 { // NDEF message
		ndefLen := data[1]
		if ndefLen > 0 && data[2] == 0xD1 { // Well-known type
			textLen := data[6]
			if int(textLen) <= len(data)-7 {
				text = string(data[7 : 7+textLen])
			}
		}
	}

	return node.Message{
		Payload: map[string]interface{}{
			"ndef_text": text,
			"raw_data":  data,
			"sensor":    "PN532",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// writeNDEF writes NDEF text message
func (e *PN532Executor) writeNDEF(text string) (node.Message, error) {
	if len(text) > 10 {
		text = text[:10] // Limit to fit in one block
	}

	// Build NDEF message
	data := make([]byte, 16)
	data[0] = 0x03                  // NDEF message
	data[1] = byte(5 + len(text))   // Length
	data[2] = 0xD1                  // MB=1, ME=1, CF=0, SR=1, IL=0, TNF=1
	data[3] = 0x01                  // Type length
	data[4] = byte(1 + len(text))   // Payload length
	data[5] = 'T'                   // Type (Text)
	data[6] = 0x02                  // Status byte (UTF-8, 2-char lang code)
	copy(data[7:], []byte("en"))    // Language
	copy(data[9:], []byte(text))

	return e.writeBlock(4, "A", nil, data)
}

// getFirmwareVersion returns the firmware version
func (e *PN532Executor) getFirmwareVersion() (node.Message, error) {
	response, err := e.sendCommand(pn532CmdGetFirmwareVersion, nil)
	if err != nil {
		return node.Message{}, err
	}

	if len(response) < 4 {
		return node.Message{}, fmt.Errorf("invalid firmware response")
	}

	return node.Message{
		Payload: map[string]interface{}{
			"ic":         response[0],
			"version":    response[1],
			"revision":   response[2],
			"support":    response[3],
			"version_str": fmt.Sprintf("%d.%d", response[1], response[2]),
			"sensor":     "PN532",
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

// sendCommand sends a command and receives response
func (e *PN532Executor) sendCommand(cmd byte, data []byte) ([]byte, error) {
	// Build frame
	dataLen := byte(len(data) + 2)
	frame := make([]byte, 0, 8+len(data))
	frame = append(frame, pn532Preamble, pn532StartCode1, pn532StartCode2)
	frame = append(frame, dataLen, ^dataLen+1) // LEN, LCS
	frame = append(frame, pn532HostToPn532, cmd)
	frame = append(frame, data...)

	// Calculate DCS
	dcs := pn532HostToPn532 + cmd
	for _, b := range data {
		dcs += b
	}
	frame = append(frame, ^dcs+1, pn532Postamble)

	// Send
	if _, err := e.dev.Write(frame); err != nil {
		return nil, err
	}

	time.Sleep(50 * time.Millisecond)

	// Read ACK
	ack := make([]byte, 6)
	if err := e.dev.Tx(nil, ack); err != nil {
		return nil, err
	}

	time.Sleep(50 * time.Millisecond)

	// Read response
	response := make([]byte, 64)
	if err := e.dev.Tx(nil, response); err != nil {
		return nil, err
	}
	n := len(response)

	// Parse response frame
	if n < 7 {
		return nil, fmt.Errorf("response too short")
	}

	// Find start of data
	for i := 0; i < n-5; i++ {
		if response[i] == 0x00 && response[i+1] == 0xFF {
			dataLen := response[i+2]
			if int(i+5+int(dataLen)) <= n {
				return response[i+5 : i+5+int(dataLen)-2], nil
			}
		}
	}

	return nil, fmt.Errorf("invalid response frame")
}

// getTagType determines tag type from SENS_RES and SEL_RES
func (e *PN532Executor) getTagType(sensRes uint16, selRes byte) string {
	switch selRes {
	case 0x00:
		return "MIFARE_Ultralight"
	case 0x08:
		return "MIFARE_Classic_1K"
	case 0x09:
		return "MIFARE_Mini"
	case 0x18:
		return "MIFARE_Classic_4K"
	case 0x20:
		return "MIFARE_Plus"
	case 0x28:
		return "JCOP"
	case 0x60:
		return "NTAG"
	default:
		return "unknown"
	}
}

// uidToDecimal converts UID to decimal
func (e *PN532Executor) uidToDecimal(uid []byte) uint64 {
	var result uint64
	for i, b := range uid {
		result |= uint64(b) << (8 * (len(uid) - 1 - i))
	}
	return result
}

// Cleanup releases resources
func (e *PN532Executor) Cleanup() error {
	if e.bus != nil {
		e.bus.Close()
		e.bus = nil
	}
	e.initialized = false
	return nil
}
