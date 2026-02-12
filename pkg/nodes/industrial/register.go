package industrial

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterNodes registers all industrial nodes with the registry
func RegisterNodes(registry *node.Registry) error {
	// Modbus TCP Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "modbus-tcp",
		Name:        "Modbus TCP",
		Category:    node.NodeTypeInput,
		Description: "Modbus TCP client for industrial PLC communication",
		Icon:        "cpu",
		Color:       "#FF6B35",
		Properties: []node.PropertySchema{
			{
				Name:        "host",
				Label:       "Host",
				Type:        "string",
				Default:     "127.0.0.1",
				Required:    true,
				Description: "Modbus TCP server hostname or IP",
			},
			{
				Name:        "port",
				Label:       "Port",
				Type:        "number",
				Default:     502,
				Required:    true,
				Description: "Modbus TCP port (default 502)",
			},
			{
				Name:        "unitId",
				Label:       "Unit ID",
				Type:        "number",
				Default:     1,
				Required:    true,
				Description: "Modbus unit/slave ID (1-247)",
			},
			{
				Name:        "operation",
				Label:       "Operation",
				Type:        "select",
				Default:     "read_holding",
				Required:    true,
				Description: "Modbus operation to perform",
				Options:     []string{"read_coils", "read_discrete", "read_holding", "read_input", "write_coil", "write_register", "write_coils", "write_registers"},
			},
			{
				Name:        "address",
				Label:       "Start Address",
				Type:        "number",
				Default:     0,
				Required:    true,
				Description: "Starting register/coil address",
			},
			{
				Name:        "quantity",
				Label:       "Quantity",
				Type:        "number",
				Default:     1,
				Required:    false,
				Description: "Number of registers/coils to read",
			},
			{
				Name:        "timeout",
				Label:       "Timeout (ms)",
				Type:        "number",
				Default:     5000,
				Required:    false,
				Description: "Connection and read timeout in milliseconds",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or data input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Modbus response data"},
		},
		Factory: NewModbusTCPExecutor,
	}); err != nil {
		return err
	}

	// Modbus RTU Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "modbus-rtu",
		Name:        "Modbus RTU",
		Category:    node.NodeTypeInput,
		Description: "Modbus RTU client over serial port for industrial devices",
		Icon:        "cpu",
		Color:       "#FF6B35",
		Properties: []node.PropertySchema{
			{
				Name:        "port",
				Label:       "Serial Port",
				Type:        "string",
				Default:     "/dev/ttyUSB0",
				Required:    true,
				Description: "Serial port path (e.g., /dev/ttyUSB0, COM1)",
			},
			{
				Name:        "baudRate",
				Label:       "Baud Rate",
				Type:        "select",
				Default:     "9600",
				Required:    true,
				Description: "Serial communication speed",
				Options:     []string{"1200", "2400", "4800", "9600", "19200", "38400", "57600", "115200"},
			},
			{
				Name:        "dataBits",
				Label:       "Data Bits",
				Type:        "select",
				Default:     "8",
				Required:    true,
				Description: "Number of data bits",
				Options:     []string{"7", "8"},
			},
			{
				Name:        "stopBits",
				Label:       "Stop Bits",
				Type:        "select",
				Default:     "1",
				Required:    true,
				Description: "Number of stop bits",
				Options:     []string{"1", "2"},
			},
			{
				Name:        "parity",
				Label:       "Parity",
				Type:        "select",
				Default:     "none",
				Required:    true,
				Description: "Parity checking mode",
				Options:     []string{"none", "odd", "even"},
			},
			{
				Name:        "unitId",
				Label:       "Unit ID",
				Type:        "number",
				Default:     1,
				Required:    true,
				Description: "Modbus unit/slave ID (1-247)",
			},
			{
				Name:        "operation",
				Label:       "Operation",
				Type:        "select",
				Default:     "read_holding",
				Required:    true,
				Description: "Modbus operation to perform",
				Options:     []string{"read_coils", "read_discrete", "read_holding", "read_input", "write_coil", "write_register", "write_coils", "write_registers"},
			},
			{
				Name:        "address",
				Label:       "Start Address",
				Type:        "number",
				Default:     0,
				Required:    true,
				Description: "Starting register/coil address",
			},
			{
				Name:        "quantity",
				Label:       "Quantity",
				Type:        "number",
				Default:     1,
				Required:    false,
				Description: "Number of registers/coils to read",
			},
			{
				Name:        "timeout",
				Label:       "Timeout (ms)",
				Type:        "number",
				Default:     1000,
				Required:    false,
				Description: "Communication timeout in milliseconds",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or data input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Modbus response data"},
		},
		Factory: NewModbusRTUExecutor,
	}); err != nil {
		return err
	}

	// OPC-UA Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "opcua",
		Name:        "OPC-UA",
		Category:    node.NodeTypeInput,
		Description: "OPC-UA client for industrial automation (Industry 4.0)",
		Icon:        "server",
		Color:       "#4A90D9",
		Properties: []node.PropertySchema{
			{
				Name:        "endpoint",
				Label:       "Endpoint URL",
				Type:        "string",
				Default:     "opc.tcp://localhost:4840",
				Required:    true,
				Description: "OPC-UA server endpoint URL",
			},
			{
				Name:        "securityMode",
				Label:       "Security Mode",
				Type:        "select",
				Default:     "none",
				Required:    true,
				Description: "Message security mode",
				Options:     []string{"none", "sign", "signandencrypt"},
			},
			{
				Name:        "securityPolicy",
				Label:       "Security Policy",
				Type:        "select",
				Default:     "none",
				Required:    true,
				Description: "Security policy for encryption",
				Options:     []string{"none", "basic128rsa15", "basic256", "basic256sha256"},
			},
			{
				Name:        "username",
				Label:       "Username",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Username for authentication (leave empty for anonymous)",
			},
			{
				Name:        "password",
				Label:       "Password",
				Type:        "password",
				Default:     "",
				Required:    false,
				Description: "Password for authentication",
			},
			{
				Name:        "operation",
				Label:       "Operation",
				Type:        "select",
				Default:     "read",
				Required:    true,
				Description: "OPC-UA operation to perform",
				Options:     []string{"read", "write", "browse", "read_multiple", "write_multiple", "get_endpoints", "get_namespace"},
			},
			{
				Name:        "nodeId",
				Label:       "Node ID",
				Type:        "string",
				Default:     "ns=0;i=2258",
				Required:    false,
				Description: "OPC-UA Node ID (e.g., ns=2;s=MyVariable)",
			},
			{
				Name:        "timeout",
				Label:       "Timeout (ms)",
				Type:        "number",
				Default:     10000,
				Required:    false,
				Description: "Connection and operation timeout",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or data input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "OPC-UA response data"},
		},
		Factory: NewOPCUAExecutor,
	}); err != nil {
		return err
	}

	// BACnet Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "bacnet",
		Name:        "BACnet",
		Category:    node.NodeTypeInput,
		Description: "BACnet/IP client for building automation systems",
		Icon:        "building",
		Color:       "#2E7D32",
		Properties: []node.PropertySchema{
			{Name: "host", Label: "Host", Type: "string", Default: "127.0.0.1", Required: true, Description: "BACnet device IP address"},
			{Name: "port", Label: "Port", Type: "number", Default: 47808, Description: "BACnet/IP port (default 47808/0xBAC0)"},
			{Name: "deviceId", Label: "Device ID", Type: "number", Default: 0, Description: "Target BACnet device instance"},
			{Name: "operation", Label: "Operation", Type: "select", Default: "read_property", Required: true, Description: "BACnet operation", Options: []string{"read_property", "write_property", "who_is"}},
			{Name: "objectType", Label: "Object Type", Type: "select", Default: "analog_input", Description: "BACnet object type", Options: []string{"analog_input", "analog_output", "binary_input", "binary_output", "multi_state_input", "multi_state_output"}},
			{Name: "objectInstance", Label: "Object Instance", Type: "number", Default: 0, Description: "Object instance number"},
			{Name: "timeout", Label: "Timeout (ms)", Type: "number", Default: 5000, Description: "Operation timeout"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or data input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "BACnet response data"},
		},
		Factory: NewBACnetExecutor,
	}); err != nil {
		return err
	}

	// PROFINET Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "profinet",
		Name:        "PROFINET",
		Category:    node.NodeTypeInput,
		Description: "PROFINET DCP discovery and I/O communication",
		Icon:        "network",
		Color:       "#1565C0",
		Properties: []node.PropertySchema{
			{Name: "interface", Label: "Network Interface", Type: "string", Default: "eth0", Required: true, Description: "Network interface for PROFINET"},
			{Name: "deviceName", Label: "Device Name", Type: "string", Default: "", Description: "Target PROFINET device name"},
			{Name: "operation", Label: "Operation", Type: "select", Default: "discover", Required: true, Description: "PROFINET operation", Options: []string{"discover", "identify", "read_io", "write_io", "get_diagnosis"}},
			{Name: "timeout", Label: "Timeout (ms)", Type: "number", Default: 5000, Description: "Operation timeout"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or data input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "PROFINET response data"},
		},
		Factory: NewProfinetExecutor,
	}); err != nil {
		return err
	}

	// CAN Bus Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "can-bus",
		Name:        "CAN Bus",
		Category:    node.NodeTypeInput,
		Description: "CAN bus communication via SocketCAN",
		Icon:        "cpu",
		Color:       "#E65100",
		Properties: []node.PropertySchema{
			{Name: "interface", Label: "CAN Interface", Type: "string", Default: "can0", Required: true, Description: "CAN interface name (e.g., can0, vcan0)"},
			{Name: "bitrate", Label: "Bitrate", Type: "select", Default: "500000", Description: "CAN bus bitrate", Options: []string{"125000", "250000", "500000", "1000000"}},
			{Name: "operation", Label: "Operation", Type: "select", Default: "receive", Required: true, Description: "CAN operation", Options: []string{"send", "receive", "listen", "status"}},
			{Name: "canId", Label: "CAN ID", Type: "number", Default: 0, Description: "CAN message ID for send"},
			{Name: "extended", Label: "Extended Frame", Type: "boolean", Default: false, Description: "Use 29-bit extended CAN ID"},
			{Name: "filterIds", Label: "Filter IDs", Type: "string", Default: "", Description: "Comma-separated CAN IDs to filter"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or data input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "CAN frame data"},
		},
		Factory: NewCANBusExecutor,
	}); err != nil {
		return err
	}

	return nil
}

// init registers industrial nodes with the global registry
func init() {
	registry := node.GetGlobalRegistry()
	RegisterNodes(registry)
}
