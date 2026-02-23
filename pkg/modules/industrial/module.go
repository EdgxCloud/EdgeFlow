// Package industrial provides the industrial module for EdgeFlow
// This module provides Modbus and OPC-UA nodes for industrial protocol support
package industrial

import (
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/EdgxCloud/EdgeFlow/internal/plugin"
)

// IndustrialModule is the industrial module that provides industrial protocol nodes
type IndustrialModule struct {
	*plugin.BasePlugin
	loaded bool
}

// NewIndustrialModule creates a new industrial module
func NewIndustrialModule() *IndustrialModule {
	metadata := plugin.Metadata{
		Name:        "industrial",
		Version:     "1.0.0",
		Description: "Industrial protocol nodes - Modbus RTU/TCP, OPC-UA",
		Author:      "EdgeFlow Team",
		Category:    plugin.CategoryIndustrial,
		License:     "MIT",
		Keywords:    []string{"industrial", "modbus", "opcua", "plc", "scada", "automation"},
		MinEdgeFlow: "0.1.0",
		Config: map[string]interface{}{
			"modbus_timeout":    5,
			"modbus_retries":    3,
			"opcua_timeout":     10,
			"opcua_security":    "none",
		},
	}

	return &IndustrialModule{
		BasePlugin: plugin.NewBasePlugin(metadata),
		loaded:     false,
	}
}

// wrapFactory wraps a config-based factory function to match NodeFactory signature
func wrapFactory(fn func(config map[string]interface{}) (node.Executor, error)) node.NodeFactory {
	return func() node.Executor {
		exec, err := fn(make(map[string]interface{}))
		if err != nil {
			return nil
		}
		return exec
	}
}

// Load loads the industrial module
func (m *IndustrialModule) Load() error {
	m.SetStatus(plugin.StatusLoading)

	registry := node.GetGlobalRegistry()

	// Register Modbus Read node
	registry.Register(&node.NodeInfo{
		Type:        "modbus-read",
		Name:        "Modbus Read",
		Category:    node.NodeTypeProcessing,
		Description: "Read from Modbus RTU/TCP devices (coils, inputs, holding registers)",
		Icon:        "server",
		Color:       "#FF6F00",
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger message"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Read result"},
		},
		Factory: wrapFactory(NewModbusReadExecutor),
	})

	// Register Modbus Write node
	registry.Register(&node.NodeInfo{
		Type:        "modbus-write",
		Name:        "Modbus Write",
		Category:    node.NodeTypeProcessing,
		Description: "Write to Modbus RTU/TCP devices (coils, holding registers)",
		Icon:        "edit",
		Color:       "#FF6F00",
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to write"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Write result"},
		},
		Factory: wrapFactory(NewModbusWriteExecutor),
	})

	// Register OPC-UA Read node
	registry.Register(&node.NodeInfo{
		Type:        "opcua-read",
		Name:        "OPC-UA Read",
		Category:    node.NodeTypeProcessing,
		Description: "Read from OPC-UA server nodes",
		Icon:        "download",
		Color:       "#1565C0",
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger message"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Read result"},
		},
		Factory: wrapFactory(NewOPCUAReadExecutor),
	})

	// Register OPC-UA Write node
	registry.Register(&node.NodeInfo{
		Type:        "opcua-write",
		Name:        "OPC-UA Write",
		Category:    node.NodeTypeProcessing,
		Description: "Write to OPC-UA server nodes",
		Icon:        "upload",
		Color:       "#1565C0",
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to write"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Write result"},
		},
		Factory: wrapFactory(NewOPCUAWriteExecutor),
	})

	m.loaded = true
	m.SetStatus(plugin.StatusLoaded)
	return nil
}

// Unload unloads the industrial module
func (m *IndustrialModule) Unload() error {
	m.SetStatus(plugin.StatusUnloading)
	m.loaded = false
	m.SetStatus(plugin.StatusNotLoaded)
	return nil
}

// IsLoaded returns whether the module is loaded
func (m *IndustrialModule) IsLoaded() bool {
	return m.loaded
}

// Nodes returns the node definitions provided by this module
func (m *IndustrialModule) Nodes() []plugin.NodeDefinition {
	return []plugin.NodeDefinition{
		{
			Type:        "modbus-read",
			Name:        "Modbus Read",
			Category:    "industrial",
			Description: "Read from Modbus RTU/TCP devices (coils, inputs, holding registers)",
			Icon:        "server",
			Color:       "#FF6F00",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "modbus-write",
			Name:        "Modbus Write",
			Category:    "industrial",
			Description: "Write to Modbus RTU/TCP devices (coils, holding registers)",
			Icon:        "edit",
			Color:       "#FF6F00",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "opcua-read",
			Name:        "OPC-UA Read",
			Category:    "industrial",
			Description: "Read from OPC-UA server nodes",
			Icon:        "download",
			Color:       "#1565C0",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "opcua-write",
			Name:        "OPC-UA Write",
			Category:    "industrial",
			Description: "Write to OPC-UA server nodes",
			Icon:        "upload",
			Color:       "#1565C0",
			Inputs:      1,
			Outputs:     1,
		},
	}
}

// RequiredMemory returns the memory requirement in bytes
func (m *IndustrialModule) RequiredMemory() uint64 {
	return 50 * 1024 * 1024 // 50 MB
}

// RequiredDisk returns the disk requirement in bytes
func (m *IndustrialModule) RequiredDisk() uint64 {
	return 20 * 1024 * 1024 // 20 MB (includes protocol libraries)
}

// Dependencies returns the list of required plugins
func (m *IndustrialModule) Dependencies() []string {
	return []string{"core", "network"}
}

// init registers the industrial module with the global registry
func init() {
	registry := plugin.GetRegistry()
	module := NewIndustrialModule()
	registry.Register(module)
}
