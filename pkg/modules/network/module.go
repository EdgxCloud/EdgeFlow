// Package network provides the network module for EdgeFlow
// This module provides HTTP, MQTT, WebSocket, TCP, UDP, and other network nodes
package network

import (
	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/edgeflow/edgeflow/internal/plugin"
	"github.com/edgeflow/edgeflow/pkg/nodes/network"
)

// NetworkModule is the network module that provides network communication nodes
type NetworkModule struct {
	*plugin.BasePlugin
	loaded bool
}

// NewNetworkModule creates a new network module
func NewNetworkModule() *NetworkModule {
	metadata := plugin.Metadata{
		Name:        "network",
		Version:     "1.0.0",
		Description: "Network communication nodes - HTTP, MQTT, WebSocket, TCP, UDP, Serial",
		Author:      "EdgeFlow Team",
		Category:    plugin.CategoryNetwork,
		License:     "MIT",
		Keywords:    []string{"network", "http", "mqtt", "websocket", "tcp", "udp", "serial"},
		MinEdgeFlow: "0.1.0",
		Config: map[string]interface{}{
			"http_timeout":       30,
			"mqtt_keep_alive":    60,
			"websocket_ping":     30,
			"tcp_buffer_size":    4096,
			"serial_buffer_size": 1024,
		},
	}

	return &NetworkModule{
		BasePlugin: plugin.NewBasePlugin(metadata),
		loaded:     false,
	}
}

// Load loads the network module
func (m *NetworkModule) Load() error {
	m.SetStatus(plugin.StatusLoading)

	// Register all network nodes using the centralized RegisterAllNodes function
	registry := node.GetGlobalRegistry()
	network.RegisterAllNodes(registry)

	m.loaded = true
	m.SetStatus(plugin.StatusLoaded)
	return nil
}

// Unload unloads the network module
func (m *NetworkModule) Unload() error {
	m.SetStatus(plugin.StatusUnloading)
	m.loaded = false
	m.SetStatus(plugin.StatusNotLoaded)
	return nil
}

// IsLoaded returns whether the module is loaded
func (m *NetworkModule) IsLoaded() bool {
	return m.loaded
}

// Nodes returns the node definitions provided by this module
func (m *NetworkModule) Nodes() []plugin.NodeDefinition {
	return []plugin.NodeDefinition{
		// HTTP Nodes
		{
			Type:        "http-request",
			Name:        "HTTP Request",
			Category:    "network",
			Description: "Make HTTP requests (GET, POST, PUT, DELETE, PATCH)",
			Icon:        "globe",
			Color:       "#4CAF50",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "http-webhook",
			Name:        "HTTP Webhook",
			Category:    "network",
			Description: "Receive HTTP webhooks",
			Icon:        "webhook",
			Color:       "#4CAF50",
			Inputs:      0,
			Outputs:     1,
		},
		{
			Type:        "http-response",
			Name:        "HTTP Response",
			Category:    "network",
			Description: "Send HTTP response for incoming requests",
			Icon:        "arrow-left",
			Color:       "#4CAF50",
			Inputs:      1,
			Outputs:     0,
		},
		// MQTT Nodes
		{
			Type:        "mqtt-in",
			Name:        "MQTT In",
			Category:    "network",
			Description: "Subscribe to MQTT topics",
			Icon:        "radio",
			Color:       "#9C27B0",
			Inputs:      0,
			Outputs:     1,
		},
		{
			Type:        "mqtt-out",
			Name:        "MQTT Out",
			Category:    "network",
			Description: "Publish to MQTT topics",
			Icon:        "send",
			Color:       "#9C27B0",
			Inputs:      1,
			Outputs:     0,
		},
		// WebSocket Node
		{
			Type:        "websocket-client",
			Name:        "WebSocket Client",
			Category:    "network",
			Description: "WebSocket client connection",
			Icon:        "plug",
			Color:       "#2196F3",
			Inputs:      1,
			Outputs:     1,
		},
		// TCP/UDP Nodes
		{
			Type:        "tcp-client",
			Name:        "TCP Client",
			Category:    "network",
			Description: "TCP client connection",
			Icon:        "server",
			Color:       "#FF9800",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "udp",
			Name:        "UDP",
			Category:    "network",
			Description: "UDP send/receive",
			Icon:        "radio-tower",
			Color:       "#FF9800",
			Inputs:      1,
			Outputs:     1,
		},
		// Serial Nodes
		{
			Type:        "serial-in",
			Name:        "Serial In",
			Category:    "network",
			Description: "Read from serial port",
			Icon:        "usb",
			Color:       "#795548",
			Inputs:      0,
			Outputs:     1,
		},
		{
			Type:        "serial-out",
			Name:        "Serial Out",
			Category:    "network",
			Description: "Write to serial port",
			Icon:        "usb",
			Color:       "#795548",
			Inputs:      1,
			Outputs:     0,
		},
		// Parser Nodes
		{
			Type:        "json-parser",
			Name:        "JSON Parser",
			Category:    "network",
			Description: "Parse/stringify JSON data",
			Icon:        "braces",
			Color:       "#607D8B",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "xml-parser",
			Name:        "XML Parser",
			Category:    "network",
			Description: "Convert between XML and JSON",
			Icon:        "code",
			Color:       "#607D8B",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "csv-parser",
			Name:        "CSV Parser",
			Category:    "network",
			Description: "Parse/generate CSV data",
			Icon:        "table",
			Color:       "#607D8B",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "yaml-parser",
			Name:        "YAML Parser",
			Category:    "network",
			Description: "Convert between YAML and JSON",
			Icon:        "file-code",
			Color:       "#607D8B",
			Inputs:      1,
			Outputs:     1,
		},
		// File Nodes
		{
			Type:        "file-in",
			Name:        "File In",
			Category:    "network",
			Description: "Read file contents",
			Icon:        "file-input",
			Color:       "#009688",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "file-out",
			Name:        "File Out",
			Category:    "network",
			Description: "Write to file",
			Icon:        "file-output",
			Color:       "#009688",
			Inputs:      1,
			Outputs:     0,
		},
		{
			Type:        "watch",
			Name:        "Watch",
			Category:    "network",
			Description: "Watch file/directory for changes",
			Icon:        "eye",
			Color:       "#009688",
			Inputs:      0,
			Outputs:     1,
		},
	}
}

// RequiredMemory returns the memory requirement in bytes
func (m *NetworkModule) RequiredMemory() uint64 {
	return 50 * 1024 * 1024 // 50 MB
}

// RequiredDisk returns the disk requirement in bytes
func (m *NetworkModule) RequiredDisk() uint64 {
	return 10 * 1024 * 1024 // 10 MB
}

// Dependencies returns the list of required plugins
func (m *NetworkModule) Dependencies() []string {
	return []string{"core"}
}

// init registers the network module with the global registry
func init() {
	registry := plugin.GetRegistry()
	module := NewNetworkModule()
	registry.Register(module)
}
