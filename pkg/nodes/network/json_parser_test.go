package network

import (
	"context"
	"testing"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONParserParse(t *testing.T) {
	t.Run("Parse valid JSON string", func(t *testing.T) {
		parser := NewJSONParserNode()
		err := parser.Init(map[string]interface{}{
			"action": "parse",
		})
		require.NoError(t, err)

		msg := node.Message{
			Payload: `{"name": "test", "value": 123}`,
		}

		result, err := parser.Execute(context.Background(), msg)
		require.NoError(t, err)

		payload, ok := result.Payload.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "test", payload["name"])
		assert.Equal(t, float64(123), payload["value"])
	})

	t.Run("Parse JSON array", func(t *testing.T) {
		parser := NewJSONParserNode()
		err := parser.Init(map[string]interface{}{
			"action": "parse",
		})
		require.NoError(t, err)

		msg := node.Message{
			Payload: `[1, 2, 3, 4, 5]`,
		}

		result, err := parser.Execute(context.Background(), msg)
		require.NoError(t, err)

		payload, ok := result.Payload.([]interface{})
		require.True(t, ok)
		assert.Len(t, payload, 5)
	})

	t.Run("Handle invalid JSON", func(t *testing.T) {
		parser := NewJSONParserNode()
		err := parser.Init(map[string]interface{}{
			"action": "parse",
		})
		require.NoError(t, err)

		msg := node.Message{
			Payload: `{invalid json}`,
		}

		_, err = parser.Execute(context.Background(), msg)
		assert.Error(t, err)
	})

	t.Run("Parse nested JSON", func(t *testing.T) {
		parser := NewJSONParserNode()
		err := parser.Init(map[string]interface{}{
			"action": "parse",
		})
		require.NoError(t, err)

		msg := node.Message{
			Payload: `{"user": {"name": "John", "age": 30}, "active": true}`,
		}

		result, err := parser.Execute(context.Background(), msg)
		require.NoError(t, err)

		payload, ok := result.Payload.(map[string]interface{})
		require.True(t, ok)

		user, ok := payload["user"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "John", user["name"])
	})
}

func TestJSONParserStringify(t *testing.T) {
	t.Run("Stringify object to JSON", func(t *testing.T) {
		parser := NewJSONParserNode()
		err := parser.Init(map[string]interface{}{
			"action": "stringify",
		})
		require.NoError(t, err)

		msg := node.Message{
			Payload: map[string]interface{}{
				"name":   "test",
				"value":  123,
				"active": true,
			},
		}

		result, err := parser.Execute(context.Background(), msg)
		require.NoError(t, err)

		payload, ok := result.Payload.(string)
		require.True(t, ok)
		assert.Contains(t, payload, `"name":"test"`)
		assert.Contains(t, payload, `"value":123`)
	})

	t.Run("Stringify array to JSON", func(t *testing.T) {
		parser := NewJSONParserNode()
		err := parser.Init(map[string]interface{}{
			"action": "stringify",
		})
		require.NoError(t, err)

		msg := node.Message{
			Payload: []interface{}{1, 2, 3, 4, 5},
		}

		result, err := parser.Execute(context.Background(), msg)
		require.NoError(t, err)

		payload, ok := result.Payload.(string)
		require.True(t, ok)
		assert.Equal(t, "[1,2,3,4,5]", payload)
	})

	t.Run("Stringify with pretty print", func(t *testing.T) {
		parser := NewJSONParserNode()
		err := parser.Init(map[string]interface{}{
			"action": "stringify",
			"pretty": true,
		})
		require.NoError(t, err)

		msg := node.Message{
			Payload: map[string]interface{}{
				"name": "test",
			},
		}

		result, err := parser.Execute(context.Background(), msg)
		require.NoError(t, err)

		payload, ok := result.Payload.(string)
		require.True(t, ok)
		assert.Contains(t, payload, "\n")
		assert.Contains(t, payload, "  ")
	})
}

func TestJSONParserBidirectional(t *testing.T) {
	t.Run("Parse then stringify returns same data", func(t *testing.T) {
		original := map[string]interface{}{
			"name":  "test",
			"value": 123,
			"items": []interface{}{1, 2, 3},
		}

		// Stringify
		stringify := NewJSONParserNode()
		err := stringify.Init(map[string]interface{}{"action": "stringify"})
		require.NoError(t, err)

		stringified, err := stringify.Execute(context.Background(), node.Message{Payload: original})
		require.NoError(t, err)

		// Parse back
		parse := NewJSONParserNode()
		err = parse.Init(map[string]interface{}{"action": "parse"})
		require.NoError(t, err)

		parsed, err := parse.Execute(context.Background(), stringified)
		require.NoError(t, err)

		result, ok := parsed.Payload.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "test", result["name"])
		assert.Equal(t, float64(123), result["value"])
	})
}
