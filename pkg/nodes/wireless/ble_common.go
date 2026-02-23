// Package wireless provides nodes for wireless communication protocols
// Supports BLE, Zigbee, Z-Wave, LoRa and other wireless technologies
package wireless

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// BLENode implements Bluetooth Low Energy communication
type BLENode struct {
	adapter      string // Adapter name (e.g., hci0)
	operation    string // scan, connect, read, write, subscribe, disconnect
	deviceAddr   string // Target device MAC address
	serviceUUID  string // Service UUID
	charUUID     string // Characteristic UUID
	scanDuration time.Duration
	timeout      time.Duration
	mu           sync.Mutex
	connected    bool
}

// BLEDevice represents a discovered BLE device
type BLEDevice struct {
	Address     string            `json:"address"`
	Name        string            `json:"name"`
	RSSI        int               `json:"rssi"`
	Connectable bool              `json:"connectable"`
	Services    []string          `json:"services,omitempty"`
	ManufData   map[uint16][]byte `json:"manufacturerData,omitempty"`
}

// BLECharacteristic represents a BLE characteristic
type BLECharacteristic struct {
	UUID       string   `json:"uuid"`
	Properties []string `json:"properties"` // read, write, notify, indicate
	Value      []byte   `json:"value,omitempty"`
}

// NewBLENode creates a new BLE node
func NewBLENode() *BLENode {
	return &BLENode{
		adapter:      "hci0",
		operation:    "scan",
		scanDuration: 10 * time.Second,
		timeout:      30 * time.Second,
	}
}

// Init initializes the BLE node
func (n *BLENode) Init(config map[string]interface{}) error {
	if adapter, ok := config["adapter"].(string); ok {
		n.adapter = adapter
	}
	if op, ok := config["operation"].(string); ok {
		n.operation = op
	}
	if addr, ok := config["deviceAddress"].(string); ok {
		n.deviceAddr = addr
	}
	if svc, ok := config["serviceUuid"].(string); ok {
		n.serviceUUID = svc
	}
	if char, ok := config["characteristicUuid"].(string); ok {
		n.charUUID = char
	}
	if dur, ok := config["scanDuration"].(float64); ok {
		n.scanDuration = time.Duration(dur) * time.Second
	}
	if timeout, ok := config["timeout"].(float64); ok {
		n.timeout = time.Duration(timeout) * time.Millisecond
	}

	return nil
}

// Execute performs BLE operation
func (n *BLENode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Get parameters from message or config
	operation := n.operation
	deviceAddr := n.deviceAddr
	serviceUUID := n.serviceUUID
	charUUID := n.charUUID

	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}
	if addr, ok := msg.Payload["deviceAddress"].(string); ok {
		deviceAddr = addr
	}
	if svc, ok := msg.Payload["serviceUuid"].(string); ok {
		serviceUUID = svc
	}
	if char, ok := msg.Payload["characteristicUuid"].(string); ok {
		charUUID = char
	}

	var result interface{}
	var err error

	switch operation {
	case "scan":
		result, err = n.scanDevices(ctx)
	case "connect":
		err = n.connect(ctx, deviceAddr)
		result = map[string]interface{}{"connected": err == nil, "address": deviceAddr}
	case "disconnect":
		err = n.disconnect(ctx, deviceAddr)
		result = map[string]interface{}{"disconnected": err == nil, "address": deviceAddr}
	case "read":
		result, err = n.readCharacteristic(ctx, deviceAddr, serviceUUID, charUUID)
	case "write":
		value, _ := msg.Payload["value"].([]byte)
		if value == nil {
			if str, ok := msg.Payload["value"].(string); ok {
				value = []byte(str)
			}
		}
		err = n.writeCharacteristic(ctx, deviceAddr, serviceUUID, charUUID, value)
		result = map[string]interface{}{"success": err == nil, "address": deviceAddr}
	case "subscribe":
		err = n.subscribeCharacteristic(ctx, deviceAddr, serviceUUID, charUUID)
		result = map[string]interface{}{"subscribed": err == nil, "address": deviceAddr}
	case "discover_services":
		result, err = n.discoverServices(ctx, deviceAddr)
	case "discover_characteristics":
		result, err = n.discoverCharacteristics(ctx, deviceAddr, serviceUUID)
	default:
		err = fmt.Errorf("unknown operation: %s", operation)
	}

	if err != nil {
		return msg, err
	}

	msg.Payload["result"] = result
	msg.Payload["operation"] = operation
	msg.Payload["adapter"] = n.adapter

	return msg, nil
}

// Cleanup releases BLE resources
func (n *BLENode) Cleanup() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.connected = false
	return nil
}

// NewBLEExecutor creates a new BLE executor
func NewBLEExecutor() node.Executor {
	return NewBLENode()
}
