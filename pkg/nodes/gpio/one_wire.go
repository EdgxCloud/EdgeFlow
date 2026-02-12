//go:build linux

package gpio

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// OneWireNode provides generic 1-Wire bus communication via sysfs
type OneWireNode struct {
	busPath   string
	deviceId  string
	operation string
}

// NewOneWireExecutor creates a new one-wire node executor
func NewOneWireExecutor() node.Executor {
	return &OneWireNode{
		busPath:   "/sys/bus/w1/devices/",
		operation: "scan",
	}
}

// Init initializes the one-wire node
func (n *OneWireNode) Init(config map[string]interface{}) error {
	if busPath, ok := config["busPath"].(string); ok && busPath != "" {
		n.busPath = busPath
	}
	if deviceId, ok := config["deviceId"].(string); ok {
		n.deviceId = deviceId
	}
	if operation, ok := config["operation"].(string); ok {
		n.operation = operation
	}
	return nil
}

// Execute performs the configured 1-Wire operation
func (n *OneWireNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	operation := n.operation
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}

	deviceId := n.deviceId
	if did, ok := msg.Payload["deviceId"].(string); ok {
		deviceId = did
	}

	switch operation {
	case "scan":
		return n.scanDevices()
	case "read":
		if deviceId == "" {
			return node.Message{}, fmt.Errorf("deviceId is required for read operation")
		}
		return n.readDevice(deviceId)
	case "read_temperature":
		if deviceId == "" {
			return node.Message{}, fmt.Errorf("deviceId is required for read_temperature operation")
		}
		return n.readTemperature(deviceId)
	default:
		return node.Message{}, fmt.Errorf("unknown 1-Wire operation: %s", operation)
	}
}

func (n *OneWireNode) scanDevices() (node.Message, error) {
	entries, err := os.ReadDir(n.busPath)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to scan 1-Wire bus at %s: %w", n.busPath, err)
	}

	var devices []map[string]interface{}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "w1_bus_master") {
			continue
		}
		parts := strings.SplitN(name, "-", 2)
		family := ""
		deviceType := "unknown"
		if len(parts) >= 1 {
			family = parts[0]
			switch family {
			case "28":
				deviceType = "DS18B20 temperature sensor"
			case "10":
				deviceType = "DS18S20 temperature sensor"
			case "3b":
				deviceType = "DS1825 temperature sensor"
			case "42":
				deviceType = "DS28EA00 temperature sensor"
			case "22":
				deviceType = "DS1822 temperature sensor"
			case "3a":
				deviceType = "DS2413 dual I/O"
			case "29":
				deviceType = "DS2408 8-channel I/O"
			case "26":
				deviceType = "DS2438 battery monitor"
			}
		}
		devices = append(devices, map[string]interface{}{
			"id":     name,
			"family": family,
			"type":   deviceType,
		})
	}

	return node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"devices":   devices,
			"count":     len(devices),
			"bus_path":  n.busPath,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}, nil
}

func (n *OneWireNode) readDevice(deviceId string) (node.Message, error) {
	slavePath := filepath.Join(n.busPath, deviceId, "w1_slave")
	data, err := os.ReadFile(slavePath)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read device %s: %w", deviceId, err)
	}

	return node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"device_id": deviceId,
			"raw_data":  string(data),
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}, nil
}

func (n *OneWireNode) readTemperature(deviceId string) (node.Message, error) {
	slavePath := filepath.Join(n.busPath, deviceId, "w1_slave")
	data, err := os.ReadFile(slavePath)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read temperature from %s: %w", deviceId, err)
	}

	content := string(data)
	lines := strings.Split(strings.TrimSpace(content), "\n")

	valid := false
	if len(lines) >= 1 && strings.HasSuffix(strings.TrimSpace(lines[0]), "YES") {
		valid = true
	}

	var tempC, tempF float64
	if len(lines) >= 2 {
		idx := strings.Index(lines[1], "t=")
		if idx >= 0 {
			tempStr := lines[1][idx+2:]
			if rawTemp, err := strconv.ParseFloat(strings.TrimSpace(tempStr), 64); err == nil {
				tempC = rawTemp / 1000.0
				tempF = tempC*9.0/5.0 + 32.0
			}
		}
	}

	return node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"device_id":     deviceId,
			"temperature_c": tempC,
			"temperature_f": tempF,
			"unit":          "celsius",
			"valid":         valid,
			"raw_data":      content,
			"timestamp":     time.Now().Format(time.RFC3339),
		},
	}, nil
}

// Cleanup releases resources
func (n *OneWireNode) Cleanup() error {
	return nil
}
