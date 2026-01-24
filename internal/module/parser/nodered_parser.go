package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// NodeRedParser parses Node-RED modules
type NodeRedParser struct{}

// NewNodeRedParser creates a new Node-RED parser
func NewNodeRedParser() *NodeRedParser {
	return &NodeRedParser{}
}

// Format returns the module format this parser handles
func (p *NodeRedParser) Format() ModuleFormat {
	return FormatNodeRED
}

// CanParse checks if this parser can handle the given module
func (p *NodeRedParser) CanParse(path string) bool {
	format, err := DetectFormat(path)
	if err != nil {
		return false
	}
	return format == FormatNodeRED
}

// Parse parses a Node-RED module and returns metadata
func (p *NodeRedParser) Parse(path string) (*ModuleInfo, error) {
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
		Format:      FormatNodeRED,
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

	// Parse nodes from node-red config
	if pkg.NodeRed != nil && pkg.NodeRed.Nodes != nil {
		for nodeType, nodeFile := range pkg.NodeRed.Nodes {
			nodeInfo, err := p.parseNodeFile(basePath, nodeType, nodeFile)
			if err != nil {
				// Log warning but continue
				continue
			}
			info.Nodes = append(info.Nodes, *nodeInfo)
		}
	}

	// If no nodes found via config, scan for .js files
	if len(info.Nodes) == 0 {
		info.Nodes = p.scanForNodes(basePath)
	}

	return info, nil
}

// parseNodeFile parses a Node-RED node definition file
func (p *NodeRedParser) parseNodeFile(basePath, nodeType, nodeFile string) (*NodeInfo, error) {
	jsPath := filepath.Join(basePath, nodeFile)
	htmlPath := strings.TrimSuffix(jsPath, ".js") + ".html"

	node := &NodeInfo{
		Type:       nodeType,
		Name:       formatNodeName(nodeType),
		Category:   "function", // Default category
		SourceFile: nodeFile,
		Inputs:     1,
		Outputs:    1,
		Properties: []PropertyInfo{},
	}

	// Try to parse JS file for node registration
	if data, err := os.ReadFile(jsPath); err == nil {
		p.parseJSFile(node, string(data))
	}

	// Try to parse HTML file for UI definition
	if _, err := os.Stat(htmlPath); err == nil {
		node.UIFile = strings.TrimSuffix(nodeFile, ".js") + ".html"
		if data, err := os.ReadFile(htmlPath); err == nil {
			p.parseHTMLFile(node, string(data))
		}
	}

	return node, nil
}

// parseJSFile extracts node info from JavaScript source
func (p *NodeRedParser) parseJSFile(node *NodeInfo, content string) {
	// Look for RED.nodes.registerType call
	// Pattern: RED.nodes.registerType("type-name", { ... })
	registerPattern := regexp.MustCompile(`RED\.nodes\.registerType\s*\(\s*["']([^"']+)["']`)
	if matches := registerPattern.FindStringSubmatch(content); len(matches) > 1 {
		node.Type = matches[1]
		node.Name = formatNodeName(matches[1])
	}

	// Look for category in defaults
	categoryPattern := regexp.MustCompile(`category\s*:\s*["']([^"']+)["']`)
	if matches := categoryPattern.FindStringSubmatch(content); len(matches) > 1 {
		node.Category = matches[1]
	}

	// Look for inputs/outputs
	inputsPattern := regexp.MustCompile(`inputs\s*:\s*(\d+)`)
	if matches := inputsPattern.FindStringSubmatch(content); len(matches) > 1 {
		fmt.Sscanf(matches[1], "%d", &node.Inputs)
	}

	outputsPattern := regexp.MustCompile(`outputs\s*:\s*(\d+)`)
	if matches := outputsPattern.FindStringSubmatch(content); len(matches) > 1 {
		fmt.Sscanf(matches[1], "%d", &node.Outputs)
	}

	// Look for color
	colorPattern := regexp.MustCompile(`color\s*:\s*["']([^"']+)["']`)
	if matches := colorPattern.FindStringSubmatch(content); len(matches) > 1 {
		node.Color = matches[1]
	}

	// Look for icon
	iconPattern := regexp.MustCompile(`icon\s*:\s*["']([^"']+)["']`)
	if matches := iconPattern.FindStringSubmatch(content); len(matches) > 1 {
		node.Icon = matches[1]
	}
}

// parseHTMLFile extracts node info from HTML template
func (p *NodeRedParser) parseHTMLFile(node *NodeInfo, content string) {
	// Look for script template with data-template-name
	templatePattern := regexp.MustCompile(`data-template-name=["']([^"']+)["']`)
	if matches := templatePattern.FindStringSubmatch(content); len(matches) > 1 {
		node.Type = matches[1]
	}

	// Look for help text
	helpPattern := regexp.MustCompile(`data-help-name=["'][^"']+["'][^>]*>([\s\S]*?)</script>`)
	if matches := helpPattern.FindStringSubmatch(content); len(matches) > 1 {
		// Extract first paragraph as description
		descPattern := regexp.MustCompile(`<p>([^<]+)</p>`)
		if descMatches := descPattern.FindStringSubmatch(matches[1]); len(descMatches) > 1 {
			node.Description = strings.TrimSpace(descMatches[1])
		}
	}

	// Parse defaults from script
	defaultsPattern := regexp.MustCompile(`defaults\s*:\s*\{([^}]+)\}`)
	if matches := defaultsPattern.FindStringSubmatch(content); len(matches) > 1 {
		node.Properties = p.parseDefaults(matches[1])
	}

	// Look for category
	categoryPattern := regexp.MustCompile(`category\s*:\s*["']([^"']+)["']`)
	if matches := categoryPattern.FindStringSubmatch(content); len(matches) > 1 {
		node.Category = matches[1]
	}

	// Look for color
	colorPattern := regexp.MustCompile(`color\s*:\s*["']([^"']+)["']`)
	if matches := colorPattern.FindStringSubmatch(content); len(matches) > 1 {
		node.Color = matches[1]
	}

	// Look for label function or string
	labelPattern := regexp.MustCompile(`label\s*:\s*(?:function\s*\(\)\s*\{\s*return\s*["']([^"']+)["']|["']([^"']+)["'])`)
	if matches := labelPattern.FindStringSubmatch(content); len(matches) > 0 {
		for i := 1; i < len(matches); i++ {
			if matches[i] != "" {
				node.Name = matches[i]
				break
			}
		}
	}
}

// parseDefaults parses the defaults object from Node-RED node definition
func (p *NodeRedParser) parseDefaults(defaultsStr string) []PropertyInfo {
	props := []PropertyInfo{}

	// Match property definitions: name: { value: "default", required: true }
	propPattern := regexp.MustCompile(`(\w+)\s*:\s*\{([^}]+)\}`)
	matches := propPattern.FindAllStringSubmatch(defaultsStr, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		propName := match[1]
		propDef := match[2]

		prop := PropertyInfo{
			Name:  propName,
			Label: formatPropertyLabel(propName),
			Type:  "string", // Default type
		}

		// Parse value (default)
		valuePattern := regexp.MustCompile(`value\s*:\s*["']?([^,"'\s}]+)["']?`)
		if valMatch := valuePattern.FindStringSubmatch(propDef); len(valMatch) > 1 {
			prop.Default = valMatch[1]
		}

		// Parse required
		requiredPattern := regexp.MustCompile(`required\s*:\s*(true|false)`)
		if reqMatch := requiredPattern.FindStringSubmatch(propDef); len(reqMatch) > 1 {
			prop.Required = reqMatch[1] == "true"
		}

		// Parse type
		typePattern := regexp.MustCompile(`type\s*:\s*["']?(\w+)["']?`)
		if typeMatch := typePattern.FindStringSubmatch(propDef); len(typeMatch) > 1 {
			prop.Type = typeMatch[1]
		}

		props = append(props, prop)
	}

	return props
}

// scanForNodes scans directory for Node-RED node files
func (p *NodeRedParser) scanForNodes(basePath string) []NodeInfo {
	nodes := []NodeInfo{}

	// Common locations for node files
	searchDirs := []string{
		basePath,
		filepath.Join(basePath, "nodes"),
		filepath.Join(basePath, "src"),
	}

	for _, dir := range searchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasSuffix(name, ".js") {
				continue
			}

			// Skip common non-node files
			if name == "index.js" || strings.HasSuffix(name, ".spec.js") || strings.HasSuffix(name, ".test.js") {
				continue
			}

			relPath, _ := filepath.Rel(basePath, filepath.Join(dir, name))
			nodeType := strings.TrimSuffix(name, ".js")

			nodeInfo, err := p.parseNodeFile(basePath, nodeType, relPath)
			if err == nil {
				nodes = append(nodes, *nodeInfo)
			}
		}
	}

	return nodes
}

// findPackageJSON finds package.json in the given path
func findPackageJSON(path string) string {
	// Direct path
	pkgPath := filepath.Join(path, "package.json")
	if _, err := os.Stat(pkgPath); err == nil {
		return pkgPath
	}

	// One level deeper (for extracted archives)
	entries, err := os.ReadDir(path)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pkgPath = filepath.Join(path, entry.Name(), "package.json")
		if _, err := os.Stat(pkgPath); err == nil {
			return pkgPath
		}
	}

	return ""
}

// formatNodeName converts node type to display name
func formatNodeName(nodeType string) string {
	// Remove common prefixes
	name := strings.TrimPrefix(nodeType, "node-red-")
	name = strings.TrimPrefix(name, "contrib-")
	name = strings.TrimPrefix(name, "node-")

	// Replace separators with spaces
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")

	// Title case
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}

	return strings.Join(words, " ")
}

// formatPropertyLabel converts property name to display label
func formatPropertyLabel(propName string) string {
	// camelCase to Title Case
	var result strings.Builder
	for i, r := range propName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		if i == 0 {
			result.WriteRune(rune(strings.ToUpper(string(r))[0]))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
