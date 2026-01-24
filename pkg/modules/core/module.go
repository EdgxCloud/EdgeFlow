// Package core provides the core module for EdgeFlow
// This module is always loaded and provides essential nodes like inject, debug, function, etc.
package core

import (
	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/edgeflow/edgeflow/internal/plugin"
	"github.com/edgeflow/edgeflow/pkg/nodes/core"
)

// CoreModule is the core module that provides essential nodes
type CoreModule struct {
	*plugin.BasePlugin
	loaded bool
}

// NewCoreModule creates a new core module
func NewCoreModule() *CoreModule {
	metadata := plugin.Metadata{
		Name:        "core",
		Version:     "1.0.0",
		Description: "Core nodes for EdgeFlow - inject, debug, function, if, delay, and more",
		Author:      "EdgeFlow Team",
		Category:    plugin.CategoryCore,
		License:     "MIT",
		Keywords:    []string{"core", "essential", "inject", "debug", "function"},
		MinEdgeFlow: "0.1.0",
		Config:      map[string]interface{}{},
	}

	return &CoreModule{
		BasePlugin: plugin.NewBasePlugin(metadata),
		loaded:     false,
	}
}

// Load loads the core module
func (m *CoreModule) Load() error {
	m.SetStatus(plugin.StatusLoading)

	// Use the centralized registration in pkg/nodes/core/register.go
	registry := node.GetGlobalRegistry()
	if err := core.RegisterAllNodes(registry); err != nil {
		m.SetStatus(plugin.StatusError)
		return err
	}

	m.loaded = true
	m.SetStatus(plugin.StatusLoaded)
	return nil
}

// Unload unloads the core module
func (m *CoreModule) Unload() error {
	m.SetStatus(plugin.StatusUnloading)

	// Core module cannot be fully unloaded - it's required
	// Just mark as unloaded for status
	m.loaded = false
	m.SetStatus(plugin.StatusNotLoaded)
	return nil
}

// IsLoaded returns whether the module is loaded
func (m *CoreModule) IsLoaded() bool {
	return m.loaded
}

// Nodes returns the node definitions provided by this module
func (m *CoreModule) Nodes() []plugin.NodeDefinition {
	return []plugin.NodeDefinition{
		{
			Type:        "inject",
			Name:        "Inject",
			Category:    "core",
			Description: "Periodic trigger node that injects messages at intervals",
			Icon:        "play",
			Color:       "#87CEEB",
			Inputs:      0,
			Outputs:     1,
		},
		{
			Type:        "debug",
			Name:        "Debug",
			Category:    "core",
			Description: "Debug output node for inspecting messages",
			Icon:        "bug",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     0,
		},
		{
			Type:        "function",
			Name:        "Function",
			Category:    "core",
			Description: "Execute custom JavaScript code",
			Icon:        "code",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "if",
			Name:        "If",
			Category:    "core",
			Description: "Conditional routing based on expressions",
			Icon:        "git-branch",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     2,
		},
		{
			Type:        "delay",
			Name:        "Delay",
			Category:    "core",
			Description: "Delay message delivery by specified time",
			Icon:        "clock",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "template",
			Name:        "Template",
			Category:    "core",
			Description: "Mustache template rendering",
			Icon:        "file-text",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "switch",
			Name:        "Switch",
			Category:    "core",
			Description: "Multi-way routing based on rules",
			Icon:        "git-merge",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     4,
		},
		{
			Type:        "change",
			Name:        "Change",
			Category:    "core",
			Description: "Set, change, delete, or move message properties",
			Icon:        "edit",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "range",
			Name:        "Range",
			Category:    "core",
			Description: "Map values from one range to another",
			Icon:        "sliders",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "split",
			Name:        "Split",
			Category:    "core",
			Description: "Split arrays, objects, or strings into sequence",
			Icon:        "scissors",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "join",
			Name:        "Join",
			Category:    "core",
			Description: "Join message sequences into single message",
			Icon:        "link",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "catch",
			Name:        "Catch",
			Category:    "core",
			Description: "Catch errors from other nodes",
			Icon:        "alert-triangle",
			Color:       "#FF6B6B",
			Inputs:      0,
			Outputs:     1,
		},
		{
			Type:        "status",
			Name:        "Status",
			Category:    "core",
			Description: "Monitor node status changes",
			Icon:        "activity",
			Color:       "#87CEEB",
			Inputs:      0,
			Outputs:     1,
		},
		{
			Type:        "complete",
			Name:        "Complete",
			Category:    "core",
			Description: "Trigger on node completion",
			Icon:        "check-circle",
			Color:       "#87CEEB",
			Inputs:      0,
			Outputs:     1,
		},
		{
			Type:        "exec",
			Name:        "Exec",
			Category:    "core",
			Description: "Execute shell commands",
			Icon:        "terminal",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "python",
			Name:        "Python",
			Category:    "core",
			Description: "Execute Python code",
			Icon:        "code",
			Color:       "#87CEEB",
			Inputs:      1,
			Outputs:     1,
		},
	}
}

// RequiredMemory returns the memory requirement in bytes
func (m *CoreModule) RequiredMemory() uint64 {
	return 10 * 1024 * 1024 // 10 MB
}

// RequiredDisk returns the disk requirement in bytes
func (m *CoreModule) RequiredDisk() uint64 {
	return 5 * 1024 * 1024 // 5 MB
}

// Dependencies returns the list of required plugins
func (m *CoreModule) Dependencies() []string {
	return []string{} // No dependencies - this is the base module
}

// init registers the core module with the global registry
func init() {
	registry := plugin.GetRegistry()
	module := NewCoreModule()
	registry.Register(module)
}
