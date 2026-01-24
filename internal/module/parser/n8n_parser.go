package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// N8NParser parses n8n modules
type N8NParser struct{}

// NewN8NParser creates a new n8n parser
func NewN8NParser() *N8NParser {
	return &N8NParser{}
}

// Format returns the module format this parser handles
func (p *N8NParser) Format() ModuleFormat {
	return FormatN8N
}

// CanParse checks if this parser can handle the given module
func (p *N8NParser) CanParse(path string) bool {
	format, err := DetectFormat(path)
	if err != nil {
		return false
	}
	return format == FormatN8N
}

// Parse parses an n8n module and returns metadata
func (p *N8NParser) Parse(path string) (*ModuleInfo, error) {
	// Handle archives
	if isArchive(path) {
		extractPath, err := extractArchive(path)
		if err != nil {
			return nil, fmt.Errorf("failed to extract: %w", err)
		}
		defer os.RemoveAll(extractPath)
		path = extractPath
	}

	// Find package.json
	pkgPath := findPackageJSON(path)
	if pkgPath == "" {
		return nil, ErrMissingPackageJSON
	}

	// Read and parse package.json
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("invalid package.json: %w", err)
	}

	basePath := filepath.Dir(pkgPath)

	// Create module info
	info := &ModuleInfo{
		Format:      FormatN8N,
		Name:        pkg.Name,
		Version:     pkg.Version,
		Description: pkg.Description,
		Author:      ParseAuthor(pkg.Author),
		License:     pkg.License,
		Homepage:    pkg.Homepage,
		Repository:  ParseRepository(pkg.Repository),
		Keywords:    pkg.Keywords,
		SourcePath:  basePath,
		Nodes:       []NodeInfo{},
		Config:      make(map[string]interface{}),
	}

	// Parse nodes from n8n config
	if pkg.N8N != nil && len(pkg.N8N.Nodes) > 0 {
		for _, nodeFile := range pkg.N8N.Nodes {
			nodeInfo, err := p.parseNodeFile(basePath, nodeFile)
			if err != nil {
				continue
			}
			info.Nodes = append(info.Nodes, *nodeInfo)
		}
	}

	// If no nodes found via config, scan for node files
	if len(info.Nodes) == 0 {
		info.Nodes = p.scanForNodes(basePath)
	}

	return info, nil
}

// N8NNodeDescription represents n8n node description
type N8NNodeDescription struct {
	DisplayName     string            `json:"displayName"`
	Name            string            `json:"name"`
	Group           []string          `json:"group"`
	Version         int               `json:"version"`
	Description     string            `json:"description"`
	Defaults        map[string]interface{} `json:"defaults"`
	Inputs          []string          `json:"inputs"`
	Outputs         []string          `json:"outputs"`
	Properties      []N8NProperty     `json:"properties"`
	Icon            string            `json:"icon"`
	Color           string            `json:"color"`
}

// N8NProperty represents an n8n node property
type N8NProperty struct {
	DisplayName     string        `json:"displayName"`
	Name            string        `json:"name"`
	Type            string        `json:"type"`
	Default         interface{}   `json:"default"`
	Required        bool          `json:"required"`
	Description     string        `json:"description"`
	Options         []N8NOption   `json:"options,omitempty"`
	DisplayOptions  interface{}   `json:"displayOptions,omitempty"`
}

// N8NOption represents an option for a property
type N8NOption struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
}

// parseNodeFile parses an n8n node definition file
func (p *N8NParser) parseNodeFile(basePath, nodeFile string) (*NodeInfo, error) {
	fullPath := filepath.Join(basePath, nodeFile)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// Try with dist prefix
		fullPath = filepath.Join(basePath, "dist", nodeFile)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("node file not found: %s", nodeFile)
		}
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read node file: %w", err)
	}

	content := string(data)
	node := &NodeInfo{
		Type:       strings.TrimSuffix(filepath.Base(nodeFile), filepath.Ext(nodeFile)),
		SourceFile: nodeFile,
		Category:   "function",
		Inputs:     1,
		Outputs:    1,
		Properties: []PropertyInfo{},
	}

	// Parse TypeScript/JavaScript for n8n node description
	p.parseN8NSource(node, content)

	return node, nil
}

// parseN8NSource extracts node info from n8n TypeScript/JavaScript source
func (p *N8NParser) parseN8NSource(node *NodeInfo, content string) {
	// Look for description object
	descPattern := regexp.MustCompile(`description\s*[=:]\s*\{([\s\S]*?)\n\t?\}[,;]`)
	if matches := descPattern.FindStringSubmatch(content); len(matches) > 1 {
		descContent := matches[1]

		// Extract displayName
		displayNamePattern := regexp.MustCompile(`displayName\s*:\s*["']([^"']+)["']`)
		if m := displayNamePattern.FindStringSubmatch(descContent); len(m) > 1 {
			node.Name = m[1]
		}

		// Extract name
		namePattern := regexp.MustCompile(`name\s*:\s*["']([^"']+)["']`)
		if m := namePattern.FindStringSubmatch(descContent); len(m) > 1 {
			node.Type = m[1]
		}

		// Extract description
		descTextPattern := regexp.MustCompile(`description\s*:\s*["']([^"']+)["']`)
		if m := descTextPattern.FindStringSubmatch(descContent); len(m) > 1 {
			node.Description = m[1]
		}

		// Extract group/category
		groupPattern := regexp.MustCompile(`group\s*:\s*\[["']([^"']+)["']\]`)
		if m := groupPattern.FindStringSubmatch(descContent); len(m) > 1 {
			node.Category = m[1]
		}

		// Extract icon
		iconPattern := regexp.MustCompile(`icon\s*:\s*["']([^"']+)["']`)
		if m := iconPattern.FindStringSubmatch(descContent); len(m) > 1 {
			node.Icon = m[1]
		}

		// Extract color
		colorPattern := regexp.MustCompile(`color\s*:\s*["']([^"']+)["']`)
		if m := colorPattern.FindStringSubmatch(descContent); len(m) > 1 {
			node.Color = m[1]
		}

		// Count inputs
		inputsPattern := regexp.MustCompile(`inputs\s*:\s*\[(.*?)\]`)
		if m := inputsPattern.FindStringSubmatch(descContent); len(m) > 1 {
			node.Inputs = countArrayItems(m[1])
		}

		// Count outputs
		outputsPattern := regexp.MustCompile(`outputs\s*:\s*\[(.*?)\]`)
		if m := outputsPattern.FindStringSubmatch(descContent); len(m) > 1 {
			node.Outputs = countArrayItems(m[1])
		}
	}

	// Parse properties array
	propsPattern := regexp.MustCompile(`properties\s*:\s*\[([\s\S]*?)\n\t?\][,;]?`)
	if matches := propsPattern.FindStringSubmatch(content); len(matches) > 1 {
		node.Properties = p.parseN8NProperties(matches[1])
	}
}

// parseN8NProperties parses n8n property definitions
func (p *N8NParser) parseN8NProperties(propsContent string) []PropertyInfo {
	props := []PropertyInfo{}

	// Match individual property objects
	propPattern := regexp.MustCompile(`\{\s*([\s\S]*?)\s*\}[,\s]*`)
	propMatches := propPattern.FindAllStringSubmatch(propsContent, -1)

	for _, match := range propMatches {
		if len(match) < 2 {
			continue
		}

		propDef := match[1]
		prop := PropertyInfo{
			Type: "string",
		}

		// Extract displayName
		displayNamePattern := regexp.MustCompile(`displayName\s*:\s*["']([^"']+)["']`)
		if m := displayNamePattern.FindStringSubmatch(propDef); len(m) > 1 {
			prop.Label = m[1]
		}

		// Extract name
		namePattern := regexp.MustCompile(`name\s*:\s*["']([^"']+)["']`)
		if m := namePattern.FindStringSubmatch(propDef); len(m) > 1 {
			prop.Name = m[1]
		}

		// Extract type
		typePattern := regexp.MustCompile(`type\s*:\s*["']([^"']+)["']`)
		if m := typePattern.FindStringSubmatch(propDef); len(m) > 1 {
			prop.Type = mapN8NType(m[1])
		}

		// Extract description
		descPattern := regexp.MustCompile(`description\s*:\s*["']([^"']+)["']`)
		if m := descPattern.FindStringSubmatch(propDef); len(m) > 1 {
			prop.Description = m[1]
		}

		// Extract required
		requiredPattern := regexp.MustCompile(`required\s*:\s*(true|false)`)
		if m := requiredPattern.FindStringSubmatch(propDef); len(m) > 1 {
			prop.Required = m[1] == "true"
		}

		// Extract default
		defaultPattern := regexp.MustCompile(`default\s*:\s*["']?([^"',}\s]+)["']?`)
		if m := defaultPattern.FindStringSubmatch(propDef); len(m) > 1 {
			prop.Default = m[1]
		}

		if prop.Name != "" {
			props = append(props, prop)
		}
	}

	return props
}

// scanForNodes scans directory for n8n node files
func (p *N8NParser) scanForNodes(basePath string) []NodeInfo {
	nodes := []NodeInfo{}

	// Common locations for n8n node files
	searchDirs := []string{
		basePath,
		filepath.Join(basePath, "nodes"),
		filepath.Join(basePath, "dist", "nodes"),
		filepath.Join(basePath, "src", "nodes"),
	}

	for _, dir := range searchDirs {
		p.scanDirectory(dir, basePath, &nodes)
	}

	return nodes
}

// scanDirectory recursively scans for n8n node files
func (p *N8NParser) scanDirectory(dir, basePath string, nodes *[]NodeInfo) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recursively scan subdirectories
			p.scanDirectory(fullPath, basePath, nodes)
			continue
		}

		name := entry.Name()

		// Look for n8n node patterns
		isNodeFile := strings.HasSuffix(name, ".node.ts") ||
			strings.HasSuffix(name, ".node.js") ||
			(strings.HasSuffix(name, ".ts") && !strings.Contains(name, ".test.") && !strings.Contains(name, ".spec."))

		if !isNodeFile {
			continue
		}

		relPath, _ := filepath.Rel(basePath, fullPath)
		nodeInfo, err := p.parseNodeFile(basePath, relPath)
		if err == nil && nodeInfo.Name != "" {
			*nodes = append(*nodes, *nodeInfo)
		}
	}
}

// mapN8NType maps n8n property types to EdgeFlow types
func mapN8NType(n8nType string) string {
	typeMap := map[string]string{
		"string":     "string",
		"number":     "number",
		"boolean":    "boolean",
		"options":    "select",
		"multiOptions": "multiselect",
		"collection": "object",
		"fixedCollection": "object",
		"json":       "json",
		"color":      "color",
		"dateTime":   "datetime",
		"hidden":     "hidden",
	}

	if mapped, ok := typeMap[n8nType]; ok {
		return mapped
	}
	return "string"
}

// countArrayItems counts comma-separated items in an array string
func countArrayItems(arrayContent string) int {
	if strings.TrimSpace(arrayContent) == "" {
		return 0
	}
	// Count quoted strings
	items := regexp.MustCompile(`["'][^"']+["']`).FindAllString(arrayContent, -1)
	if len(items) > 0 {
		return len(items)
	}
	// Count by commas
	return len(strings.Split(arrayContent, ","))
}
