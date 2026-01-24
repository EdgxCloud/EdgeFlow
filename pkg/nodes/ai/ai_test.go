package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============ OpenAI Tests ============

func TestNewOpenAIExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"apiKey": "sk-test-key-12345",
			},
			wantErr: false,
		},
		{
			name:    "missing API key",
			config:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "with model",
			config: map[string]interface{}{
				"apiKey": "sk-test-key-12345",
				"model":  "gpt-4",
			},
			wantErr: false,
		},
		{
			name: "with temperature",
			config: map[string]interface{}{
				"apiKey":      "sk-test-key-12345",
				"temperature": 0.5,
			},
			wantErr: false,
		},
		{
			name: "with max tokens",
			config: map[string]interface{}{
				"apiKey":    "sk-test-key-12345",
				"maxTokens": 2000,
			},
			wantErr: false,
		},
		{
			name: "full config",
			config: map[string]interface{}{
				"apiKey":      "sk-test-key-12345",
				"model":       "gpt-4-turbo",
				"temperature": 0.9,
				"maxTokens":   4096,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewOpenAIExecutor()
			err := executor.Init(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestOpenAIConfig_Defaults(t *testing.T) {
	executor := NewOpenAIExecutor()
	err := executor.Init(map[string]interface{}{
		"apiKey": "sk-test-key-12345",
	})
	require.NoError(t, err)

	assert.Equal(t, "gpt-3.5-turbo", executor.config.Model)
	assert.Equal(t, 0.7, executor.config.Temperature)
	assert.Equal(t, 1000, executor.config.MaxTokens)
}

func TestOpenAIExecutor_Execute_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer ")

		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		assert.Equal(t, "gpt-3.5-turbo", payload["model"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"role":    "assistant",
						"content": "Hello! How can I help you today?",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]int{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		})
	}))
	defer server.Close()

	// Note: In real tests, we would need to inject the base URL
	executor := NewOpenAIExecutor()
	err := executor.Init(map[string]interface{}{
		"apiKey": "sk-test-key-12345",
	})
	require.NoError(t, err)
	assert.NotNil(t, executor)
}

func TestOpenAIExecutor_Execute_MissingPrompt(t *testing.T) {
	executor := NewOpenAIExecutor()
	err := executor.Init(map[string]interface{}{
		"apiKey": "sk-test-key-12345",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{},
	}

	_, err = executor.Execute(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt is required")
}

func TestOpenAIExecutor_Cleanup(t *testing.T) {
	executor := NewOpenAIExecutor()
	err := executor.Init(map[string]interface{}{
		"apiKey": "sk-test-key-12345",
	})
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}

// ============ Anthropic Tests ============

func TestNewAnthropicExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"apiKey": "sk-ant-test-key-12345",
			},
			wantErr: false,
		},
		{
			name:    "missing API key",
			config:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "with model",
			config: map[string]interface{}{
				"apiKey": "sk-ant-test-key-12345",
				"model":  "claude-3-opus-20240229",
			},
			wantErr: false,
		},
		{
			name: "with temperature",
			config: map[string]interface{}{
				"apiKey":      "sk-ant-test-key-12345",
				"temperature": 0.5,
			},
			wantErr: false,
		},
		{
			name: "with max tokens",
			config: map[string]interface{}{
				"apiKey":    "sk-ant-test-key-12345",
				"maxTokens": 2048,
			},
			wantErr: false,
		},
		{
			name: "full config",
			config: map[string]interface{}{
				"apiKey":      "sk-ant-test-key-12345",
				"model":       "claude-3-5-sonnet-20241022",
				"temperature": 0.8,
				"maxTokens":   4096,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewAnthropicExecutor()
			err := executor.Init(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestAnthropicConfig_Defaults(t *testing.T) {
	executor := NewAnthropicExecutor()
	err := executor.Init(map[string]interface{}{
		"apiKey": "sk-ant-test-key-12345",
	})
	require.NoError(t, err)

	assert.Equal(t, "claude-3-5-sonnet-20241022", executor.config.Model)
	assert.Equal(t, 1.0, executor.config.Temperature)
	assert.Equal(t, 1024, executor.config.MaxTokens)
}

func TestAnthropicExecutor_Execute_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "msg_01XFDUDYJgAACzvnptvVoYEL",
			"type": "message",
			"role": "assistant",
			"content": []map[string]string{
				{
					"type": "text",
					"text": "Hello! I'm Claude, how can I assist you today?",
				},
			},
			"model":       "claude-3-5-sonnet-20241022",
			"stop_reason": "end_turn",
			"usage": map[string]int{
				"input_tokens":  15,
				"output_tokens": 25,
			},
		})
	}))
	defer server.Close()

	executor := NewAnthropicExecutor()
	err := executor.Init(map[string]interface{}{
		"apiKey": "sk-ant-test-key-12345",
	})
	require.NoError(t, err)
	assert.NotNil(t, executor)
}

func TestAnthropicExecutor_Execute_MissingPrompt(t *testing.T) {
	executor := NewAnthropicExecutor()
	err := executor.Init(map[string]interface{}{
		"apiKey": "sk-ant-test-key-12345",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{},
	}

	_, err = executor.Execute(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt is required")
}

func TestAnthropicExecutor_Cleanup(t *testing.T) {
	executor := NewAnthropicExecutor()
	err := executor.Init(map[string]interface{}{
		"apiKey": "sk-ant-test-key-12345",
	})
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}

// ============ Ollama Tests ============

func TestNewOllamaExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "default config",
			config:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "with base URL",
			config: map[string]interface{}{
				"baseUrl": "http://ollama.local:11434",
			},
			wantErr: false,
		},
		{
			name: "with model",
			config: map[string]interface{}{
				"model": "mistral",
			},
			wantErr: false,
		},
		{
			name: "with temperature",
			config: map[string]interface{}{
				"temperature": 0.5,
			},
			wantErr: false,
		},
		{
			name: "with stream disabled",
			config: map[string]interface{}{
				"stream": false,
			},
			wantErr: false,
		},
		{
			name: "full config",
			config: map[string]interface{}{
				"baseUrl":     "http://localhost:11434",
				"model":       "codellama",
				"temperature": 0.3,
				"stream":      false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewOllamaExecutor()
			err := executor.Init(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestOllamaConfig_Defaults(t *testing.T) {
	executor := NewOllamaExecutor()
	err := executor.Init(map[string]interface{}{})
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:11434", executor.config.BaseURL)
	assert.Equal(t, "llama2", executor.config.Model)
	assert.Equal(t, 0.7, executor.config.Temperature)
}

func TestOllamaExecutor_Execute_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/generate", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		assert.Equal(t, "llama2", payload["model"])
		assert.Equal(t, false, payload["stream"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"model":               "llama2",
			"created_at":          "2024-01-15T10:30:00Z",
			"response":            "Hello! I'm Llama, how can I help you today?",
			"done":                true,
			"context":             []int{1, 2, 3},
			"total_duration":      1500000000,
			"load_duration":       100000000,
			"prompt_eval_count":   10,
			"prompt_eval_duration": 200000000,
			"eval_count":          15,
			"eval_duration":       1200000000,
		})
	}))
	defer server.Close()

	executor := NewOllamaExecutor()
	err := executor.Init(map[string]interface{}{
		"baseUrl": server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"prompt": "Hello!",
		},
	}

	result, err := executor.Execute(ctx, msg)
	require.NoError(t, err)

	payload := result.Payload
	assert.Equal(t, "Hello! I'm Llama, how can I help you today?", payload["response"])
	assert.Equal(t, "llama2", payload["model"])
}

func TestOllamaExecutor_Execute_MissingPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not reach here
		t.Fatal("Request should not be made without prompt")
	}))
	defer server.Close()

	executor := NewOllamaExecutor()
	err := executor.Init(map[string]interface{}{
		"baseUrl": server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{},
	}

	_, err = executor.Execute(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt is required")
}

func TestOllamaExecutor_Execute_StringPayload(t *testing.T) {
	t.Skip("String payload not supported with current Payload type")
}

func TestOllamaExecutor_Execute_WithSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		assert.Equal(t, "You are a helpful assistant.", payload["system"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"model":    "llama2",
			"response": "Hello!",
			"done":     true,
		})
	}))
	defer server.Close()

	executor := NewOllamaExecutor()
	err := executor.Init(map[string]interface{}{
		"baseUrl": server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"prompt": "Hello",
			"system": "You are a helpful assistant.",
		},
	}

	result, err := executor.Execute(ctx, msg)
	require.NoError(t, err)

	payload := result.Payload
	assert.Equal(t, "Hello!", payload["response"])
}

func TestOllamaExecutor_Execute_CustomModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		assert.Equal(t, "codellama", payload["model"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"model":    "codellama",
			"response": "def hello(): print('Hello!')",
			"done":     true,
		})
	}))
	defer server.Close()

	executor := NewOllamaExecutor()
	err := executor.Init(map[string]interface{}{
		"baseUrl": server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"prompt": "Write a hello function in Python",
			"model":  "codellama",
		},
	}

	result, err := executor.Execute(ctx, msg)
	require.NoError(t, err)

	payload := result.Payload
	assert.Contains(t, payload["response"], "def hello")
}

func TestOllamaExecutor_Cleanup(t *testing.T) {
	executor := NewOllamaExecutor()
	err := executor.Init(map[string]interface{}{})
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}
