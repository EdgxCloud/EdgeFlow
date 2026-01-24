package subflow

import (
	"encoding/json"
	"fmt"
	"time"
)

// SubflowDefinition represents a reusable subflow template
type SubflowDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Category    string                 `json:"category,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Color       string                 `json:"color,omitempty"`
	Info        string                 `json:"info,omitempty"`
	InputPorts  []PortDefinition       `json:"in"`
	OutputPorts []PortDefinition       `json:"out"`
	Nodes       []NodeDefinition       `json:"nodes"`
	Connections []ConnectionDefinition `json:"connections"`
	Properties  []PropertyDefinition   `json:"properties"`
	Env         []EnvVar               `json:"env,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Version     string                 `json:"version,omitempty"`
	Author      string                 `json:"author,omitempty"`
	License     string                 `json:"license,omitempty"`
}

// PortDefinition defines an input or output port
type PortDefinition struct {
	Type   string            `json:"type"` // "input" or "output"
	Name   string            `json:"name,omitempty"`
	Label  string            `json:"label,omitempty"`
	Index  int               `json:"index"`
	Wires  []WireDestination `json:"wires,omitempty"` // For output ports
	Config map[string]any    `json:"config,omitempty"`
}

// WireDestination represents a connection from a port to a node
type WireDestination struct {
	NodeID string `json:"nodeId"`
	Port   int    `json:"port"`
}

// NodeDefinition represents a node within a subflow
type NodeDefinition struct {
	ID     string         `json:"id"`
	Type   string         `json:"type"`
	Name   string         `json:"name,omitempty"`
	X      float64        `json:"x"`
	Y      float64        `json:"y"`
	Z      string         `json:"z"` // Subflow ID
	Config map[string]any `json:"config"`
	Wires  [][]string     `json:"wires,omitempty"`
}

// ConnectionDefinition represents a wire connection within a subflow
type ConnectionDefinition struct {
	Source      string `json:"source"`
	SourcePort  int    `json:"sourcePort"`
	Target      string `json:"target"`
	TargetPort  int    `json:"targetPort"`
	Passthrough bool   `json:"passthrough,omitempty"` // For subflow input/output connections
}

// PropertyDefinition defines a configurable property
type PropertyDefinition struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // string, number, boolean, json, etc.
	Label        string `json:"label,omitempty"`
	DefaultValue any    `json:"default,omitempty"`
	Required     bool   `json:"required,omitempty"`
	Description  string `json:"description,omitempty"`
	Options      []any  `json:"options,omitempty"` // For select/enum types
}

// EnvVar represents an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
	UI    struct {
		Label string `json:"label,omitempty"`
		Type  string `json:"type,omitempty"` // input, select, etc.
	} `json:"ui,omitempty"`
}

// SubflowInstance represents an instance of a subflow in a flow
type SubflowInstance struct {
	ID             string         `json:"id"`
	Type           string         `json:"type"` // Always "subflow:<subflowId>"
	SubflowID      string         `json:"subflowId"`
	Name           string         `json:"name,omitempty"`
	X              float64        `json:"x"`
	Y              float64        `json:"y"`
	Z              string         `json:"z"` // Flow ID
	Config         map[string]any `json:"config,omitempty"`
	Env            map[string]any `json:"env,omitempty"`
	Wires          [][]string     `json:"wires,omitempty"`
	RuntimeContext *RuntimeContext `json:"-"` // Runtime execution context
}

// RuntimeContext holds runtime state for a subflow instance
type RuntimeContext struct {
	InstanceID   string
	SubflowID    string
	NodeStates   map[string]any
	FlowContext  map[string]any
	GlobalContext map[string]any
	ActiveNodes  map[string]bool
}

// Validate validates a subflow definition
func (sd *SubflowDefinition) Validate() error {
	if sd.ID == "" {
		return fmt.Errorf("subflow ID is required")
	}
	if sd.Name == "" {
		return fmt.Errorf("subflow name is required")
	}

	// Validate input ports
	inputIndices := make(map[int]bool)
	for _, port := range sd.InputPorts {
		if port.Type != "input" {
			return fmt.Errorf("invalid port type in InputPorts: %s", port.Type)
		}
		if inputIndices[port.Index] {
			return fmt.Errorf("duplicate input port index: %d", port.Index)
		}
		inputIndices[port.Index] = true
	}

	// Validate output ports
	outputIndices := make(map[int]bool)
	for _, port := range sd.OutputPorts {
		if port.Type != "output" {
			return fmt.Errorf("invalid port type in OutputPorts: %s", port.Type)
		}
		if outputIndices[port.Index] {
			return fmt.Errorf("duplicate output port index: %d", port.Index)
		}
		outputIndices[port.Index] = true
	}

	// Validate nodes
	nodeIDs := make(map[string]bool)
	for _, node := range sd.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node ID is required")
		}
		if node.Type == "" {
			return fmt.Errorf("node type is required for node %s", node.ID)
		}
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true
	}

	// Validate connections reference existing nodes
	for i, conn := range sd.Connections {
		if !nodeIDs[conn.Source] && !isPortReference(conn.Source) {
			return fmt.Errorf("connection %d: invalid source node: %s", i, conn.Source)
		}
		if !nodeIDs[conn.Target] && !isPortReference(conn.Target) {
			return fmt.Errorf("connection %d: invalid target node: %s", i, conn.Target)
		}
	}

	// Validate properties
	propNames := make(map[string]bool)
	for _, prop := range sd.Properties {
		if prop.Name == "" {
			return fmt.Errorf("property name is required")
		}
		if propNames[prop.Name] {
			return fmt.Errorf("duplicate property name: %s", prop.Name)
		}
		propNames[prop.Name] = true
	}

	return nil
}

// isPortReference checks if a node ID is a subflow port reference
func isPortReference(nodeID string) bool {
	return len(nodeID) > 8 && (nodeID[:8] == "subflow-" || nodeID[:5] == "port-")
}

// Clone creates a deep copy of the subflow definition
func (sd *SubflowDefinition) Clone() (*SubflowDefinition, error) {
	data, err := json.Marshal(sd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subflow: %w", err)
	}

	var clone SubflowDefinition
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subflow: %w", err)
	}

	return &clone, nil
}

// GetInputPort retrieves an input port by index
func (sd *SubflowDefinition) GetInputPort(index int) *PortDefinition {
	for i := range sd.InputPorts {
		if sd.InputPorts[i].Index == index {
			return &sd.InputPorts[i]
		}
	}
	return nil
}

// GetOutputPort retrieves an output port by index
func (sd *SubflowDefinition) GetOutputPort(index int) *PortDefinition {
	for i := range sd.OutputPorts {
		if sd.OutputPorts[i].Index == index {
			return &sd.OutputPorts[i]
		}
	}
	return nil
}

// GetNode retrieves a node by ID
func (sd *SubflowDefinition) GetNode(nodeID string) *NodeDefinition {
	for i := range sd.Nodes {
		if sd.Nodes[i].ID == nodeID {
			return &sd.Nodes[i]
		}
	}
	return nil
}

// ToJSON serializes the subflow definition to JSON
func (sd *SubflowDefinition) ToJSON() ([]byte, error) {
	return json.MarshalIndent(sd, "", "  ")
}

// FromJSON deserializes a subflow definition from JSON
func FromJSON(data []byte) (*SubflowDefinition, error) {
	var sd SubflowDefinition
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subflow: %w", err)
	}

	if err := sd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid subflow: %w", err)
	}

	return &sd, nil
}

// CreateInstance creates a new subflow instance from this definition
func (sd *SubflowDefinition) CreateInstance(instanceID, flowID string, x, y float64) *SubflowInstance {
	return &SubflowInstance{
		ID:        instanceID,
		Type:      fmt.Sprintf("subflow:%s", sd.ID),
		SubflowID: sd.ID,
		Name:      sd.Name,
		X:         x,
		Y:         y,
		Z:         flowID,
		Config:    make(map[string]any),
		Env:       make(map[string]any),
		Wires:     make([][]string, len(sd.OutputPorts)),
		RuntimeContext: &RuntimeContext{
			InstanceID:    instanceID,
			SubflowID:     sd.ID,
			NodeStates:    make(map[string]any),
			FlowContext:   make(map[string]any),
			GlobalContext: make(map[string]any),
			ActiveNodes:   make(map[string]bool),
		},
	}
}

// Validate validates a subflow instance
func (si *SubflowInstance) Validate() error {
	if si.ID == "" {
		return fmt.Errorf("instance ID is required")
	}
	if si.SubflowID == "" {
		return fmt.Errorf("subflow ID is required")
	}
	return nil
}
