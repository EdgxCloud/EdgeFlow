package storage

import (
	"fmt"
)

// Storage defines the interface for persisting flows
type Storage interface {
	// Flow operations
	SaveFlow(flow *Flow) error
	GetFlow(id string) (*Flow, error)
	ListFlows() ([]*Flow, error)
	DeleteFlow(id string) error
	UpdateFlow(flow *Flow) error

	// Close closes the storage connection
	Close() error
}

// StorageType defines the type of storage backend
type StorageType string

const (
	StorageTypeSQLite     StorageType = "sqlite"
	StorageTypePostgreSQL StorageType = "postgres"
	StorageTypeFile       StorageType = "file"
)

// Config holds storage configuration
type Config struct {
	Type StorageType
	Path string
	// Additional fields for different storage types
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// New creates a new storage instance based on configuration
func New(config Config) (Storage, error) {
	switch config.Type {
	case StorageTypeSQLite:
		return NewSQLiteStorage(config.Path)
	case StorageTypeFile:
		return NewFileStorage(config.Path)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}
