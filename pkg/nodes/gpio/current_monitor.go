package gpio

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/ads1x15"
	"periph.io/x/conn/v3/i2c/i2creg"
)

type CurrentMonitorNode struct {
	i2cBus       string
	address      uint16
	channel      int
	gain         int
	shuntResistor float64
	maxCurrent   float64

	device *ads1x15.Dev
}

func NewCurrentMonitorNode() *CurrentMonitorNode {
	return &CurrentMonitorNode{
		i2cBus:       "1",
		address:      0x48,
		channel:      0,
		gain:         1,
		shuntResistor: 0.1,
		maxCurrent:   10.0,
	}
}

func (n *CurrentMonitorNode) Init(config map[string]interface{}) error {
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
	if shunt, ok := config["shuntResistor"].(float64); ok {
		n.shuntResistor = shunt
	}
	if maxCurrent, ok := config["maxCurrent"].(float64); ok {
		n.maxCurrent = maxCurrent
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

func (n *CurrentMonitorNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Use 4.096V range for current sensing through shunt resistor
	c, err := n.device.PinForChannel(ads1x15.Channel(n.channel), 4096*physic.MilliVolt, 128*physic.Hertz, ads1x15.BestQuality)
	if err != nil {
		return msg, fmt.Errorf("failed to get channel: %w", err)
	}

	sample, err := c.Read()
	if err != nil {
		return msg, fmt.Errorf("failed to read voltage: %w", err)
	}

	voltage := float64(sample.V) / float64(physic.Volt)
	current := voltage / n.shuntResistor
	power := current * voltage

	msg.Payload = map[string]interface{}{
		"current": current,
		"voltage": voltage,
		"power":   power,
		"channel": n.channel,
	}

	if current > n.maxCurrent {
		msg.Payload["alert"] = "overcurrent detected"
	}

	return msg, nil
}

func (n *CurrentMonitorNode) Cleanup() error {
	if n.device != nil {
		return n.device.Halt()
	}
	return nil
}

func NewCurrentMonitorExecutor(config map[string]interface{}) (node.Executor, error) {
	executor := NewCurrentMonitorNode()
	if err := executor.Init(config); err != nil {
		return nil, err
	}
	return executor, nil
}
