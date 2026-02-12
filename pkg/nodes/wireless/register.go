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

	// LoRa Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "lora",
		Name:        "LoRa",
		Category:    node.NodeTypeInput,
		Description: "LoRa SX1276/SX1278 long-range wireless communication",
		Icon:        "radio",
		Color:       "#7B1FA2",
		Properties: []node.PropertySchema{
			{Name: "frequency", Label: "Frequency (MHz)", Type: "number", Default: 433.0, Required: true, Description: "Operating frequency in MHz"},
			{Name: "spreadingFactor", Label: "Spreading Factor", Type: "select", Default: "7", Description: "LoRa spreading factor", Options: []string{"6", "7", "8", "9", "10", "11", "12"}},
			{Name: "bandwidth", Label: "Bandwidth (kHz)", Type: "select", Default: "125", Description: "Signal bandwidth", Options: []string{"7.8", "10.4", "15.6", "20.8", "31.25", "41.7", "62.5", "125", "250", "500"}},
			{Name: "txPower", Label: "TX Power (dBm)", Type: "number", Default: 17, Description: "Transmit power 2-20 dBm"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to send"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "LoRa data"},
		},
		Factory: NewLoRaExecutor,
	}); err != nil {
		return err
	}

	// RF 433MHz Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "rf433",
		Name:        "RF 433MHz",
		Category:    node.NodeTypeInput,
		Description: "433MHz RF wireless transceiver for simple devices",
		Icon:        "radio",
		Color:       "#F57C00",
		Properties: []node.PropertySchema{
			{Name: "dataPin", Label: "Data Pin", Type: "number", Default: 0, Required: true, Description: "GPIO pin for RF data"},
			{Name: "protocol", Label: "Protocol", Type: "number", Default: 1, Description: "RF protocol number (1-6)"},
			{Name: "pulseLength", Label: "Pulse Length (us)", Type: "number", Default: 350, Description: "Pulse length in microseconds"},
			{Name: "repeatCount", Label: "Repeat Count", Type: "number", Default: 10, Description: "Transmission repeat count"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to send"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "RF data"},
		},
		Factory: NewRF433Executor,
	}); err != nil {
		return err
	}

	// NRF24L01 Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "nrf24",
		Name:        "NRF24L01",
		Category:    node.NodeTypeInput,
		Description: "NRF24L01 2.4GHz wireless transceiver",
		Icon:        "radio",
		Color:       "#00897B",
		Properties: []node.PropertySchema{
			{Name: "cePin", Label: "CE Pin", Type: "number", Default: 0, Description: "GPIO CE pin"},
			{Name: "csnPin", Label: "CSN Pin", Type: "number", Default: 0, Description: "GPIO CSN pin"},
			{Name: "channel", Label: "Channel", Type: "number", Default: 76, Description: "Radio channel (0-125)"},
			{Name: "dataRate", Label: "Data Rate", Type: "select", Default: "1Mbps", Description: "Data rate", Options: []string{"250kbps", "1Mbps", "2Mbps"}},
			{Name: "payloadSize", Label: "Payload Size", Type: "number", Default: 32, Description: "Payload size in bytes (1-32)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to send"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "NRF24 data"},
		},
		Factory: NewNRF24Executor,
	}); err != nil {
		return err
	}

	// RFID RC522 Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "rfid",
		Name:        "RFID RC522",
		Category:    node.NodeTypeInput,
		Description: "RFID RC522 reader for MIFARE cards (13.56MHz)",
		Icon:        "credit-card",
		Color:       "#AD1457",
		Properties: []node.PropertySchema{
			{Name: "spiBus", Label: "SPI Bus", Type: "number", Default: 0, Description: "SPI bus number"},
			{Name: "spiDevice", Label: "SPI Device", Type: "number", Default: 0, Description: "SPI device number"},
			{Name: "resetPin", Label: "Reset Pin", Type: "number", Default: 0, Description: "GPIO reset pin"},
			{Name: "antennaGain", Label: "Antenna Gain", Type: "number", Default: 4, Description: "Antenna gain (0-7)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or data input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "RFID card data"},
		},
		Factory: NewRFIDExecutor,
	}); err != nil {
		return err
	}

	// NFC PN532 Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "nfc",
		Name:        "NFC PN532",
		Category:    node.NodeTypeInput,
		Description: "NFC PN532 reader/writer (13.56MHz)",
		Icon:        "smartphone",
		Color:       "#0277BD",
		Properties: []node.PropertySchema{
			{Name: "interfaceType", Label: "Interface", Type: "select", Default: "i2c", Description: "Communication interface", Options: []string{"i2c", "spi", "uart"}},
			{Name: "i2cBus", Label: "I2C Bus", Type: "number", Default: 1, Description: "I2C bus number"},
			{Name: "i2cAddress", Label: "I2C Address", Type: "number", Default: 0x24, Description: "I2C device address"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or data input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "NFC tag data"},
		},
		Factory: NewNFCExecutor,
	}); err != nil {
		return err
	}

	// IR (Infrared) Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "ir",
		Name:        "IR Transceiver",
		Category:    node.NodeTypeInput,
		Description: "Infrared transmit/receive with NEC, RC5, RC6, raw protocols",
		Icon:        "radio",
		Color:       "#B71C1C",
		Properties: []node.PropertySchema{
			{Name: "txPin", Label: "TX Pin", Type: "number", Default: 0, Description: "GPIO pin for IR transmitter"},
			{Name: "rxPin", Label: "RX Pin", Type: "number", Default: 0, Description: "GPIO pin for IR receiver"},
			{Name: "protocol", Label: "Protocol", Type: "select", Default: "NEC", Description: "IR protocol", Options: []string{"NEC", "RC5", "RC6", "Sony", "Samsung", "raw"}},
			{Name: "operation", Label: "Operation", Type: "select", Default: "send", Description: "IR operation", Options: []string{"send", "receive", "learn"}},
			{Name: "frequency", Label: "Carrier Frequency (Hz)", Type: "number", Default: 38000, Description: "Carrier frequency in Hz"},
			{Name: "repeatCount", Label: "Repeat Count", Type: "number", Default: 1, Description: "Number of transmission repeats"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "IR command data"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "IR signal data"},
		},
		Factory: NewIRExecutor,
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
