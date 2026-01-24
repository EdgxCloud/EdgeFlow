package network

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

func TestNewHTTPRequestExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config with URL",
			config: map[string]interface{}{
				"url":    "https://api.example.com",
				"method": "GET",
			},
			wantErr: false,
		},
		{
			name:    "empty config uses defaults",
			config:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "with timeout",
			config: map[string]interface{}{
				"url":     "https://api.example.com",
				"timeout": 5000,
			},
			wantErr: false,
		},
		{
			name: "with headers",
			config: map[string]interface{}{
				"url": "https://api.example.com",
				"headers": map[string]interface{}{
					"Authorization": "Bearer token123",
					"Content-Type":  "application/json",
				},
			},
			wantErr: false,
		},
		{
			name: "with retry config",
			config: map[string]interface{}{
				"url":        "https://api.example.com",
				"retries":    3,
				"retryDelay": 1000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewHTTPRequestExecutor(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestHTTPRequestExecutor_Execute_GET(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/data", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []int{1, 2, 3},
		})
	}))
	defer server.Close()

	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"url":    server.URL + "/api/data",
		"method": "GET",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{}

	result, err := executor.Execute(ctx, msg)
	require.NoError(t, err)

	// Check response
	payload := result.Payload.(map[string]interface{})
	assert.Equal(t, 200, int(payload["statusCode"].(float64)))
}

func TestHTTPRequestExecutor_Execute_POST(t *testing.T) {
	// Create mock server
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		json.NewDecoder(r.Body).Decode(&receivedBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      123,
			"created": true,
		})
	}))
	defer server.Close()

	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"url":    server.URL + "/api/create",
		"method": "POST",
		"headers": map[string]interface{}{
			"Content-Type": "application/json",
		},
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"name":  "test",
			"value": 42,
		},
	}

	result, err := executor.Execute(ctx, msg)
	require.NoError(t, err)

	payload := result.Payload.(map[string]interface{})
	assert.Equal(t, 201, int(payload["statusCode"].(float64)))
}

func TestHTTPRequestExecutor_Execute_PUT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"updated": true})
	}))
	defer server.Close()

	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"url":    server.URL + "/api/update",
		"method": "PUT",
	})
	require.NoError(t, err)

	result, err := executor.Execute(context.Background(), node.Message{})
	require.NoError(t, err)

	payload := result.Payload.(map[string]interface{})
	assert.Equal(t, 200, int(payload["statusCode"].(float64)))
}

func TestHTTPRequestExecutor_Execute_DELETE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"url":    server.URL + "/api/delete/1",
		"method": "DELETE",
	})
	require.NoError(t, err)

	result, err := executor.Execute(context.Background(), node.Message{})
	require.NoError(t, err)

	payload := result.Payload.(map[string]interface{})
	assert.Equal(t, 204, int(payload["statusCode"].(float64)))
}

func TestHTTPRequestExecutor_Execute_WithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer secret-token", r.Header.Get("Authorization"))
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"url":    server.URL,
		"method": "GET",
		"headers": map[string]interface{}{
			"Authorization":   "Bearer secret-token",
			"X-Custom-Header": "custom-value",
		},
	})
	require.NoError(t, err)

	_, err = executor.Execute(context.Background(), node.Message{})
	require.NoError(t, err)
}

func TestHTTPRequestExecutor_Execute_WithQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "value1", r.URL.Query().Get("param1"))
		assert.Equal(t, "value2", r.URL.Query().Get("param2"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"url":    server.URL + "?param1=value1&param2=value2",
		"method": "GET",
	})
	require.NoError(t, err)

	_, err = executor.Execute(context.Background(), node.Message{})
	require.NoError(t, err)
}

func TestHTTPRequestExecutor_Execute_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Internal server error",
		})
	}))
	defer server.Close()

	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"url":    server.URL,
		"method": "GET",
	})
	require.NoError(t, err)

	result, err := executor.Execute(context.Background(), node.Message{})
	// Should not error, but return error status code
	require.NoError(t, err)

	payload := result.Payload.(map[string]interface{})
	assert.Equal(t, 500, int(payload["statusCode"].(float64)))
}

func TestHTTPRequestExecutor_Execute_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't respond, causing timeout
		select {}
	}))
	defer server.Close()

	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"url":     server.URL,
		"method":  "GET",
		"timeout": 100, // 100ms timeout
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 200*1000000) // 200ms
	defer cancel()

	_, err = executor.Execute(ctx, node.Message{})
	// Should timeout
	assert.Error(t, err)
}

func TestHTTPRequestExecutor_Execute_URLFromMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dynamic/path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"method": "GET",
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: map[string]interface{}{
			"url": server.URL + "/dynamic/path",
		},
	}

	_, err = executor.Execute(context.Background(), msg)
	require.NoError(t, err)
}

func TestHTTPRequestExecutor_Execute_MethodFromMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"url": server.URL,
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: map[string]interface{}{
			"method": "PATCH",
		},
	}

	_, err = executor.Execute(context.Background(), msg)
	require.NoError(t, err)
}

func TestHTTPRequestExecutor_Execute_BasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "testuser", username)
		assert.Equal(t, "testpass", password)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"url":               server.URL,
		"method":            "GET",
		"basicAuthUsername": "testuser",
		"basicAuthPassword": "testpass",
	})
	require.NoError(t, err)

	_, err = executor.Execute(context.Background(), node.Message{})
	require.NoError(t, err)
}

func TestHTTPRequestExecutor_Cleanup(t *testing.T) {
	executor, err := NewHTTPRequestExecutor(map[string]interface{}{
		"url": "https://api.example.com",
	})
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}
