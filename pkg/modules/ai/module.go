// Package ai provides the AI module for EdgeFlow
// This module provides OpenAI, Anthropic, and Ollama nodes for AI/LLM integration
package ai

import (
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/EdgxCloud/EdgeFlow/internal/plugin"
	"github.com/EdgxCloud/EdgeFlow/pkg/nodes/ai"
)

// AIModule is the AI module that provides AI/LLM integration nodes
type AIModule struct {
	*plugin.BasePlugin
	loaded bool
}

// NewAIModule creates a new AI module
func NewAIModule() *AIModule {
	metadata := plugin.Metadata{
		Name:        "ai",
		Version:     "1.0.0",
		Description: "AI and LLM integration nodes - OpenAI, Anthropic, Ollama",
		Author:      "EdgeFlow Team",
		Category:    plugin.CategoryAI,
		License:     "MIT",
		Keywords:    []string{"ai", "openai", "anthropic", "ollama", "llm", "gpt", "claude"},
		MinEdgeFlow: "0.1.0",
		Config: map[string]interface{}{
			"default_timeout":     120,
			"max_tokens":          4096,
			"default_temperature": 0.7,
		},
	}

	return &AIModule{
		BasePlugin: plugin.NewBasePlugin(metadata),
		loaded:     false,
	}
}

// Load loads the AI module
func (m *AIModule) Load() error {
	m.SetStatus(plugin.StatusLoading)

	// Register all AI nodes using the centralized RegisterAllNodes function
	registry := node.GetGlobalRegistry()
	if err := ai.RegisterAllNodes(registry); err != nil {
		m.SetStatus(plugin.StatusError)
		return err
	}

	m.loaded = true
	m.SetStatus(plugin.StatusLoaded)
	return nil
}

// Unload unloads the AI module
func (m *AIModule) Unload() error {
	m.SetStatus(plugin.StatusUnloading)
	m.loaded = false
	m.SetStatus(plugin.StatusNotLoaded)
	return nil
}

// IsLoaded returns whether the module is loaded
func (m *AIModule) IsLoaded() bool {
	return m.loaded
}

// Nodes returns the node definitions provided by this module
func (m *AIModule) Nodes() []plugin.NodeDefinition {
	return []plugin.NodeDefinition{
		{
			Type:        "openai",
			Name:        "OpenAI",
			Category:    "ai",
			Description: "OpenAI GPT models (GPT-4, GPT-3.5, etc.)",
			Icon:        "brain",
			Color:       "#10A37F",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "anthropic",
			Name:        "Anthropic",
			Category:    "ai",
			Description: "Anthropic Claude models (Claude 3 Opus, Sonnet, Haiku)",
			Icon:        "brain",
			Color:       "#D4A574",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "ollama",
			Name:        "Ollama",
			Category:    "ai",
			Description: "Local LLMs via Ollama (Llama, Mistral, CodeLlama, etc.)",
			Icon:        "box",
			Color:       "#000000",
			Inputs:      1,
			Outputs:     1,
		},
	}
}

// RequiredMemory returns the memory requirement in bytes
func (m *AIModule) RequiredMemory() uint64 {
	return 100 * 1024 * 1024 // 100 MB (for API responses and streaming)
}

// RequiredDisk returns the disk requirement in bytes
func (m *AIModule) RequiredDisk() uint64 {
	return 10 * 1024 * 1024 // 10 MB
}

// Dependencies returns the list of required plugins
func (m *AIModule) Dependencies() []string {
	return []string{"core", "network"}
}

// init registers the AI module with the global registry
func init() {
	registry := plugin.GetRegistry()
	module := NewAIModule()
	registry.Register(module)
}
