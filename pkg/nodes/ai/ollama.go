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

// OllamaConfig configuration for the Ollama node
type OllamaConfig struct {
	BaseURL     string  `json:"baseUrl"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	Stream      bool    `json:"stream"`
}

// OllamaExecutor executor for the Ollama node
type OllamaExecutor struct {
	config OllamaConfig
	client *http.Client
}

// NewOllamaExecutor creates a new OllamaExecutor
func NewOllamaExecutor() *OllamaExecutor {
	return &OllamaExecutor{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

// Init initializes the executor with configuration
func (e *OllamaExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var ollamaConfig OllamaConfig
	if err := json.Unmarshal(configJSON, &ollamaConfig); err != nil {
		return fmt.Errorf("invalid ollama config: %w", err)
	}

	// Default values
	if ollamaConfig.BaseURL == "" {
		ollamaConfig.BaseURL = "http://localhost:11434"
	}
	if ollamaConfig.Model == "" {
		ollamaConfig.Model = "gemma3:1b"
	}
	if ollamaConfig.Temperature == 0 {
		ollamaConfig.Temperature = 0.7
	}

	e.config = ollamaConfig
	return nil
}

// Execute executes the node
func (e *OllamaExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	var prompt string
	var systemPrompt string
	var model string
	var temperature float64

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

	// Prepare request
	payload := map[string]interface{}{
		"model":       model,
		"prompt":      prompt,
		"temperature": temperature,
		"stream":      false, // Always false for node execution
	}

	// Add system prompt if provided
	if systemPrompt != "" {
		payload["system"] = systemPrompt
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return node.Message{}, err
	}

	url := fmt.Sprintf("%s/api/generate", e.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return node.Message{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := e.client.Do(req)
	if err != nil {
		return node.Message{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return node.Message{}, fmt.Errorf("Ollama API error: %s", string(body))
	}

	// Parse response
	var result struct {
		Model              string `json:"model"`
		CreatedAt          string `json:"created_at"`
		Response           string `json:"response"`
		Done               bool   `json:"done"`
		Context            []int  `json:"context"`
		TotalDuration      int64  `json:"total_duration"`
		LoadDuration       int64  `json:"load_duration"`
		PromptEvalCount    int    `json:"prompt_eval_count"`
		PromptEvalDuration int64  `json:"prompt_eval_duration"`
		EvalCount          int    `json:"eval_count"`
		EvalDuration       int64  `json:"eval_duration"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return node.Message{}, err
	}

	if !result.Done {
		return node.Message{}, fmt.Errorf("incomplete response from Ollama")
	}

	return node.Message{
		Payload: map[string]interface{}{
			"response": result.Response,
			"prompt":   prompt,
			"model":    result.Model,
			"usage": map[string]interface{}{
				"promptTokens":     result.PromptEvalCount,
				"completionTokens": result.EvalCount,
				"totalTokens":      result.PromptEvalCount + result.EvalCount,
			},
			"duration": map[string]interface{}{
				"total":      result.TotalDuration,
				"load":       result.LoadDuration,
				"prompt":     result.PromptEvalDuration,
				"evaluation": result.EvalDuration,
			},
		},
	}, nil
}

// Cleanup cleans up resources
func (e *OllamaExecutor) Cleanup() error {
	return nil
}
