//go:build integration
// +build integration

package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests that download real Node-RED packages
// Run with: go test -tags=integration ./internal/module/parser/...

func downloadNPMTarball(packageName string, destDir string) (string, error) {
	// Get package info from npm registry
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", packageName))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var pkgInfo struct {
		DistTags map[string]string `json:"dist-tags"`
		Versions map[string]struct {
			Dist struct {
				Tarball string `json:"tarball"`
			} `json:"dist"`
		} `json:"versions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pkgInfo); err != nil {
		return "", err
	}

	latest := pkgInfo.DistTags["latest"]
	tarballURL := pkgInfo.Versions[latest].Dist.Tarball

	// Download tarball
	resp, err = http.Get(tarballURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Save to temp file
	tmpFile, err := os.CreateTemp(destDir, "module_*.tgz")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

func TestParseNodeRedContribTelegrambot(t *testing.T) {
	tmpDir := t.TempDir()

	// Download node-red-contrib-telegrambot
	tarball, err := downloadNPMTarball("node-red-contrib-telegrambot", tmpDir)
	require.NoError(t, err)
	defer os.Remove(tarball)

	t.Logf("Downloaded tarball: %s", tarball)

	// Detect format
	format, err := DetectFormat(tarball)
	require.NoError(t, err)
	assert.Equal(t, FormatNodeRED, format)

	// Parse module
	parser := NewNodeRedParser()
	info, err := parser.Parse(tarball)
	require.NoError(t, err)

	// Verify module info
	assert.Equal(t, "node-red-contrib-telegrambot", info.Name)
	assert.NotEmpty(t, info.Version)
	assert.NotEmpty(t, info.Description)
	assert.Equal(t, FormatNodeRED, info.Format)

	t.Logf("Module: %s v%s", info.Name, info.Version)
	t.Logf("Description: %s", info.Description)
	t.Logf("Author: %s", info.Author)
	t.Logf("Nodes found: %d", len(info.Nodes))

	// Should have nodes
	assert.Greater(t, len(info.Nodes), 0, "Should have at least one node")

	for _, node := range info.Nodes {
		t.Logf("  - Node: %s (type: %s, category: %s)", node.Name, node.Type, node.Category)
		t.Logf("    Inputs: %d, Outputs: %d", node.Inputs, node.Outputs)
		if len(node.Properties) > 0 {
			t.Logf("    Properties: %d", len(node.Properties))
			for _, prop := range node.Properties {
				t.Logf("      - %s (%s): default=%v, required=%v", prop.Name, prop.Type, prop.Default, prop.Required)
			}
		}
	}
}

func TestParseNodeRedNodeRandom(t *testing.T) {
	tmpDir := t.TempDir()

	// Download node-red-node-random (simple, well-known package)
	tarball, err := downloadNPMTarball("node-red-node-random", tmpDir)
	require.NoError(t, err)
	defer os.Remove(tarball)

	// Parse module
	parser := NewNodeRedParser()
	info, err := parser.Parse(tarball)
	require.NoError(t, err)

	assert.Equal(t, "node-red-node-random", info.Name)
	assert.NotEmpty(t, info.Version)
	assert.Equal(t, FormatNodeRED, info.Format)

	t.Logf("Module: %s v%s", info.Name, info.Version)
	t.Logf("Nodes: %d", len(info.Nodes))

	// Verify at least basic node info
	assert.Greater(t, len(info.Nodes), 0)
	if len(info.Nodes) > 0 {
		node := info.Nodes[0]
		t.Logf("First node: %s (type: %s)", node.Name, node.Type)
		assert.NotEmpty(t, node.Type)
	}
}

func TestParseNodeRedNodePing(t *testing.T) {
	tmpDir := t.TempDir()

	// Download node-red-node-ping
	tarball, err := downloadNPMTarball("node-red-node-ping", tmpDir)
	require.NoError(t, err)
	defer os.Remove(tarball)

	// Parse
	parser := NewNodeRedParser()
	info, err := parser.Parse(tarball)
	require.NoError(t, err)

	assert.Equal(t, "node-red-node-ping", info.Name)
	t.Logf("Module: %s v%s - %d nodes", info.Name, info.Version, len(info.Nodes))

	for _, node := range info.Nodes {
		t.Logf("  Node: %s (type: %s, inputs: %d, outputs: %d)",
			node.Name, node.Type, node.Inputs, node.Outputs)
	}
}

func TestParseNodeRedContribMqttBroker(t *testing.T) {
	tmpDir := t.TempDir()

	// Download node-red-contrib-aedes
	tarball, err := downloadNPMTarball("node-red-contrib-aedes", tmpDir)
	require.NoError(t, err)
	defer os.Remove(tarball)

	// Parse
	parser := NewNodeRedParser()
	info, err := parser.Parse(tarball)
	require.NoError(t, err)

	t.Logf("Module: %s v%s", info.Name, info.Version)
	t.Logf("Description: %s", info.Description)
	t.Logf("Nodes: %d", len(info.Nodes))

	// Verify nodes and properties
	for _, node := range info.Nodes {
		t.Logf("  Node: %s", node.Type)
		t.Logf("    Category: %s", node.Category)
		t.Logf("    Inputs: %d, Outputs: %d", node.Inputs, node.Outputs)
		t.Logf("    Properties: %d", len(node.Properties))
	}
}

func TestParseMultipleNodesPackage(t *testing.T) {
	tmpDir := t.TempDir()

	// Download node-red-contrib-influxdb which has multiple explicit node entries
	tarball, err := downloadNPMTarball("node-red-dashboard", tmpDir)
	require.NoError(t, err)
	defer os.Remove(tarball)

	parser := NewNodeRedParser()
	info, err := parser.Parse(tarball)
	require.NoError(t, err)

	t.Logf("Module: %s v%s", info.Name, info.Version)
	t.Logf("Total nodes: %d", len(info.Nodes))

	// This package has many nodes
	for i, node := range info.Nodes {
		t.Logf("  %d. %s (type: %s, category: %s)", i+1, node.Name, node.Type, node.Category)
	}

	// Should have at least some nodes
	assert.GreaterOrEqual(t, len(info.Nodes), 1, "Should have at least one node")
}

func TestParseBundledModuleCorrectly(t *testing.T) {
	// Test that bundled modules (single entry point) are parsed correctly
	tmpDir := t.TempDir()

	// Home Assistant uses bundled approach: "all": "dist/index.js"
	tarball, err := downloadNPMTarball("node-red-contrib-home-assistant-websocket", tmpDir)
	require.NoError(t, err)
	defer os.Remove(tarball)

	parser := NewNodeRedParser()
	info, err := parser.Parse(tarball)
	require.NoError(t, err)

	t.Logf("Module: %s v%s", info.Name, info.Version)
	t.Logf("Description: %s", info.Description)
	t.Logf("Nodes (entry points): %d", len(info.Nodes))

	// For bundled modules, we at least get the entry point
	assert.GreaterOrEqual(t, len(info.Nodes), 1)
	assert.Equal(t, "node-red-contrib-home-assistant-websocket", info.Name)
	assert.NotEmpty(t, info.Version)

	// Module metadata should be parsed correctly
	assert.NotEmpty(t, info.Description)
}

func TestNodePropertiesToPropertySchema(t *testing.T) {
	tmpDir := t.TempDir()

	// Download a simple package
	tarball, err := downloadNPMTarball("node-red-node-random", tmpDir)
	require.NoError(t, err)
	defer os.Remove(tarball)

	parser := NewNodeRedParser()
	info, err := parser.Parse(tarball)
	require.NoError(t, err)

	// Check that PropertyInfo can map to our PropertySchema
	for _, node := range info.Nodes {
		t.Logf("Node: %s", node.Type)
		for _, prop := range node.Properties {
			// These should be compatible with our PropertySchema structure
			assert.NotEmpty(t, prop.Name, "Property should have name")
			assert.NotEmpty(t, prop.Label, "Property should have label")
			assert.NotEmpty(t, prop.Type, "Property should have type")

			t.Logf("  Property: name=%s, label=%s, type=%s, default=%v, required=%v",
				prop.Name, prop.Label, prop.Type, prop.Default, prop.Required)
		}
	}
}

func TestExtractAndVerifyStructure(t *testing.T) {
	tmpDir := t.TempDir()

	tarball, err := downloadNPMTarball("node-red-node-random", tmpDir)
	require.NoError(t, err)

	// Extract manually to verify structure
	extractDir := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(extractDir, 0755)

	err = extractTarGz(tarball, extractDir)
	require.NoError(t, err)

	// List extracted files
	t.Log("Extracted structure:")
	filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(extractDir, path)
		if info.IsDir() {
			t.Logf("  [DIR]  %s", relPath)
		} else {
			t.Logf("  [FILE] %s (%d bytes)", relPath, info.Size())
		}
		return nil
	})

	// Find package.json
	pkgPath := findPackageJSON(extractDir)
	assert.NotEmpty(t, pkgPath, "Should find package.json")
	t.Logf("Found package.json at: %s", pkgPath)
}
