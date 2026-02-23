// Package database provides the database module for EdgeFlow
// This module provides MySQL, PostgreSQL, MongoDB, and Redis nodes
package database

import (
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/EdgxCloud/EdgeFlow/internal/plugin"
	"github.com/EdgxCloud/EdgeFlow/pkg/nodes/database"
)

// DatabaseModule is the database module that provides database interaction nodes
type DatabaseModule struct {
	*plugin.BasePlugin
	loaded bool
}

// NewDatabaseModule creates a new database module
func NewDatabaseModule() *DatabaseModule {
	metadata := plugin.Metadata{
		Name:        "database",
		Version:     "1.0.0",
		Description: "Database nodes - MySQL, PostgreSQL, MongoDB, Redis",
		Author:      "EdgeFlow Team",
		Category:    plugin.CategoryDatabase,
		License:     "MIT",
		Keywords:    []string{"database", "mysql", "postgresql", "mongodb", "redis", "sql", "nosql"},
		MinEdgeFlow: "0.1.0",
		Config: map[string]interface{}{
			"connection_timeout": 10,
			"max_connections":    10,
			"idle_timeout":       300,
		},
	}

	return &DatabaseModule{
		BasePlugin: plugin.NewBasePlugin(metadata),
		loaded:     false,
	}
}

// Load loads the database module
func (m *DatabaseModule) Load() error {
	m.SetStatus(plugin.StatusLoading)

	// Register all database nodes using the centralized RegisterAllNodes function
	registry := node.GetGlobalRegistry()
	database.RegisterAllNodes(registry)

	m.loaded = true
	m.SetStatus(plugin.StatusLoaded)
	return nil
}

// Unload unloads the database module
func (m *DatabaseModule) Unload() error {
	m.SetStatus(plugin.StatusUnloading)
	m.loaded = false
	m.SetStatus(plugin.StatusNotLoaded)
	return nil
}

// IsLoaded returns whether the module is loaded
func (m *DatabaseModule) IsLoaded() bool {
	return m.loaded
}

// Nodes returns the node definitions provided by this module
func (m *DatabaseModule) Nodes() []plugin.NodeDefinition {
	return []plugin.NodeDefinition{
		{
			Type:        "mysql",
			Name:        "MySQL",
			Category:    "database",
			Description: "MySQL database queries and operations",
			Icon:        "database",
			Color:       "#4479A1",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "postgresql",
			Name:        "PostgreSQL",
			Category:    "database",
			Description: "PostgreSQL database queries and operations",
			Icon:        "database",
			Color:       "#336791",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "mongodb",
			Name:        "MongoDB",
			Category:    "database",
			Description: "MongoDB document operations (find, insert, update, delete)",
			Icon:        "leaf",
			Color:       "#47A248",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "redis",
			Name:        "Redis",
			Category:    "database",
			Description: "Redis key-value operations (get, set, del, pub/sub)",
			Icon:        "box",
			Color:       "#DC382D",
			Inputs:      1,
			Outputs:     1,
		},
	}
}

// RequiredMemory returns the memory requirement in bytes
func (m *DatabaseModule) RequiredMemory() uint64 {
	return 40 * 1024 * 1024 // 40 MB
}

// RequiredDisk returns the disk requirement in bytes
func (m *DatabaseModule) RequiredDisk() uint64 {
	return 15 * 1024 * 1024 // 15 MB (includes database drivers)
}

// Dependencies returns the list of required plugins
func (m *DatabaseModule) Dependencies() []string {
	return []string{"core"}
}

// init registers the database module with the global registry
func init() {
	registry := plugin.GetRegistry()
	module := NewDatabaseModule()
	registry.Register(module)
}
