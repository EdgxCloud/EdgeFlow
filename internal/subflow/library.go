package subflow

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Library manages subflow import/export and catalog
type Library struct {
	basePath string
	registry *Registry
}

// NewLibrary creates a new subflow library
func NewLibrary(basePath string, registry *Registry) *Library {
	return &Library{
		basePath: basePath,
		registry: registry,
	}
}

// Export exports a subflow to a JSON file
func (l *Library) Export(subflowID, outputPath string) error {
	def, err := l.registry.GetDefinition(subflowID)
	if err != nil {
		return fmt.Errorf("failed to get subflow: %w", err)
	}

	// Update export timestamp
	def.UpdatedAt = time.Now()

	data, err := def.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize subflow: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Import imports a subflow from a JSON file
func (l *Library) Import(inputPath string) (*SubflowDefinition, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	def, err := FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subflow: %w", err)
	}

	// Register in registry
	if err := l.registry.RegisterDefinition(def); err != nil {
		return nil, fmt.Errorf("failed to register subflow: %w", err)
	}

	return def, nil
}

// ImportFromReader imports a subflow from an io.Reader
func (l *Library) ImportFromReader(reader io.Reader) (*SubflowDefinition, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	def, err := FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subflow: %w", err)
	}

	if err := l.registry.RegisterDefinition(def); err != nil {
		return nil, fmt.Errorf("failed to register subflow: %w", err)
	}

	return def, nil
}

// ExportToWriter exports a subflow to an io.Writer
func (l *Library) ExportToWriter(subflowID string, writer io.Writer) error {
	def, err := l.registry.GetDefinition(subflowID)
	if err != nil {
		return fmt.Errorf("failed to get subflow: %w", err)
	}

	def.UpdatedAt = time.Now()

	data, err := def.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize subflow: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// SaveToLibrary saves a subflow to the library directory
func (l *Library) SaveToLibrary(subflowID string) error {
	def, err := l.registry.GetDefinition(subflowID)
	if err != nil {
		return fmt.Errorf("failed to get subflow: %w", err)
	}

	// Create library directory structure: basePath/category/subflowName.json
	category := def.Category
	if category == "" {
		category = "general"
	}

	filename := sanitizeFilename(def.Name) + ".json"
	outputPath := filepath.Join(l.basePath, category, filename)

	return l.Export(subflowID, outputPath)
}

// LoadFromLibrary loads a subflow from the library directory
func (l *Library) LoadFromLibrary(category, name string) (*SubflowDefinition, error) {
	filename := sanitizeFilename(name) + ".json"
	inputPath := filepath.Join(l.basePath, category, filename)

	return l.Import(inputPath)
}

// ListLibrary lists all subflows in the library
func (l *Library) ListLibrary() ([]LibraryEntry, error) {
	var entries []LibraryEntry

	if _, err := os.Stat(l.basePath); os.IsNotExist(err) {
		return entries, nil
	}

	err := filepath.Walk(l.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files that can't be read
		}

		var def SubflowDefinition
		if err := json.Unmarshal(data, &def); err != nil {
			return nil // Skip invalid files
		}

		relPath, _ := filepath.Rel(l.basePath, path)
		category := filepath.Dir(relPath)
		if category == "." {
			category = "general"
		}

		entries = append(entries, LibraryEntry{
			ID:          def.ID,
			Name:        def.Name,
			Description: def.Description,
			Category:    category,
			Path:        path,
			Version:     def.Version,
			Author:      def.Author,
			UpdatedAt:   def.UpdatedAt,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk library: %w", err)
	}

	return entries, nil
}

// LibraryEntry represents a subflow in the library catalog
type LibraryEntry struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Path        string    `json:"path"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	result := filepath.Base(name)
	result = filepath.Clean(result)

	if result == "" || result == "." || result == ".." {
		result = "unnamed"
	}

	return result
}

// ExportPackage exports multiple subflows as a package
func (l *Library) ExportPackage(subflowIDs []string, outputPath string) error {
	pkg := SubflowPackage{
		Version:   "1.0",
		Subflows:  make([]*SubflowDefinition, 0, len(subflowIDs)),
		CreatedAt: time.Now(),
	}

	for _, id := range subflowIDs {
		def, err := l.registry.GetDefinition(id)
		if err != nil {
			return fmt.Errorf("failed to get subflow %s: %w", id, err)
		}
		pkg.Subflows = append(pkg.Subflows, def)
	}

	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package: %w", err)
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write package: %w", err)
	}

	return nil
}

// ImportPackage imports multiple subflows from a package
func (l *Library) ImportPackage(inputPath string) ([]string, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package: %w", err)
	}

	var pkg SubflowPackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package: %w", err)
	}

	var imported []string
	for _, def := range pkg.Subflows {
		if err := def.Validate(); err != nil {
			return imported, fmt.Errorf("invalid subflow %s: %w", def.ID, err)
		}

		if err := l.registry.RegisterDefinition(def); err != nil {
			return imported, fmt.Errorf("failed to register subflow %s: %w", def.ID, err)
		}

		imported = append(imported, def.ID)
	}

	return imported, nil
}

// SubflowPackage represents a collection of subflows
type SubflowPackage struct {
	Version     string               `json:"version"`
	Subflows    []*SubflowDefinition `json:"subflows"`
	Description string               `json:"description,omitempty"`
	Author      string               `json:"author,omitempty"`
	CreatedAt   time.Time            `json:"createdAt"`
}

// Clone creates a new subflow by cloning an existing one
func (l *Library) Clone(sourceID, newID, newName string) (*SubflowDefinition, error) {
	source, err := l.registry.GetDefinition(sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source subflow: %w", err)
	}

	clone, err := source.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone subflow: %w", err)
	}

	clone.ID = newID
	clone.Name = newName
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = time.Now()

	if err := l.registry.RegisterDefinition(clone); err != nil {
		return nil, fmt.Errorf("failed to register clone: %w", err)
	}

	return clone, nil
}

// Delete removes a subflow from the library and registry
func (l *Library) Delete(subflowID string) error {
	def, err := l.registry.GetDefinition(subflowID)
	if err != nil {
		return fmt.Errorf("failed to get subflow: %w", err)
	}

	// Remove from registry
	if err := l.registry.UnregisterDefinition(subflowID); err != nil {
		return fmt.Errorf("failed to unregister: %w", err)
	}

	// Remove from file system if exists
	category := def.Category
	if category == "" {
		category = "general"
	}

	filename := sanitizeFilename(def.Name) + ".json"
	filePath := filepath.Join(l.basePath, category, filename)

	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	return nil
}

// GetMetadata retrieves metadata for a subflow without loading the full definition
func (l *Library) GetMetadata(category, name string) (*LibraryEntry, error) {
	filename := sanitizeFilename(name) + ".json"
	path := filepath.Join(l.basePath, category, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var def SubflowDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse subflow: %w", err)
	}

	return &LibraryEntry{
		ID:          def.ID,
		Name:        def.Name,
		Description: def.Description,
		Category:    category,
		Path:        path,
		Version:     def.Version,
		Author:      def.Author,
		UpdatedAt:   def.UpdatedAt,
	}, nil
}
