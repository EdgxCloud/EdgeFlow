package subflow

import (
	"fmt"
	"sync"
)

// Registry manages subflow definitions and instances
type Registry struct {
	mu          sync.RWMutex
	definitions map[string]*SubflowDefinition
	instances   map[string]*SubflowInstance
}

// NewRegistry creates a new subflow registry
func NewRegistry() *Registry {
	return &Registry{
		definitions: make(map[string]*SubflowDefinition),
		instances:   make(map[string]*SubflowInstance),
	}
}

// RegisterDefinition registers a subflow definition
func (r *Registry) RegisterDefinition(def *SubflowDefinition) error {
	if err := def.Validate(); err != nil {
		return fmt.Errorf("invalid subflow definition: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.definitions[def.ID] = def
	return nil
}

// UnregisterDefinition removes a subflow definition
func (r *Registry) UnregisterDefinition(subflowID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if any instances exist
	for _, instance := range r.instances {
		if instance.SubflowID == subflowID {
			return fmt.Errorf("cannot unregister subflow %s: instances exist", subflowID)
		}
	}

	delete(r.definitions, subflowID)
	return nil
}

// GetDefinition retrieves a subflow definition by ID
func (r *Registry) GetDefinition(subflowID string) (*SubflowDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, ok := r.definitions[subflowID]
	if !ok {
		return nil, fmt.Errorf("subflow not found: %s", subflowID)
	}

	return def, nil
}

// ListDefinitions returns all registered subflow definitions
func (r *Registry) ListDefinitions() []*SubflowDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]*SubflowDefinition, 0, len(r.definitions))
	for _, def := range r.definitions {
		defs = append(defs, def)
	}

	return defs
}

// RegisterInstance registers a subflow instance
func (r *Registry) RegisterInstance(instance *SubflowInstance) error {
	if err := instance.Validate(); err != nil {
		return fmt.Errorf("invalid subflow instance: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify definition exists
	if _, ok := r.definitions[instance.SubflowID]; !ok {
		return fmt.Errorf("subflow definition not found: %s", instance.SubflowID)
	}

	r.instances[instance.ID] = instance
	return nil
}

// UnregisterInstance removes a subflow instance
func (r *Registry) UnregisterInstance(instanceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.instances, instanceID)
}

// GetInstance retrieves a subflow instance by ID
func (r *Registry) GetInstance(instanceID string) (*SubflowInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instance, ok := r.instances[instanceID]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}

	return instance, nil
}

// ListInstances returns all registered subflow instances
func (r *Registry) ListInstances() []*SubflowInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances := make([]*SubflowInstance, 0, len(r.instances))
	for _, instance := range r.instances {
		instances = append(instances, instance)
	}

	return instances
}

// GetInstancesBySubflow returns all instances of a specific subflow
func (r *Registry) GetInstancesBySubflow(subflowID string) []*SubflowInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances := make([]*SubflowInstance, 0)
	for _, instance := range r.instances {
		if instance.SubflowID == subflowID {
			instances = append(instances, instance)
		}
	}

	return instances
}

// UpdateDefinition updates an existing subflow definition
func (r *Registry) UpdateDefinition(def *SubflowDefinition) error {
	if err := def.Validate(); err != nil {
		return fmt.Errorf("invalid subflow definition: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.definitions[def.ID]; !ok {
		return fmt.Errorf("subflow not found: %s", def.ID)
	}

	r.definitions[def.ID] = def
	return nil
}

// Clear removes all definitions and instances
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.definitions = make(map[string]*SubflowDefinition)
	r.instances = make(map[string]*SubflowInstance)
}

// Stats returns registry statistics
func (r *Registry) Stats() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]int{
		"definitions": len(r.definitions),
		"instances":   len(r.instances),
	}
}

// Global registry instance
var globalRegistry = NewRegistry()

// GlobalRegistry returns the global subflow registry
func GlobalRegistry() *Registry {
	return globalRegistry
}
