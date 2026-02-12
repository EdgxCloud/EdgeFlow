package wireless

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// IR protocol timing constants (microseconds)
const (
	irNECAGCPulse    = 9000
	irNECAGCSpace    = 4500
	irNECBitPulse    = 562
	irNEC0Space      = 562
	irNEC1Space      = 1687
	irNECRepeatSpace = 2250

	irRC5BitTime   = 889
	irRC6HeaderPulse = 2666
	irRC6HeaderSpace = 889
)

// IRExecutor provides infrared transmit/receive functionality
type IRExecutor struct {
	txPin       int
	rxPin       int
	protocol    string // NEC, RC5, RC6, Sony, Samsung, raw
	operation   string // send, receive, learn
	frequency   int    // carrier frequency Hz (default 38000)
	repeatCount int
}

// NewIRExecutor creates a new IR executor
func NewIRExecutor() node.Executor {
	return &IRExecutor{
		protocol:    "NEC",
		operation:   "send",
		frequency:   38000,
		repeatCount: 1,
	}
}

func (e *IRExecutor) Init(config map[string]interface{}) error {
	if tp, ok := config["txPin"].(float64); ok {
		e.txPin = int(tp)
	}
	if rp, ok := config["rxPin"].(float64); ok {
		e.rxPin = int(rp)
	}
	if p, ok := config["protocol"].(string); ok {
		e.protocol = p
	}
	if op, ok := config["operation"].(string); ok {
		e.operation = op
	}
	if f, ok := config["frequency"].(float64); ok {
		e.frequency = int(f)
	}
	if rc, ok := config["repeatCount"].(float64); ok {
		e.repeatCount = int(rc)
	}
	return nil
}

func (e *IRExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	operation := e.operation
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}

	protocol := e.protocol
	if p, ok := msg.Payload["protocol"].(string); ok {
		protocol = p
	}

	switch operation {
	case "send":
		return e.send(msg, protocol)
	case "receive":
		return e.receive(protocol)
	case "learn":
		return e.learn()
	default:
		return node.Message{}, fmt.Errorf("unknown IR operation: %s", operation)
	}
}

func (e *IRExecutor) send(msg node.Message, protocol string) (node.Message, error) {
	address := uint16(0)
	command := uint16(0)

	if a, ok := msg.Payload["address"].(float64); ok {
		address = uint16(a)
	}
	if c, ok := msg.Payload["command"].(float64); ok {
		command = uint16(c)
	}

	// Encode IR signal based on protocol
	var timings []int
	switch protocol {
	case "NEC":
		timings = encodeNEC(address, command)
	case "RC5":
		timings = encodeRC5(address, command)
	default:
		// Check for raw timings
		if raw, ok := msg.Payload["raw_timings"].([]interface{}); ok {
			for _, t := range raw {
				if v, ok := t.(float64); ok {
					timings = append(timings, int(v))
				}
			}
		} else {
			return node.Message{}, fmt.Errorf("unsupported IR protocol: %s (use raw_timings for custom protocols)", protocol)
		}
	}

	// On non-Linux or without GPIO, return the encoded signal info
	return node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"protocol":     protocol,
			"address":      address,
			"command":      command,
			"frequency_hz": e.frequency,
			"repeat_count": e.repeatCount,
			"timings_us":   timings,
			"timings_count": len(timings),
			"tx_pin":       e.txPin,
			"sent":         false,
			"note":         "IR transmission requires Linux with GPIO support",
		},
	}, nil
}

func (e *IRExecutor) receive(protocol string) (node.Message, error) {
	return node.Message{}, fmt.Errorf("IR receive requires Linux with GPIO support (rxPin: %d)", e.rxPin)
}

func (e *IRExecutor) learn() (node.Message, error) {
	return node.Message{}, fmt.Errorf("IR learn mode requires Linux with GPIO support (rxPin: %d)", e.rxPin)
}

// encodeNEC encodes an NEC protocol IR signal
func encodeNEC(address, command uint16) []int {
	var timings []int

	// AGC burst
	timings = append(timings, irNECAGCPulse, irNECAGCSpace)

	// Address byte
	for i := 0; i < 8; i++ {
		timings = append(timings, irNECBitPulse)
		if (address>>uint(i))&1 == 1 {
			timings = append(timings, irNEC1Space)
		} else {
			timings = append(timings, irNEC0Space)
		}
	}

	// Inverted address
	invAddr := ^address
	for i := 0; i < 8; i++ {
		timings = append(timings, irNECBitPulse)
		if (invAddr>>uint(i))&1 == 1 {
			timings = append(timings, irNEC1Space)
		} else {
			timings = append(timings, irNEC0Space)
		}
	}

	// Command byte
	for i := 0; i < 8; i++ {
		timings = append(timings, irNECBitPulse)
		if (command>>uint(i))&1 == 1 {
			timings = append(timings, irNEC1Space)
		} else {
			timings = append(timings, irNEC0Space)
		}
	}

	// Inverted command
	invCmd := ^command
	for i := 0; i < 8; i++ {
		timings = append(timings, irNECBitPulse)
		if (invCmd>>uint(i))&1 == 1 {
			timings = append(timings, irNEC1Space)
		} else {
			timings = append(timings, irNEC0Space)
		}
	}

	// Stop bit
	timings = append(timings, irNECBitPulse)

	return timings
}

// encodeRC5 encodes an RC5 protocol IR signal
func encodeRC5(address, command uint16) []int {
	var timings []int

	// RC5 uses Manchester encoding
	// Start bits (2x '1'), toggle bit, 5-bit address, 6-bit command
	frame := uint16(0x3000) // Two start bits
	frame |= (address & 0x1F) << 6
	frame |= command & 0x3F

	for i := 13; i >= 0; i-- {
		bit := (frame >> uint(i)) & 1
		if bit == 1 {
			timings = append(timings, irRC5BitTime, irRC5BitTime) // low then high
		} else {
			timings = append(timings, irRC5BitTime, irRC5BitTime) // high then low
		}
	}

	return timings
}

// Cleanup releases resources
func (e *IRExecutor) Cleanup() error {
	return nil
}
