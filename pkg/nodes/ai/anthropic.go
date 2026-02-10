package ai

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

// AnthropicConfig configuration for the Anthropic node
type AnthropicConfig struct {
	APIKey      string  `json:"apiKey"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxTokens"`
}

// AnthropicExecutor executor for the Anthropic node
type AnthropicExecutor struct {
	config AnthropicConfig
	client *http.Client
}

// NewAnthropicExecutor creates a new AnthropicExecutor
func NewAnthropicExecutor() *AnthropicExecutor {
	return &AnthropicExecutor{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Init initializes the executor with configuration
func (e *AnthropicExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var anthropicConfig AnthropicConfig
	if err := json.Unmarshal(configJSON, &anthropicConfig); err != nil {
		return fmt.Errorf("invalid anthropic config: %w", err)
	}

	if anthropicConfig.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Default values
	if anthropicConfig.Model == "" {
		anthropicConfig.Model = "claude-3-5-sonnet-20241022"
	}
	if anthropicConfig.Temperature == 0 {
		anthropicConfig.Temperature = 1.0
	}
	if anthropicConfig.MaxTokens == 0 {
		anthropicConfig.MaxTokens = 1024
	}

	e.config = anthropicConfig
	return nil
}

// Execute executes the node
func (e *AnthropicExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	var prompt string
	var systemPrompt string
	var model string
	var temperature float64
	var maxTokens int

	// Get parameters from message
	if p, ok := msg.Payload["prompt"].(string); ok {
		prompt = p
	} else if p, ok := msg.Payload["message"].(string); ok {
		prompt = p
	} else if p, ok := msg.Payload["text"].(string); ok {
		prompt = p
	}

	if s, ok := msg.Payload["system"].(string); ok {
		systemPrompt = s
	}
	if m, ok := msg.Payload["model"].(string); ok {
		model = m
	}
	if t, ok := msg.Payload["temperature"].(float64); ok {
		temperature = t
	}
	if mt, ok := msg.Payload["maxTokens"].(float64); ok {
		maxTokens = int(mt)
	}

	if prompt == "" {
		return node.Message{}, fmt.Errorf("prompt is required")
	}

	// Use config defaults if not provided
	if model == "" {
		model = e.config.Model
	}
	if temperature == 0 {
		temperature = e.config.Temperature
	}
	if maxTokens == 0 {
		maxTokens = e.config.MaxTokens
	}

	// Prepare messages
	messages := []map[string]string{
		{
			"role":    "user",
			"content": prompt,
		},
	}

	// Prepare request
	payload := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
	}

	// Add system prompt if provided
	if systemPrompt != "" {
		payload["system"] = systemPrompt
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return node.Message{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return node.Message{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", e.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	resp, err := e.client.Do(req)
	if err != nil {
		return node.Message{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return node.Message{}, fmt.Errorf("Anthropic API error: %s", string(body))
	}

	// Parse response
	var result struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Model        string `json:"model"`
		StopReason   string `json:"stop_reason"`
		StopSequence string `json:"stop_sequence"`
		Usage        struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return node.Message{}, err
	}

	if len(result.Content) == 0 {
		return node.Message{}, fmt.Errorf("no response from Anthropic")
	}

	response := result.Content[0].Text

	return node.Message{
		Payload: map[string]interface{}{
			"response":   response,
			"prompt":     prompt,
			"model":      result.Model,
			"usage":      result.Usage,
			"stopReason": result.StopReason,
			"id":         result.ID,
		},
	}, nil
}

// Cleanup cleans up resources
func (e *AnthropicExecutor) Cleanup() error {
	return nil
}
