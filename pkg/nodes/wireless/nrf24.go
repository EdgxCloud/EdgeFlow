package wireless

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// NRF24Executor provides NRF24L01 2.4GHz wireless communication
type NRF24Executor struct {
	cePin       int
	csnPin      int
	spiBus      int
	channel     int
	dataRate    string
	payloadSize int
}

// NewNRF24Executor creates a new NRF24 executor
func NewNRF24Executor() node.Executor {
	return &NRF24Executor{
		channel:     76,
		dataRate:    "1Mbps",
		payloadSize: 32,
	}
}

func (e *NRF24Executor) Init(config map[string]interface{}) error {
	if ce, ok := config["cePin"].(float64); ok {
		e.cePin = int(ce)
	}
	if csn, ok := config["csnPin"].(float64); ok {
		e.csnPin = int(csn)
	}
	if sb, ok := config["spiBus"].(float64); ok {
		e.spiBus = int(sb)
	}
	if ch, ok := config["channel"].(float64); ok {
		e.channel = int(ch)
	}
	if dr, ok := config["dataRate"].(string); ok {
		e.dataRate = dr
	}
	if ps, ok := config["payloadSize"].(float64); ok {
		e.payloadSize = int(ps)
	}
	return nil
}

func (e *NRF24Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	action, _ := msg.Payload["action"].(string)
	switch action {
	case "send", "receive", "listen":
		return msg, fmt.Errorf("NRF24L01 hardware requires Linux with SPI support (use nrf24l01 GPIO node)")
	default:
		msg.Payload["info"] = "NRF24L01 2.4GHz wireless transceiver"
		msg.Payload["channel"] = e.channel
		msg.Payload["data_rate"] = e.dataRate
		msg.Payload["payload_size"] = e.payloadSize
		msg.Payload["actions"] = []string{"send", "receive", "listen", "set_channel", "set_power"}
		msg.Payload["note"] = "Configure nrf24l01 GPIO node for hardware access on Linux"
		return msg, nil
	}
}

func (e *NRF24Executor) Cleanup() error { return nil }
