package gpio

import (
	"context"
	"fmt"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/ads1x15"
	"periph.io/x/conn/v3/i2c/i2creg"
)

type VoltageMonitorNode struct {
	i2cBus    string
	address   uint16
	channel   int
	gain      int
	dataRate  int
	threshold float64

	device *ads1x15.Dev
}

func NewVoltageMonitorNode() *VoltageMonitorNode {
	return &VoltageMonitorNode{
		i2cBus:   "1",
		address:  0x48,
		channel:  0,
		gain:     1,
		dataRate: 128,
	}
}

func (n *VoltageMonitorNode) Init(config map[string]interface{}) error {
	if bus, ok := config["i2cBus"].(string); ok {
		n.i2cBus = bus
	}
	if addr, ok := config["address"].(float64); ok {
		n.address = uint16(addr)
	} else if addr, ok := config["address"].(int); ok {
		n.address = uint16(addr)
	}
	if channel, ok := config["channel"].(float64); ok {
		n.channel = int(channel)
	} else if channel, ok := config["channel"].(int); ok {
		n.channel = channel
	}
	if gain, ok := config["gain"].(float64); ok {
		n.gain = int(gain)
	} else if gain, ok := config["gain"].(int); ok {
		n.gain = gain
	}
	if threshold, ok := config["threshold"].(float64); ok {
		n.threshold = threshold
	}

	bus, err := i2creg.Open(n.i2cBus)
	if err != nil {
		return fmt.Errorf("failed to open I2C bus: %w", err)
	}

	device, err := ads1x15.NewADS1115(bus, &ads1x15.DefaultOpts)
	if err != nil {
		return fmt.Errorf("failed to initialize ADS1115: %w", err)
	}

	n.device = device
	return nil
}

func (n *VoltageMonitorNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	c, err := n.device.PinForChannel(ads1x15.Channel(n.channel), 4096*physic.MilliVolt, 128*physic.Hertz, ads1x15.BestQuality)
	if err != nil {
		return msg, fmt.Errorf("failed to get channel: %w", err)
	}

	sample, err := c.Read()
	if err != nil {
		return msg, fmt.Errorf("failed to read voltage: %w", err)
	}

	voltage := float64(sample.V) / 1000000.0

	msg.Payload = map[string]interface{}{
		"voltage": voltage,
		"channel": n.channel,
		"raw":     sample.Raw,
	}

	if n.threshold > 0 && voltage > n.threshold {
		msg.Payload["alert"] = "threshold exceeded"
	}

	return msg, nil
}

func (n *VoltageMonitorNode) Cleanup() error {
	if n.device != nil {
		return n.device.Halt()
	}
	return nil
}

func NewVoltageMonitorExecutor(config map[string]interface{}) (node.Executor, error) {
	executor := NewVoltageMonitorNode()
	if err := executor.Init(config); err != nil {
		return nil, err
	}
	return executor, nil
}
