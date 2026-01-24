package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// SlackConfig نود Slack
type SlackConfig struct {
	WebhookURL string `json:"webhookUrl"` // Slack Webhook URL
	BotToken   string `json:"botToken"`   // Bot Token (for API mode)
	Channel    string `json:"channel"`    // Default channel
	Username   string `json:"username"`   // Bot username
	IconEmoji  string `json:"iconEmoji"`  // Bot icon emoji
	Mode       string `json:"mode"`       // webhook or api
}

// SlackExecutor اجراکننده نود Slack
type SlackExecutor struct {
	config SlackConfig
	client *http.Client
}

// Init initializes the executor with configuration
func (e *SlackExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var slackConfig SlackConfig
	if err := json.Unmarshal(configJSON, &slackConfig); err != nil {
		return fmt.Errorf("invalid slack config: %w", err)
	}

	// Default values
	if slackConfig.Mode == "" {
		slackConfig.Mode = "webhook"
	}
	if slackConfig.Username == "" {
		slackConfig.Username = "EdgeFlow Bot"
	}
	if slackConfig.IconEmoji == "" {
		slackConfig.IconEmoji = ":robot_face:"
	}

	// Validate based on mode
	if slackConfig.Mode == "webhook" && slackConfig.WebhookURL == "" {
		return fmt.Errorf("webhookUrl is required for webhook mode")
	}
	if slackConfig.Mode == "api" && slackConfig.BotToken == "" {
		return fmt.Errorf("botToken is required for api mode")
	}

	e.config = slackConfig
	e.client = &http.Client{Timeout: 30 * time.Second}
	return nil
}

// Execute اجرای نود
func (e *SlackExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	var text, channel, username, iconEmoji string
	var attachments []map[string]interface{}
	var blocks []map[string]interface{}

	// Get parameters from message
	if t, ok := msg.Payload["text"].(string); ok {
		text = t
	} else if t, ok := msg.Payload["message"].(string); ok {
		text = t
	}
	if c, ok := msg.Payload["channel"].(string); ok {
		channel = c
	}
	if u, ok := msg.Payload["username"].(string); ok {
		username = u
	}
	if i, ok := msg.Payload["iconEmoji"].(string); ok {
		iconEmoji = i
	}
	if a, ok := msg.Payload["attachments"].([]map[string]interface{}); ok {
		attachments = a
	}
	if b, ok := msg.Payload["blocks"].([]map[string]interface{}); ok {
		blocks = b
	}

	if text == "" && len(attachments) == 0 && len(blocks) == 0 {
		return node.Message{}, fmt.Errorf("text, attachments, or blocks required")
	}

	// Use config defaults if not provided
	if channel == "" {
		channel = e.config.Channel
	}
	if username == "" {
		username = e.config.Username
	}
	if iconEmoji == "" {
		iconEmoji = e.config.IconEmoji
	}

	if e.config.Mode == "webhook" {
		return e.sendViaWebhook(ctx, text, channel, username, iconEmoji, attachments, blocks)
	}
	return e.sendViaAPI(ctx, text, channel, username, iconEmoji, attachments, blocks)
}

// sendViaWebhook ارسال از طریق Webhook
func (e *SlackExecutor) sendViaWebhook(ctx context.Context, text, channel, username, iconEmoji string, attachments, blocks []map[string]interface{}) (node.Message, error) {
	payload := map[string]interface{}{
		"text":       text,
		"username":   username,
		"icon_emoji": iconEmoji,
	}

	if channel != "" {
		payload["channel"] = channel
	}
	if len(attachments) > 0 {
		payload["attachments"] = attachments
	}
	if len(blocks) > 0 {
		payload["blocks"] = blocks
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

	if resp.StatusCode != 200 {
		return node.Message{}, fmt.Errorf("Slack webhook error: %s", string(body))
	}

	return node.Message{
		Payload: map[string]interface{}{
			"sent":    true,
			"text":    text,
			"channel": channel,
			"mode":    "webhook",
		},
	}, nil
}

// sendViaAPI ارسال از طریق API
func (e *SlackExecutor) sendViaAPI(ctx context.Context, text, channel, username, iconEmoji string, attachments, blocks []map[string]interface{}) (node.Message, error) {
	if channel == "" {
		return node.Message{}, fmt.Errorf("channel is required for API mode")
	}

	payload := map[string]interface{}{
		"channel":    channel,
		"text":       text,
		"username":   username,
		"icon_emoji": iconEmoji,
	}

	if len(attachments) > 0 {
		payload["attachments"] = attachments
	}
	if len(blocks) > 0 {
		payload["blocks"] = blocks
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return node.Message{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(jsonData))
	if err != nil {
		return node.Message{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.config.BotToken)

	resp, err := e.client.Do(req)
	if err != nil {
		return node.Message{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		TS    string `json:"ts"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return node.Message{}, err
	}

	if !result.OK {
		return node.Message{}, fmt.Errorf("Slack API error: %s", result.Error)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"sent":      true,
			"text":      text,
			"channel":   channel,
			"mode":      "api",
			"timestamp": result.TS,
		},
	}, nil
}

// Cleanup پاکسازی منابع
func (e *SlackExecutor) Cleanup() error {
	return nil
}
