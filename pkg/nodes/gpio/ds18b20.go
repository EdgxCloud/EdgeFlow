package gpio

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

type DS18B20Node struct {
	deviceID string
	unit     string // "celsius" or "fahrenheit"
}

func NewDS18B20Node() *DS18B20Node {
	return &DS18B20Node{
		unit: "celsius",
	}
}

func (n *DS18B20Node) Init(config map[string]interface{}) error {
	if deviceID, ok := config["deviceId"].(string); ok {
		n.deviceID = deviceID
	}
	if unit, ok := config["unit"].(string); ok {
		n.unit = unit
	}
	return nil
}

func (n *DS18B20Node) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	devicePath := filepath.Join("/sys/bus/w1/devices", n.deviceID, "w1_slave")

	data, err := os.ReadFile(devicePath)
	if err != nil {
		return msg, fmt.Errorf("failed to read sensor: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return msg, fmt.Errorf("invalid sensor data")
	}

	if !strings.Contains(lines[0], "YES") {
		return msg, fmt.Errorf("CRC check failed")
	}

	tempStr := strings.Split(lines[1], "t=")
	if len(tempStr) < 2 {
		return msg, fmt.Errorf("temperature not found")
	}

	tempInt, err := strconv.Atoi(strings.TrimSpace(tempStr[1]))
	if err != nil {
		return msg, fmt.Errorf("failed to parse temperature: %w", err)
	}

	tempCelsius := float64(tempInt) / 1000.0

	var temperature float64
	if n.unit == "fahrenheit" {
		temperature = (tempCelsius * 9.0 / 5.0) + 32.0
	} else {
		temperature = tempCelsius
	}

	msg.Payload = map[string]interface{}{
		"temperature": temperature,
		"unit":        n.unit,
		"deviceId":    n.deviceID,
	}

	return msg, nil
}

func (n *DS18B20Node) Cleanup() error {
	return nil
}

func NewDS18B20Executor(config map[string]interface{}) (node.Executor, error) {
	executor := NewDS18B20Node()
	if err := executor.Init(config); err != nil {
		return nil, err
	}
	return executor, nil
}
