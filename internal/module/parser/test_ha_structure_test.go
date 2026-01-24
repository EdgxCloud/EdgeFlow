//go:build integration
// +build integration

package parser

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestHomeAssistantPackageStructure(t *testing.T) {
	tmpDir := t.TempDir()

	// Download Home Assistant package
	resp, err := http.Get("https://registry.npmjs.org/node-red-contrib-home-assistant-websocket")
	if err != nil {
		t.Fatalf("Failed to get package info: %v", err)
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
		t.Fatalf("Failed to decode package info: %v", err)
	}

	latest := pkgInfo.DistTags["latest"]
	tarballURL := pkgInfo.Versions[latest].Dist.Tarball

	t.Logf("Downloading version %s from %s", latest, tarballURL)

	// Download tarball
	resp, err = http.Get(tarballURL)
	if err != nil {
		t.Fatalf("Failed to download tarball: %v", err)
	}
	defer resp.Body.Close()

	// Save to temp file
	tarball := filepath.Join(tmpDir, "package.tgz")
	f, _ := os.Create(tarball)
	io.Copy(f, resp.Body)
	f.Close()

	// Extract
	extractDir := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(extractDir, 0755)
	if err := extractTarGz(tarball, extractDir); err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	// Find and read package.json
	pkgPath := findPackageJSON(extractDir)
	t.Logf("Package.json at: %s", pkgPath)

	data, _ := os.ReadFile(pkgPath)
	var pkg PackageJSON
	json.Unmarshal(data, &pkg)

	t.Logf("Package name: %s", pkg.Name)
	t.Logf("Version: %s", pkg.Version)

	if pkg.NodeRed != nil {
		t.Logf("Node-RED config found!")
		t.Logf("Node-RED version: %s", pkg.NodeRed.Version)
		t.Logf("Nodes in config: %d", len(pkg.NodeRed.Nodes))
		for nodeType, nodeFile := range pkg.NodeRed.Nodes {
			t.Logf("  - %s -> %s", nodeType, nodeFile)
		}
	} else {
		t.Log("No node-red config in package.json")
	}

	// List all .js files
	t.Log("\nAll .js files in package:")
	basePath := filepath.Dir(pkgPath)
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".js" {
			relPath, _ := filepath.Rel(basePath, path)
			t.Logf("  %s (%d bytes)", relPath, info.Size())
		}
		return nil
	})

	// List all .html files
	t.Log("\nAll .html files in package:")
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			relPath, _ := filepath.Rel(basePath, path)
			t.Logf("  %s (%d bytes)", relPath, info.Size())
		}
		return nil
	})

	// Print raw package.json node-red section
	t.Log("\nRaw package.json content (first 2000 chars):")
	if len(data) > 2000 {
		t.Logf("%s...", string(data[:2000]))
	} else {
		t.Log(string(data))
	}
}

func TestSimpleNodeRedPackage(t *testing.T) {
	// Test with node-red-contrib-influxdb which has multiple defined nodes
	tmpDir := t.TempDir()

	tarball, err := downloadNPMTarball("node-red-contrib-influxdb", tmpDir)
	if err != nil {
		t.Fatalf("Failed to download: %v", err)
	}

	// Extract
	extractDir := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(extractDir, 0755)
	if err := extractTarGz(tarball, extractDir); err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	// Find and read package.json
	pkgPath := findPackageJSON(extractDir)
	data, _ := os.ReadFile(pkgPath)
	var pkg PackageJSON
	json.Unmarshal(data, &pkg)

	t.Logf("Package: %s", pkg.Name)

	if pkg.NodeRed != nil && len(pkg.NodeRed.Nodes) > 0 {
		t.Logf("Node-RED nodes defined: %d", len(pkg.NodeRed.Nodes))
		for nodeType, nodeFile := range pkg.NodeRed.Nodes {
			t.Logf("  - %s -> %s", nodeType, nodeFile)
		}
	}

	// Now parse with our parser
	parser := NewNodeRedParser()
	info, err := parser.Parse(tarball)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	t.Logf("Parsed nodes: %d", len(info.Nodes))
	for _, node := range info.Nodes {
		t.Logf("  - %s (type: %s)", node.Name, node.Type)
	}

	// Compare
	if pkg.NodeRed != nil {
		expected := len(pkg.NodeRed.Nodes)
		actual := len(info.Nodes)
		if actual != expected {
			t.Errorf("Expected %d nodes, got %d", expected, actual)
		}
	}
}
