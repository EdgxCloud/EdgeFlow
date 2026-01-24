package engine

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlowCreation(t *testing.T) {
	flow := NewFlow("test-flow", "Test Flow")

	assert.Equal(t, "test-flow", flow.ID)
	assert.Equal(t, "Test Flow", flow.Name)
	assert.Equal(t, FlowStatusStopped, flow.Status)
	assert.Empty(t, flow.Nodes)
	assert.Empty(t, flow.Connections)
}

func TestFlowAddNode(t *testing.T) {
	flow := NewFlow("test-flow", "Test Flow")

	nodeInstance := &node.Node{
		ID:   "node-1",
		Type: "test",
		Name: "Test Node",
	}

	err := flow.AddNode(nodeInstance)
	require.NoError(t, err)

	assert.Len(t, flow.Nodes, 1)
	assert.Equal(t, nodeInstance, flow.Nodes["node-1"])

	err = flow.AddNode(nodeInstance)
	assert.Error(t, err)
}

func TestFlowRemoveNode(t *testing.T) {
	flow := NewFlow("test-flow", "Test Flow")

	nodeInstance := &node.Node{
		ID:   "node-1",
		Type: "test",
		Name: "Test Node",
	}

	flow.AddNode(nodeInstance)
	assert.Len(t, flow.Nodes, 1)

	err := flow.RemoveNode("node-1")
	require.NoError(t, err)
	assert.Empty(t, flow.Nodes)

	err = flow.RemoveNode("non-existent")
	assert.Error(t, err)
}

func TestFlowAddConnection(t *testing.T) {
	flow := NewFlow("test-flow", "Test Flow")

	node1 := &node.Node{ID: "node-1", Type: "test"}
	node2 := &node.Node{ID: "node-2", Type: "test"}

	flow.AddNode(node1)
	flow.AddNode(node2)

	conn := Connection{
		ID:     "conn-1",
		Source: "node-1",
		Target: "node-2",
	}

	err := flow.AddConnection(conn)
	require.NoError(t, err)
	assert.Len(t, flow.Connections, 1)

	err = flow.AddConnection(conn)
	assert.Error(t, err)
}

func TestFlowRemoveConnection(t *testing.T) {
	flow := NewFlow("test-flow", "Test Flow")

	node1 := &node.Node{ID: "node-1", Type: "test"}
	node2 := &node.Node{ID: "node-2", Type: "test"}

	flow.AddNode(node1)
	flow.AddNode(node2)

	conn := Connection{
		ID:     "conn-1",
		Source: "node-1",
		Target: "node-2",
	}

	flow.AddConnection(conn)
	assert.Len(t, flow.Connections, 1)

	err := flow.RemoveConnection("conn-1")
	require.NoError(t, err)
	assert.Empty(t, flow.Connections)

	err = flow.RemoveConnection("non-existent")
	assert.Error(t, err)
}

func TestFlowValidation(t *testing.T) {
	flow := NewFlow("test-flow", "Test Flow")

	err := flow.Validate()
	assert.NoError(t, err)

	node1 := &node.Node{ID: "node-1", Type: "test"}
	flow.AddNode(node1)

	conn := Connection{
		ID:     "conn-1",
		Source: "node-1",
		Target: "non-existent",
	}

	flow.Connections = append(flow.Connections, conn)

	err = flow.Validate()
	assert.Error(t, err)
}

func TestFlowStartStop(t *testing.T) {
	flow := NewFlow("test-flow", "Test Flow")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := flow.Start(ctx)
	require.NoError(t, err)
	assert.Equal(t, FlowStatusRunning, flow.Status)

	err = flow.Stop()
	require.NoError(t, err)
	assert.Equal(t, FlowStatusStopped, flow.Status)
}

func TestFlowConcurrentAccess(t *testing.T) {
	flow := NewFlow("test-flow", "Test Flow")

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			nodeInstance := &node.Node{
				ID:   fmt.Sprintf("node-%d", id),
				Type: "test",
			}
			flow.AddNode(nodeInstance)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Len(t, flow.Nodes, 10)
}
