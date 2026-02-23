package gpio

import (
	"context"
	"fmt"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
)

type AHT20Node struct {
	i2cBus  string
	address uint16
	bus     i2c.BusCloser
	device  *i2c.Dev
}

func NewAHT20Node() *AHT20Node {
	return &AHT20Node{
		i2cBus:  "1",
		address: 0x38,
	}
}

func (n *AHT20Node) Init(config map[string]interface{}) error {
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

	n.bus = bus
	n.device = &i2c.Dev{Bus: bus, Addr: n.address}

	if err := n.reset(); err != nil {
		return fmt.Errorf("failed to reset sensor: %w", err)
	}

	time.Sleep(10 * time.Millisecond)
	return nil
}

func (n *AHT20Node) reset() error {
	return n.device.Tx([]byte{0xBA}, nil)
}

func (n *AHT20Node) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if err := n.device.Tx([]byte{0xAC, 0x33, 0x00}, nil); err != nil {
		return msg, fmt.Errorf("failed to trigger measurement: %w", err)
	}

	time.Sleep(80 * time.Millisecond)

	data := make([]byte, 7)
	if err := n.device.Tx(nil, data); err != nil {
		return msg, fmt.Errorf("failed to read data: %w", err)
	}

	humidity := ((uint32(data[1]) << 12) | (uint32(data[2]) << 4) | (uint32(data[3]) >> 4))
	temperature := (((uint32(data[3]) & 0x0F) << 16) | (uint32(data[4]) << 8) | uint32(data[5]))

	humidityPercent := (float64(humidity) / 1048576.0) * 100.0
	temperatureCelsius := ((float64(temperature) / 1048576.0) * 200.0) - 50.0

	msg.Payload = map[string]interface{}{
		"temperature": temperatureCelsius,
		"humidity":    humidityPercent,
	}

	return msg, nil
}

func (n *AHT20Node) Cleanup() error {
	if n.bus != nil {
		return n.bus.Close()
	}
	return nil
}

func NewAHT20Executor(config map[string]interface{}) (node.Executor, error) {
	executor := NewAHT20Node()
	if err := executor.Init(config); err != nil {
		return nil, err
	}
	return executor, nil
}
