package messaging

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterAllNodes registers all messaging nodes
func RegisterAllNodes(registry *node.Registry) error {
	// Telegram
	if err := registry.Register(&node.NodeInfo{
		Type:        "telegram",
		Name:        "Telegram",
		Category:    node.NodeTypeOutput,
		Description: "Send and receive messages via Telegram Bot",
		Icon:        "send",
		Color:       "#0088cc",
		Properties: []node.PropertySchema{
			{
				Name:        "botToken",
				Label:       "Bot Token",
				Type:        "password",
				Default:     "",
				Required:    true,
				Description: "Telegram bot token",
			},
			{
				Name:        "chatId",
				Label:       "Chat ID",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Default chat ID",
			},
			{
				Name:        "mode",
				Label:       "Mode",
				Type:        "select",
				Default:     "send",
				Options:     []string{"send", "receive"},
				Description: "Send or receive mode",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return &TelegramExecutor{}
		},
	}); err != nil {
		return err
	}

	// Email
	if err := registry.Register(&node.NodeInfo{
		Type:        "email",
		Name:        "Email",
		Category:    node.NodeTypeOutput,
		Description: "Send email via SMTP",
		Icon:        "mail",
		Color:       "#ea4335",
		Properties: []node.PropertySchema{
			{
				Name:        "host",
				Label:       "SMTP Host",
				Type:        "string",
				Default:     "",
				Required:    true,
				Description: "SMTP server host",
			},
			{
				Name:        "port",
				Label:       "SMTP Port",
				Type:        "number",
				Default:     587,
				Required:    true,
				Description: "SMTP server port",
			},
			{
				Name:        "username",
				Label:       "Username",
				Type:        "string",
				Default:     "",
				Required:    true,
				Description: "SMTP username",
			},
			{
				Name:        "password",
				Label:       "Password",
				Type:        "password",
				Default:     "",
				Required:    true,
				Description: "SMTP password",
			},
			{
				Name:        "from",
				Label:       "From Address",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "From email address",
			},
			{
				Name:        "to",
				Label:       "To Address",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Default to email address",
			},
			{
				Name:        "useTls",
				Label:       "Use TLS",
				Type:        "boolean",
				Default:     true,
				Required:    false,
				Description: "Use TLS connection",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return &EmailExecutor{}
		},
	}); err != nil {
		return err
	}

	// Slack
	if err := registry.Register(&node.NodeInfo{
		Type:        "slack",
		Name:        "Slack",
		Category:    node.NodeTypeOutput,
		Description: "Send messages to Slack via Webhook or API",
		Icon:        "message-square",
		Color:       "#4a154b",
		Properties: []node.PropertySchema{
			{
				Name:        "webhookUrl",
				Label:       "Webhook URL",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Slack webhook URL",
			},
			{
				Name:        "botToken",
				Label:       "Bot Token",
				Type:        "password",
				Default:     "",
				Required:    false,
				Description: "Slack bot token",
			},
			{
				Name:        "channel",
				Label:       "Channel",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Default channel",
			},
			{
				Name:        "username",
				Label:       "Username",
				Type:        "string",
				Default:     "EdgeFlow Bot",
				Required:    false,
				Description: "Bot username",
			},
			{
				Name:        "iconEmoji",
				Label:       "Icon Emoji",
				Type:        "string",
				Default:     ":robot_face:",
				Required:    false,
				Description: "Bot icon emoji",
			},
			{
				Name:        "mode",
				Label:       "Mode",
				Type:        "select",
				Default:     "webhook",
				Options:     []string{"webhook", "api"},
				Description: "Webhook or API mode",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return &SlackExecutor{}
		},
	}); err != nil {
		return err
	}

	// Discord
	if err := registry.Register(&node.NodeInfo{
		Type:        "discord",
		Name:        "Discord",
		Category:    node.NodeTypeOutput,
		Description: "Send messages to Discord via Webhook",
		Icon:        "message-circle",
		Color:       "#5865f2",
		Properties: []node.PropertySchema{
			{
				Name:        "webhookUrl",
				Label:       "Webhook URL",
				Type:        "string",
				Default:     "",
				Required:    true,
				Description: "Discord webhook URL",
			},
			{
				Name:        "username",
				Label:       "Username",
				Type:        "string",
				Default:     "EdgeFlow Bot",
				Required:    false,
				Description: "Bot username",
			},
			{
				Name:        "avatarUrl",
				Label:       "Avatar URL",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Bot avatar URL",
			},
			{
				Name:        "tts",
				Label:       "Text-to-Speech",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Enable text-to-speech",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return &DiscordExecutor{}
		},
	}); err != nil {
		return err
	}

	return nil
}
