package plugin

import (
	"fmt"
	"sort"
	"sync"

	"github.com/EdgxCloud/EdgeFlow/internal/logger"
	"go.uber.org/zap"
)

// Loader plugin loader with dependency resolution
type Loader struct {
	mu sync.RWMutex
}

// NewLoader create new loader
func NewLoader() *Loader {
	return &Loader{}
}

// DependencyGraph dependency graph
type DependencyGraph struct {
	nodes map[string]*GraphNode
	edges map[string][]string // from -> []to
}

// GraphNode graph node
type GraphNode struct {
	Name         string
	Plugin       Plugin
	Dependencies []string
	Visited      bool
	InStack      bool
}

// BuildDependencyGraph build dependency graph
func (l *Loader) BuildDependencyGraph(plugins []Plugin) *DependencyGraph {
	graph := &DependencyGraph{
		nodes: make(map[string]*GraphNode),
		edges: make(map[string][]string),
	}

	// Add nodes
	for _, p := range plugins {
		node := &GraphNode{
			Name:         p.Name(),
			Plugin:       p,
			Dependencies: p.Dependencies(),
		}
		graph.nodes[p.Name()] = node
	}

	// Add edges
	for _, node := range graph.nodes {
		for _, dep := range node.Dependencies {
			graph.edges[node.Name] = append(graph.edges[node.Name], dep)
		}
	}

	return graph
}

// TopologicalSort topological sort for load order
func (l *Loader) TopologicalSort(graph *DependencyGraph) ([]string, error) {
	result := make([]string, 0, len(graph.nodes))
	visited := make(map[string]bool)

	var visit func(name string) error
	visit = func(name string) error {
		if visited[name] {
			return nil
		}

		node, exists := graph.nodes[name]
		if !exists {
			return fmt.Errorf("plugin '%s' not found in graph", name)
		}

		// Check for circular dependencies
		if node.InStack {
			return fmt.Errorf("circular dependency detected: %s", name)
		}

		node.InStack = true

		// Visit dependencies (in reverse order)
		for _, dep := range node.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		node.InStack = false
		visited[name] = true
		node.Visited = true

		// Add to result (dependencies first)
		result = append(result, name)
		return nil
	}

	// Visit all nodes
	names := make([]string, 0, len(graph.nodes))
	for name := range graph.nodes {
		names = append(names, name)
	}
	sort.Strings(names) // Fixed order for reproducibility

	for _, name := range names {
		if !visited[name] {
			if err := visit(name); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// ValidateDependencies validate dependencies
func (l *Loader) ValidateDependencies(plugins []Plugin) error {
	pluginMap := make(map[string]Plugin)
	for _, p := range plugins {
		pluginMap[p.Name()] = p
	}

	for _, p := range plugins {
		for _, dep := range p.Dependencies() {
			if _, exists := pluginMap[dep]; !exists {
				return fmt.Errorf("plugin '%s' requires '%s' which is not available", p.Name(), dep)
			}
		}
	}

	return nil
}

// ResolveLoadOrder determine load order
func (l *Loader) ResolveLoadOrder(plugins []Plugin) ([]string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Validate dependencies
	if err := l.ValidateDependencies(plugins); err != nil {
		return nil, fmt.Errorf("dependency validation failed: %w", err)
	}

	// Build graph
	graph := l.BuildDependencyGraph(plugins)

	// Topological sort
	order, err := l.TopologicalSort(graph)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve load order: %w", err)
	}

	logger.Debug("Plugin load order resolved", zap.Strings("order", order))
	return order, nil
}

// GetMissingDependencies get missing dependencies
func (l *Loader) GetMissingDependencies(plugins []Plugin) map[string][]string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	pluginMap := make(map[string]bool)
	for _, p := range plugins {
		pluginMap[p.Name()] = true
	}

	missing := make(map[string][]string)
	for _, p := range plugins {
		for _, dep := range p.Dependencies() {
			if !pluginMap[dep] {
				missing[p.Name()] = append(missing[p.Name()], dep)
			}
		}
	}

	return missing
}

// GetDependents get plugins that depend on a given plugin
func (l *Loader) GetDependents(pluginName string, plugins []Plugin) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	dependents := make([]string, 0)
	for _, p := range plugins {
		for _, dep := range p.Dependencies() {
			if dep == pluginName {
				dependents = append(dependents, p.Name())
				break
			}
		}
	}

	return dependents
}

// CanUnload check if plugin can be unloaded
func (l *Loader) CanUnload(pluginName string, loadedPlugins []Plugin) (bool, string) {
	dependents := l.GetDependents(pluginName, loadedPlugins)
	if len(dependents) > 0 {
		return false, fmt.Sprintf("required by: %v", dependents)
	}
	return true, ""
}

// OptimizeLoadOrder optimize load order based on priority
func (l *Loader) OptimizeLoadOrder(order []string, priorities map[Category]int) []string {
	// TODO: Optimize based on category priorities
	// For now, returns the same topological order
	return order
}

// LoadOrderInfo load order information
type LoadOrderInfo struct {
	Order              []string            `json:"order"`
	TotalPlugins       int                 `json:"total_plugins"`
	MissingDependencies map[string][]string `json:"missing_dependencies,omitempty"`
	CircularDependency bool                `json:"circular_dependency"`
	Error              string              `json:"error,omitempty"`
}

// AnalyzeLoadOrder analyze load order
func (l *Loader) AnalyzeLoadOrder(plugins []Plugin) LoadOrderInfo {
	info := LoadOrderInfo{
		TotalPlugins: len(plugins),
	}

	// Check for missing dependencies
	missing := l.GetMissingDependencies(plugins)
	if len(missing) > 0 {
		info.MissingDependencies = missing
		info.Error = "missing dependencies detected"
		return info
	}

	// Determine order
	order, err := l.ResolveLoadOrder(plugins)
	if err != nil {
		info.Error = err.Error()
		// Check for circular dependency
		if contains(err.Error(), "circular dependency") {
			info.CircularDependency = true
		}
		return info
	}

	info.Order = order
	return info
}

// contains check for substring existence
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
