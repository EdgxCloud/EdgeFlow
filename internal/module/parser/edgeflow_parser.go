// Package parser provides the EdgeFlow native module parser
package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// EdgeFlowParser parses native EdgeFlow modules
type EdgeFlowParser struct{}

// NewEdgeFlowParser creates a new EdgeFlow parser
func NewEdgeFlowParser() *EdgeFlowParser {
	return &EdgeFlowParser{}
}

// Format returns the module format
func (p *EdgeFlowParser) Format() ModuleFormat {
	return FormatEdgeFlow
}

// CanParse checks if this parser can handle the given module
func (p *EdgeFlowParser) CanParse(path string) bool {
	// Check for edgeflow.json
	manifestPath := filepath.Join(path, "edgeflow.json")
	if _, err := os.Stat(manifestPath); err == nil {
		return true
	}

	// Check one level deeper (for extracted archives with root folder)
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			nestedPath := filepath.Join(path, entry.Name(), "edgeflow.json")
			if _, err := os.Stat(nestedPath); err == nil {
				return true
			}
		}
	}

	return false
}

// Parse parses an EdgeFlow native module
func (p *EdgeFlowParser) Parse(path string) (*ModuleInfo, error) {
	// Handle archives
	if isArchive(path) {
		extractPath, err := extractArchive(path)
		if err != nil {
			return nil, fmt.Errorf("failed to extract archive: %w", err)
		}
		defer os.RemoveAll(extractPath)
		path = extractPath
	}

	// Find edgeflow.json
	manifestPath := p.findManifest(path)
	if manifestPath == "" {
		return nil, fmt.Errorf("edgeflow.json not found")
	}

	// Read and parse manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read edgeflow.json: %w", err)
	}

	var manifest EdgeFlowManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("invalid edgeflow.json: %w", err)
	}

	// Validate required fields
	if manifest.Name == "" {
		return nil, fmt.Errorf("module name is required in edgeflow.json")
	}
	if manifest.Version == "" {
		manifest.Version = "1.0.0"
	}

	// Get the module directory (parent of edgeflow.json)
	moduleDir := filepath.Dir(manifestPath)

	// Convert to ModuleInfo
	info := &ModuleInfo{
		Format:      FormatEdgeFlow,
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		License:     manifest.License,
		Homepage:    manifest.Homepage,
		Repository:  manifest.Repository,
		Keywords:    manifest.Keywords,
		SourcePath:  moduleDir,
		Config: map[string]interface{}{
			"platform":    manifest.Platform,
			"arch":        manifest.Arch,
			"go_version":  manifest.GoVersion,
			"binary":      manifest.Binary,
			"entry_point": manifest.EntryPoint,
		},
	}

	// Convert nodes
	for _, nodeDef := range manifest.Nodes {
		nodeInfo := NodeInfo{
			Type:        nodeDef.Type,
			Name:        nodeDef.Name,
			Category:    nodeDef.Category,
			Description: nodeDef.Description,
			Icon:        nodeDef.Icon,
			Color:       nodeDef.Color,
			Inputs:      len(nodeDef.Inputs),
			Outputs:     len(nodeDef.Outputs),
			Properties:  nodeDef.Properties,
			Config:      nodeDef.Config,
		}

		// Set source file based on entry point or binary
		if manifest.Binary != "" {
			nodeInfo.SourceFile = manifest.Binary
		} else if manifest.EntryPoint != "" {
			nodeInfo.SourceFile = manifest.EntryPoint
		}

		info.Nodes = append(info.Nodes, nodeInfo)
	}

	return info, nil
}

// findManifest finds the edgeflow.json file in the given path
func (p *EdgeFlowParser) findManifest(path string) string {
	// Direct path
	manifestPath := filepath.Join(path, "edgeflow.json")
	if _, err := os.Stat(manifestPath); err == nil {
		return manifestPath
	}

	// Check subdirectories (for extracted archives with root folder like "package/")
	entries, err := os.ReadDir(path)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			nestedPath := filepath.Join(path, entry.Name(), "edgeflow.json")
			if _, err := os.Stat(nestedPath); err == nil {
				return nestedPath
			}
		}
	}

	return ""
}

// ValidatePlatform checks if the module is compatible with the current platform
func ValidatePlatform(manifest *EdgeFlowManifest, currentPlatform, currentArch string) error {
	// Check platform
	if manifest.Platform != "" && manifest.Platform != "all" {
		if manifest.Platform != currentPlatform {
			return fmt.Errorf("module requires platform %s, but running on %s", manifest.Platform, currentPlatform)
		}
	}

	// Check architecture
	if len(manifest.Arch) > 0 {
		archValid := false
		for _, arch := range manifest.Arch {
			if arch == currentArch || arch == "all" {
				archValid = true
				break
			}
		}
		if !archValid {
			return fmt.Errorf("module requires architecture %v, but running on %s", manifest.Arch, currentArch)
		}
	}

	return nil
}
