// Package manager provides module lifecycle management
// Handles installation, loading, unloading, and updates of modules
package manager

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/module/adapter"
	"github.com/edgeflow/edgeflow/internal/module/parser"
	"github.com/edgeflow/edgeflow/internal/module/validator"
	"github.com/edgeflow/edgeflow/internal/node"
)

// ModuleStatus represents the status of a module
type ModuleStatus string

const (
	StatusPending    ModuleStatus = "pending"
	StatusInstalled  ModuleStatus = "installed"
	StatusLoaded     ModuleStatus = "loaded"
	StatusError      ModuleStatus = "error"
	StatusDisabled   ModuleStatus = "disabled"
)

// InstalledModule represents an installed module
type InstalledModule struct {
	Info          *parser.ModuleInfo          `json:"info"`
	Status        ModuleStatus                `json:"status"`
	InstalledAt   time.Time                   `json:"installed_at"`
	UpdatedAt     time.Time                   `json:"updated_at"`
	Enabled       bool                        `json:"enabled"`
	Error         string                      `json:"error,omitempty"`
	Validation    *validator.ValidationResult `json:"validation,omitempty"`
	LoadedNodes   []string                    `json:"loaded_nodes,omitempty"`
	LicenseInfo   *validator.LicenseInfo      `json:"license_info,omitempty"`
	Attribution   *LicenseAttribution         `json:"attribution,omitempty"`
}

// LicenseAttribution stores license attribution information
type LicenseAttribution struct {
	ModuleName    string    `json:"module_name"`
	ModuleVersion string    `json:"module_version"`
	License       string    `json:"license"`
	Author        string    `json:"author,omitempty"`
	Homepage      string    `json:"homepage,omitempty"`
	Repository    string    `json:"repository,omitempty"`
	Copyright     string    `json:"copyright,omitempty"`
	LicenseText   string    `json:"license_text,omitempty"`
	AcceptedAt    time.Time `json:"accepted_at"`
	AcceptedBy    string    `json:"accepted_by,omitempty"`
}

// ModuleManager manages imported modules
type ModuleManager struct {
	mu            sync.RWMutex
	modulesDir    string
	modules       map[string]*InstalledModule
	nodeRegistry  *node.Registry
	adapterReg    *adapter.AdapterRegistry
	validator     *validator.Validator
	manifestPath  string
}

// NewModuleManager creates a new module manager
func NewModuleManager(modulesDir string) (*ModuleManager, error) {
	// Ensure modules directory exists
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create modules directory: %w", err)
	}

	m := &ModuleManager{
		modulesDir:   modulesDir,
		modules:      make(map[string]*InstalledModule),
		nodeRegistry: node.GetGlobalRegistry(),
		adapterReg:   adapter.GetAdapterRegistry(),
		validator:    validator.NewValidator(),
		manifestPath: filepath.Join(modulesDir, "modules.json"),
	}

	// Load existing modules manifest
	if err := m.loadManifest(); err != nil {
		// Not a fatal error - may be first run
	}

	return m, nil
}

// Install installs a module from a path (directory or archive)
func (m *ModuleManager) Install(sourcePath string) (*InstalledModule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If it's an archive, extract it first to a persistent temp location
	workPath := sourcePath
	var cleanupPath string
	if isArchiveFile(sourcePath) {
		extractedPath, cleanup, err := extractArchiveToTemp(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract archive: %w", err)
		}
		cleanupPath = cleanup
		defer os.RemoveAll(cleanupPath)
		workPath = extractedPath
	}

	// Detect module format from the working path
	format, err := parser.DetectFormat(workPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect module format: %w", err)
	}

	// Get appropriate parser
	var moduleParser parser.Parser
	switch format {
	case parser.FormatEdgeFlow:
		moduleParser = parser.NewEdgeFlowParser()
	case parser.FormatNodeRED:
		moduleParser = parser.NewNodeRedParser()
	case parser.FormatN8N:
		moduleParser = parser.NewN8NParser()
	default:
		return nil, fmt.Errorf("unsupported module format: %s", format)
	}

	// Parse module from the working path
	info, err := moduleParser.Parse(workPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse module: %w", err)
	}

	// Validate module (info.SourcePath now points to valid extracted files)
	validationResult := m.validator.Validate(info)
	if !validationResult.Valid {
		return nil, fmt.Errorf("module validation failed: %v", validationResult.Errors)
	}

	// Check if module already exists
	if existing, ok := m.modules[info.Name]; ok {
		// Update existing module - use info.SourcePath which has the valid files
		return m.updateModule(existing, info, info.SourcePath)
	}

	// Copy module to modules directory from the source path in info
	moduleDir := filepath.Join(m.modulesDir, info.Name)
	if err := copyDirectory(info.SourcePath, moduleDir); err != nil {
		return nil, fmt.Errorf("failed to copy module: %w", err)
	}

	// Update source path to installed location
	info.SourcePath = moduleDir

	// Read license file if exists
	licenseText := readLicenseFile(moduleDir)

	// Create license attribution
	attribution := &LicenseAttribution{
		ModuleName:    info.Name,
		ModuleVersion: info.Version,
		License:       info.License,
		Author:        info.Author,
		Homepage:      info.Homepage,
		Repository:    info.Repository,
		LicenseText:   licenseText,
		AcceptedAt:    time.Now(),
	}

	// Create installed module record
	installed := &InstalledModule{
		Info:        info,
		Status:      StatusInstalled,
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		Enabled:     true,
		Validation:  validationResult,
		LicenseInfo: validationResult.LicenseInfo,
		Attribution: attribution,
	}

	m.modules[info.Name] = installed

	// Save manifest
	if err := m.saveManifest(); err != nil {
		return nil, fmt.Errorf("failed to save manifest: %w", err)
	}

	return installed, nil
}

// updateModule updates an existing module
func (m *ModuleManager) updateModule(existing *InstalledModule, newInfo *parser.ModuleInfo, sourcePath string) (*InstalledModule, error) {
	// Unload existing module first
	if existing.Status == StatusLoaded {
		if err := m.unloadModuleNodes(existing); err != nil {
			return nil, fmt.Errorf("failed to unload existing module: %w", err)
		}
	}

	// Backup existing module
	moduleDir := filepath.Join(m.modulesDir, existing.Info.Name)
	backupDir := moduleDir + ".bak"
	if err := os.Rename(moduleDir, backupDir); err != nil {
		return nil, fmt.Errorf("failed to backup existing module: %w", err)
	}

	// Copy new module
	if err := copyDirectory(sourcePath, moduleDir); err != nil {
		// Restore backup on failure
		os.Rename(backupDir, moduleDir)
		return nil, fmt.Errorf("failed to copy new module: %w", err)
	}

	// Remove backup
	os.RemoveAll(backupDir)

	// Update info
	newInfo.SourcePath = moduleDir
	existing.Info = newInfo
	existing.UpdatedAt = time.Now()
	existing.Status = StatusInstalled
	existing.Error = ""

	// Re-validate
	existing.Validation = m.validator.Validate(newInfo)

	// Save manifest
	if err := m.saveManifest(); err != nil {
		return nil, fmt.Errorf("failed to save manifest: %w", err)
	}

	return existing, nil
}

// Uninstall removes a module
func (m *ModuleManager) Uninstall(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	module, ok := m.modules[name]
	if !ok {
		return fmt.Errorf("module not found: %s", name)
	}

	// Unload nodes first
	if module.Status == StatusLoaded {
		if err := m.unloadModuleNodes(module); err != nil {
			return fmt.Errorf("failed to unload module: %w", err)
		}
	}

	// Remove module directory
	moduleDir := filepath.Join(m.modulesDir, name)
	if err := os.RemoveAll(moduleDir); err != nil {
		return fmt.Errorf("failed to remove module directory: %w", err)
	}

	delete(m.modules, name)

	// Save manifest
	return m.saveManifest()
}

// Load loads a module and registers its nodes
func (m *ModuleManager) Load(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	module, ok := m.modules[name]
	if !ok {
		return fmt.Errorf("module not found: %s", name)
	}

	if module.Status == StatusLoaded {
		return nil // Already loaded
	}

	if !module.Enabled {
		return fmt.Errorf("module is disabled: %s", name)
	}

	// Get adapter for module format
	adapterInstance, ok := m.adapterReg.Get(module.Info.Format)
	if !ok {
		return fmt.Errorf("no adapter for format: %s", module.Info.Format)
	}

	loadedNodes := []string{}

	// Register each node
	for _, nodeInfo := range module.Info.Nodes {
		// Read source code
		sourcePath := filepath.Join(module.Info.SourcePath, nodeInfo.SourceFile)
		sourceCode, err := os.ReadFile(sourcePath)
		if err != nil {
			module.Error = fmt.Sprintf("failed to read node source: %s", err)
			module.Status = StatusError
			m.saveManifest()
			return err
		}

		// Create node info copy for closure
		ni := nodeInfo
		sc := string(sourceCode)

		// Create factory function
		factory := func() node.Executor {
			executor, err := adapterInstance.CreateExecutor(&ni, sc)
			if err != nil {
				return nil
			}
			return executor
		}

		// Register node type with full node info
		nodeType := fmt.Sprintf("%s/%s", module.Info.Name, nodeInfo.Type)

		// Create node.NodeInfo for registration
		regInfo := &node.NodeInfo{
			Type:        nodeType,
			Name:        nodeInfo.Name,
			Category:    node.NodeTypeFunction,
			Description: nodeInfo.Description,
			Icon:        nodeInfo.Icon,
			Color:       nodeInfo.Color,
			Factory:     factory,
		}

		if err := m.nodeRegistry.Register(regInfo); err != nil {
			// Node type may already exist, continue
		}
		loadedNodes = append(loadedNodes, nodeType)
	}

	module.LoadedNodes = loadedNodes
	module.Status = StatusLoaded
	module.Error = ""

	return m.saveManifest()
}

// Unload unloads a module
func (m *ModuleManager) Unload(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	module, ok := m.modules[name]
	if !ok {
		return fmt.Errorf("module not found: %s", name)
	}

	return m.unloadModuleNodes(module)
}

// unloadModuleNodes unloads nodes for a module (must hold lock)
func (m *ModuleManager) unloadModuleNodes(module *InstalledModule) error {
	// Unregister nodes
	for _, nodeType := range module.LoadedNodes {
		m.nodeRegistry.Unregister(nodeType)
	}

	module.LoadedNodes = nil
	module.Status = StatusInstalled

	return m.saveManifest()
}

// Enable enables a module
func (m *ModuleManager) Enable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	module, ok := m.modules[name]
	if !ok {
		return fmt.Errorf("module not found: %s", name)
	}

	module.Enabled = true
	module.Status = StatusInstalled

	return m.saveManifest()
}

// Disable disables a module
func (m *ModuleManager) Disable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	module, ok := m.modules[name]
	if !ok {
		return fmt.Errorf("module not found: %s", name)
	}

	// Unload if loaded
	if module.Status == StatusLoaded {
		m.unloadModuleNodes(module)
	}

	module.Enabled = false
	module.Status = StatusDisabled

	return m.saveManifest()
}

// List returns all installed modules
func (m *ModuleManager) List() []*InstalledModule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	modules := make([]*InstalledModule, 0, len(m.modules))
	for _, mod := range m.modules {
		modules = append(modules, mod)
	}
	return modules
}

// Get returns a specific module
func (m *ModuleManager) Get(name string) (*InstalledModule, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mod, ok := m.modules[name]
	return mod, ok
}

// LoadAll loads all enabled modules
func (m *ModuleManager) LoadAll() error {
	m.mu.RLock()
	names := make([]string, 0, len(m.modules))
	for name, mod := range m.modules {
		if mod.Enabled {
			names = append(names, name)
		}
	}
	m.mu.RUnlock()

	var lastErr error
	for _, name := range names {
		if err := m.Load(name); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// loadManifest loads the modules manifest
func (m *ModuleManager) loadManifest() error {
	data, err := os.ReadFile(m.manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &m.modules)
}

// saveManifest saves the modules manifest
func (m *ModuleManager) saveManifest() error {
	data, err := json.MarshalIndent(m.modules, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.manifestPath, data, 0644)
}

// copyDirectory copies a directory recursively
func copyDirectory(src, dst string) error {
	// Handle archive sources
	ext := filepath.Ext(src)
	if ext == ".zip" || ext == ".tgz" || ext == ".tar" || ext == ".gz" {
		// Extract to destination
		return extractToDirectory(src, dst)
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// extractToDirectory extracts an archive to a directory
func extractToDirectory(archive, dst string) error {
	// Create temp directory for extraction
	tmpDir, err := os.MkdirTemp("", "edgeflow_extract_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Extract archive to temp directory
	ext := filepath.Ext(archive)
	switch ext {
	case ".tgz", ".gz":
		if err := extractTarGzToDir(archive, tmpDir); err != nil {
			return fmt.Errorf("failed to extract tgz: %w", err)
		}
	case ".zip":
		if err := extractZipToDir(archive, tmpDir); err != nil {
			return fmt.Errorf("failed to extract zip: %w", err)
		}
	case ".tar":
		if err := extractTarToDir(archive, tmpDir); err != nil {
			return fmt.Errorf("failed to extract tar: %w", err)
		}
	default:
		return fmt.Errorf("unsupported archive format: %s", ext)
	}

	// Find the actual content (may be nested in a "package" folder for npm packs)
	contentDir := tmpDir
	entries, err := os.ReadDir(tmpDir)
	if err == nil && len(entries) == 1 && entries[0].IsDir() {
		// Single directory inside - use that as content
		contentDir = filepath.Join(tmpDir, entries[0].Name())
	}

	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Copy extracted content to destination
	return filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(contentDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// Global module manager instance
var (
	globalManager *ModuleManager
	managerOnce   sync.Once
)

// GetModuleManager returns the global module manager
func GetModuleManager() *ModuleManager {
	managerOnce.Do(func() {
		// Default to ./modules directory
		var err error
		globalManager, err = NewModuleManager("./modules")
		if err != nil {
			panic(fmt.Sprintf("failed to create module manager: %v", err))
		}
	})
	return globalManager
}

// InitModuleManager initializes the global module manager with a custom path
func InitModuleManager(modulesDir string) error {
	manager, err := NewModuleManager(modulesDir)
	if err != nil {
		return err
	}
	globalManager = manager
	return nil
}

// readLicenseFile reads the LICENSE file from a module directory
func readLicenseFile(moduleDir string) string {
	// Common license file names
	licenseFiles := []string{
		"LICENSE",
		"LICENSE.txt",
		"LICENSE.md",
		"LICENSE.MIT",
		"LICENSE.APACHE",
		"license",
		"license.txt",
		"LICENCE",
		"LICENCE.txt",
		"COPYING",
		"COPYING.txt",
	}

	for _, filename := range licenseFiles {
		licensePath := filepath.Join(moduleDir, filename)
		if data, err := os.ReadFile(licensePath); err == nil {
			// Limit license text to 10KB
			if len(data) > 10*1024 {
				return string(data[:10*1024]) + "\n... (truncated)"
			}
			return string(data)
		}
	}

	// Check one level deeper (for npm packages with nested structure)
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDir := filepath.Join(moduleDir, entry.Name())
		for _, filename := range licenseFiles {
			licensePath := filepath.Join(subDir, filename)
			if data, err := os.ReadFile(licensePath); err == nil {
				if len(data) > 10*1024 {
					return string(data[:10*1024]) + "\n... (truncated)"
				}
				return string(data)
			}
		}
	}

	return ""
}

// GetAllAttributions returns license attributions for all installed modules
func (m *ModuleManager) GetAllAttributions() []*LicenseAttribution {
	m.mu.RLock()
	defer m.mu.RUnlock()

	attributions := make([]*LicenseAttribution, 0, len(m.modules))
	for _, mod := range m.modules {
		if mod.Attribution != nil {
			attributions = append(attributions, mod.Attribution)
		}
	}
	return attributions
}

// GetLicenseInfo returns license info for a specific module
func (m *ModuleManager) GetLicenseInfo(name string) (*validator.LicenseInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mod, ok := m.modules[name]
	if !ok {
		return nil, fmt.Errorf("module not found: %s", name)
	}

	return mod.LicenseInfo, nil
}

// isArchiveFile checks if the path is an archive file
func isArchiveFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".zip" || ext == ".tgz" || ext == ".gz" || ext == ".tar"
}

// extractArchiveToTemp extracts an archive to a temp directory and returns:
// - extractedPath: the path to the actual module content (may be nested)
// - cleanupPath: the parent temp directory to clean up
// The caller is responsible for cleaning up the cleanupPath directory
func extractArchiveToTemp(archivePath string) (extractedPath string, cleanupPath string, err error) {
	tmpDir, err := os.MkdirTemp("", "edgeflow_install_*")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(archivePath))
	switch ext {
	case ".tgz", ".gz":
		if err := extractTarGzToDir(archivePath, tmpDir); err != nil {
			os.RemoveAll(tmpDir)
			return "", "", fmt.Errorf("failed to extract tgz: %w", err)
		}
	case ".zip":
		if err := extractZipToDir(archivePath, tmpDir); err != nil {
			os.RemoveAll(tmpDir)
			return "", "", fmt.Errorf("failed to extract zip: %w", err)
		}
	case ".tar":
		if err := extractTarToDir(archivePath, tmpDir); err != nil {
			os.RemoveAll(tmpDir)
			return "", "", fmt.Errorf("failed to extract tar: %w", err)
		}
	default:
		os.RemoveAll(tmpDir)
		return "", "", fmt.Errorf("unsupported archive format: %s", ext)
	}

	// Check if there's a single directory inside (like "package/")
	entries, err := os.ReadDir(tmpDir)
	if err == nil && len(entries) == 1 && entries[0].IsDir() {
		// Return the nested directory path as extractedPath, but tmpDir as cleanupPath
		return filepath.Join(tmpDir, entries[0].Name()), tmpDir, nil
	}

	return tmpDir, tmpDir, nil
}

// extractTarGzToDir extracts a .tgz or .tar.gz file to a directory
func extractTarGzToDir(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	return extractTarReader(tar.NewReader(gzr), dst)
}

// extractTarToDir extracts a .tar file to a directory
func extractTarToDir(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	return extractTarReader(tar.NewReader(file), dst)
}

// extractTarReader extracts files from a tar reader
func extractTarReader(tr *tar.Reader, dst string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dst, header.Name)

		// Security check: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

// extractZipToDir extracts a .zip file to a directory
func extractZipToDir(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dst, f.Name)

		// Security check: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
