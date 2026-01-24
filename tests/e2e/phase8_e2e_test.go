package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/edgeflow/edgeflow/internal/api"
	"github.com/edgeflow/edgeflow/internal/engine"
	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/edgeflow/edgeflow/pkg/nodes/core"
	"github.com/edgeflow/edgeflow/pkg/nodes/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEndFlowExecution(t *testing.T) {
	registry := node.NewRegistry()
	core.RegisterAllNodes(registry)
	network.RegisterAllNodes(registry)

	t.Run("Complete Flow: Inject -> Template -> HTTP Request -> Debug", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data":   "test response",
			})
		}))
		defer mockServer.Close()

		injectNode, _ := registry.Create("inject")
		templateNode, _ := registry.Create("template")
		httpNode, _ := registry.Create("http-request")
		debugNode, _ := registry.Create("debug")

		injectNode.Init(map[string]interface{}{
			"payload": map[string]interface{}{
				"user": "Alice",
			},
		})

		templateNode.Init(map[string]interface{}{
			"template": "User: {{user}}",
			"syntax":   "mustache",
		})

		httpNode.Init(map[string]interface{}{
			"method": "GET",
			"url":    mockServer.URL,
		})

		debugNode.Init(map[string]interface{}{
			"console": true,
		})

		ctx := context.Background()

		msg, err := injectNode.Execute(ctx, node.Message{})
		require.NoError(t, err)

		msg, err = templateNode.Execute(ctx, msg)
		require.NoError(t, err)

		msg, err = httpNode.Execute(ctx, msg)
		require.NoError(t, err)

		_, err = debugNode.Execute(ctx, msg)
		require.NoError(t, err)
	})

	t.Run("Split and Join Flow", func(t *testing.T) {
		splitNode, _ := registry.Create("split")
		joinNode, _ := registry.Create("join")

		splitNode.Init(map[string]interface{}{})
		joinNode.Init(map[string]interface{}{
			"mode": "auto",
		})

		ctx := context.Background()

		msg := node.Message{
			Payload: []interface{}{1, 2, 3, 4, 5},
		}

		messages := []node.Message{}

		for i := 0; i < 5; i++ {
			result, err := splitNode.Execute(ctx, msg)
			require.NoError(t, err)
			messages = append(messages, result)
		}

		var finalMsg node.Message
		for _, m := range messages {
			finalMsg, _ = joinNode.Execute(ctx, m)
		}

		payload, ok := finalMsg.Payload.([]interface{})
		require.True(t, ok)
		assert.Equal(t, 5, len(payload))
	})

	t.Run("Error Handling Flow with Catch", func(t *testing.T) {
		functionNode, _ := registry.Create("function")
		catchNode, _ := registry.Create("catch")

		functionNode.Init(map[string]interface{}{
			"code": "throw new Error('Test error')",
		})

		catchNode.Init(map[string]interface{}{
			"scope": "all",
		})

		ctx := context.Background()

		msg := node.Message{
			Payload: "test",
		}

		_, err := functionNode.Execute(ctx, msg)
		assert.Error(t, err)

		errorMsg := node.Message{
			Metadata: map[string]interface{}{
				"error": map[string]interface{}{
					"message": err.Error(),
					"source":  "function",
				},
			},
		}

		catchResult, err := catchNode.Execute(ctx, errorMsg)
		require.NoError(t, err)
		assert.NotNil(t, catchResult.Payload)
	})
}

func TestDeployAPI(t *testing.T) {
	registry := node.NewRegistry()
	core.RegisterAllNodes(registry)

	server := api.NewServer(registry)
	router := server.SetupRoutes()

	t.Run("Deploy Full Flow", func(t *testing.T) {
		deployReq := map[string]interface{}{
			"mode": "full",
			"flows": []interface{}{
				map[string]interface{}{
					"id":   "test-flow-1",
					"name": "Test Flow 1",
					"nodes": []interface{}{
						map[string]interface{}{
							"id":   "inject-1",
							"type": "inject",
							"config": map[string]interface{}{
								"payload": "Hello",
							},
						},
					},
				},
			},
		}

		body, _ := json.Marshal(deployReq)
		req := httptest.NewRequest("POST", "/api/flows/deploy", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		json.NewDecoder(rec.Body).Decode(&response)
		assert.True(t, response["success"].(bool))
	})

	t.Run("Deploy Modified Flow", func(t *testing.T) {
		deployReq := map[string]interface{}{
			"mode": "modified",
			"flows": []interface{}{
				map[string]interface{}{
					"id":   "test-flow-2",
					"name": "Test Flow 2",
					"nodes": []interface{}{
						map[string]interface{}{
							"id":   "debug-1",
							"type": "debug",
						},
					},
				},
			},
		}

		body, _ := json.Marshal(deployReq)
		req := httptest.NewRequest("POST", "/api/flows/deploy", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestWebSocketRealtime(t *testing.T) {
	t.Skip("WebSocket testing requires special setup")
}

func TestLongRunningFlow(t *testing.T) {
	registry := node.NewRegistry()
	core.RegisterAllNodes(registry)

	t.Run("Flow with Delay Node", func(t *testing.T) {
		delayNode, _ := registry.Create("delay")
		delayNode.Init(map[string]interface{}{
			"timeout": 100,
		})

		ctx := context.Background()
		msg := node.Message{
			Payload: "test",
		}

		start := time.Now()
		_, err := delayNode.Execute(ctx, msg)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, duration.Milliseconds(), int64(100))
	})
}

func TestComplexDataTransformation(t *testing.T) {
	registry := node.NewRegistry()
	core.RegisterAllNodes(registry)

	t.Run("JSON Parsing and Template", func(t *testing.T) {
		jsonNode, _ := registry.Create("json-parser")
		templateNode, _ := registry.Create("template")

		jsonNode.Init(map[string]interface{}{
			"action": "parse",
		})

		templateNode.Init(map[string]interface{}{
			"template": "Name: {{name}}, Age: {{age}}",
			"syntax":   "mustache",
		})

		ctx := context.Background()

		msg := node.Message{
			Payload: `{"name": "Bob", "age": 30}`,
		}

		msg, err := jsonNode.Execute(ctx, msg)
		require.NoError(t, err)

		msg, err = templateNode.Execute(ctx, msg)
		require.NoError(t, err)

		assert.Equal(t, "Name: Bob, Age: 30", msg.Payload)
	})
}
