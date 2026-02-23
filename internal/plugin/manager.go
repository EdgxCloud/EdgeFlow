package plugin

import (
	"context"
	"fmt"
	"sync"

	"github.com/EdgxCloud/EdgeFlow/internal/logger"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/EdgxCloud/EdgeFlow/internal/resources"
	"go.uber.org/zap"
)

// Manager plugin manager
type Manager struct {
	registry        *Registry
	nodeRegistry    *node.Registry
	resourceMonitor *resources.Monitor
	loader          *Loader

	loadOrder       []string          // load order
	enabledPlugins  map[string]bool   // enabled plugins

	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewManager create plugin manager
func NewManager(nodeRegistry *node.Registry, resourceMonitor *resources.Monitor) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		registry:        GetRegistry(),
		nodeRegistry:    nodeRegistry,
		resourceMonitor: resourceMonitor,
		loader:          NewLoader(),
		loadOrder:       make([]string, 0),
		enabledPlugins:  make(map[string]bool),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// LoadPlugin load plugin
func (m *Manager) LoadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get plugin from registry
	plugin, err := m.registry.Get(name)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// Check that it's not already loaded
	if plugin.IsLoaded() {
		return fmt.Errorf("plugin '%s' already loaded", name)
	}

	// Check system resources
	if m.resourceMonitor != nil {
		canLoad, reason := m.resourceMonitor.CanLoadModule(name, plugin.RequiredMemory())
		if !canLoad {
			return fmt.Errorf("cannot load plugin: %s", reason)
		}
	}

	// Check and load dependencies
	for _, dep := range plugin.Dependencies() {
		depPlugin, err := m.registry.Get(dep)
		if err != nil {
			return fmt.Errorf("dependency '%s' not found: %w", dep, err)
		}

		if !depPlugin.IsLoaded() {
			logger.Info("Loading plugin dependency", zap.String("dependency", dep))
			if err := m.loadPluginInternal(depPlugin); err != nil {
				return fmt.Errorf("failed to load dependency '%s': %w", dep, err)
			}
		}
	}

	// Load plugin
	if err := m.loadPluginInternal(plugin); err != nil {
		return err
	}

	// Add to enabled list
	m.enabledPlugins[name] = true
	m.loadOrder = append(m.loadOrder, name)

	// Register in resource monitor
	if m.resourceMonitor != nil {
		m.resourceMonitor.EnableModule(name)
	}

	logger.Info("Plugin loaded", zap.String("name", name), zap.String("version", plugin.Version()), zap.Int("nodes", len(plugin.Nodes())))
	return nil
}

// loadPluginInternal internal plugin loading
func (m *Manager) loadPluginInternal(plugin Plugin) error {
	// Set status
	if bp, ok := plugin.(*BasePlugin); ok {
		bp.SetStatus(StatusLoading)
	}

	// Load plugin
	if err := plugin.Load(); err != nil {
		if bp, ok := plugin.(*BasePlugin); ok {
			bp.SetError(err)
		}
		return fmt.Errorf("plugin load failed: %w", err)
	}

	// Register nodes in node registry
	for _, nodeDef := range plugin.Nodes() {
		nodeInfo := &node.NodeInfo{
			Type:        nodeDef.Type,
			Name:        nodeDef.Name,
			Category:    node.NodeType(nodeDef.Category),
			Description: nodeDef.Description,
			Icon:        nodeDef.Icon,
			Color:       nodeDef.Color,
			Factory:     nodeDef.Factory,
		}
		if err := m.nodeRegistry.Register(nodeInfo); err != nil {
			logger.Warn("Failed to register plugin node", zap.String("type", nodeDef.Type), zap.Error(err))
		}
	}

	// Set loaded status
	if bp, ok := plugin.(*BasePlugin); ok {
		bp.SetStatus(StatusLoaded)
	}

	return nil
}

// UnloadPlugin unload plugin from memory
func (m *Manager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get plugin
	plugin, err := m.registry.Get(name)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// Check if loaded
	if !plugin.IsLoaded() {
		return fmt.Errorf("plugin '%s' not loaded", name)
	}

	// Check reverse dependencies
	for _, p := range m.registry.List() {
		if !p.IsLoaded() {
			continue
		}

		for _, dep := range p.Dependencies() {
			if dep == name {
				return fmt.Errorf("cannot unload: plugin '%s' depends on it", p.Name())
			}
		}
	}

	// Unload from memory
	if err := m.unloadPluginInternal(plugin); err != nil {
		return err
	}

	// Remove from enabled list
	delete(m.enabledPlugins, name)

	// Remove from load order
	for i, pname := range m.loadOrder {
		if pname == name {
			m.loadOrder = append(m.loadOrder[:i], m.loadOrder[i+1:]...)
			break
		}
	}

	// Disable in resource monitor
	if m.resourceMonitor != nil {
		m.resourceMonitor.DisableModule(name)
	}

	logger.Info("Plugin unloaded", zap.String("name", name))
	return nil
}

// unloadPluginInternal internal unloading from memory
func (m *Manager) unloadPluginInternal(plugin Plugin) error {
	// Set status
	if bp, ok := plugin.(*BasePlugin); ok {
		bp.SetStatus(StatusUnloading)
	}

	// Remove nodes from registry
	for _, nodeDef := range plugin.Nodes() {
		// TODO: Unregister nodes (requires changes to node.Registry)
		_ = nodeDef
	}

	// Unload plugin
	if err := plugin.Unload(); err != nil {
		if bp, ok := plugin.(*BasePlugin); ok {
			bp.SetError(err)
		}
		return fmt.Errorf("plugin unload failed: %w", err)
	}

	// Set status
	if bp, ok := plugin.(*BasePlugin); ok {
		bp.SetStatus(StatusNotLoaded)
	}

	return nil
}

// EnablePlugin enable plugin
func (m *Manager) EnablePlugin(name string) error {
	m.mu.Lock()
	m.enabledPlugins[name] = true
	m.mu.Unlock()

	return m.LoadPlugin(name)
}

// DisablePlugin disable plugin
func (m *Manager) DisablePlugin(name string) error {
	m.mu.Lock()
	m.enabledPlugins[name] = false
	m.mu.Unlock()

	return m.UnloadPlugin(name)
}

// IsEnabled check if plugin is enabled
func (m *Manager) IsEnabled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabledPlugins[name]
}

// LoadAll load all enabled plugins
func (m *Manager) LoadAll() error {
	m.mu.RLock()
	enabled := make([]string, 0, len(m.enabledPlugins))
	for name, isEnabled := range m.enabledPlugins {
		if isEnabled {
			enabled = append(enabled, name)
		}
	}
	m.mu.RUnlock()

	// Load in order
	for _, name := range enabled {
		if err := m.LoadPlugin(name); err != nil {
			logger.Error("Failed to load plugin", zap.String("name", name), zap.Error(err))
			// Continue with other plugins
		}
	}

	return nil
}

// UnloadAll unload all plugins from memory
func (m *Manager) UnloadAll() error {
	m.mu.RLock()
	loadOrder := make([]string, len(m.loadOrder))
	copy(loadOrder, m.loadOrder)
	m.mu.RUnlock()

	// Unload in reverse order
	for i := len(loadOrder) - 1; i >= 0; i-- {
		name := loadOrder[i]
		if err := m.UnloadPlugin(name); err != nil {
			logger.Error("Failed to unload plugin", zap.String("name", name), zap.Error(err))
		}
	}

	return nil
}

// ReloadPlugin reload plugin
func (m *Manager) ReloadPlugin(name string) error {
	// Unload
	if err := m.UnloadPlugin(name); err != nil {
		return err
	}

	// Reload
	return m.LoadPlugin(name)
}

// CheckCompatibility check plugin compatibility with system
func (m *Manager) CheckCompatibility(name string) (bool, string) {
	plugin, err := m.registry.Get(name)
	if err != nil {
		return false, "plugin not found"
	}

	// Check resources
	if m.resourceMonitor != nil {
		canLoad, reason := m.resourceMonitor.CanLoadModule(name, plugin.RequiredMemory())
		if !canLoad {
			return false, reason
		}
	}

	// Check dependencies
	for _, dep := range plugin.Dependencies() {
		_, err := m.registry.Get(dep)
		if err != nil {
			return false, fmt.Sprintf("missing dependency: %s", dep)
		}
	}

	return true, "compatible"
}

// GetPluginInfo get plugin information
func (m *Manager) GetPluginInfo(name string) (PluginInfo, error) {
	plugin, err := m.registry.Get(name)
	if err != nil {
		return PluginInfo{}, err
	}

	compatible, reason := m.CheckCompatibility(name)
	return ToPluginInfo(plugin, compatible, reason), nil
}

// ListPlugins list all plugins with information
func (m *Manager) ListPlugins() []PluginInfo {
	plugins := m.registry.List()
	infos := make([]PluginInfo, 0, len(plugins))

	for _, plugin := range plugins {
		compatible, reason := m.CheckCompatibility(plugin.Name())
		infos = append(infos, ToPluginInfo(plugin, compatible, reason))
	}

	return infos
}

// GetStats plugin statistics
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalNodes := 0
	for _, plugin := range m.registry.ListLoaded() {
		totalNodes += len(plugin.Nodes())
	}

	return map[string]interface{}{
		"total_plugins":      m.registry.Count(),
		"loaded_plugins":     m.registry.CountLoaded(),
		"enabled_plugins":    len(m.enabledPlugins),
		"total_nodes":        totalNodes,
		"load_order":         m.loadOrder,
		"memory_available_mb": m.getAvailableMemoryMB(),
	}
}

// getAvailableMemoryMB get available memory
func (m *Manager) getAvailableMemoryMB() uint64 {
	if m.resourceMonitor == nil {
		return 0
	}

	stats := m.resourceMonitor.GetStats()
	return stats.MemoryAvailable / 1024 / 1024
}

// Shutdown shutdown manager
func (m *Manager) Shutdown() error {
	m.cancel()
	return m.UnloadAll()
}

// SetResourceMonitor set resource monitor
func (m *Manager) SetResourceMonitor(monitor *resources.Monitor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resourceMonitor = monitor
}

// GetLoadOrder get load order
func (m *Manager) GetLoadOrder() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	order := make([]string, len(m.loadOrder))
	copy(order, m.loadOrder)
	return order
}
