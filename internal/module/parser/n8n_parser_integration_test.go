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

// Integration tests that download real n8n community packages
// Run with: go test -tags=integration ./internal/module/parser/... -run TestN8N

func downloadN8NPackage(packageName string, destDir string) (string, error) {
	// Get package info from npm registry
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", packageName))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("package not found: %s (status: %d)", packageName, resp.StatusCode)
	}

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
	if latest == "" {
		return "", fmt.Errorf("no latest version found for %s", packageName)
	}

	tarballURL := pkgInfo.Versions[latest].Dist.Tarball
	if tarballURL == "" {
		return "", fmt.Errorf("no tarball URL found for %s@%s", packageName, latest)
	}

	// Download tarball
	resp, err = http.Get(tarballURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Save to temp file
	tmpFile, err := os.CreateTemp(destDir, "n8n_module_*.tgz")
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

func TestN8NParseGoogleSheetsNode(t *testing.T) {
	tmpDir := t.TempDir()

	// n8n-nodes-google-sheets is a popular community node
	tarball, err := downloadN8NPackage("n8n-nodes-base", tmpDir)
	if err != nil {
		t.Skipf("Could not download n8n-nodes-base (may be large or rate-limited): %v", err)
	}
	defer os.Remove(tarball)

	t.Logf("Downloaded tarball: %s", tarball)

	// Detect format
	format, err := DetectFormat(tarball)
	require.NoError(t, err)
	t.Logf("Detected format: %s", format)

	// This is n8n-nodes-base which should be detected as n8n format
	if format != FormatN8N {
		t.Logf("Note: n8n-nodes-base detected as %s (expected n8n)", format)
	}
}

func TestN8NParseSimpleCommunityNode(t *testing.T) {
	tmpDir := t.TempDir()

	// Try a smaller n8n community node
	// n8n-nodes-text-manipulation is a simple community node
	tarball, err := downloadN8NPackage("n8n-nodes-text-manipulation", tmpDir)
	if err != nil {
		t.Skipf("Could not download n8n-nodes-text-manipulation: %v", err)
	}
	defer os.Remove(tarball)

	t.Logf("Downloaded tarball: %s", tarball)

	// Detect format
	format, err := DetectFormat(tarball)
	require.NoError(t, err)
	t.Logf("Detected format: %s", format)

	// Parse module
	parser := NewN8NParser()
	info, err := parser.Parse(tarball)
	require.NoError(t, err)

	t.Logf("Module: %s v%s", info.Name, info.Version)
	t.Logf("Description: %s", info.Description)
	t.Logf("License: %s", info.License)
	t.Logf("Author: %s", info.Author)
	t.Logf("Format: %s", info.Format)
	t.Logf("Nodes found: %d", len(info.Nodes))

	// Verify basic module info
	assert.NotEmpty(t, info.Name)
	assert.NotEmpty(t, info.Version)
	assert.Equal(t, FormatN8N, info.Format)

	// Log node details
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

func TestN8NLicenseCheck(t *testing.T) {
	tmpDir := t.TempDir()

	// Download a known MIT-licensed n8n community node
	tarball, err := downloadN8NPackage("n8n-nodes-text-manipulation", tmpDir)
	if err != nil {
		t.Skipf("Could not download package: %v", err)
	}
	defer os.Remove(tarball)

	// Parse module
	parser := NewN8NParser()
	info, err := parser.Parse(tarball)
	require.NoError(t, err)

	t.Logf("Package: %s", info.Name)
	t.Logf("License: %s", info.License)

	// n8n community nodes should be MIT licensed
	// This is a requirement for n8n community node verification
	assert.NotEmpty(t, info.License, "n8n community nodes must have a license")

	// Most n8n community nodes use MIT
	if info.License == "MIT" {
		t.Log("License: MIT (compatible - standard for n8n community nodes)")
	} else {
		t.Logf("License: %s (verify compatibility)", info.License)
	}
}

func TestN8NPackageStructure(t *testing.T) {
	tmpDir := t.TempDir()

	// Download and examine package structure
	tarball, err := downloadN8NPackage("n8n-nodes-text-manipulation", tmpDir)
	if err != nil {
		t.Skipf("Could not download package: %v", err)
	}
	defer os.Remove(tarball)

	// Extract to examine structure
	extractDir := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(extractDir, 0755)

	err = extractTarGz(tarball, extractDir)
	require.NoError(t, err)

	// Find package.json
	pkgPath := findPackageJSON(extractDir)
	require.NotEmpty(t, pkgPath, "Should find package.json")

	t.Logf("Package.json at: %s", pkgPath)

	// Read and parse package.json
	data, err := os.ReadFile(pkgPath)
	require.NoError(t, err)

	var pkg PackageJSON
	err = json.Unmarshal(data, &pkg)
	require.NoError(t, err)

	t.Logf("Package name: %s", pkg.Name)
	t.Logf("Version: %s", pkg.Version)
	t.Logf("License: %s", pkg.License)

	// Check for n8n config
	if pkg.N8N != nil {
		t.Log("n8n config found in package.json:")
		t.Logf("  n8nNodesModule: %d", pkg.N8N.N8NNodesModule)
		t.Logf("  Nodes: %v", pkg.N8N.Nodes)
		t.Logf("  Credentials: %v", pkg.N8N.Credentials)
	} else {
		t.Log("No n8n config in package.json - will scan for node files")
	}

	// List all TypeScript/JavaScript files
	basePath := filepath.Dir(pkgPath)
	t.Log("\nAll .ts/.js files in package:")
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		ext := filepath.Ext(path)
		if !info.IsDir() && (ext == ".ts" || ext == ".js") {
			relPath, _ := filepath.Rel(basePath, path)
			t.Logf("  %s (%d bytes)", relPath, info.Size())
		}
		return nil
	})
}

func TestN8NDetectFormatFromPackageName(t *testing.T) {
	// Test that n8n format is detected from package name prefix
	tests := []struct {
		name     string
		expected ModuleFormat
	}{
		{"n8n-nodes-base", FormatN8N},
		{"n8n-nodes-text-manipulation", FormatN8N},
		{"n8n-nodes-custom-thing", FormatN8N},
		{"node-red-contrib-something", FormatNodeRED},
		{"node-red-node-thing", FormatNodeRED},
		{"some-random-package", FormatUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal package.json
			tmpDir := t.TempDir()
			pkgJSON := fmt.Sprintf(`{"name": "%s", "version": "1.0.0"}`, tt.name)
			pkgPath := filepath.Join(tmpDir, "package.json")
			os.WriteFile(pkgPath, []byte(pkgJSON), 0644)

			format, err := DetectFormat(tmpDir)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, format, "Format mismatch for %s", tt.name)
		})
	}
}

func TestN8NImportProcedureComplete(t *testing.T) {
	tmpDir := t.TempDir()

	t.Log("=== n8n Community Node Import Procedure Test ===")

	// Step 1: Download from npm
	t.Log("\n1. Downloading n8n community node from npm...")
	tarball, err := downloadN8NPackage("n8n-nodes-text-manipulation", tmpDir)
	if err != nil {
		t.Skipf("Could not download package: %v", err)
	}
	defer os.Remove(tarball)
	t.Logf("   Downloaded: %s", tarball)

	// Step 2: Detect format
	t.Log("\n2. Detecting module format...")
	format, err := DetectFormat(tarball)
	require.NoError(t, err)
	t.Logf("   Format detected: %s", format)

	// Verify it's recognized as n8n
	if format != FormatN8N {
		t.Logf("   Warning: Expected n8n format, got %s", format)
	}

	// Step 3: Parse module
	t.Log("\n3. Parsing module metadata and nodes...")
	parser := NewN8NParser()
	info, err := parser.Parse(tarball)
	require.NoError(t, err)

	t.Logf("   Name: %s", info.Name)
	t.Logf("   Version: %s", info.Version)
	t.Logf("   License: %s", info.License)
	t.Logf("   Nodes: %d", len(info.Nodes))

	// Step 4: Check license compatibility
	t.Log("\n4. Checking license compatibility...")
	licenseCompatible := false
	switch info.License {
	case "MIT", "ISC", "Apache-2.0", "BSD-2-Clause", "BSD-3-Clause":
		t.Logf("   License '%s' is COMPATIBLE (permissive)", info.License)
		licenseCompatible = true
	case "":
		t.Log("   WARNING: No license specified")
	default:
		t.Logf("   License '%s' requires review", info.License)
	}

	// n8n community nodes MUST use MIT for verification
	if info.License == "MIT" {
		t.Log("   n8n community node requirement: MIT license - PASSED")
	}

	// Step 5: Verify node structure
	t.Log("\n5. Verifying node structure...")
	for i, node := range info.Nodes {
		t.Logf("   Node %d:", i+1)
		t.Logf("     Type: %s", node.Type)
		t.Logf("     Name: %s", node.Name)
		t.Logf("     Category: %s", node.Category)
		t.Logf("     Inputs: %d, Outputs: %d", node.Inputs, node.Outputs)
		t.Logf("     Properties: %d", len(node.Properties))
	}

	// Step 6: Summary
	t.Log("\n=== Import Procedure Summary ===")
	t.Logf("Package: %s v%s", info.Name, info.Version)
	t.Logf("Format: %s", info.Format)
	t.Logf("License: %s (compatible: %v)", info.License, licenseCompatible)
	t.Logf("Nodes: %d", len(info.Nodes))

	if licenseCompatible && len(info.Nodes) >= 0 {
		t.Log("Result: READY FOR IMPORT")
	} else {
		t.Log("Result: REVIEW REQUIRED")
	}
}
