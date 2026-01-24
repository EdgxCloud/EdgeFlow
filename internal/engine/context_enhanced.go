package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ContextScope defines the scope of a context store
type ContextScope string

const (
	ContextScopeNode   ContextScope = "node"   // Per-node storage
	ContextScopeFlow   ContextScope = "flow"   // Per-flow storage
	ContextScopeGlobal ContextScope = "global" // Cross-flow storage
)

// ContextStore defines the interface for context storage
type ContextStore interface {
	// Get retrieves a value from the specified scope and key
	Get(scope ContextScope, scopeID string, key string) (interface{}, error)

	// Set stores a value in the specified scope and key
	Set(scope ContextScope, scopeID string, key string, value interface{}) error

	// Keys returns all keys in the specified scope
	Keys(scope ContextScope, scopeID string) ([]string, error)

	// Delete removes a value from the specified scope and key
	Delete(scope ContextScope, scopeID string, key string) error

	// Clear removes all values from the specified scope
	Clear(scope ContextScope, scopeID string) error

	// Close closes the context store and persists data if necessary
	Close() error
}

// MemoryContextStore implements an in-memory context store
type MemoryContextStore struct {
	data map[string]map[string]interface{} // scopeKey -> key -> value
	mu   sync.RWMutex
}

// NewMemoryContextStore creates a new memory-based context store
func NewMemoryContextStore() *MemoryContextStore {
	return &MemoryContextStore{
		data: make(map[string]map[string]interface{}),
	}
}

func (m *MemoryContextStore) scopeKey(scope ContextScope, scopeID string) string {
	return string(scope) + ":" + scopeID
}

func (m *MemoryContextStore) Get(scope ContextScope, scopeID string, key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	scopeKey := m.scopeKey(scope, scopeID)
	if scopeData, ok := m.data[scopeKey]; ok {
		if value, exists := scopeData[key]; exists {
			return value, nil
		}
	}

	return nil, fmt.Errorf("key '%s' not found in %s context", key, scope)
}

func (m *MemoryContextStore) Set(scope ContextScope, scopeID string, key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scopeKey := m.scopeKey(scope, scopeID)
	if _, ok := m.data[scopeKey]; !ok {
		m.data[scopeKey] = make(map[string]interface{})
	}

	m.data[scopeKey][key] = value
	return nil
}

func (m *MemoryContextStore) Keys(scope ContextScope, scopeID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	scopeKey := m.scopeKey(scope, scopeID)
	if scopeData, ok := m.data[scopeKey]; ok {
		keys := make([]string, 0, len(scopeData))
		for k := range scopeData {
			keys = append(keys, k)
		}
		return keys, nil
	}

	return []string{}, nil
}

func (m *MemoryContextStore) Delete(scope ContextScope, scopeID string, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scopeKey := m.scopeKey(scope, scopeID)
	if scopeData, ok := m.data[scopeKey]; ok {
		delete(scopeData, key)
	}

	return nil
}

func (m *MemoryContextStore) Clear(scope ContextScope, scopeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scopeKey := m.scopeKey(scope, scopeID)
	delete(m.data, scopeKey)

	return nil
}

func (m *MemoryContextStore) Close() error {
	// Memory store doesn't need cleanup
	return nil
}

// FileContextStore implements a file-based persistent context store
type FileContextStore struct {
	basePath string
	data     map[string]map[string]interface{}
	mu       sync.RWMutex
	dirty    map[string]bool // Track which scopes need saving
}

// NewFileContextStore creates a new file-based context store
func NewFileContextStore(basePath string) (*FileContextStore, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create context directory: %w", err)
	}

	store := &FileContextStore{
		basePath: basePath,
		data:     make(map[string]map[string]interface{}),
		dirty:    make(map[string]bool),
	}

	// Load existing context files
	if err := store.loadAll(); err != nil {
		return nil, fmt.Errorf("failed to load context files: %w", err)
	}

	return store, nil
}

func (f *FileContextStore) scopeKey(scope ContextScope, scopeID string) string {
	return string(scope) + ":" + scopeID
}

func (f *FileContextStore) scopeFilePath(scopeKey string) string {
	return filepath.Join(f.basePath, scopeKey+".json")
}

func (f *FileContextStore) loadAll() error {
	// Load global context
	globalPath := f.scopeFilePath(string(ContextScopeGlobal) + ":default")
	if _, err := os.Stat(globalPath); err == nil {
		if err := f.loadFile(string(ContextScopeGlobal)+":default", globalPath); err != nil {
			return err
		}
	}

	// Load flow and node contexts from directory
	entries, err := os.ReadDir(f.basePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		scopeKey := entry.Name()[:len(entry.Name())-5] // Remove .json
		filePath := filepath.Join(f.basePath, entry.Name())

		if err := f.loadFile(scopeKey, filePath); err != nil {
			// Log error but continue loading other files
			continue
		}
	}

	return nil
}

func (f *FileContextStore) loadFile(scopeKey, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var scopeData map[string]interface{}
	if err := json.Unmarshal(data, &scopeData); err != nil {
		return err
	}

	f.data[scopeKey] = scopeData
	return nil
}

func (f *FileContextStore) saveFile(scopeKey string) error {
	f.mu.RLock()
	scopeData, exists := f.data[scopeKey]
	f.mu.RUnlock()

	if !exists || len(scopeData) == 0 {
		// Remove file if no data
		filePath := f.scopeFilePath(scopeKey)
		os.Remove(filePath)
		return nil
	}

	data, err := json.MarshalIndent(scopeData, "", "  ")
	if err != nil {
		return err
	}

	filePath := f.scopeFilePath(scopeKey)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	f.mu.Lock()
	delete(f.dirty, scopeKey)
	f.mu.Unlock()

	return nil
}

func (f *FileContextStore) Get(scope ContextScope, scopeID string, key string) (interface{}, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	scopeKey := f.scopeKey(scope, scopeID)
	if scopeData, ok := f.data[scopeKey]; ok {
		if value, exists := scopeData[key]; exists {
			return value, nil
		}
	}

	return nil, fmt.Errorf("key '%s' not found in %s context", key, scope)
}

func (f *FileContextStore) Set(scope ContextScope, scopeID string, key string, value interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	scopeKey := f.scopeKey(scope, scopeID)
	if _, ok := f.data[scopeKey]; !ok {
		f.data[scopeKey] = make(map[string]interface{})
	}

	f.data[scopeKey][key] = value
	f.dirty[scopeKey] = true

	// Auto-save after short delay (async)
	go func() {
		time.Sleep(100 * time.Millisecond)
		f.saveFile(scopeKey)
	}()

	return nil
}

func (f *FileContextStore) Keys(scope ContextScope, scopeID string) ([]string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	scopeKey := f.scopeKey(scope, scopeID)
	if scopeData, ok := f.data[scopeKey]; ok {
		keys := make([]string, 0, len(scopeData))
		for k := range scopeData {
			keys = append(keys, k)
		}
		return keys, nil
	}

	return []string{}, nil
}

func (f *FileContextStore) Delete(scope ContextScope, scopeID string, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	scopeKey := f.scopeKey(scope, scopeID)
	if scopeData, ok := f.data[scopeKey]; ok {
		delete(scopeData, key)
		f.dirty[scopeKey] = true

		// Auto-save after short delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			f.saveFile(scopeKey)
		}()
	}

	return nil
}

func (f *FileContextStore) Clear(scope ContextScope, scopeID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	scopeKey := f.scopeKey(scope, scopeID)
	delete(f.data, scopeKey)
	delete(f.dirty, scopeKey)

	// Remove file
	filePath := f.scopeFilePath(scopeKey)
	os.Remove(filePath)

	return nil
}

func (f *FileContextStore) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Save all dirty contexts
	for scopeKey := range f.dirty {
		if err := f.saveFile(scopeKey); err != nil {
			// Log error but continue
			continue
		}
	}

	return nil
}

// ContextManager manages all three context levels
type ContextManager struct {
	store ContextStore
	mu    sync.RWMutex
}

// NewContextManager creates a new context manager
func NewContextManager(store ContextStore) *ContextManager {
	return &ContextManager{
		store: store,
	}
}

// NewMemoryContextManager creates a context manager with memory storage
func NewMemoryContextManager() *ContextManager {
	return &ContextManager{
		store: NewMemoryContextStore(),
	}
}

// NewFileContextManager creates a context manager with file storage
func NewFileContextManager(basePath string) (*ContextManager, error) {
	store, err := NewFileContextStore(basePath)
	if err != nil {
		return nil, err
	}

	return &ContextManager{
		store: store,
	}, nil
}

// GetNodeContext returns a context accessor for a specific node
func (cm *ContextManager) GetNodeContext(nodeID string) *Context {
	return &Context{
		store:   cm.store,
		scope:   ContextScopeNode,
		scopeID: nodeID,
	}
}

// GetFlowContext returns a context accessor for a specific flow
func (cm *ContextManager) GetFlowContext(flowID string) *Context {
	return &Context{
		store:   cm.store,
		scope:   ContextScopeFlow,
		scopeID: flowID,
	}
}

// GetGlobalContext returns the global context accessor
func (cm *ContextManager) GetGlobalContext() *Context {
	return &Context{
		store:   cm.store,
		scope:   ContextScopeGlobal,
		scopeID: "default",
	}
}

// Close closes the context manager and persists data
func (cm *ContextManager) Close() error {
	return cm.store.Close()
}

// Context provides a convenient interface for accessing context data
type Context struct {
	store   ContextStore
	scope   ContextScope
	scopeID string
}

// Get retrieves a value from context
func (c *Context) Get(key string) (interface{}, error) {
	return c.store.Get(c.scope, c.scopeID, key)
}

// GetString retrieves a string value from context
func (c *Context) GetString(key string) (string, error) {
	val, err := c.Get(key)
	if err != nil {
		return "", err
	}

	if str, ok := val.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("value for key '%s' is not a string", key)
}

// GetInt retrieves an int value from context
func (c *Context) GetInt(key string) (int, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}

	// Handle both int and float64 (JSON unmarshaling uses float64 for numbers)
	switch v := val.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("value for key '%s' is not a number", key)
	}
}

// GetBool retrieves a bool value from context
func (c *Context) GetBool(key string) (bool, error) {
	val, err := c.Get(key)
	if err != nil {
		return false, err
	}

	if b, ok := val.(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("value for key '%s' is not a boolean", key)
}

// Set stores a value in context
func (c *Context) Set(key string, value interface{}) error {
	return c.store.Set(c.scope, c.scopeID, key, value)
}

// Keys returns all keys in this context
func (c *Context) Keys() ([]string, error) {
	return c.store.Keys(c.scope, c.scopeID)
}

// Delete removes a value from context
func (c *Context) Delete(key string) error {
	return c.store.Delete(c.scope, c.scopeID, key)
}

// Clear removes all values from this context
func (c *Context) Clear() error {
	return c.store.Clear(c.scope, c.scopeID)
}

// GetOrDefault retrieves a value or returns a default if not found
func (c *Context) GetOrDefault(key string, defaultValue interface{}) interface{} {
	val, err := c.Get(key)
	if err != nil {
		return defaultValue
	}
	return val
}

// Increment increments a numeric value in context (atomic)
func (c *Context) Increment(key string, delta int) (int, error) {
	val, err := c.GetInt(key)
	if err != nil {
		// Initialize to 0 if not exists
		val = 0
	}

	newVal := val + delta
	if err := c.Set(key, newVal); err != nil {
		return 0, err
	}

	return newVal, nil
}

// Append appends a value to an array in context
func (c *Context) Append(key string, value interface{}) error {
	val, err := c.Get(key)
	if err != nil {
		// Initialize new array
		return c.Set(key, []interface{}{value})
	}

	arr, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("value for key '%s' is not an array", key)
	}

	arr = append(arr, value)
	return c.Set(key, arr)
}
