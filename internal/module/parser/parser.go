// Package parser provides parsers for importing external modules
// Supports Node-RED, n8n, and native EdgeFlow module formats
package parser

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ModuleFormat represents the source platform of a module
type ModuleFormat string

const (
	FormatNodeRED   ModuleFormat = "node-red"
	FormatN8N       ModuleFormat = "n8n"
	FormatEdgeFlow  ModuleFormat = "edgeflow"
	FormatUnknown   ModuleFormat = "unknown"
)

// ModuleInfo contains parsed module metadata
type ModuleInfo struct {
	Format      ModuleFormat           `json:"format"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	License     string                 `json:"license"`
	Homepage    string                 `json:"homepage"`
	Repository  string                 `json:"repository"`
	Keywords    []string               `json:"keywords"`
	Nodes       []NodeInfo             `json:"nodes"`
	Config      map[string]interface{} `json:"config"`
	SourcePath  string                 `json:"source_path"`
}

// NodeInfo contains parsed node definition
type NodeInfo struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Color       string                 `json:"color"`
	Inputs      int                    `json:"inputs"`
	Outputs     int                    `json:"outputs"`
	Properties  []PropertyInfo         `json:"properties"`
	SourceFile  string                 `json:"source_file"`
	UIFile      string                 `json:"ui_file"`
	Config      map[string]interface{} `json:"config"`
}

// PropertyInfo describes a node property/configuration
type PropertyInfo struct {
	Name        string      `json:"name"`
	Label       string      `json:"label"`
	Type        string      `json:"type"`
	Default     interface{} `json:"default"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Options     []string    `json:"options,omitempty"`
}

// Parser interface for module parsers
type Parser interface {
	// CanParse checks if this parser can handle the given module
	CanParse(path string) bool

	// Parse parses the module and returns metadata
	Parse(path string) (*ModuleInfo, error)

	// Format returns the module format this parser handles
	Format() ModuleFormat
}

// PackageJSON represents a Node.js package.json file
type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Author          interface{}       `json:"author"` // Can be string or object
	License         string            `json:"license"`
	Homepage        string            `json:"homepage"`
	Repository      interface{}       `json:"repository"` // Can be string or object
	Keywords        []string          `json:"keywords"`
	Main            string            `json:"main"`
	NodeRed         *NodeRedConfig    `json:"node-red,omitempty"`
	N8N             *N8NConfig        `json:"n8n,omitempty"`
	EdgeFlow        *EdgeFlowConfig   `json:"edgeflow,omitempty"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// NodeRedConfig is the node-red configuration in package.json
type NodeRedConfig struct {
	Version string            `json:"version"`
	Nodes   map[string]string `json:"nodes"` // type -> file path
}

// N8NConfig is the n8n configuration in package.json
type N8NConfig struct {
	N8NNodesModule  int      `json:"n8nNodesModule"`
	Nodes           []string `json:"nodes"`
	Credentials     []string `json:"credentials"`
}

// EdgeFlowConfig is the edgeflow configuration in package.json
type EdgeFlowConfig struct {
	Version string            `json:"version"`
	Nodes   map[string]string `json:"nodes"`
}

// EdgeFlowManifest is the native EdgeFlow module manifest (edgeflow.json)
type EdgeFlowManifest struct {
	Name        string                   `json:"name"`
	Version     string                   `json:"version"`
	Description string                   `json:"description"`
	Author      string                   `json:"author"`
	License     string                   `json:"license"`
	Homepage    string                   `json:"homepage"`
	Repository  string                   `json:"repository"`
	Keywords    []string                 `json:"keywords"`
	Platform    string                   `json:"platform"`    // "raspberry-pi", "linux", "windows", "all"
	Arch        []string                 `json:"arch"`        // ["arm64", "amd64", "arm"]
	GoVersion   string                   `json:"go_version"`  // Minimum Go version required
	Nodes       []EdgeFlowNodeDefinition `json:"nodes"`
	Binary      string                   `json:"binary"`      // Pre-compiled binary name (optional)
	EntryPoint  string                   `json:"entry_point"` // Main Go file if not pre-compiled
}

// EdgeFlowNodeDefinition defines a node in EdgeFlow native format
type EdgeFlowNodeDefinition struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Color       string                 `json:"color"`
	Inputs      []PortDefinition       `json:"inputs"`
	Outputs     []PortDefinition       `json:"outputs"`
	Properties  []PropertyInfo         `json:"properties"`
	Config      map[string]interface{} `json:"config"`
}

// PortDefinition defines input/output ports for a node
type PortDefinition struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// DetectFormat detects the module format from a directory or archive
func DetectFormat(path string) (ModuleFormat, error) {
	// Check if it's an archive
	if isArchive(path) {
		// Extract and detect
		extractPath, err := extractArchive(path)
		if err != nil {
			return FormatUnknown, fmt.Errorf("failed to extract archive: %w", err)
		}
		defer os.RemoveAll(extractPath)
		path = extractPath
	}

	// Helper to find files in path or one level deeper
	findFile := func(filename string) string {
		// First try direct path
		filePath := filepath.Join(path, filename)
		if _, err := os.Stat(filePath); err == nil {
			return filePath
		}
		// Try one level deeper (some archives have a root folder like "package/")
		entries, err := os.ReadDir(path)
		if err == nil && len(entries) >= 1 {
			for _, entry := range entries {
				if entry.IsDir() {
					nestedPath := filepath.Join(path, entry.Name(), filename)
					if _, err := os.Stat(nestedPath); err == nil {
						return nestedPath
					}
				}
			}
		}
		return ""
	}

	// First, check for native EdgeFlow module (edgeflow.json)
	edgeflowPath := findFile("edgeflow.json")
	if edgeflowPath != "" {
		data, err := os.ReadFile(edgeflowPath)
		if err == nil {
			var manifest EdgeFlowManifest
			if err := json.Unmarshal(data, &manifest); err == nil && manifest.Name != "" {
				return FormatEdgeFlow, nil
			}
		}
	}

	// Look for package.json (Node-RED, n8n, or JS-based EdgeFlow modules)
	pkgPath := findFile("package.json")
	if pkgPath == "" {
		// No manifest files found
		return FormatUnknown, fmt.Errorf("no edgeflow.json or package.json found")
	}

	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return FormatUnknown, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return FormatUnknown, fmt.Errorf("invalid package.json: %w", err)
	}

	// Detect format based on package.json contents
	if pkg.EdgeFlow != nil {
		return FormatEdgeFlow, nil
	}
	if pkg.NodeRed != nil && len(pkg.NodeRed.Nodes) > 0 {
		return FormatNodeRED, nil
	}
	if pkg.N8N != nil && len(pkg.N8N.Nodes) > 0 {
		return FormatN8N, nil
	}

	// Check for Node-RED by convention (node-red-contrib-* or node-red-node-*)
	if strings.HasPrefix(pkg.Name, "node-red-contrib-") ||
		strings.HasPrefix(pkg.Name, "node-red-node-") {
		return FormatNodeRED, nil
	}

	// Check for n8n by convention (n8n-nodes-*)
	if strings.HasPrefix(pkg.Name, "n8n-nodes-") {
		return FormatN8N, nil
	}

	return FormatUnknown, nil
}

// ParseAuthor extracts author string from various formats
func ParseAuthor(author interface{}) string {
	if author == nil {
		return ""
	}
	if s, ok := author.(string); ok {
		return s
	}
	if m, ok := author.(map[string]interface{}); ok {
		name, _ := m["name"].(string)
		email, _ := m["email"].(string)
		if email != "" {
			return fmt.Sprintf("%s <%s>", name, email)
		}
		return name
	}
	return ""
}

// ParseRepository extracts repository URL from various formats
func ParseRepository(repo interface{}) string {
	if repo == nil {
		return ""
	}
	if s, ok := repo.(string); ok {
		return s
	}
	if m, ok := repo.(map[string]interface{}); ok {
		url, _ := m["url"].(string)
		return url
	}
	return ""
}

// isArchive checks if the path is an archive file
func isArchive(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".zip" || ext == ".tgz" || ext == ".gz" || ext == ".tar"
}

// extractArchive extracts an archive to a temporary directory
func extractArchive(archivePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(archivePath))

	tmpDir, err := os.MkdirTemp("", "edgeflow_module_*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	switch ext {
	case ".zip":
		return tmpDir, extractZip(archivePath, tmpDir)
	case ".tgz", ".gz":
		return tmpDir, extractTarGz(archivePath, tmpDir)
	case ".tar":
		return tmpDir, extractTar(archivePath, tmpDir)
	default:
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("unsupported archive format: %s", ext)
	}
}

// extractZip extracts a zip archive
func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Security check: prevent path traversal
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
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

// extractTarGz extracts a tar.gz archive
func extractTarGz(src, dest string) error {
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

	return extractTarReader(tar.NewReader(gzr), dest)
}

// extractTar extracts a tar archive
func extractTar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	return extractTarReader(tar.NewReader(file), dest)
}

// extractTarReader extracts files from a tar reader
func extractTarReader(tr *tar.Reader, dest string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		// Security check: prevent path traversal
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
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

// Errors
var (
	ErrInvalidModule    = errors.New("invalid module")
	ErrUnsupportedFormat = errors.New("unsupported module format")
	ErrMissingPackageJSON = errors.New("missing package.json")
	ErrParseError       = errors.New("parse error")
)
