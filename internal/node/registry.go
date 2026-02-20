package node

import (
	"fmt"
	"sync"
)

// NodeFactory is a function that creates a new node executor instance
type NodeFactory func() Executor

// NodeInfo contains metadata about a node type
type NodeInfo struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Category    NodeType               `json:"category"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Color       string                 `json:"color"`
	Properties  []PropertySchema       `json:"properties"`
	Inputs      []PortSchema           `json:"inputs"`
	Outputs     []PortSchema           `json:"outputs"`
	Factory     NodeFactory            `json:"-"`
}

// FloatPtr returns a pointer to a float64 value, used for optional Min/Max/Step fields
func FloatPtr(v float64) *float64 {
	return &v
}

// PropertySchema defines a configurable property of a node
type PropertySchema struct {
	Name        string      `json:"name"`
	Label       string      `json:"label"`
	Type        string      `json:"type"` // string, number, boolean, select, password, code, payload, etc.
	Default     interface{} `json:"default"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Options     []string    `json:"options,omitempty"`     // For select type
	Placeholder string      `json:"placeholder,omitempty"` // Placeholder text for input fields
	Min         *float64    `json:"min,omitempty"`         // Minimum value for number fields
	Max         *float64    `json:"max,omitempty"`         // Maximum value for number fields
	Step        *float64    `json:"step,omitempty"`        // Step increment for number fields
	Group       string      `json:"group,omitempty"`       // Group name for organizing properties in UI
	Validation  string      `json:"validation,omitempty"`  // Regex validation pattern
}

// PortSchema defines an input or output port
type PortSchema struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"` // any, string, number, object, etc.
	Description string `json:"description"`
}

// Registry manages all available node types
type Registry struct {
	nodes map[string]*NodeInfo
	mu    sync.RWMutex
}

// NewRegistry creates a new node registry
func NewRegistry() *Registry {
	return &Registry{
		nodes: make(map[string]*NodeInfo),
	}
}

// Register registers a new node type
func (r *Registry) Register(info *NodeInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if info.Type == "" {
		return fmt.Errorf("node type cannot be empty")
	}

	if info.Factory == nil {
		return fmt.Errorf("node factory cannot be nil")
	}

	if _, exists := r.nodes[info.Type]; exists {
		return fmt.Errorf("node type %s already registered", info.Type)
	}

	r.nodes[info.Type] = info
	return nil
}

// Get retrieves node info by type
func (r *Registry) Get(nodeType string) (*NodeInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.nodes[nodeType]
	if !exists {
		return nil, fmt.Errorf("node type %s not found", nodeType)
	}

	return info, nil
}

// List returns all registered node types
func (r *Registry) List() []*NodeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*NodeInfo, 0, len(r.nodes))
	for _, info := range r.nodes {
		list = append(list, info)
	}

	return list
}

// ListByCategory returns nodes filtered by category
func (r *Registry) ListByCategory(category NodeType) []*NodeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*NodeInfo, 0)
	for _, info := range r.nodes {
		if info.Category == category {
			list = append(list, info)
		}
	}

	return list
}

// CreateNode creates a new node instance
func (r *Registry) CreateNode(nodeType, name string) (*Node, error) {
	info, err := r.Get(nodeType)
	if err != nil {
		return nil, err
	}

	executor := info.Factory()
	node := NewNode(nodeType, name, info.Category, executor)

	return node, nil
}

// Unregister removes a node type from the registry
func (r *Registry) Unregister(nodeType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[nodeType]; !exists {
		return fmt.Errorf("node type %s not found", nodeType)
	}

	delete(r.nodes, nodeType)
	return nil
}

// Count returns the total number of registered node types
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.nodes)
}

// Global registry instance
var globalRegistry = NewRegistry()

// GetGlobalRegistry returns the global node registry
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// RegisterNode is a convenience function to register a node with the global registry
func RegisterNode(info *NodeInfo) error {
	return globalRegistry.Register(info)
}
