package network

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterAllNodes registers all network nodes
func RegisterAllNodes(registry *node.Registry) {
	// ============================================
	// HTTP NODES (3 nodes)
	// ============================================

	// HTTP Request
	registry.Register(&node.NodeInfo{
		Type:        "http-request",
		Name:        "HTTP Request",
		Category:    node.NodeTypeProcessing,
		Description: "Send HTTP requests with support for all methods",
		Icon:        "globe",
		Color:       "#3b82f6",
		Properties: []node.PropertySchema{
			{Name: "method", Label: "Method", Type: "select", Default: "GET", Required: true, Description: "HTTP method", Options: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}},
			{Name: "url", Label: "URL", Type: "string", Default: "", Description: "Target URL (can be set via msg.url)", Placeholder: "https://api.example.com/endpoint"},
			{Name: "timeout", Label: "Timeout (s)", Type: "number", Default: 30, Description: "Request timeout in seconds", Min: node.FloatPtr(1), Max: node.FloatPtr(300)},
			{Name: "retryCount", Label: "Retry Count", Type: "number", Default: 0, Description: "Number of retries on failure", Min: node.FloatPtr(0), Max: node.FloatPtr(10)},
			{Name: "retryDelay", Label: "Retry Delay (ms)", Type: "number", Default: 1000, Description: "Delay between retries in milliseconds"},
			{Name: "followRedirect", Label: "Follow Redirects", Type: "boolean", Default: true, Description: "Automatically follow HTTP redirects"},
			{Name: "insecureSkipTls", Label: "Skip TLS Verify", Type: "boolean", Default: false, Description: "Skip TLS certificate verification"},
			{Name: "basicAuthUsername", Label: "Basic Auth Username", Type: "string", Default: "", Description: "Username for HTTP Basic Authentication"},
			{Name: "basicAuthPassword", Label: "Basic Auth Password", Type: "password", Default: "", Description: "Password for HTTP Basic Authentication"},
			{Name: "proxyUrl", Label: "Proxy URL", Type: "string", Default: "", Description: "HTTP proxy server URL"},
			{Name: "caCertPath", Label: "CA Certificate Path", Type: "string", Default: "", Description: "Custom CA certificate file path"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Request data (url, headers, payload)"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Response", Type: "object", Description: "HTTP response (statusCode, headers, body)"},
		},
		Factory: NewHTTPRequestExecutor,
	})

	// HTTP Webhook
	registry.Register(&node.NodeInfo{
		Type:        "http-webhook",
		Name:        "HTTP Webhook",
		Category:    node.NodeTypeInput,
		Description: "Receive HTTP requests via webhook",
		Icon:        "webhook",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "path", Label: "Path", Type: "string", Default: "/webhook", Required: true, Description: "Webhook endpoint path (e.g. /webhook/myflow)"},
			{Name: "method", Label: "Method", Type: "select", Default: "POST", Description: "Accepted HTTP method", Options: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "ALL"}},
			{Name: "authType", Label: "Authentication", Type: "select", Default: "none", Description: "Authentication type", Options: []string{"none", "basic", "bearer", "apikey"}},
			{Name: "authValue", Label: "Auth Value", Type: "string", Default: "", Description: "Token/key value for authentication"},
			{Name: "rawBody", Label: "Raw Body", Type: "boolean", Default: false, Description: "Return raw body instead of parsed JSON"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Request", Type: "object", Description: "Incoming HTTP request (headers, body, query)"},
		},
		Factory: NewHTTPWebhookExecutor,
	})

	// HTTP Response
	registry.Register(&node.NodeInfo{
		Type:        "http-response",
		Name:        "HTTP Response",
		Category:    node.NodeTypeOutput,
		Description: "Send HTTP response",
		Icon:        "send",
		Color:       "#06b6d4",
		Properties: []node.PropertySchema{
			{Name: "statusCode", Label: "Status Code", Type: "number", Default: 200, Description: "HTTP response status code", Min: node.FloatPtr(100), Max: node.FloatPtr(599)},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Response data (payload becomes body)"},
		},
		Factory: NewHTTPResponseExecutor,
	})

	// ============================================
	// MQTT NODES (2 nodes)
	// ============================================

	// MQTT Input
	registry.Register(&node.NodeInfo{
		Type:        "mqtt-in",
		Name:        "MQTT Input",
		Category:    node.NodeTypeInput,
		Description: "Subscribe to a topic and receive messages from MQTT broker",
		Icon:        "message-square",
		Color:       "#22c55e",
		Properties: []node.PropertySchema{
			{Name: "broker", Label: "Broker URL", Type: "string", Default: "tcp://localhost:1883", Required: true, Description: "MQTT broker address (tcp://host:port)"},
			{Name: "topic", Label: "Topic", Type: "string", Default: "", Required: true, Description: "Topic to subscribe (supports +/# wildcards)"},
			{Name: "qos", Label: "QoS", Type: "select", Default: "0", Description: "Quality of Service level", Options: []string{"0", "1", "2"}},
			{Name: "clientId", Label: "Client ID", Type: "string", Default: "", Description: "MQTT client identifier (auto-generated if empty)"},
			{Name: "username", Label: "Username", Type: "string", Default: "", Description: "Broker authentication username"},
			{Name: "password", Label: "Password", Type: "password", Default: "", Description: "Broker authentication password"},
			{Name: "cleanSession", Label: "Clean Session", Type: "boolean", Default: true, Description: "Start with a clean session"},
			{Name: "autoReconnect", Label: "Auto Reconnect", Type: "boolean", Default: true, Description: "Automatically reconnect on disconnect"},
			{Name: "keepAlive", Label: "Keep Alive (s)", Type: "number", Default: 60, Description: "Keep alive interval in seconds", Min: node.FloatPtr(0), Max: node.FloatPtr(65535)},
			{Name: "connectTimeout", Label: "Connect Timeout (s)", Type: "number", Default: 30, Description: "Connection timeout in seconds"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Message", Type: "object", Description: "Received MQTT message (topic, payload, qos)"},
		},
		Factory: NewMQTTInExecutor,
	})

	// MQTT Output
	registry.Register(&node.NodeInfo{
		Type:        "mqtt-out",
		Name:        "MQTT Output",
		Category:    node.NodeTypeOutput,
		Description: "Publish messages to MQTT broker",
		Icon:        "send",
		Color:       "#16a34a",
		Properties: []node.PropertySchema{
			{Name: "broker", Label: "Broker URL", Type: "string", Default: "tcp://localhost:1883", Required: true, Description: "MQTT broker address (tcp://host:port)"},
			{Name: "topic", Label: "Topic", Type: "string", Default: "", Required: true, Description: "Topic to publish to"},
			{Name: "qos", Label: "QoS", Type: "select", Default: "0", Description: "Quality of Service level", Options: []string{"0", "1", "2"}},
			{Name: "retain", Label: "Retain", Type: "boolean", Default: false, Description: "Retain message on broker"},
			{Name: "clientId", Label: "Client ID", Type: "string", Default: "", Description: "MQTT client identifier (auto-generated if empty)"},
			{Name: "username", Label: "Username", Type: "string", Default: "", Description: "Broker authentication username"},
			{Name: "password", Label: "Password", Type: "password", Default: "", Description: "Broker authentication password"},
			{Name: "cleanSession", Label: "Clean Session", Type: "boolean", Default: true, Description: "Start with a clean session"},
			{Name: "autoReconnect", Label: "Auto Reconnect", Type: "boolean", Default: true, Description: "Automatically reconnect on disconnect"},
			{Name: "keepAlive", Label: "Keep Alive (s)", Type: "number", Default: 60, Description: "Keep alive interval in seconds", Min: node.FloatPtr(0), Max: node.FloatPtr(65535)},
			{Name: "connectTimeout", Label: "Connect Timeout (s)", Type: "number", Default: 30, Description: "Connection timeout in seconds"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Message", Type: "any", Description: "Message payload to publish"},
		},
		Factory: NewMQTTOutExecutor,
	})

	// ============================================
	// WEBSOCKET NODE (1 node)
	// ============================================

	// WebSocket Client
	registry.Register(&node.NodeInfo{
		Type:        "websocket-client",
		Name:        "WebSocket Client",
		Category:    node.NodeTypeProcessing,
		Description: "Connect to WebSocket server and send/receive messages",
		Icon:        "radio",
		Color:       "#a855f7",
		Properties: []node.PropertySchema{
			{Name: "url", Label: "URL", Type: "string", Default: "", Required: true, Description: "WebSocket server URL (ws:// or wss://)"},
			{Name: "autoReconnect", Label: "Auto Reconnect", Type: "boolean", Default: true, Description: "Automatically reconnect on disconnect"},
			{Name: "reconnectDelay", Label: "Reconnect Delay (ms)", Type: "number", Default: 5000, Description: "Delay between reconnect attempts"},
			{Name: "maxReconnectAttempts", Label: "Max Reconnect Attempts", Type: "number", Default: 0, Description: "Maximum reconnect attempts (0 = unlimited)"},
			{Name: "pingInterval", Label: "Ping Interval (s)", Type: "number", Default: 30, Description: "WebSocket ping interval in seconds"},
			{Name: "handshakeTimeout", Label: "Handshake Timeout (s)", Type: "number", Default: 10, Description: "Connection handshake timeout"},
			{Name: "enableCompression", Label: "Enable Compression", Type: "boolean", Default: false, Description: "Enable permessage-deflate compression"},
			{Name: "maxMessageSize", Label: "Max Message Size (bytes)", Type: "number", Default: 33554432, Description: "Maximum message size in bytes"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Send", Type: "any", Description: "Message to send to server"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Received", Type: "any", Description: "Message received from server"},
		},
		Factory: NewWebSocketClientExecutor,
	})

	// ============================================
	// TCP/UDP NODES (2 nodes)
	// ============================================

	// TCP Client
	registry.Register(&node.NodeInfo{
		Type:        "tcp-client",
		Name:        "TCP Client",
		Category:    node.NodeTypeProcessing,
		Description: "Connect to TCP server and send/receive data",
		Icon:        "server",
		Color:       "#f97316",
		Properties: []node.PropertySchema{
			{Name: "host", Label: "Host", Type: "string", Default: "localhost", Required: true, Description: "TCP server hostname or IP"},
			{Name: "port", Label: "Port", Type: "number", Default: 0, Required: true, Description: "TCP server port number"},
			{Name: "timeout", Label: "Timeout (s)", Type: "number", Default: 10, Description: "Connection timeout in seconds"},
			{Name: "autoReconnect", Label: "Auto Reconnect", Type: "boolean", Default: true, Description: "Automatically reconnect on disconnect"},
			{Name: "reconnectDelay", Label: "Reconnect Delay (ms)", Type: "number", Default: 5000, Description: "Delay between reconnect attempts"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Send", Type: "any", Description: "Data to send to server"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Received", Type: "any", Description: "Data received from server"},
		},
		Factory: NewTCPClientExecutor,
	})

	// UDP
	registry.Register(&node.NodeInfo{
		Type:        "udp",
		Name:        "UDP",
		Category:    node.NodeTypeProcessing,
		Description: "Send and receive data via UDP",
		Icon:        "radio",
		Color:       "#eab308",
		Properties: []node.PropertySchema{
			{Name: "mode", Label: "Mode", Type: "select", Default: "listen", Required: true, Description: "UDP operation mode", Options: []string{"listen", "send"}},
			{Name: "host", Label: "Host", Type: "string", Default: "", Description: "Target host (required for send mode)"},
			{Name: "port", Label: "Port", Type: "number", Default: 0, Required: true, Description: "UDP port number"},
			{Name: "bufferSize", Label: "Buffer Size", Type: "number", Default: 4096, Description: "Receive buffer size in bytes"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Send", Type: "any", Description: "Data to send (send mode)"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Received", Type: "any", Description: "Received UDP data"},
		},
		Factory: NewUDPExecutor,
	})

	// ============================================
	// PARSER NODES (4 nodes)
	// ============================================

	// JSON Parser
	registry.Register(&node.NodeInfo{
		Type:        "json-parser",
		Name:        "JSON Parser",
		Category:    node.NodeTypeProcessing,
		Description: "Convert JSON to object and vice versa",
		Icon:        "braces",
		Color:       "#64748b",
		Properties: []node.PropertySchema{
			{Name: "action", Label: "Action", Type: "select", Default: "parse", Required: true, Description: "Parse JSON string to object or stringify object to JSON", Options: []string{"parse", "stringify"}},
			{Name: "property", Label: "Property", Type: "string", Default: "payload", Description: "Message property to parse/stringify"},
			{Name: "target", Label: "Target", Type: "string", Default: "payload", Description: "Property to store result"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to parse or stringify"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Parsed/stringified result"},
		},
		Factory: NewJSONParserExecutor,
	})

	// XML Parser
	registry.Register(&node.NodeInfo{
		Type:        "xml-parser",
		Name:        "XML Parser",
		Category:    node.NodeTypeProcessing,
		Description: "Convert XML to JSON and vice versa",
		Icon:        "code",
		Color:       "#78716c",
		Properties: []node.PropertySchema{
			{Name: "action", Label: "Action", Type: "select", Default: "parse", Required: true, Description: "Parse XML to JSON or stringify JSON to XML", Options: []string{"parse", "stringify"}},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "XML or JSON data"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Converted result"},
		},
		Factory: NewXMLParserExecutor,
	})

	// CSV Parser
	registry.Register(&node.NodeInfo{
		Type:        "csv-parser",
		Name:        "CSV Parser",
		Category:    node.NodeTypeProcessing,
		Description: "Convert CSV to JSON and vice versa",
		Icon:        "table",
		Color:       "#059669",
		Properties: []node.PropertySchema{
			{Name: "action", Label: "Action", Type: "select", Default: "parse", Required: true, Description: "Parse CSV to JSON array or stringify to CSV", Options: []string{"parse", "stringify"}},
			{Name: "delimiter", Label: "Delimiter", Type: "string", Default: ",", Description: "Column delimiter character"},
			{Name: "hasHeader", Label: "Has Header Row", Type: "boolean", Default: true, Description: "First row contains column headers"},
			{Name: "skipRows", Label: "Skip Rows", Type: "number", Default: 0, Description: "Number of rows to skip from start"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "CSV string or JSON array"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Parsed/stringified result"},
		},
		Factory: NewCSVParserExecutor,
	})

	// YAML Parser
	registry.Register(&node.NodeInfo{
		Type:        "yaml-parser",
		Name:        "YAML Parser",
		Category:    node.NodeTypeProcessing,
		Description: "Convert YAML to JSON and vice versa",
		Icon:        "file-text",
		Color:       "#7c3aed",
		Properties: []node.PropertySchema{
			{Name: "action", Label: "Action", Type: "select", Default: "parse", Required: true, Description: "Parse YAML to JSON or stringify JSON to YAML", Options: []string{"parse", "stringify"}},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "YAML or JSON data"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Converted result"},
		},
		Factory: NewYAMLParserExecutor,
	})

	// ============================================
	// FILE NODES (3 nodes)
	// ============================================

	// File Watch
	registry.Register(&node.NodeInfo{
		Type:        "watch",
		Name:        "File Watch",
		Category:    node.NodeTypeInput,
		Description: "Monitor file/directory changes",
		Icon:        "eye",
		Color:       "#d946ef",
		Properties: []node.PropertySchema{
			{Name: "path", Label: "Path", Type: "string", Default: "", Required: true, Description: "File or directory path to watch"},
			{Name: "recursive", Label: "Recursive", Type: "boolean", Default: false, Description: "Watch subdirectories recursively"},
			{Name: "pattern", Label: "File Pattern", Type: "string", Default: "", Description: "Glob pattern to filter files (e.g. *.log)"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Event", Type: "object", Description: "File change event (path, type, timestamp)"},
		},
		Factory: NewWatchExecutor,
	})

	// File Read
	registry.Register(&node.NodeInfo{
		Type:        "file-in",
		Name:        "File Read",
		Category:    node.NodeTypeInput,
		Description: "Read file contents",
		Icon:        "file",
		Color:       "#0ea5e9",
		Properties: []node.PropertySchema{
			{Name: "filename", Label: "Filename", Type: "string", Default: "", Description: "File path to read (can be set via msg.filename)"},
			{Name: "format", Label: "Format", Type: "select", Default: "utf8", Description: "Output format", Options: []string{"utf8", "binary", "lines"}},
			{Name: "encoding", Label: "Encoding", Type: "string", Default: "utf-8", Description: "File character encoding"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger file read (filename in msg.filename)"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Content", Type: "any", Description: "File content"},
		},
		Factory: NewFileInExecutor,
	})

	// File Write
	registry.Register(&node.NodeInfo{
		Type:        "file-out",
		Name:        "File Write",
		Category:    node.NodeTypeOutput,
		Description: "Write content to file",
		Icon:        "save",
		Color:       "#f43f5e",
		Properties: []node.PropertySchema{
			{Name: "filename", Label: "Filename", Type: "string", Default: "", Description: "File path to write (can be set via msg.filename)"},
			{Name: "action", Label: "Action", Type: "select", Default: "write", Description: "Write mode", Options: []string{"write", "append"}},
			{Name: "createDir", Label: "Create Directory", Type: "boolean", Default: true, Description: "Create parent directories if not exist"},
			{Name: "encoding", Label: "Encoding", Type: "string", Default: "utf-8", Description: "File character encoding"},
			{Name: "addNewline", Label: "Add Newline", Type: "boolean", Default: false, Description: "Append newline at end of content"},
			{Name: "overwriteFile", Label: "Overwrite File", Type: "boolean", Default: true, Description: "Overwrite existing file content"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Content", Type: "any", Description: "Content to write to file"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Confirmation", Type: "object", Description: "Write confirmation (filename, bytes)"},
		},
		Factory: NewFileOutExecutor,
	})

	// ============================================
	// SERIAL NODES (2 nodes)
	// ============================================

	// Serial Input
	registry.Register(&node.NodeInfo{
		Type:        "serial-in",
		Name:        "Serial Input",
		Category:    node.NodeTypeInput,
		Description: "Receive data from serial port",
		Icon:        "terminal",
		Color:       "#ec4899",
		Properties: []node.PropertySchema{
			{Name: "port", Label: "Serial Port", Type: "string", Default: "", Required: true, Description: "Serial port name (e.g. /dev/ttyUSB0, COM3)"},
			{Name: "baudRate", Label: "Baud Rate", Type: "select", Default: "9600", Required: true, Description: "Communication speed", Options: []string{"1200", "2400", "4800", "9600", "19200", "38400", "57600", "115200"}},
			{Name: "dataBits", Label: "Data Bits", Type: "select", Default: "8", Description: "Data bits per frame", Options: []string{"5", "6", "7", "8"}},
			{Name: "stopBits", Label: "Stop Bits", Type: "select", Default: "1", Description: "Stop bits", Options: []string{"1", "2"}},
			{Name: "parity", Label: "Parity", Type: "select", Default: "none", Description: "Parity check mode", Options: []string{"none", "even", "odd"}},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Data", Type: "any", Description: "Received serial data"},
		},
		Factory: NewSerialInExecutor,
	})

	// HTTP In
	registry.Register(&node.NodeInfo{
		Type:        "http-in",
		Name:        "HTTP In",
		Category:    node.NodeTypeInput,
		Description: "Create an HTTP endpoint to receive requests",
		Icon:        "globe",
		Color:       "#2563eb",
		Properties: []node.PropertySchema{
			{Name: "path", Label: "Path", Type: "string", Default: "/api/custom", Required: true, Description: "HTTP endpoint path (supports :param for path parameters)"},
			{Name: "method", Label: "Method", Type: "select", Default: "ALL", Description: "Accepted HTTP method", Options: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "ALL"}},
			{Name: "authType", Label: "Authentication", Type: "select", Default: "none", Description: "Authentication type", Options: []string{"none", "basic", "bearer", "apikey"}},
			{Name: "authValue", Label: "Auth Value", Type: "string", Default: "", Description: "Token/key value for authentication"},
			{Name: "rawBody", Label: "Raw Body", Type: "boolean", Default: false, Description: "Return raw body instead of parsed JSON"},
			{Name: "cors", Label: "Enable CORS", Type: "boolean", Default: false, Description: "Enable CORS headers"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Request", Type: "object", Description: "Incoming HTTP request (method, path, params, query, headers, body)"},
		},
		Factory: NewHTTPInExecutor,
	})

	// WebSocket In (Server)
	registry.Register(&node.NodeInfo{
		Type:        "websocket-in",
		Name:        "WebSocket In",
		Category:    node.NodeTypeInput,
		Description: "Accept WebSocket connections and receive messages (server mode)",
		Icon:        "radio",
		Color:       "#7c3aed",
		Properties: []node.PropertySchema{
			{Name: "path", Label: "Path", Type: "string", Default: "/ws/nodes", Required: true, Description: "WebSocket endpoint path"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Message", Type: "any", Description: "Messages received from connected clients"},
		},
		Factory: NewWebSocketInExecutor,
	})

	// WebSocket Out (Server)
	registry.Register(&node.NodeInfo{
		Type:        "websocket-out",
		Name:        "WebSocket Out",
		Category:    node.NodeTypeOutput,
		Description: "Send messages to connected WebSocket clients (server mode)",
		Icon:        "send",
		Color:       "#6d28d9",
		Properties: []node.PropertySchema{
			{Name: "path", Label: "Path", Type: "string", Default: "/ws/nodes", Required: true, Description: "WebSocket endpoint path (must match websocket-in)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Message", Type: "any", Description: "Message to broadcast to connected clients"},
		},
		Factory: NewWebSocketOutExecutor,
	})

	// Serial Output
	registry.Register(&node.NodeInfo{
		Type:        "serial-out",
		Name:        "Serial Output",
		Category:    node.NodeTypeOutput,
		Description: "Send data to serial port",
		Icon:        "terminal",
		Color:       "#db2777",
		Properties: []node.PropertySchema{
			{Name: "port", Label: "Serial Port", Type: "string", Default: "", Required: true, Description: "Serial port name (e.g. /dev/ttyUSB0, COM3)"},
			{Name: "baudRate", Label: "Baud Rate", Type: "select", Default: "9600", Required: true, Description: "Communication speed", Options: []string{"1200", "2400", "4800", "9600", "19200", "38400", "57600", "115200"}},
			{Name: "dataBits", Label: "Data Bits", Type: "select", Default: "8", Description: "Data bits per frame", Options: []string{"5", "6", "7", "8"}},
			{Name: "stopBits", Label: "Stop Bits", Type: "select", Default: "1", Description: "Stop bits", Options: []string{"1", "2"}},
			{Name: "parity", Label: "Parity", Type: "select", Default: "none", Description: "Parity check mode", Options: []string{"none", "even", "odd"}},
			{Name: "addNewline", Label: "Add Newline", Type: "boolean", Default: false, Description: "Append newline to sent data"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Data", Type: "any", Description: "Data to send to serial port"},
		},
		Factory: NewSerialOutExecutor,
	})
}
