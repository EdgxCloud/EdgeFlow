package wireless

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// RFIDExecutor provides RFID reader functionality via RC522
type RFIDExecutor struct {
	spiBus      int
	spiDevice   int
	resetPin    int
	antennaGain int
}

// NewRFIDExecutor creates a new RFID executor
func NewRFIDExecutor() node.Executor {
	return &RFIDExecutor{
		antennaGain: 4,
	}
}

func (e *RFIDExecutor) Init(config map[string]interface{}) error {
	if sb, ok := config["spiBus"].(float64); ok {
		e.spiBus = int(sb)
	}
	if sd, ok := config["spiDevice"].(float64); ok {
		e.spiDevice = int(sd)
	}
	if rp, ok := config["resetPin"].(float64); ok {
		e.resetPin = int(rp)
	}
	if ag, ok := config["antennaGain"].(float64); ok {
		e.antennaGain = int(ag)
	}
	return nil
}

func (e *RFIDExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	action, _ := msg.Payload["action"].(string)
	switch action {
	case "read", "write", "detect":
		return msg, fmt.Errorf("RFID RC522 hardware requires Linux with SPI support (use rfid_rc522 GPIO node)")
	default:
		msg.Payload["info"] = "RFID RC522 reader (13.56MHz)"
		msg.Payload["antenna_gain"] = e.antennaGain
		msg.Payload["actions"] = []string{"read", "write", "detect", "authenticate"}
		msg.Payload["supported_cards"] = []string{"MIFARE Classic 1K/4K", "MIFARE Ultralight", "NTAG213/215/216"}
		msg.Payload["note"] = "Configure rfid_rc522 GPIO node for hardware access on Linux"
		return msg, nil
	}
}

func (e *RFIDExecutor) Cleanup() error { return nil }
