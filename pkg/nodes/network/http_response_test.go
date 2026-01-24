package network

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPResponseNode(t *testing.T) {
	t.Run("Send basic response", func(t *testing.T) {
		respNode := NewHTTPResponseNode()
		err := respNode.Init(map[string]interface{}{
			"statusCode": 200,
		})
		require.NoError(t, err)

		// Create mock response writer
		recorder := httptest.NewRecorder()

		msg := node.Message{
			Payload: "Hello, World!",
		}

		// Set response writer in message metadata
		if eMsg, ok := interface{}(&msg).(*node.EnhancedMessage); ok {
			eMsg.Res = recorder
		}

		_, err = respNode.Execute(context.Background(), msg)
		require.NoError(t, err)
	})

	t.Run("Send JSON response", func(t *testing.T) {
		respNode := NewHTTPResponseNode()
		err := respNode.Init(map[string]interface{}{
			"statusCode": 200,
			"headers": map[string]string{
				"Content-Type": "application/json",
			},
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()

		msg := node.Message{
			Payload: map[string]interface{}{
				"status": "success",
				"data":   "test",
			},
		}

		if eMsg, ok := interface{}(&msg).(*node.EnhancedMessage); ok {
			eMsg.Res = recorder
		}

		_, err = respNode.Execute(context.Background(), msg)
		require.NoError(t, err)
	})

	t.Run("Set custom status code", func(t *testing.T) {
		respNode := NewHTTPResponseNode()
		err := respNode.Init(map[string]interface{}{
			"statusCode": 404,
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()

		msg := node.Message{
			Payload: "Not Found",
		}

		if eMsg, ok := interface{}(&msg).(*node.EnhancedMessage); ok {
			eMsg.Res = recorder
		}

		_, err = respNode.Execute(context.Background(), msg)
		require.NoError(t, err)
	})

	t.Run("Set custom headers", func(t *testing.T) {
		respNode := NewHTTPResponseNode()
		err := respNode.Init(map[string]interface{}{
			"statusCode": 200,
			"headers": map[string]string{
				"X-Custom-Header": "custom-value",
				"Content-Type":    "text/plain",
			},
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()

		msg := node.Message{
			Payload: "Custom response",
		}

		if eMsg, ok := interface{}(&msg).(*node.EnhancedMessage); ok {
			eMsg.Res = recorder
		}

		_, err = respNode.Execute(context.Background(), msg)
		require.NoError(t, err)
	})

	t.Run("Set cookies", func(t *testing.T) {
		respNode := NewHTTPResponseNode()
		err := respNode.Init(map[string]interface{}{
			"statusCode": 200,
			"cookies": []map[string]interface{}{
				{
					"name":  "session_id",
					"value": "abc123",
				},
			},
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()

		msg := node.Message{
			Payload: "Response with cookie",
		}

		if eMsg, ok := interface{}(&msg).(*node.EnhancedMessage); ok {
			eMsg.Res = recorder
		}

		_, err = respNode.Execute(context.Background(), msg)
		require.NoError(t, err)
	})
}

func TestHTTPResponseErrorHandling(t *testing.T) {
	t.Run("Handle missing response writer", func(t *testing.T) {
		respNode := NewHTTPResponseNode()
		err := respNode.Init(map[string]interface{}{
			"statusCode": 200,
		})
		require.NoError(t, err)

		msg := node.Message{
			Payload: "Test",
		}

		_, err = respNode.Execute(context.Background(), msg)
		assert.Error(t, err)
	})

	t.Run("Handle invalid status code", func(t *testing.T) {
		respNode := NewHTTPResponseNode()
		err := respNode.Init(map[string]interface{}{
			"statusCode": 999,
		})
		assert.Error(t, err)
	})
}
