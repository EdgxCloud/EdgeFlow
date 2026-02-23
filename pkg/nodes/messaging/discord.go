package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// DiscordConfig configuration for the Discord node
type DiscordConfig struct {
	WebhookURL string `json:"webhookUrl"` // Discord Webhook URL
	Username   string `json:"username"`   // Bot username
	AvatarURL  string `json:"avatarUrl"`  // Bot avatar URL
	TTS        bool   `json:"tts"`        // Text-to-speech
}

// DiscordExecutor executor for the Discord node
type DiscordExecutor struct {
	config DiscordConfig
	client *http.Client
}

// Init initializes the executor with configuration
func (e *DiscordExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var discordConfig DiscordConfig
	if err := json.Unmarshal(configJSON, &discordConfig); err != nil {
		return fmt.Errorf("invalid discord config: %w", err)
	}

	// Validate
	if discordConfig.WebhookURL == "" {
		return fmt.Errorf("webhookUrl is required")
	}

	// Default values
	if discordConfig.Username == "" {
		discordConfig.Username = "EdgeFlow Bot"
	}

	e.config = discordConfig
	e.client = &http.Client{Timeout: 30 * time.Second}
	return nil
}

// Execute executes the node
func (e *DiscordExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	var content, username, avatarURL string
	var tts bool
	var embeds []map[string]interface{}

	// Get parameters from message
	if c, ok := msg.Payload["content"].(string); ok {
		content = c
	} else if c, ok := msg.Payload["message"].(string); ok {
		content = c
	} else if c, ok := msg.Payload["text"].(string); ok {
		content = c
	}
	if u, ok := msg.Payload["username"].(string); ok {
		username = u
	}
	if a, ok := msg.Payload["avatarUrl"].(string); ok {
		avatarURL = a
	}
	if t, ok := msg.Payload["tts"].(bool); ok {
		tts = t
	}
	if emb, ok := msg.Payload["embeds"].([]map[string]interface{}); ok {
		embeds = emb
	}

	if content == "" && len(embeds) == 0 {
		return node.Message{}, fmt.Errorf("content or embeds required")
	}

	// Use config defaults if not provided
	if username == "" {
		username = e.config.Username
	}
	if avatarURL == "" {
		avatarURL = e.config.AvatarURL
	}
	if !tts {
		tts = e.config.TTS
	}

	// Prepare payload
	payload := map[string]interface{}{
		"content":    content,
		"username":   username,
		"avatar_url": avatarURL,
		"tts":        tts,
	}

	if len(embeds) > 0 {
		payload["embeds"] = embeds
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return node.Message{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return node.Message{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return node.Message{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return node.Message{}, fmt.Errorf("Discord webhook error: %s", string(body))
	}

	return node.Message{
		Payload: map[string]interface{}{
			"sent":     true,
			"content":  content,
			"username": username,
		},
	}, nil
}

// Cleanup cleans up resources
func (e *DiscordExecutor) Cleanup() error {
	return nil
}
