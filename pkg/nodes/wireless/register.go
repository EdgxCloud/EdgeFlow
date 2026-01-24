package wireless

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterNodes registers all wireless nodes with the registry
func RegisterNodes(registry *node.Registry) error {
	// BLE Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "ble",
		Name:        "Bluetooth LE",
		Category:    node.NodeTypeInput,
		Description: "Bluetooth Low Energy communication for IoT devices and sensors",
		Icon:        "bluetooth",
		Color:       "#0082FC",
		Properties: []node.PropertySchema{
			{
				Name:        "adapter",
				Label:       "Adapter",
				Type:        "string",
				Default:     "hci0",
				Required:    true,
				Description: "Bluetooth adapter name (e.g., hci0)",
			},
			{
				Name:        "operation",
				Label:       "Operation",
				Type:        "select",
				Default:     "scan",
				Required:    true,
				Description: "BLE operation to perform",
				Options:     []string{"scan", "connect", "disconnect", "read", "write", "subscribe", "discover_services", "discover_characteristics"},
			},
			{
				Name:        "deviceAddress",
				Label:       "Device Address",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Target device MAC address (e.g., AA:BB:CC:DD:EE:FF)",
			},
			{
				Name:        "serviceUuid",
				Label:       "Service UUID",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "BLE service UUID for read/write operations",
			},
			{
				Name:        "characteristicUuid",
				Label:       "Characteristic UUID",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "BLE characteristic UUID for read/write operations",
			},
			{
				Name:        "scanDuration",
				Label:       "Scan Duration (s)",
				Type:        "number",
				Default:     10,
				Required:    false,
				Description: "Duration of BLE scan in seconds",
			},
			{
				Name:        "timeout",
				Label:       "Timeout (ms)",
				Type:        "number",
				Default:     30000,
				Required:    false,
				Description: "Operation timeout in milliseconds",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or data input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "BLE response data"},
		},
		Factory: NewBLEExecutor,
	}); err != nil {
		return err
	}

	// Zigbee Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "zigbee",
		Name:        "Zigbee",
		Category:    node.NodeTypeInput,
		Description: "Zigbee communication via zigbee2mqtt or deCONZ coordinator",
		Icon:        "radio",
		Color:       "#FFB000",
		Properties: []node.PropertySchema{
			{
				Name:        "coordinator",
				Label:       "Coordinator",
				Type:        "select",
				Default:     "zigbee2mqtt",
				Required:    true,
				Description: "Zigbee coordinator/gateway type",
				Options:     []string{"zigbee2mqtt", "deconz"},
			},
			{
				Name:        "mqttBroker",
				Label:       "MQTT Broker",
				Type:        "string",
				Default:     "mqtt://localhost:1883",
				Required:    false,
				Description: "MQTT broker URL (for zigbee2mqtt)",
			},
			{
				Name:        "mqttTopic",
				Label:       "MQTT Topic",
				Type:        "string",
				Default:     "zigbee2mqtt",
				Required:    false,
				Description: "Base MQTT topic (for zigbee2mqtt)",
			},
			{
				Name:        "apiUrl",
				Label:       "API URL",
				Type:        "string",
				Default:     "http://localhost:80",
				Required:    false,
				Description: "REST API URL (for deCONZ)",
			},
			{
				Name:        "apiKey",
				Label:       "API Key",
				Type:        "password",
				Default:     "",
				Required:    false,
				Description: "API key (for deCONZ)",
			},
			{
				Name:        "operation",
				Label:       "Operation",
				Type:        "select",
				Default:     "get_devices",
				Required:    true,
				Description: "Zigbee operation to perform",
				Options:     []string{"get_devices", "get_device", "set_state", "permit_join", "remove_device"},
			},
			{
				Name:        "deviceId",
				Label:       "Device ID",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Target device IEEE address or friendly name",
			},
			{
				Name:        "timeout",
				Label:       "Timeout (ms)",
				Type:        "number",
				Default:     30000,
				Required:    false,
				Description: "Operation timeout in milliseconds",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or data input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Zigbee response data"},
		},
		Factory: NewZigbeeExecutor,
	}); err != nil {
		return err
	}

	// Z-Wave Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "zwave",
		Name:        "Z-Wave",
		Category:    node.NodeTypeInput,
		Description: "Z-Wave communication via zwave2mqtt or zwave-js-ui",
		Icon:        "radio",
		Color:       "#00A4BD",
		Properties: []node.PropertySchema{
			{
				Name:        "controller",
				Label:       "Controller",
				Type:        "select",
				Default:     "zwave2mqtt",
				Required:    true,
				Description: "Z-Wave controller/gateway type",
				Options:     []string{"zwave2mqtt", "zwave-js"},
			},
			{
				Name:        "mqttBroker",
				Label:       "MQTT Broker",
				Type:        "string",
				Default:     "mqtt://localhost:1883",
				Required:    false,
				Description: "MQTT broker URL (for zwave2mqtt)",
			},
			{
				Name:        "mqttTopic",
				Label:       "MQTT Topic",
				Type:        "string",
				Default:     "zwave",
				Required:    false,
				Description: "Base MQTT topic (for zwave2mqtt)",
			},
			{
				Name:        "wsUrl",
				Label:       "WebSocket URL",
				Type:        "string",
				Default:     "http://localhost:8091",
				Required:    false,
				Description: "zwave-js-ui API URL",
			},
			{
				Name:        "operation",
				Label:       "Operation",
				Type:        "select",
				Default:     "get_nodes",
				Required:    true,
				Description: "Z-Wave operation to perform",
				Options:     []string{"get_nodes", "get_node", "set_value", "add_node", "remove_node", "heal_network", "heal_node"},
			},
			{
				Name:        "nodeId",
				Label:       "Node ID",
				Type:        "number",
				Default:     0,
				Required:    false,
				Description: "Target Z-Wave node ID",
			},
			{
				Name:        "commandClass",
				Label:       "Command Class",
				Type:        "number",
				Default:     0,
				Required:    false,
				Description: "Z-Wave command class ID (e.g., 37 for Switch Binary)",
			},
			{
				Name:        "property",
				Label:       "Property",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Property name within command class",
			},
			{
				Name:        "timeout",
				Label:       "Timeout (ms)",
				Type:        "number",
				Default:     30000,
				Required:    false,
				Description: "Operation timeout in milliseconds",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or data input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Z-Wave response data"},
		},
		Factory: NewZWaveExecutor,
	}); err != nil {
		return err
	}

	return nil
}

// init registers wireless nodes with the global registry
func init() {
	registry := node.GetGlobalRegistry()
	RegisterNodes(registry)
}
