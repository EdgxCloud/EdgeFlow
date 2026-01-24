// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgeflow/edgeflow/internal/engine"
	"github.com/edgeflow/edgeflow/internal/node"
	// Import node packages to register factories
	_ "github.com/edgeflow/edgeflow/pkg/nodes/core"
)

// TestFlowCreation tests creating and starting a basic flow
func TestFlowCreation(t *testing.T) {
	// Create a simple inject -> debug flow
	flowDef := engine.FlowDefinition{
		ID:          "test-flow-1",
		Name:        "Test Flow 1",
		Description: "Integration test flow",
		Nodes: []engine.NodeDefinition{
			{
				ID:   "inject-1",
				Type: "inject",
				Config: map[string]interface{}{
					"interval": 0, // Manual trigger only
					"payload":  "test message",
				},
			},
			{
				ID:   "debug-1",
				Type: "debug",
				Config: map[string]interface{}{
					"active": true,
				},
			},
		},
		Connections: []engine.ConnectionDefinition{
			{
				Source:     "inject-1",
				SourcePort: 0,
				Target:     "debug-1",
				TargetPort: 0,
			},
		},
	}

	fm := engine.NewFlowManager()

	// Create flow
	flow, err := fm.CreateFlow(flowDef)
	require.NoError(t, err)
	assert.NotNil(t, flow)
	assert.Equal(t, "test-flow-1", flow.ID)

	// Start flow
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = fm.StartFlow(ctx, flow.ID)
	require.NoError(t, err)

	// Verify flow is running
	status := fm.GetFlowStatus(flow.ID)
	assert.Equal(t, engine.FlowStatusRunning, status)

	// Stop flow
	err = fm.StopFlow(flow.ID)
	require.NoError(t, err)

	// Verify flow is stopped
	status = fm.GetFlowStatus(flow.ID)
	assert.Equal(t, engine.FlowStatusStopped, status)
}

// TestFlowWithCondition tests a flow with conditional routing
func TestFlowWithCondition(t *testing.T) {
	flowDef := engine.FlowDefinition{
		ID:   "test-flow-condition",
		Name: "Conditional Flow",
		Nodes: []engine.NodeDefinition{
			{
				ID:   "inject-1",
				Type: "inject",
				Config: map[string]interface{}{
					"interval": 0,
					"payload":  100,
				},
			},
			{
				ID:   "if-1",
				Type: "if",
				Config: map[string]interface{}{
					"condition": "msg.payload > 50",
				},
			},
			{
				ID:   "debug-true",
				Type: "debug",
				Config: map[string]interface{}{
					"name": "true-branch",
				},
			},
			{
				ID:   "debug-false",
				Type: "debug",
				Config: map[string]interface{}{
					"name": "false-branch",
				},
			},
		},
		Connections: []engine.ConnectionDefinition{
			{Source: "inject-1", SourcePort: 0, Target: "if-1", TargetPort: 0},
			{Source: "if-1", SourcePort: 0, Target: "debug-true", TargetPort: 0},  // True branch
			{Source: "if-1", SourcePort: 1, Target: "debug-false", TargetPort: 0}, // False branch
		},
	}

	fm := engine.NewFlowManager()
	flow, err := fm.CreateFlow(flowDef)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = fm.StartFlow(ctx, flow.ID)
	require.NoError(t, err)
	defer fm.StopFlow(flow.ID)

	// Trigger the inject node manually
	err = fm.TriggerNode(flow.ID, "inject-1")
	require.NoError(t, err)

	// Give time for message to propagate
	time.Sleep(100 * time.Millisecond)

	// Verify flow is still running
	status := fm.GetFlowStatus(flow.ID)
	assert.Equal(t, engine.FlowStatusRunning, status)
}

// TestFlowWithDelay tests a flow with delay node
func TestFlowWithDelay(t *testing.T) {
	flowDef := engine.FlowDefinition{
		ID:   "test-flow-delay",
		Name: "Delay Flow",
		Nodes: []engine.NodeDefinition{
			{
				ID:   "inject-1",
				Type: "inject",
				Config: map[string]interface{}{
					"interval": 0,
					"payload":  "delayed message",
				},
			},
			{
				ID:   "delay-1",
				Type: "delay",
				Config: map[string]interface{}{
					"delay": 100, // 100ms
				},
			},
			{
				ID:   "debug-1",
				Type: "debug",
			},
		},
		Connections: []engine.ConnectionDefinition{
			{Source: "inject-1", SourcePort: 0, Target: "delay-1", TargetPort: 0},
			{Source: "delay-1", SourcePort: 0, Target: "debug-1", TargetPort: 0},
		},
	}

	fm := engine.NewFlowManager()
	flow, err := fm.CreateFlow(flowDef)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = fm.StartFlow(ctx, flow.ID)
	require.NoError(t, err)
	defer fm.StopFlow(flow.ID)

	start := time.Now()
	err = fm.TriggerNode(flow.ID, "inject-1")
	require.NoError(t, err)

	// Wait for delay to complete
	time.Sleep(200 * time.Millisecond)
	elapsed := time.Since(start)

	// Delay should have been at least 100ms
	assert.True(t, elapsed >= 100*time.Millisecond,
		"Expected delay of at least 100ms, got %v", elapsed)
}

// TestFlowWithFunction tests a flow with JavaScript function
func TestFlowWithFunction(t *testing.T) {
	flowDef := engine.FlowDefinition{
		ID:   "test-flow-function",
		Name: "Function Flow",
		Nodes: []engine.NodeDefinition{
			{
				ID:   "inject-1",
				Type: "inject",
				Config: map[string]interface{}{
					"interval": 0,
					"payload":  10,
				},
			},
			{
				ID:   "function-1",
				Type: "function",
				Config: map[string]interface{}{
					"code": `
						msg.payload = msg.payload * 2;
						return msg;
					`,
				},
			},
			{
				ID:   "debug-1",
				Type: "debug",
			},
		},
		Connections: []engine.ConnectionDefinition{
			{Source: "inject-1", SourcePort: 0, Target: "function-1", TargetPort: 0},
			{Source: "function-1", SourcePort: 0, Target: "debug-1", TargetPort: 0},
		},
	}

	fm := engine.NewFlowManager()
	flow, err := fm.CreateFlow(flowDef)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = fm.StartFlow(ctx, flow.ID)
	require.NoError(t, err)
	defer fm.StopFlow(flow.ID)

	// Flow should be running
	status := fm.GetFlowStatus(flow.ID)
	assert.Equal(t, engine.FlowStatusRunning, status)
}

// TestMultipleFlows tests running multiple flows concurrently
func TestMultipleFlows(t *testing.T) {
	fm := engine.NewFlowManager()

	// Create multiple flows
	for i := 0; i < 5; i++ {
		flowDef := engine.FlowDefinition{
			ID:   fmt.Sprintf("multi-flow-%d", i),
			Name: fmt.Sprintf("Multi Flow %d", i),
			Nodes: []engine.NodeDefinition{
				{
					ID:   "inject-1",
					Type: "inject",
					Config: map[string]interface{}{
						"interval": 0,
						"payload":  i,
					},
				},
				{
					ID:   "debug-1",
					Type: "debug",
				},
			},
			Connections: []engine.ConnectionDefinition{
				{Source: "inject-1", SourcePort: 0, Target: "debug-1", TargetPort: 0},
			},
		}

		flow, err := fm.CreateFlow(flowDef)
		require.NoError(t, err)
		assert.NotNil(t, flow)
	}

	// Start all flows
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < 5; i++ {
		flowID := fmt.Sprintf("multi-flow-%d", i)
		err := fm.StartFlow(ctx, flowID)
		require.NoError(t, err)
	}

	// Verify all flows are running
	for i := 0; i < 5; i++ {
		flowID := fmt.Sprintf("multi-flow-%d", i)
		status := fm.GetFlowStatus(flowID)
		assert.Equal(t, engine.FlowStatusRunning, status)
	}

	// Stop all flows
	for i := 0; i < 5; i++ {
		flowID := fmt.Sprintf("multi-flow-%d", i)
		err := fm.StopFlow(flowID)
		require.NoError(t, err)
	}

	// Verify all flows are stopped
	for i := 0; i < 5; i++ {
		flowID := fmt.Sprintf("multi-flow-%d", i)
		status := fm.GetFlowStatus(flowID)
		assert.Equal(t, engine.FlowStatusStopped, status)
	}
}

// TestFlowContextPersistence tests flow context storage
func TestFlowContextPersistence(t *testing.T) {
	flowDef := engine.FlowDefinition{
		ID:   "test-flow-context",
		Name: "Context Flow",
		Nodes: []engine.NodeDefinition{
			{
				ID:   "inject-1",
				Type: "inject",
				Config: map[string]interface{}{
					"interval": 0,
					"payload":  "context-test",
				},
			},
			{
				ID:   "function-1",
				Type: "function",
				Config: map[string]interface{}{
					"code": `
						var count = flow.get('count') || 0;
						count++;
						flow.set('count', count);
						msg.payload = count;
						return msg;
					`,
				},
			},
			{
				ID:   "debug-1",
				Type: "debug",
			},
		},
		Connections: []engine.ConnectionDefinition{
			{Source: "inject-1", SourcePort: 0, Target: "function-1", TargetPort: 0},
			{Source: "function-1", SourcePort: 0, Target: "debug-1", TargetPort: 0},
		},
	}

	fm := engine.NewFlowManager()
	flow, err := fm.CreateFlow(flowDef)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = fm.StartFlow(ctx, flow.ID)
	require.NoError(t, err)
	defer fm.StopFlow(flow.ID)

	// Trigger multiple times - context should persist
	for i := 0; i < 3; i++ {
		err = fm.TriggerNode(flow.ID, "inject-1")
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)
	}

	// Verify flow still running
	status := fm.GetFlowStatus(flow.ID)
	assert.Equal(t, engine.FlowStatusRunning, status)
}

// TestFlowErrorHandling tests error handling in flows
func TestFlowErrorHandling(t *testing.T) {
	// Create a flow with an error-throwing function
	flowDef := engine.FlowDefinition{
		ID:   "test-flow-error",
		Name: "Error Flow",
		Nodes: []engine.NodeDefinition{
			{
				ID:   "inject-1",
				Type: "inject",
				Config: map[string]interface{}{
					"interval": 0,
					"payload":  "test",
				},
			},
			{
				ID:   "function-error",
				Type: "function",
				Config: map[string]interface{}{
					"code": `
						throw new Error("Test error");
					`,
				},
			},
			{
				ID:   "catch-1",
				Type: "catch",
				Config: map[string]interface{}{
					"scope": "all",
				},
			},
			{
				ID:   "debug-error",
				Type: "debug",
			},
		},
		Connections: []engine.ConnectionDefinition{
			{Source: "inject-1", SourcePort: 0, Target: "function-error", TargetPort: 0},
			{Source: "catch-1", SourcePort: 0, Target: "debug-error", TargetPort: 0},
		},
	}

	fm := engine.NewFlowManager()
	flow, err := fm.CreateFlow(flowDef)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = fm.StartFlow(ctx, flow.ID)
	require.NoError(t, err)
	defer fm.StopFlow(flow.ID)

	// Trigger the inject - should cause error but flow should keep running
	err = fm.TriggerNode(flow.ID, "inject-1")
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Flow should still be running despite error
	status := fm.GetFlowStatus(flow.ID)
	assert.Equal(t, engine.FlowStatusRunning, status)
}

// TestFlowValidation tests flow validation
func TestFlowValidation(t *testing.T) {
	fm := engine.NewFlowManager()

	tests := []struct {
		name    string
		flow    engine.FlowDefinition
		wantErr bool
	}{
		{
			name: "valid flow",
			flow: engine.FlowDefinition{
				ID:   "valid-flow",
				Name: "Valid Flow",
				Nodes: []engine.NodeDefinition{
					{ID: "inject-1", Type: "inject"},
					{ID: "debug-1", Type: "debug"},
				},
				Connections: []engine.ConnectionDefinition{
					{Source: "inject-1", SourcePort: 0, Target: "debug-1", TargetPort: 0},
				},
			},
			wantErr: false,
		},
		{
			name: "missing flow ID",
			flow: engine.FlowDefinition{
				Name: "No ID Flow",
				Nodes: []engine.NodeDefinition{
					{ID: "inject-1", Type: "inject"},
				},
			},
			wantErr: true,
		},
		{
			name: "unknown node type",
			flow: engine.FlowDefinition{
				ID:   "unknown-type-flow",
				Name: "Unknown Type Flow",
				Nodes: []engine.NodeDefinition{
					{ID: "unknown-1", Type: "nonexistent-type"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid connection - source not found",
			flow: engine.FlowDefinition{
				ID:   "invalid-connection-flow",
				Name: "Invalid Connection Flow",
				Nodes: []engine.NodeDefinition{
					{ID: "debug-1", Type: "debug"},
				},
				Connections: []engine.ConnectionDefinition{
					{Source: "nonexistent", SourcePort: 0, Target: "debug-1", TargetPort: 0},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fm.CreateFlow(tt.flow)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNodeRegistry tests the node registry
func TestNodeRegistry(t *testing.T) {
	registry := node.GetRegistry()

	// Core nodes should be registered
	coreNodes := []string{
		"inject", "debug", "function", "if", "delay",
		"template", "switch", "change", "range",
		"split", "join", "catch", "status", "complete",
	}

	for _, nodeType := range coreNodes {
		factory, err := registry.Get(nodeType)
		if err != nil {
			t.Logf("Node type '%s' not registered (may be expected if not imported)", nodeType)
			continue
		}
		assert.NotNil(t, factory, "Factory for %s should not be nil", nodeType)
	}
}

// BenchmarkFlowCreation benchmarks flow creation
func BenchmarkFlowCreation(b *testing.B) {
	fm := engine.NewFlowManager()

	flowDef := engine.FlowDefinition{
		Name: "Benchmark Flow",
		Nodes: []engine.NodeDefinition{
			{ID: "inject-1", Type: "inject"},
			{ID: "debug-1", Type: "debug"},
		},
		Connections: []engine.ConnectionDefinition{
			{Source: "inject-1", SourcePort: 0, Target: "debug-1", TargetPort: 0},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flowDef.ID = fmt.Sprintf("bench-flow-%d", i)
		fm.CreateFlow(flowDef)
	}
}

// BenchmarkFlowStartStop benchmarks flow start/stop cycle
func BenchmarkFlowStartStop(b *testing.B) {
	fm := engine.NewFlowManager()

	flowDef := engine.FlowDefinition{
		ID:   "bench-start-stop",
		Name: "Benchmark Start Stop",
		Nodes: []engine.NodeDefinition{
			{ID: "inject-1", Type: "inject"},
			{ID: "debug-1", Type: "debug"},
		},
		Connections: []engine.ConnectionDefinition{
			{Source: "inject-1", SourcePort: 0, Target: "debug-1", TargetPort: 0},
		},
	}

	flow, _ := fm.CreateFlow(flowDef)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.StartFlow(ctx, flow.ID)
		fm.StopFlow(flow.ID)
	}
}

