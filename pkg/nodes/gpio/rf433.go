//go:build linux
// +build linux

package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

// RF433 Protocol timing (in microseconds)
const (
	rf433PulseLengthDefault = 350  // Default pulse length
	rf433SyncFactor         = 31   // Sync bit factor
	rf433ZeroHighFactor     = 1    // Zero bit high factor
	rf433ZeroLowFactor      = 3    // Zero bit low factor
	rf433OneHighFactor      = 3    // One bit high factor
	rf433OneLowFactor       = 1    // One bit low factor
)

// RF433Config holds configuration for RF 433MHz module
type RF433Config struct {
	TxPin        string `json:"tx_pin"`
	RxPin        string `json:"rx_pin"`
	PulseLength  int    `json:"pulse_length_us"`
	Protocol     int    `json:"protocol"` // 1-6 for common protocols
	RepeatCount  int    `json:"repeat_count"`
	BitLength    int    `json:"bit_length"`
	PollInterval int    `json:"poll_interval_ms"`
}

// RF433Protocol defines timing for different RF protocols
type RF433Protocol struct {
	PulseLength   int
	SyncFactor    int
	ZeroHighFactor int
	ZeroLowFactor int
	OneHighFactor int
	OneLowFactor  int
	Inverted      bool
}

var rf433Protocols = map[int]RF433Protocol{
	1: {350, 31, 1, 3, 3, 1, false},  // Standard
	2: {650, 10, 1, 3, 3, 1, false},  // Protocol 2
	3: {100, 30, 1, 3, 3, 1, false},  // Protocol 3
	4: {380, 6, 1, 3, 3, 1, false},   // Protocol 4
	5: {500, 6, 1, 2, 2, 1, false},   // Protocol 5
	6: {450, 23, 1, 2, 2, 1, true},   // Inverted
}

// RF433Executor implements 433MHz RF transmitter/receiver
type RF433Executor struct {
	config       RF433Config
	txPin        gpio.PinIO
	rxPin        gpio.PinIO
	mu           sync.Mutex
	hostInited   bool
	initialized  bool
	protocol     RF433Protocol
	receiveData  []uint64
	lastReceived uint64
	receiveTime  time.Time
}

func (e *RF433Executor) Init(config map[string]interface{}) error {
	e.config = RF433Config{
		TxPin:        "GPIO17",
		RxPin:        "GPIO27",
		PulseLength:  350,
		Protocol:     1,
		RepeatCount:  10,
		BitLength:    24,
		PollInterval: 100,
	}

	if config != nil {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := json.Unmarshal(configJSON, &e.config); err != nil {
			return fmt.Errorf("failed to parse RF433 config: %w", err)
		}
	}

	// Set protocol
	if proto, ok := rf433Protocols[e.config.Protocol]; ok {
		e.protocol = proto
	} else {
		e.protocol = rf433Protocols[1]
	}

	// Override pulse length if specified
	if e.config.PulseLength > 0 {
		e.protocol.PulseLength = e.config.PulseLength
	}

	return nil
}

func (e *RF433Executor) initHardware() error {
	if e.initialized {
		return nil
	}

	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return fmt.Errorf("failed to initialize periph host: %w", err)
		}
		e.hostInited = true
	}

	// Initialize TX pin
	if e.config.TxPin != "" {
		txPin := gpioreg.ByName(e.config.TxPin)
		if txPin == nil {
			return fmt.Errorf("failed to find TX pin %s", e.config.TxPin)
		}
		if err := txPin.Out(gpio.Low); err != nil {
			return fmt.Errorf("failed to configure TX pin as output: %w", err)
		}
		e.txPin = txPin
	}

	// Initialize RX pin
	if e.config.RxPin != "" {
		rxPin := gpioreg.ByName(e.config.RxPin)
		if rxPin == nil {
			return fmt.Errorf("failed to find RX pin %s", e.config.RxPin)
		}
		if err := rxPin.In(gpio.PullDown, gpio.BothEdges); err != nil {
			return fmt.Errorf("failed to configure RX pin as input: %w", err)
		}
		e.rxPin = rxPin
	}

	e.initialized = true
	return nil
}

func (e *RF433Executor) transmitBit(high bool) {
	pulseLen := time.Duration(e.protocol.PulseLength) * time.Microsecond

	if e.protocol.Inverted {
		high = !high
	}

	if high {
		// Transmit '1' bit
		e.txPin.Out(gpio.High)
		time.Sleep(pulseLen * time.Duration(e.protocol.OneHighFactor))
		e.txPin.Out(gpio.Low)
		time.Sleep(pulseLen * time.Duration(e.protocol.OneLowFactor))
	} else {
		// Transmit '0' bit
		e.txPin.Out(gpio.High)
		time.Sleep(pulseLen * time.Duration(e.protocol.ZeroHighFactor))
		e.txPin.Out(gpio.Low)
		time.Sleep(pulseLen * time.Duration(e.protocol.ZeroLowFactor))
	}
}

func (e *RF433Executor) transmitSync() {
	pulseLen := time.Duration(e.protocol.PulseLength) * time.Microsecond

	e.txPin.Out(gpio.High)
	time.Sleep(pulseLen)
	e.txPin.Out(gpio.Low)
	time.Sleep(pulseLen * time.Duration(e.protocol.SyncFactor))
}

func (e *RF433Executor) transmit(code uint64, bitLength int) error {
	if e.txPin == nil {
		return fmt.Errorf("TX pin not configured")
	}

	for repeat := 0; repeat < e.config.RepeatCount; repeat++ {
		// Transmit bits MSB first
		for i := bitLength - 1; i >= 0; i-- {
			bit := (code >> i) & 1
			e.transmitBit(bit == 1)
		}

		// Transmit sync
		e.transmitSync()
	}

	e.txPin.Out(gpio.Low)
	return nil
}

func (e *RF433Executor) transmitTriState(code string) error {
	if e.txPin == nil {
		return fmt.Errorf("TX pin not configured")
	}

	for repeat := 0; repeat < e.config.RepeatCount; repeat++ {
		for _, c := range code {
			switch c {
			case '0':
				e.transmitBit(false)
				e.transmitBit(false)
			case '1':
				e.transmitBit(true)
				e.transmitBit(true)
			case 'F', 'f':
				e.transmitBit(false)
				e.transmitBit(true)
			}
		}
		e.transmitSync()
	}

	e.txPin.Out(gpio.Low)
	return nil
}

func (e *RF433Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.initHardware(); err != nil {
		return node.Message{}, err
	}

	action := "status"
	if payload, ok := msg.Payload.(map[string]interface{}); ok {
		if a, ok := payload["action"].(string); ok {
			action = a
		}
	}

	switch action {
	case "send":
		return e.handleSend(msg)
	case "send_tristate":
		return e.handleSendTriState(msg)
	case "send_binary":
		return e.handleSendBinary(msg)
	case "configure":
		return e.handleConfigure(msg)
	case "status":
		return e.handleStatus()
	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *RF433Executor) handleSend(msg node.Message) (node.Message, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	var code uint64
	bitLength := e.config.BitLength

	if c, ok := payload["code"].(float64); ok {
		code = uint64(c)
	} else if c, ok := payload["code"].(string); ok {
		// Parse hex string
		_, err := fmt.Sscanf(c, "%x", &code)
		if err != nil {
			return node.Message{}, fmt.Errorf("invalid code format: %w", err)
		}
	} else {
		return node.Message{}, fmt.Errorf("code required")
	}

	if b, ok := payload["bit_length"].(float64); ok {
		bitLength = int(b)
	}

	if err := e.transmit(code, bitLength); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":       "sent",
			"code":         code,
			"code_hex":     fmt.Sprintf("0x%X", code),
			"bit_length":   bitLength,
			"protocol":     e.config.Protocol,
			"pulse_length": e.protocol.PulseLength,
			"repeat_count": e.config.RepeatCount,
			"timestamp":    time.Now().Unix(),
		},
	}, nil
}

func (e *RF433Executor) handleSendTriState(msg node.Message) (node.Message, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	code, ok := payload["code"].(string)
	if !ok {
		return node.Message{}, fmt.Errorf("tri-state code string required")
	}

	// Validate tri-state code
	for _, c := range code {
		if c != '0' && c != '1' && c != 'F' && c != 'f' {
			return node.Message{}, fmt.Errorf("invalid tri-state character: %c (use 0, 1, or F)", c)
		}
	}

	if err := e.transmitTriState(code); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":       "sent",
			"code":         code,
			"type":         "tristate",
			"protocol":     e.config.Protocol,
			"pulse_length": e.protocol.PulseLength,
			"repeat_count": e.config.RepeatCount,
			"timestamp":    time.Now().Unix(),
		},
	}, nil
}

func (e *RF433Executor) handleSendBinary(msg node.Message) (node.Message, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	code, ok := payload["code"].(string)
	if !ok {
		return node.Message{}, fmt.Errorf("binary code string required")
	}

	// Validate binary code
	var value uint64
	for i, c := range code {
		if c != '0' && c != '1' {
			return node.Message{}, fmt.Errorf("invalid binary character at position %d: %c", i, c)
		}
		value = (value << 1) | uint64(c-'0')
	}

	if err := e.transmit(value, len(code)); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":       "sent",
			"code":         code,
			"code_decimal": value,
			"code_hex":     fmt.Sprintf("0x%X", value),
			"type":         "binary",
			"bit_length":   len(code),
			"protocol":     e.config.Protocol,
			"pulse_length": e.protocol.PulseLength,
			"repeat_count": e.config.RepeatCount,
			"timestamp":    time.Now().Unix(),
		},
	}, nil
}

func (e *RF433Executor) handleConfigure(msg node.Message) (node.Message, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	if protocol, ok := payload["protocol"].(float64); ok {
		e.config.Protocol = int(protocol)
		if proto, ok := rf433Protocols[int(protocol)]; ok {
			e.protocol = proto
		}
	}

	if pulseLength, ok := payload["pulse_length"].(float64); ok {
		e.config.PulseLength = int(pulseLength)
		e.protocol.PulseLength = int(pulseLength)
	}

	if repeatCount, ok := payload["repeat_count"].(float64); ok {
		e.config.RepeatCount = int(repeatCount)
	}

	if bitLength, ok := payload["bit_length"].(float64); ok {
		e.config.BitLength = int(bitLength)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":       "configured",
			"protocol":     e.config.Protocol,
			"pulse_length": e.protocol.PulseLength,
			"repeat_count": e.config.RepeatCount,
			"bit_length":   e.config.BitLength,
		},
	}, nil
}

func (e *RF433Executor) handleStatus() (node.Message, error) {
	return node.Message{
		Payload: map[string]interface{}{
			"tx_pin":       e.config.TxPin,
			"rx_pin":       e.config.RxPin,
			"protocol":     e.config.Protocol,
			"pulse_length": e.protocol.PulseLength,
			"repeat_count": e.config.RepeatCount,
			"bit_length":   e.config.BitLength,
			"initialized":  e.initialized,
			"tx_enabled":   e.txPin != nil,
			"rx_enabled":   e.rxPin != nil,
		},
	}, nil
}

func (e *RF433Executor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized {
		if e.txPin != nil {
			e.txPin.Out(gpio.Low)
			e.txPin.Halt()
		}
		if e.rxPin != nil {
			e.rxPin.Halt()
		}
		e.initialized = false
	}
	return nil
}
