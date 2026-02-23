package integration

import (
	"context"
	"testing"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/engine"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/EdgxCloud/EdgeFlow/pkg/nodes/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase8Integration(t *testing.T) {
	registry := node.NewRegistry()

	core.RegisterAllNodes(registry)

	t.Run("Template Node Integration", func(t *testing.T) {
		node, err := registry.Create("template")
		require.NoError(t, err)

		config := map[string]interface{}{
			"template": "Hello {{name}}!",
			"syntax":   "mustache",
		}

		err = node.Init(config)
		require.NoError(t, err)

		msg := node.Message{
			Payload: map[string]interface{}{
				"name": "World",
			},
		}

		result, err := node.Execute(context.Background(), msg)
		require.NoError(t, err)
		assert.Equal(t, "Hello World!", result.Payload)
	})

	t.Run("Change Node Integration", func(t *testing.T) {
		node, err := registry.Create("change")
		require.NoError(t, err)

		config := map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{
					"t":  "set",
					"p":  "payload.message",
					"to": "Hello",
				},
			},
		}

		err = node.Init(config)
		require.NoError(t, err)

		msg := node.Message{
			Payload: map[string]interface{}{},
		}

		result, err := node.Execute(context.Background(), msg)
		require.NoError(t, err)

		payload, ok := result.Payload.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Hello", payload["message"])
	})

	t.Run("Switch Node Integration", func(t *testing.T) {
		node, err := registry.Create("switch")
		require.NoError(t, err)

		config := map[string]interface{}{
			"property": "payload.temperature",
			"rules": []interface{}{
				map[string]interface{}{
					"t": "gt",
					"v": float64(25),
				},
			},
		}

		err = node.Init(config)
		require.NoError(t, err)

		msg := node.Message{
			Payload: map[string]interface{}{
				"temperature": float64(30),
			},
		}

		result, err := node.Execute(context.Background(), msg)
		require.NoError(t, err)
		assert.NotNil(t, result.Payload)
	})

	t.Run("Split and Join Integration", func(t *testing.T) {
		splitNode, err := registry.Create("split")
		require.NoError(t, err)

		err = splitNode.Init(map[string]interface{}{})
		require.NoError(t, err)

		msg := node.Message{
			Payload: []interface{}{"a", "b", "c"},
		}

		result, err := splitNode.Execute(context.Background(), msg)
		require.NoError(t, err)

		parts, ok := result.Metadata["parts"]
		require.True(t, ok)
		assert.NotNil(t, parts)
	})

	t.Run("Catch Node Integration", func(t *testing.T) {
		catchNode, err := registry.Create("catch")
		require.NoError(t, err)

		config := map[string]interface{}{
			"scope": "all",
		}

		err = catchNode.Init(config)
		require.NoError(t, err)

		msg := node.Message{
			Metadata: map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Test error",
					"source":  "test-node",
				},
			},
		}

		result, err := catchNode.Execute(context.Background(), msg)
		require.NoError(t, err)
		assert.NotNil(t, result.Payload)
	})
}

func TestDeployManagerIntegration(t *testing.T) {
	t.Run("Full Deployment", func(t *testing.T) {
		dm := engine.NewDeployManager()

		flows := []engine.Flow{
			{
				ID:    "flow1",
				Name:  "Test Flow 1",
				Nodes: []engine.FlowNode{},
			},
		}

		req := engine.DeployRequest{
			Mode:  engine.DeployModeFull,
			Flows: flows,
		}

		ctx := context.Background()
		result, err := dm.Deploy(ctx, req)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 1, result.FlowsDeployed)
	})

	t.Run("Modified Deployment", func(t *testing.T) {
		dm := engine.NewDeployManager()

		flows := []engine.Flow{
			{
				ID:    "flow1",
				Name:  "Test Flow 1",
				Nodes: []engine.FlowNode{},
			},
		}

		req := engine.DeployRequest{
			Mode:  engine.DeployModeModified,
			Flows: flows,
		}

		ctx := context.Background()
		result, err := dm.Deploy(ctx, req)
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("Flow Deployment", func(t *testing.T) {
		dm := engine.NewDeployManager()

		flows := []engine.Flow{
			{
				ID:    "flow1",
				Name:  "Test Flow 1",
				Nodes: []engine.FlowNode{},
			},
		}

		req := engine.DeployRequest{
			Mode:   engine.DeployModeFlow,
			Flows:  flows,
			FlowID: "flow1",
		}

		ctx := context.Background()
		result, err := dm.Deploy(ctx, req)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 1, result.FlowsDeployed)
	})
}

func TestRedisContextIntegration(t *testing.T) {
	t.Skip("Requires Redis running")

	storage, err := engine.NewRedisContextStorage("localhost:6379", "", "edgeflow")
	require.NoError(t, err)
	defer storage.Close()

	t.Run("Set and Get Node Context", func(t *testing.T) {
		err := storage.Set(context.Background(), engine.ScopeNode, "node1", "key1", "value1", 0)
		require.NoError(t, err)

		value, err := storage.Get(context.Background(), engine.ScopeNode, "node1", "key1")
		require.NoError(t, err)
		assert.Equal(t, "value1", value)
	})

	t.Run("Set with TTL", func(t *testing.T) {
		err := storage.Set(context.Background(), engine.ScopeNode, "node2", "key2", "value2", 1*time.Second)
		require.NoError(t, err)

		value, err := storage.Get(context.Background(), engine.ScopeNode, "node2", "key2")
		require.NoError(t, err)
		assert.Equal(t, "value2", value)

		time.Sleep(2 * time.Second)

		value, err = storage.Get(context.Background(), engine.ScopeNode, "node2", "key2")
		assert.Error(t, err)
	})

	t.Run("Delete Context", func(t *testing.T) {
		err := storage.Set(context.Background(), engine.ScopeNode, "node3", "key3", "value3", 0)
		require.NoError(t, err)

		err = storage.Delete(context.Background(), engine.ScopeNode, "node3", "key3")
		require.NoError(t, err)

		_, err = storage.Get(context.Background(), engine.ScopeNode, "node3", "key3")
		assert.Error(t, err)
	})

	t.Run("Get All Keys", func(t *testing.T) {
		err := storage.Set(context.Background(), engine.ScopeFlow, "flow1", "key1", "value1", 0)
		require.NoError(t, err)
		err = storage.Set(context.Background(), engine.ScopeFlow, "flow1", "key2", "value2", 0)
		require.NoError(t, err)

		keys, err := storage.GetAll(context.Background(), engine.ScopeFlow, "flow1")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(keys), 2)
	})
}

func TestFlowExecution(t *testing.T) {
	registry := node.NewRegistry()
	core.RegisterAllNodes(registry)

	t.Run("Simple Flow with Inject and Debug", func(t *testing.T) {
		injectNode, err := registry.Create("inject")
		require.NoError(t, err)

		debugNode, err := registry.Create("debug")
		require.NoError(t, err)

		err = injectNode.Init(map[string]interface{}{
			"payload": "Hello World",
		})
		require.NoError(t, err)

		err = debugNode.Init(map[string]interface{}{
			"console": true,
		})
		require.NoError(t, err)

		msg, err := injectNode.Execute(context.Background(), node.Message{})
		require.NoError(t, err)

		_, err = debugNode.Execute(context.Background(), msg)
		require.NoError(t, err)
	})

	t.Run("Flow with Template and Change", func(t *testing.T) {
		templateNode, err := registry.Create("template")
		require.NoError(t, err)

		changeNode, err := registry.Create("change")
		require.NoError(t, err)

		err = templateNode.Init(map[string]interface{}{
			"template": "User: {{name}}",
			"syntax":   "mustache",
		})
		require.NoError(t, err)

		err = changeNode.Init(map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{
					"t":  "set",
					"p":  "payload.processed",
					"to": true,
				},
			},
		})
		require.NoError(t, err)

		msg := node.Message{
			Payload: map[string]interface{}{
				"name": "Alice",
			},
		}

		msg, err = templateNode.Execute(context.Background(), msg)
		require.NoError(t, err)

		msg, err = changeNode.Execute(context.Background(), msg)
		require.NoError(t, err)

		payload, ok := msg.Payload.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, true, payload["processed"])
	})
}
