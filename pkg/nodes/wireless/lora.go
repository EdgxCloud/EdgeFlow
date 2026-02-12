package wireless

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// LoRaExecutor provides LoRa wireless communication via SX1276/SX1278
type LoRaExecutor struct {
	frequency    float64
	spreadFactor int
	bandwidth    int
	txPower      int
	syncWord     int
	spiBus       int
	spiDevice    int
	resetPin     int
	dio0Pin      int
}

// NewLoRaExecutor creates a new LoRa executor
func NewLoRaExecutor() node.Executor {
	return &LoRaExecutor{
		frequency:    868.0,
		spreadFactor: 7,
		bandwidth:    125000,
		txPower:      17,
		syncWord:     0x12,
	}
}

func (e *LoRaExecutor) Init(config map[string]interface{}) error {
	if f, ok := config["frequency"].(float64); ok {
		e.frequency = f
	}
	if sf, ok := config["spreadFactor"].(float64); ok {
		e.spreadFactor = int(sf)
	}
	if bw, ok := config["bandwidth"].(float64); ok {
		e.bandwidth = int(bw)
	}
	if tp, ok := config["txPower"].(float64); ok {
		e.txPower = int(tp)
	}
	if sw, ok := config["syncWord"].(float64); ok {
		e.syncWord = int(sw)
	}
	if sb, ok := config["spiBus"].(float64); ok {
		e.spiBus = int(sb)
	}
	if sd, ok := config["spiDevice"].(float64); ok {
		e.spiDevice = int(sd)
	}
	if rp, ok := config["resetPin"].(float64); ok {
		e.resetPin = int(rp)
	}
	if dp, ok := config["dio0Pin"].(float64); ok {
		e.dio0Pin = int(dp)
	}
	return nil
}

func (e *LoRaExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	action, _ := msg.Payload["action"].(string)
	switch action {
	case "send", "transmit":
		return msg, fmt.Errorf("LoRa hardware requires Linux with SPI support (use lora_sx1276 GPIO node)")
	case "receive", "receive_start":
		return msg, fmt.Errorf("LoRa hardware requires Linux with SPI support (use lora_sx1276 GPIO node)")
	default:
		msg.Payload["info"] = "LoRa SX1276/SX1278 wireless transceiver"
		msg.Payload["frequency_mhz"] = e.frequency
		msg.Payload["spread_factor"] = e.spreadFactor
		msg.Payload["bandwidth_hz"] = e.bandwidth
		msg.Payload["tx_power_dbm"] = e.txPower
		msg.Payload["actions"] = []string{"send", "receive", "receive_start", "set_frequency", "set_power", "cad"}
		msg.Payload["note"] = "Configure lora_sx1276 GPIO node for hardware access on Linux"
		return msg, nil
	}
}

func (e *LoRaExecutor) Cleanup() error { return nil }
