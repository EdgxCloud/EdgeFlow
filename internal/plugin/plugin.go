package plugin

import (
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// Plugin main interface for all plugins
type Plugin interface {
	// Plugin information
	Name() string
	Version() string
	Description() string
	Author() string
	Category() Category

	// Lifecycle management
	Load() error
	Unload() error
	IsLoaded() bool

	// Provided nodes
	Nodes() []NodeDefinition

	// System requirements
	RequiredMemory() uint64  // bytes
	RequiredDisk() uint64    // bytes
	Dependencies() []string  // names of dependent plugins

	// Configuration
	DefaultConfig() map[string]interface{}
	ValidateConfig(config map[string]interface{}) error
}

// Category plugin category
type Category string

const (
	CategoryCore       Category = "core"
	CategoryNetwork    Category = "network"
	CategoryGPIO       Category = "gpio"
	CategoryDatabase   Category = "database"
	CategoryMessaging  Category = "messaging"
	CategoryAI         Category = "ai"
	CategoryIndustrial Category = "industrial"
	CategoryAdvanced   Category = "advanced"
	CategoryUI         Category = "ui"
)

// NodeDefinition definition of a node provided by a plugin
type NodeDefinition struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Color       string                 `json:"color"`
	Inputs      int                    `json:"inputs"`
	Outputs     int                    `json:"outputs"`
	Factory     node.NodeFactory       `json:"-"`
	Config      map[string]interface{} `json:"config"`
}

// Metadata plugin metadata
type Metadata struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author"`
	Category     Category               `json:"category"`
	License      string                 `json:"license"`
	Homepage     string                 `json:"homepage"`
	Repository   string                 `json:"repository"`
	Keywords     []string               `json:"keywords"`
	MinEdgeFlow  string                 `json:"min_edgeflow"`
	Config       map[string]interface{} `json:"config"`
	Dependencies []string               `json:"dependencies"`
}

// Status plugin status
type Status string

const (
	StatusNotLoaded Status = "not_loaded"
	StatusLoading   Status = "loading"
	StatusLoaded    Status = "loaded"
	StatusUnloading Status = "unloading"
	StatusError     Status = "error"
)

// BasePlugin base implementation for plugins
type BasePlugin struct {
	metadata     Metadata
	status       Status
	loadedAt     time.Time
	unloadedAt   time.Time
	errorMessage string
	config       map[string]interface{}
	mu           sync.RWMutex
}

// NewBasePlugin create base plugin
func NewBasePlugin(metadata Metadata) *BasePlugin {
	return &BasePlugin{
		metadata: metadata,
		status:   StatusNotLoaded,
		config:   make(map[string]interface{}),
	}
}

// Name plugin name
func (p *BasePlugin) Name() string {
	return p.metadata.Name
}

// Version plugin version
func (p *BasePlugin) Version() string {
	return p.metadata.Version
}

// Description plugin description
func (p *BasePlugin) Description() string {
	return p.metadata.Description
}

// Author plugin author
func (p *BasePlugin) Author() string {
	return p.metadata.Author
}

// Category plugin category
func (p *BasePlugin) Category() Category {
	return p.metadata.Category
}

// GetMetadata get all metadata
func (p *BasePlugin) GetMetadata() Metadata {
	return p.metadata
}

// SetStatus set status
func (p *BasePlugin) SetStatus(status Status) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = status

	if status == StatusLoaded {
		p.loadedAt = time.Now()
	} else if status == StatusNotLoaded {
		p.unloadedAt = time.Now()
	}
}

// GetStatus get status
func (p *BasePlugin) GetStatus() Status {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

// IsLoaded check if loaded
func (p *BasePlugin) IsLoaded() bool {
	return p.GetStatus() == StatusLoaded
}

// SetError set error
func (p *BasePlugin) SetError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = StatusError
	if err != nil {
		p.errorMessage = err.Error()
	}
}

// GetError get error
func (p *BasePlugin) GetError() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.errorMessage
}

// SetConfig set configuration
func (p *BasePlugin) SetConfig(config map[string]interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = config
}

// GetConfig get configuration
func (p *BasePlugin) GetConfig() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// LoadedAt load time
func (p *BasePlugin) LoadedAt() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.loadedAt
}

// UnloadedAt unload time
func (p *BasePlugin) UnloadedAt() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.unloadedAt
}

// DefaultConfig default configuration
func (p *BasePlugin) DefaultConfig() map[string]interface{} {
	return p.metadata.Config
}

// ValidateConfig validate configuration
func (p *BasePlugin) ValidateConfig(config map[string]interface{}) error {
	// Base implementation - plugins can override
	return nil
}

// Dependencies list of dependent plugins
func (p *BasePlugin) Dependencies() []string {
	return p.metadata.Dependencies
}

// Load load plugin (default implementation - plugins can override)
func (p *BasePlugin) Load() error {
	return nil
}

// Unload unload plugin from memory (default implementation - plugins can override)
func (p *BasePlugin) Unload() error {
	return nil
}

// Nodes list of provided nodes (default implementation - plugins should override)
func (p *BasePlugin) Nodes() []NodeDefinition {
	return []NodeDefinition{}
}

// RequiredMemory required memory (default implementation)
func (p *BasePlugin) RequiredMemory() uint64 {
	return 0
}

// RequiredDisk required disk space (default implementation)
func (p *BasePlugin) RequiredDisk() uint64 {
	return 0
}

// PluginInfo complete plugin information for API
type PluginInfo struct {
	Name             string                 `json:"name"`
	Version          string                 `json:"version"`
	Description      string                 `json:"description"`
	Author           string                 `json:"author"`
	Category         Category               `json:"category"`
	Status           Status                 `json:"status"`
	LoadedAt         *time.Time             `json:"loaded_at,omitempty"`
	UnloadedAt       *time.Time             `json:"unloaded_at,omitempty"`
	Error            string                 `json:"error,omitempty"`
	RequiredMemoryMB uint64                 `json:"required_memory_mb"`
	RequiredDiskMB   uint64                 `json:"required_disk_mb"`
	Dependencies     []string               `json:"dependencies"`
	Nodes            []NodeDefinition       `json:"nodes"`
	Config           map[string]interface{} `json:"config"`
	Compatible       bool                   `json:"compatible"`
	CompatibleReason string                 `json:"compatible_reason,omitempty"`
}

// ToPluginInfo convert to PluginInfo
func ToPluginInfo(p Plugin, compatible bool, reason string) PluginInfo {
	info := PluginInfo{
		Name:             p.Name(),
		Version:          p.Version(),
		Description:      p.Description(),
		Author:           p.Author(),
		Category:         p.Category(),
		RequiredMemoryMB: p.RequiredMemory() / 1024 / 1024,
		RequiredDiskMB:   p.RequiredDisk() / 1024 / 1024,
		Dependencies:     p.Dependencies(),
		Nodes:            p.Nodes(),
		Config:           p.DefaultConfig(),
		Compatible:       compatible,
		CompatibleReason: reason,
	}

	// Status information if it's a BasePlugin
	if bp, ok := p.(*BasePlugin); ok {
		info.Status = bp.GetStatus()
		info.Error = bp.GetError()

		if !bp.LoadedAt().IsZero() {
			t := bp.LoadedAt()
			info.LoadedAt = &t
		}
		if !bp.UnloadedAt().IsZero() {
			t := bp.UnloadedAt()
			info.UnloadedAt = &t
		}
	} else {
		if p.IsLoaded() {
			info.Status = StatusLoaded
		} else {
			info.Status = StatusNotLoaded
		}
	}

	return info
}

// Registry global plugin registry
type Registry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// globalRegistry global registry
var globalRegistry = &Registry{
	plugins: make(map[string]Plugin),
}

// GetRegistry get global registry
func GetRegistry() *Registry {
	return globalRegistry
}

// Register register plugin in registry
func (r *Registry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

// Unregister remove plugin from registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	delete(r.plugins, name)
	return nil
}

// Get get plugin
func (r *Registry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	return plugin, nil
}

// List list all plugins
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}

	return plugins
}

// ListByCategory list plugins by category
func (r *Registry) ListByCategory(category Category) []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0)
	for _, p := range r.plugins {
		if p.Category() == category {
			plugins = append(plugins, p)
		}
	}

	return plugins
}

// ListLoaded list loaded plugins
func (r *Registry) ListLoaded() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0)
	for _, p := range r.plugins {
		if p.IsLoaded() {
			plugins = append(plugins, p)
		}
	}

	return plugins
}

// Count number of plugins
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.plugins)
}

// CountLoaded number of loaded plugins
func (r *Registry) CountLoaded() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, p := range r.plugins {
		if p.IsLoaded() {
			count++
		}
	}
	return count
}
