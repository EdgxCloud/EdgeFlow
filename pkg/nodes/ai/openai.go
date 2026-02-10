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

// OpenAIConfig configuration for the OpenAI node
type OpenAIConfig struct {
	APIKey      string  `json:"apiKey"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxTokens"`
}

// OpenAIExecutor executor for the OpenAI node
type OpenAIExecutor struct {
	config OpenAIConfig
	client *http.Client
}

// NewOpenAIExecutor creates a new OpenAIExecutor
func NewOpenAIExecutor() *OpenAIExecutor {
	return &OpenAIExecutor{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Init initializes the executor with configuration
func (e *OpenAIExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var openaiConfig OpenAIConfig
	if err := json.Unmarshal(configJSON, &openaiConfig); err != nil {
		return fmt.Errorf("invalid openai config: %w", err)
	}

	if openaiConfig.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Default values
	if openaiConfig.Model == "" {
		openaiConfig.Model = "gpt-3.5-turbo"
	}
	if openaiConfig.Temperature == 0 {
		openaiConfig.Temperature = 0.7
	}
	if openaiConfig.MaxTokens == 0 {
		openaiConfig.MaxTokens = 1000
	}

	e.config = openaiConfig
	return nil
}

// Execute executes the node
func (e *OpenAIExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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
	messages := []map[string]string{}
	if systemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": prompt,
	})

	// Prepare request
	payload := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return node.Message{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return node.Message{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.config.APIKey)

	// Execute request
	resp, err := e.client.Do(req)
	if err != nil {
		return node.Message{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return node.Message{}, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	// Parse response
	var result struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return node.Message{}, err
	}

	if len(result.Choices) == 0 {
		return node.Message{}, fmt.Errorf("no response from OpenAI")
	}

	response := result.Choices[0].Message.Content

	return node.Message{
		Payload: map[string]interface{}{
			"response":    response,
			"prompt":      prompt,
			"model":       model,
			"usage":       result.Usage,
			"finishReason": result.Choices[0].FinishReason,
		},
	}, nil
}

// Cleanup cleans up resources
func (e *OpenAIExecutor) Cleanup() error {
	return nil
}
