package subflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock node executor for testing
type mockNodeExecutor struct {
	executions []executionRecord
}

type executionRecord struct {
	nodeID  string
	config  map[string]any
	payload any
}

func (m *mockNodeExecutor) Execute(ctx context.Context, nodeID string, config map[string]any, msg *Message) ([]*Message, error) {
	m.executions = append(m.executions, executionRecord{
		nodeID:  nodeID,
		config:  config,
		payload: msg.Payload,
	})

	// Echo the message back
	return []*Message{msg}, nil
}

func TestSubflowDefinition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		subflow *SubflowDefinition
		wantErr bool
	}{
		{
			name: "valid subflow",
			subflow: &SubflowDefinition{
				ID:   "subflow-1",
				Name: "Test Subflow",
				InputPorts: []PortDefinition{
					{Type: "input", Index: 0},
				},
				OutputPorts: []PortDefinition{
					{Type: "output", Index: 0},
				},
				Nodes: []NodeDefinition{
					{ID: "node-1", Type: "function"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			subflow: &SubflowDefinition{
				Name: "Test",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			subflow: &SubflowDefinition{
				ID: "subflow-1",
			},
			wantErr: true,
		},
		{
			name: "duplicate input port index",
			subflow: &SubflowDefinition{
				ID:   "subflow-1",
				Name: "Test",
				InputPorts: []PortDefinition{
					{Type: "input", Index: 0},
					{Type: "input", Index: 0},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate node ID",
			subflow: &SubflowDefinition{
				ID:   "subflow-1",
				Name: "Test",
				Nodes: []NodeDefinition{
					{ID: "node-1", Type: "function"},
					{ID: "node-1", Type: "debug"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.subflow.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSubflowDefinition_CreateInstance(t *testing.T) {
	def := &SubflowDefinition{
		ID:   "subflow-1",
		Name: "Test Subflow",
		OutputPorts: []PortDefinition{
			{Type: "output", Index: 0},
			{Type: "output", Index: 1},
		},
	}

	instance := def.CreateInstance("instance-1", "flow-1", 100, 200)

	assert.Equal(t, "instance-1", instance.ID)
	assert.Equal(t, "subflow:subflow-1", instance.Type)
	assert.Equal(t, "subflow-1", instance.SubflowID)
	assert.Equal(t, "flow-1", instance.Z)
	assert.Equal(t, 100.0, instance.X)
	assert.Equal(t, 200.0, instance.Y)
	assert.Len(t, instance.Wires, 2)
	assert.NotNil(t, instance.RuntimeContext)
}

func TestRegistry_RegisterDefinition(t *testing.T) {
	registry := NewRegistry()

	def := &SubflowDefinition{
		ID:   "subflow-1",
		Name: "Test Subflow",
	}

	err := registry.RegisterDefinition(def)
	assert.NoError(t, err)

	retrieved, err := registry.GetDefinition("subflow-1")
	assert.NoError(t, err)
	assert.Equal(t, "Test Subflow", retrieved.Name)
}

func TestRegistry_UnregisterDefinition(t *testing.T) {
	registry := NewRegistry()

	def := &SubflowDefinition{
		ID:   "subflow-1",
		Name: "Test Subflow",
	}

	registry.RegisterDefinition(def)

	// Should succeed when no instances exist
	err := registry.UnregisterDefinition("subflow-1")
	assert.NoError(t, err)

	// Should fail to get after unregister
	_, err = registry.GetDefinition("subflow-1")
	assert.Error(t, err)
}

func TestRegistry_UnregisterDefinition_WithInstances(t *testing.T) {
	registry := NewRegistry()

	def := &SubflowDefinition{
		ID:   "subflow-1",
		Name: "Test Subflow",
	}

	instance := &SubflowInstance{
		ID:        "instance-1",
		SubflowID: "subflow-1",
	}

	registry.RegisterDefinition(def)
	registry.RegisterInstance(instance)

	// Should fail when instances exist
	err := registry.UnregisterDefinition("subflow-1")
	assert.Error(t, err)
}

func TestRegistry_RegisterInstance(t *testing.T) {
	registry := NewRegistry()

	def := &SubflowDefinition{
		ID:   "subflow-1",
		Name: "Test Subflow",
	}

	instance := &SubflowInstance{
		ID:        "instance-1",
		SubflowID: "subflow-1",
	}

	registry.RegisterDefinition(def)
	err := registry.RegisterInstance(instance)
	assert.NoError(t, err)

	retrieved, err := registry.GetInstance("instance-1")
	assert.NoError(t, err)
	assert.Equal(t, "subflow-1", retrieved.SubflowID)
}

func TestRegistry_GetInstancesBySubflow(t *testing.T) {
	registry := NewRegistry()

	def := &SubflowDefinition{
		ID:   "subflow-1",
		Name: "Test Subflow",
	}

	registry.RegisterDefinition(def)
	registry.RegisterInstance(&SubflowInstance{ID: "instance-1", SubflowID: "subflow-1"})
	registry.RegisterInstance(&SubflowInstance{ID: "instance-2", SubflowID: "subflow-1"})
	registry.RegisterInstance(&SubflowInstance{ID: "instance-3", SubflowID: "subflow-2"})

	instances := registry.GetInstancesBySubflow("subflow-1")
	assert.Len(t, instances, 2)
}

func TestExecutor_Execute(t *testing.T) {
	registry := NewRegistry()
	mockExec := &mockNodeExecutor{}
	executor := NewExecutor(registry, mockExec)

	// Create subflow definition
	def := &SubflowDefinition{
		ID:   "subflow-1",
		Name: "Test Subflow",
		InputPorts: []PortDefinition{
			{Type: "input", Index: 0},
		},
		OutputPorts: []PortDefinition{
			{Type: "output", Index: 0},
		},
		Nodes: []NodeDefinition{
			{ID: "node-1", Type: "function", Config: map[string]any{"code": "return msg"}},
		},
		Connections: []ConnectionDefinition{
			{Source: "port-input-0", Target: "node-1", SourcePort: 0, TargetPort: 0},
			{Source: "node-1", Target: "port-output-0", SourcePort: 0, TargetPort: 0},
		},
	}

	registry.RegisterDefinition(def)

	// Create instance
	instance := def.CreateInstance("instance-1", "flow-1", 0, 0)
	registry.RegisterInstance(instance)

	// Execute
	ctx := context.Background()
	msg := CreateMessage(map[string]any{"value": 42}, "test/topic")

	outputs, err := executor.Execute(ctx, "instance-1", 0, msg)
	require.NoError(t, err)
	assert.NotEmpty(t, outputs)

	// Verify node was executed
	assert.Len(t, mockExec.executions, 1)
	assert.Equal(t, "node-1", mockExec.executions[0].nodeID)
}

func TestExecutor_StopInstance(t *testing.T) {
	registry := NewRegistry()
	mockExec := &mockNodeExecutor{}
	executor := NewExecutor(registry, mockExec)

	def := &SubflowDefinition{
		ID:   "subflow-1",
		Name: "Test Subflow",
		InputPorts: []PortDefinition{
			{Type: "input", Index: 0},
		},
	}

	registry.RegisterDefinition(def)
	instance := def.CreateInstance("instance-1", "flow-1", 0, 0)
	registry.RegisterInstance(instance)

	// Start execution
	ctx := context.Background()
	msg := CreateMessage("test", "")
	executor.Execute(ctx, "instance-1", 0, msg)

	// Stop instance
	err := executor.StopInstance("instance-1")
	assert.NoError(t, err)

	// Should fail to stop again
	err = executor.StopInstance("instance-1")
	assert.Error(t, err)
}

func TestLibrary_ExportImport(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry()
	library := NewLibrary(tmpDir, registry)

	def := &SubflowDefinition{
		ID:          "subflow-1",
		Name:        "Test Subflow",
		Description: "Test description",
		Category:    "test",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	registry.RegisterDefinition(def)

	// Export
	outputPath := filepath.Join(tmpDir, "export", "test.json")
	err := library.Export("subflow-1", outputPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(outputPath)
	assert.NoError(t, err)

	// Clear registry
	registry.Clear()

	// Import
	imported, err := library.Import(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "Test Subflow", imported.Name)
	assert.Equal(t, "Test description", imported.Description)
}

func TestLibrary_SaveToLibrary(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry()
	library := NewLibrary(tmpDir, registry)

	def := &SubflowDefinition{
		ID:       "subflow-1",
		Name:     "My Subflow",
		Category: "utilities",
	}

	registry.RegisterDefinition(def)

	// Save to library
	err := library.SaveToLibrary("subflow-1")
	require.NoError(t, err)

	// Verify file exists in correct location
	expectedPath := filepath.Join(tmpDir, "utilities", "My Subflow.json")
	_, err = os.Stat(expectedPath)
	assert.NoError(t, err)
}

func TestLibrary_LoadFromLibrary(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry()
	library := NewLibrary(tmpDir, registry)

	def := &SubflowDefinition{
		ID:       "subflow-1",
		Name:     "Test Subflow",
		Category: "test",
	}

	registry.RegisterDefinition(def)
	library.SaveToLibrary("subflow-1")

	// Clear registry
	registry.Clear()

	// Load from library
	loaded, err := library.LoadFromLibrary("test", "Test Subflow")
	require.NoError(t, err)
	assert.Equal(t, "Test Subflow", loaded.Name)
}

func TestLibrary_ListLibrary(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry()
	library := NewLibrary(tmpDir, registry)

	// Create multiple subflows
	for i := 1; i <= 3; i++ {
		def := &SubflowDefinition{
			ID:       string(rune('a' + i)),
			Name:     string(rune('A' + i)),
			Category: "test",
		}
		registry.RegisterDefinition(def)
		library.SaveToLibrary(def.ID)
	}

	entries, err := library.ListLibrary()
	require.NoError(t, err)
	assert.Len(t, entries, 3)
}

func TestLibrary_ExportPackage(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry()
	library := NewLibrary(tmpDir, registry)

	// Register multiple subflows
	def1 := &SubflowDefinition{ID: "subflow-1", Name: "Subflow 1"}
	def2 := &SubflowDefinition{ID: "subflow-2", Name: "Subflow 2"}

	registry.RegisterDefinition(def1)
	registry.RegisterDefinition(def2)

	// Export package
	packagePath := filepath.Join(tmpDir, "package.json")
	err := library.ExportPackage([]string{"subflow-1", "subflow-2"}, packagePath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(packagePath)
	assert.NoError(t, err)
}

func TestLibrary_ImportPackage(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry()
	library := NewLibrary(tmpDir, registry)

	def1 := &SubflowDefinition{ID: "subflow-1", Name: "Subflow 1"}
	def2 := &SubflowDefinition{ID: "subflow-2", Name: "Subflow 2"}

	registry.RegisterDefinition(def1)
	registry.RegisterDefinition(def2)

	packagePath := filepath.Join(tmpDir, "package.json")
	library.ExportPackage([]string{"subflow-1", "subflow-2"}, packagePath)

	// Clear registry
	registry.Clear()

	// Import package
	imported, err := library.ImportPackage(packagePath)
	require.NoError(t, err)
	assert.Len(t, imported, 2)

	// Verify subflows are registered
	_, err = registry.GetDefinition("subflow-1")
	assert.NoError(t, err)
	_, err = registry.GetDefinition("subflow-2")
	assert.NoError(t, err)
}

func TestLibrary_Clone(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry()
	library := NewLibrary(tmpDir, registry)

	original := &SubflowDefinition{
		ID:          "subflow-1",
		Name:        "Original",
		Description: "Original description",
		Nodes: []NodeDefinition{
			{ID: "node-1", Type: "function"},
		},
	}

	registry.RegisterDefinition(original)

	// Clone
	clone, err := library.Clone("subflow-1", "subflow-2", "Cloned")
	require.NoError(t, err)

	assert.Equal(t, "subflow-2", clone.ID)
	assert.Equal(t, "Cloned", clone.Name)
	assert.Equal(t, "Original description", clone.Description)
	assert.Len(t, clone.Nodes, 1)

	// Verify both exist in registry
	_, err = registry.GetDefinition("subflow-1")
	assert.NoError(t, err)
	_, err = registry.GetDefinition("subflow-2")
	assert.NoError(t, err)
}

func TestCreateMessage(t *testing.T) {
	msg := CreateMessage(map[string]any{"value": 42}, "test/topic")

	assert.NotNil(t, msg)
	assert.Equal(t, "test/topic", msg.Topic)
	assert.NotNil(t, msg.Metadata)
	assert.Contains(t, msg.Metadata, "_msgid")
	assert.NotNil(t, msg.Context.Variables)
}

func TestCloneMessage(t *testing.T) {
	original := CreateMessage("test payload", "test/topic")
	original.Metadata["custom"] = "value"
	original.Context.Variables["var1"] = "value1"

	clone := CloneMessage(original)

	assert.Equal(t, original.Payload, clone.Payload)
	assert.Equal(t, original.Topic, clone.Topic)
	assert.NotEqual(t, original.Metadata["_msgid"], clone.Metadata["_msgid"])
	assert.Equal(t, "value", clone.Metadata["custom"])
	assert.Equal(t, "value1", clone.Context.Variables["var1"])

	// Modify clone shouldn't affect original
	clone.Context.Variables["var2"] = "value2"
	assert.NotContains(t, original.Context.Variables, "var2")
}

func TestSubflowDefinition_Clone(t *testing.T) {
	original := &SubflowDefinition{
		ID:          "subflow-1",
		Name:        "Original",
		Description: "Test",
		Nodes: []NodeDefinition{
			{ID: "node-1", Type: "function"},
		},
		Properties: []PropertyDefinition{
			{Name: "prop1", Type: "string"},
		},
	}

	clone, err := original.Clone()
	require.NoError(t, err)

	assert.Equal(t, original.ID, clone.ID)
	assert.Equal(t, original.Name, clone.Name)
	assert.Len(t, clone.Nodes, 1)
	assert.Len(t, clone.Properties, 1)

	// Modify clone shouldn't affect original
	clone.Name = "Modified"
	assert.Equal(t, "Original", original.Name)
}
