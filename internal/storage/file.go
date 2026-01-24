package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileStorage implements Storage using the filesystem
type FileStorage struct {
	basePath string
	mu       sync.RWMutex
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(basePath string) (*FileStorage, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &FileStorage{
		basePath: basePath,
	}, nil
}

// SaveFlow saves a flow to disk
func (s *FileStorage) SaveFlow(flow *Flow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	flow.CreatedAt = time.Now()
	flow.UpdatedAt = time.Now()

	filePath := filepath.Join(s.basePath, flow.ID+".json")

	data, err := json.MarshalIndent(flow, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal flow: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write flow file: %w", err)
	}

	return nil
}

// GetFlow retrieves a flow from disk
func (s *FileStorage) GetFlow(id string) (*Flow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := filepath.Join(s.basePath, id+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("flow not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read flow file: %w", err)
	}

	var flow Flow
	if err := json.Unmarshal(data, &flow); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flow: %w", err)
	}

	return &flow, nil
}

// ListFlows returns all flows from disk
func (s *FileStorage) ListFlows() ([]*Flow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	files, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	flows := []*Flow{}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(s.basePath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files that can't be read
		}

		var flow Flow
		if err := json.Unmarshal(data, &flow); err != nil {
			continue // Skip invalid files
		}

		flows = append(flows, &flow)
	}

	return flows, nil
}

// DeleteFlow removes a flow from disk
func (s *FileStorage) DeleteFlow(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.basePath, id+".json")

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("flow not found: %s", id)
		}
		return fmt.Errorf("failed to delete flow file: %w", err)
	}

	return nil
}

// UpdateFlow updates an existing flow on disk
func (s *FileStorage) UpdateFlow(flow *Flow) error {
	flow.UpdatedAt = time.Now()
	// For file storage, update is the same as save
	return s.SaveFlow(flow)
}

// Close closes the storage (no-op for file storage)
func (s *FileStorage) Close() error {
	return nil
}
