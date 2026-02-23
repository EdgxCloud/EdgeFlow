package plugin

import (
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	coreNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/core"
)

// CorePlugin core plugin (always loaded)
type CorePlugin struct {
	*BasePlugin
}

// NewCorePlugin create core plugin
func NewCorePlugin() *CorePlugin {
	metadata := Metadata{
		Name:        "core",
		Version:     "1.0.0",
		Description: "Core nodes for EdgeFlow - always required",
		Author:      "EdgeFlow Team",
		Category:    CategoryCore,
		License:     "Apache-2.0",
		Config:      make(map[string]interface{}),
	}

	return &CorePlugin{
		BasePlugin: NewBasePlugin(metadata),
	}
}

// Load load plugin
func (p *CorePlugin) Load() error {
	// No special logic needed - nodes are defined in Nodes()
	return nil
}

// Unload unload from memory
func (p *CorePlugin) Unload() error {
	// Core plugin cannot be unloaded
	return nil
}

// Nodes list of provided nodes
func (p *CorePlugin) Nodes() []NodeDefinition {
	return []NodeDefinition{
		{
			Type:        "inject",
			Name:        "Inject",
			Category:    "core",
			Description: "Send periodic or manual messages",
			Icon:        "play",
			Color:       "#3FADB5",
			Inputs:      0,
			Outputs:     1,
			Factory:     func() node.Executor { return coreNodes.NewInjectNode() },
			Config: map[string]interface{}{
				"payload":  "",
				"interval": 1000,
				"repeat":   true,
			},
		},
		{
			Type:        "debug",
			Name:        "Debug",
			Category:    "core",
			Description: "Display message in console",
			Icon:        "bug",
			Color:       "#87A980",
			Inputs:      1,
			Outputs:     0,
			Factory:     func() node.Executor { return coreNodes.NewDebugNode() },
			Config: map[string]interface{}{
				"complete": false,
			},
		},
		{
			Type:        "function",
			Name:        "Function",
			Category:    "core",
			Description:   "Execute JavaScript code",
			Icon:        "code",
			Color:       "#E7E7AE",
			Inputs:      1,
			Outputs:     1,
			Factory:     func() node.Executor { return coreNodes.NewFunctionNode() },
			Config: map[string]interface{}{
				"func": "return msg;",
			},
		},
		{
			Type:        "if",
			Name:        "If",
			Category:    "core",
			Description: "Conditional routing",
			Icon:        "git-branch",
			Color:       "#C0DEED",
			Inputs:      1,
			Outputs:     2,
			Factory:     func() node.Executor { return coreNodes.NewIfNode() },
			Config: map[string]interface{}{
				"condition": "msg.payload > 0",
			},
		},
		{
			Type:        "delay",
			Name:        "Delay",
			Category:    "core",
			Description: "Delay message sending",
			Icon:        "clock",
			Color:       "#C0C0C0",
			Inputs:      1,
			Outputs:     1,
			Factory:     func() node.Executor { return coreNodes.NewDelayNode() },
			Config: map[string]interface{}{
				"delay": 1000,
			},
		},
	}
}

// RequiredMemory required memory (MB)
func (p *CorePlugin) RequiredMemory() uint64 {
	return 5 * 1024 * 1024 // 5MB
}

// RequiredDisk required disk space
func (p *CorePlugin) RequiredDisk() uint64 {
	return 1 * 1024 * 1024 // 1MB
}

// Dependencies dependencies
func (p *CorePlugin) Dependencies() []string {
	return []string{} // No dependencies
}
