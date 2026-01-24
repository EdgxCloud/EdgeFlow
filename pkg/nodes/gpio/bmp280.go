package gpio

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
)

type BMP280Node struct {
	i2cBus  string
	address uint16
	device  *bmxx80.Dev
}

func NewBMP280Node() *BMP280Node {
	return &BMP280Node{
		i2cBus:  "1",
		address: 0x76,
	}
}

func (n *BMP280Node) Init(config map[string]interface{}) error {
	if bus, ok := config["i2cBus"].(string); ok {
		n.i2cBus = bus
	}
	if addr, ok := config["address"].(float64); ok {
		n.address = uint16(addr)
	} else if addr, ok := config["address"].(int); ok {
		n.address = uint16(addr)
	}

	bus, err := i2creg.Open(n.i2cBus)
	if err != nil {
		return fmt.Errorf("failed to open I2C bus: %w", err)
	}

	device, err := bmxx80.NewI2C(bus, n.address, &bmxx80.DefaultOpts)
	if err != nil {
		return fmt.Errorf("failed to initialize BMP280: %w", err)
	}

	n.device = device
	return nil
}

func (n *BMP280Node) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	var env physic.Env
	if err := n.device.Sense(&env); err != nil {
		return msg, fmt.Errorf("failed to read sensor: %w", err)
	}

	msg.Payload = map[string]interface{}{
		"temperature": env.Temperature.Celsius(),
		"pressure":    float64(env.Pressure) / float64(physic.Pascal) / 100.0, // Convert to hPa
	}

	return msg, nil
}

func (n *BMP280Node) Cleanup() error {
	if n.device != nil {
		return n.device.Halt()
	}
	return nil
}

func NewBMP280Executor(config map[string]interface{}) (node.Executor, error) {
	executor := NewBMP280Node()
	if err := executor.Init(config); err != nil {
		return nil, err
	}
	return executor, nil
}
