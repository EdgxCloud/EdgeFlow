package node

import (
	"context"
	"testing"
	"time"
)

// MockExecutor is a mock implementation of the Executor interface for testing
type MockExecutor struct {
	initCalled    bool
	executeCalled bool
	cleanupCalled bool
	executeFunc   func(ctx context.Context, msg Message) (Message, error)
}

func (m *MockExecutor) Init(config map[string]interface{}) error {
	m.initCalled = true
	return nil
}

func (m *MockExecutor) Execute(ctx context.Context, msg Message) (Message, error) {
	m.executeCalled = true
	if m.executeFunc != nil {
		return m.executeFunc(ctx, msg)
	}
	return msg, nil
}

func (m *MockExecutor) Cleanup() error {
	m.cleanupCalled = true
	return nil
}

func TestNewNode(t *testing.T) {
	executor := &MockExecutor{}
	node := NewNode("test-type", "Test Node", NodeTypeProcessing, executor)

	if node.Type != "test-type" {
		t.Errorf("Expected type 'test-type', got '%s'", node.Type)
	}

	if node.Name != "Test Node" {
		t.Errorf("Expected name 'Test Node', got '%s'", node.Name)
	}

	if node.Category != NodeTypeProcessing {
		t.Errorf("Expected category NodeTypeProcessing, got '%s'", node.Category)
	}

	if node.Status != NodeStatusIdle {
		t.Errorf("Expected status NodeStatusIdle, got '%s'", node.Status)
	}

	if node.ID == "" {
		t.Error("Expected node ID to be set")
	}
}

func TestNodeStartStop(t *testing.T) {
	executor := &MockExecutor{}
	node := NewNode("test", "Test", NodeTypeProcessing, executor)

	ctx := context.Background()

	// Test start
	err := node.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}

	if !executor.initCalled {
		t.Error("Expected Init to be called on start")
	}

	if node.GetStatus() != NodeStatusRunning {
		t.Errorf("Expected status NodeStatusRunning, got %s", node.GetStatus())
	}

	// Test starting already running node
	err = node.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting already running node")
	}

	// Test stop
	err = node.Stop()
	if err != nil {
		t.Fatalf("Failed to stop node: %v", err)
	}

	if !executor.cleanupCalled {
		t.Error("Expected Cleanup to be called on stop")
	}

	if node.GetStatus() != NodeStatusIdle {
		t.Errorf("Expected status NodeStatusIdle after stop, got %s", node.GetStatus())
	}
}

func TestNodeSend(t *testing.T) {
	executor := &MockExecutor{
		executeFunc: func(ctx context.Context, msg Message) (Message, error) {
			// Echo the message back
			return msg, nil
		},
	}
	node := NewNode("test", "Test", NodeTypeProcessing, executor)

	ctx := context.Background()
	err := node.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}
	defer node.Stop()

	msg := Message{
		Type: MessageTypeData,
		Payload: map[string]interface{}{
			"test": "value",
		},
	}

	err = node.Send(msg)
	if err != nil {
		t.Errorf("Failed to send message: %v", err)
	}

	// Give some time for message processing
	time.Sleep(100 * time.Millisecond)

	if !executor.executeCalled {
		t.Error("Expected Execute to be called")
	}
}

func TestNodeConnect(t *testing.T) {
	executor1 := &MockExecutor{}
	executor2 := &MockExecutor{}

	node1 := NewNode("test1", "Test 1", NodeTypeProcessing, executor1)
	node2 := NewNode("test2", "Test 2", NodeTypeProcessing, executor2)

	node1.Connect(node2)

	if len(node1.Outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(node1.Outputs))
	}

	if node1.Outputs[0] != node2.ID {
		t.Errorf("Expected output to be node2 ID, got %s", node1.Outputs[0])
	}

	if len(node2.Inputs) != 1 {
		t.Errorf("Expected 1 input, got %d", len(node2.Inputs))
	}

	if node2.Inputs[0] != node1.ID {
		t.Errorf("Expected input to be node1 ID, got %s", node2.Inputs[0])
	}
}

func TestNodeUpdateConfig(t *testing.T) {
	executor := &MockExecutor{}
	node := NewNode("test", "Test", NodeTypeProcessing, executor)

	config := map[string]interface{}{
		"key": "value",
	}

	err := node.UpdateConfig(config)
	if err != nil {
		t.Errorf("Failed to update config: %v", err)
	}

	if node.Config["key"] != "value" {
		t.Error("Config was not updated")
	}
}
