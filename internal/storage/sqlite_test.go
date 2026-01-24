package storage

import (
	"os"
	"testing"

	"github.com/edgeflow/edgeflow/internal/engine"
	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteStorage_SaveAndGetFlow(t *testing.T) {
	// Create temp database
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	storage, err := NewSQLiteStorage(tmpFile.Name())
	require.NoError(t, err)
	defer storage.Close()

	// Create a flow
	flow := engine.NewFlow("test-flow-1", "Test Flow 1")
	flow.Description = "A test flow"

	// Save the flow
	err = storage.SaveFlow(flow)
	require.NoError(t, err)

	// Get the flow
	retrieved, err := storage.GetFlow("test-flow-1")
	require.NoError(t, err)

	assert.Equal(t, flow.ID, retrieved.ID)
	assert.Equal(t, flow.Name, retrieved.Name)
	assert.Equal(t, flow.Description, retrieved.Description)
}

func TestSQLiteStorage_ListFlows(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	storage, err := NewSQLiteStorage(tmpFile.Name())
	require.NoError(t, err)
	defer storage.Close()

	// Create and save multiple flows
	flow1 := engine.NewFlow("flow-1", "Flow 1")
	flow2 := engine.NewFlow("flow-2", "Flow 2")
	flow3 := engine.NewFlow("flow-3", "Flow 3")

	require.NoError(t, storage.SaveFlow(flow1))
	require.NoError(t, storage.SaveFlow(flow2))
	require.NoError(t, storage.SaveFlow(flow3))

	// List flows
	flows, err := storage.ListFlows()
	require.NoError(t, err)

	assert.Len(t, flows, 3)
}

func TestSQLiteStorage_DeleteFlow(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	storage, err := NewSQLiteStorage(tmpFile.Name())
	require.NoError(t, err)
	defer storage.Close()

	// Create and save a flow
	flow := engine.NewFlow("delete-test", "Delete Test")
	require.NoError(t, storage.SaveFlow(flow))

	// Verify it exists
	_, err = storage.GetFlow("delete-test")
	require.NoError(t, err)

	// Delete the flow
	err = storage.DeleteFlow("delete-test")
	require.NoError(t, err)

	// Verify it's deleted
	_, err = storage.GetFlow("delete-test")
	assert.Error(t, err)
}

func TestSQLiteStorage_DeleteNonExistentFlow(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	storage, err := NewSQLiteStorage(tmpFile.Name())
	require.NoError(t, err)
	defer storage.Close()

	err = storage.DeleteFlow("non-existent")
	assert.Error(t, err)
}

func TestSQLiteStorage_UpdateFlow(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	storage, err := NewSQLiteStorage(tmpFile.Name())
	require.NoError(t, err)
	defer storage.Close()

	// Create and save a flow
	flow := engine.NewFlow("update-test", "Original Name")
	require.NoError(t, storage.SaveFlow(flow))

	// Update the flow
	flow.Name = "Updated Name"
	flow.Description = "Updated description"
	require.NoError(t, storage.UpdateFlow(flow))

	// Get and verify
	retrieved, err := storage.GetFlow("update-test")
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, "Updated description", retrieved.Description)
}

func TestSQLiteStorage_GetNonExistentFlow(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	storage, err := NewSQLiteStorage(tmpFile.Name())
	require.NoError(t, err)
	defer storage.Close()

	_, err = storage.GetFlow("non-existent")
	assert.Error(t, err)
}

func TestSQLiteStorage_FlowWithNodes(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	storage, err := NewSQLiteStorage(tmpFile.Name())
	require.NoError(t, err)
	defer storage.Close()

	// Create a flow with nodes
	flow := engine.NewFlow("flow-with-nodes", "Flow With Nodes")

	node1 := &node.Node{
		ID:   "node-1",
		Type: "inject",
		Name: "Inject Node",
	}
	node2 := &node.Node{
		ID:   "node-2",
		Type: "debug",
		Name: "Debug Node",
	}

	flow.AddNode(node1)
	flow.AddNode(node2)

	// Add a connection
	conn := engine.Connection{
		ID:     "conn-1",
		Source: "node-1",
		Target: "node-2",
	}
	flow.AddConnection(conn)

	// Save the flow
	require.NoError(t, storage.SaveFlow(flow))

	// Get and verify
	retrieved, err := storage.GetFlow("flow-with-nodes")
	require.NoError(t, err)

	assert.Len(t, retrieved.Nodes, 2)
	assert.Len(t, retrieved.Connections, 1)
}

func TestSQLiteStorage_EmptyDatabase(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	storage, err := NewSQLiteStorage(tmpFile.Name())
	require.NoError(t, err)
	defer storage.Close()

	flows, err := storage.ListFlows()
	require.NoError(t, err)

	assert.Empty(t, flows)
}

func TestSQLiteStorage_InvalidPath(t *testing.T) {
	// Try to create storage with invalid path
	_, err := NewSQLiteStorage("/invalid/path/that/does/not/exist/test.db")
	// This might not fail immediately on some systems, so we don't assert error
	// but the test ensures no panic occurs
	if err != nil {
		t.Logf("Expected error for invalid path: %v", err)
	}
}
