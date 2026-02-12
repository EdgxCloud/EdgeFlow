package wireless

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// NFCExecutor provides NFC reader functionality via PN532
type NFCExecutor struct {
	interfaceType string // i2c, spi, uart
	i2cBus        int
	i2cAddress    int
	irqPin        int
	resetPin      int
}

// NewNFCExecutor creates a new NFC executor
func NewNFCExecutor() node.Executor {
	return &NFCExecutor{
		interfaceType: "i2c",
		i2cBus:        1,
		i2cAddress:    0x24,
	}
}

func (e *NFCExecutor) Init(config map[string]interface{}) error {
	if it, ok := config["interfaceType"].(string); ok {
		e.interfaceType = it
	}
	if ib, ok := config["i2cBus"].(float64); ok {
		e.i2cBus = int(ib)
	}
	if ia, ok := config["i2cAddress"].(float64); ok {
		e.i2cAddress = int(ia)
	}
	if ip, ok := config["irqPin"].(float64); ok {
		e.irqPin = int(ip)
	}
	if rp, ok := config["resetPin"].(float64); ok {
		e.resetPin = int(rp)
	}
	return nil
}

func (e *NFCExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	action, _ := msg.Payload["action"].(string)
	switch action {
	case "read", "write", "detect", "emulate":
		return msg, fmt.Errorf("NFC PN532 hardware requires Linux with I2C/SPI support (use nfc_pn532 GPIO node)")
	default:
		msg.Payload["info"] = "NFC PN532 reader/writer (13.56MHz)"
		msg.Payload["interface"] = e.interfaceType
		msg.Payload["actions"] = []string{"read", "write", "detect", "emulate", "peer_to_peer"}
		msg.Payload["supported_tags"] = []string{"ISO14443A/B", "MIFARE", "FeliCa", "ISO18092"}
		msg.Payload["note"] = "Configure nfc_pn532 GPIO node for hardware access on Linux"
		return msg, nil
	}
}

func (e *NFCExecutor) Cleanup() error { return nil }
