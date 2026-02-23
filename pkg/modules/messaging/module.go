// Package messaging provides the messaging module for EdgeFlow
// This module provides Telegram, Email, Slack, and Discord nodes
package messaging

import (
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/EdgxCloud/EdgeFlow/internal/plugin"
	"github.com/EdgxCloud/EdgeFlow/pkg/nodes/messaging"
)

// MessagingModule is the messaging module that provides notification and messaging nodes
type MessagingModule struct {
	*plugin.BasePlugin
	loaded bool
}

// NewMessagingModule creates a new messaging module
func NewMessagingModule() *MessagingModule {
	metadata := plugin.Metadata{
		Name:        "messaging",
		Version:     "1.0.0",
		Description: "Messaging and notification nodes - Telegram, Email, Slack, Discord",
		Author:      "EdgeFlow Team",
		Category:    plugin.CategoryMessaging,
		License:     "MIT",
		Keywords:    []string{"messaging", "telegram", "email", "slack", "discord", "notification"},
		MinEdgeFlow: "0.1.0",
		Config: map[string]interface{}{
			"telegram_timeout": 30,
			"email_timeout":    60,
			"slack_timeout":    30,
			"discord_timeout":  30,
		},
	}

	return &MessagingModule{
		BasePlugin: plugin.NewBasePlugin(metadata),
		loaded:     false,
	}
}

// Load loads the messaging module
func (m *MessagingModule) Load() error {
	m.SetStatus(plugin.StatusLoading)

	// Register all messaging nodes using the centralized RegisterAllNodes function
	registry := node.GetGlobalRegistry()
	if err := messaging.RegisterAllNodes(registry); err != nil {
		m.SetStatus(plugin.StatusError)
		return err
	}

	m.loaded = true
	m.SetStatus(plugin.StatusLoaded)
	return nil
}

// Unload unloads the messaging module
func (m *MessagingModule) Unload() error {
	m.SetStatus(plugin.StatusUnloading)
	m.loaded = false
	m.SetStatus(plugin.StatusNotLoaded)
	return nil
}

// IsLoaded returns whether the module is loaded
func (m *MessagingModule) IsLoaded() bool {
	return m.loaded
}

// Nodes returns the node definitions provided by this module
func (m *MessagingModule) Nodes() []plugin.NodeDefinition {
	return []plugin.NodeDefinition{
		{
			Type:        "telegram",
			Name:        "Telegram",
			Category:    "messaging",
			Description: "Send messages to Telegram chats/groups",
			Icon:        "send",
			Color:       "#0088CC",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "email",
			Name:        "Email",
			Category:    "messaging",
			Description: "Send emails via SMTP",
			Icon:        "mail",
			Color:       "#EA4335",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "slack",
			Name:        "Slack",
			Category:    "messaging",
			Description: "Send messages to Slack channels",
			Icon:        "hash",
			Color:       "#4A154B",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "discord",
			Name:        "Discord",
			Category:    "messaging",
			Description: "Send messages to Discord channels via webhook",
			Icon:        "message-circle",
			Color:       "#5865F2",
			Inputs:      1,
			Outputs:     1,
		},
	}
}

// RequiredMemory returns the memory requirement in bytes
func (m *MessagingModule) RequiredMemory() uint64 {
	return 20 * 1024 * 1024 // 20 MB
}

// RequiredDisk returns the disk requirement in bytes
func (m *MessagingModule) RequiredDisk() uint64 {
	return 5 * 1024 * 1024 // 5 MB
}

// Dependencies returns the list of required plugins
func (m *MessagingModule) Dependencies() []string {
	return []string{"core", "network"}
}

// init registers the messaging module with the global registry
func init() {
	registry := plugin.GetRegistry()
	module := NewMessagingModule()
	registry.Register(module)
}
