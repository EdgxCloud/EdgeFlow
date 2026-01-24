package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetermineExtension(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		contentType string
		expected    string
	}{
		{"zip url", "https://example.com/module.zip", "", ".zip"},
		{"tgz url", "https://example.com/module.tgz", "", ".tgz"},
		{"tar.gz url", "https://example.com/module.tar.gz", "", ".tgz"},
		{"zip content-type", "https://example.com/module", "application/zip", ".zip"},
		{"gzip content-type", "https://example.com/module", "application/gzip", ".tgz"},
		{"x-gzip content-type", "https://example.com/module", "application/x-gzip", ".tgz"},
		{"default to tgz", "https://example.com/module", "application/octet-stream", ".tgz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineExtension(tt.url, tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseNPMPackageName(t *testing.T) {
	tests := []struct {
		input           string
		expectedName    string
		expectedVersion string
	}{
		{"node-red-contrib-test", "node-red-contrib-test", ""},
		{"node-red-contrib-test@1.0.0", "node-red-contrib-test", "1.0.0"},
		{"node-red-contrib-test@latest", "node-red-contrib-test", "latest"},
		{"@scope/package", "@scope/package", ""},
		{"@scope/package@2.0.0", "@scope/package", "2.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, version := parseNPMPackageName(tt.input)
			assert.Equal(t, tt.expectedName, name)
			assert.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestParseGitHubRepo(t *testing.T) {
	tests := []struct {
		input          string
		expectedOwner  string
		expectedRepo   string
		expectedBranch string
	}{
		{"owner/repo", "owner", "repo", ""},
		{"owner/repo@main", "owner", "repo", "main"},
		{"owner/repo@develop", "owner", "repo", "develop"},
		{"my-org/my-repo@v1.0.0", "my-org", "my-repo", "v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			owner, repo, branch := parseGitHubRepo(tt.input)
			assert.Equal(t, tt.expectedOwner, owner)
			assert.Equal(t, tt.expectedRepo, repo)
			assert.Equal(t, tt.expectedBranch, branch)
		})
	}
}

func TestCopyToTemp(t *testing.T) {
	tmpDir := t.TempDir()

	content := "test module content"
	reader := strings.NewReader(content)

	tmpPath, err := CopyToTemp(reader, tmpDir, "module.zip")
	require.NoError(t, err)
	defer os.Remove(tmpPath)

	// Verify file was created
	assert.FileExists(t, tmpPath)
	assert.True(t, strings.HasPrefix(filepath.Base(tmpPath), "module.zip_"))

	// Verify content
	data, err := os.ReadFile(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestDownloadModule(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("PK\x03\x04fake zip content"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	// Test download - URL ends with .zip so extension should be .zip
	path, err := downloadModule(server.URL+"/module.zip", tmpDir)
	require.NoError(t, err)
	defer os.Remove(path)

	// Verify file was downloaded
	assert.FileExists(t, path)
	// The file is saved with prefix "module.zip_*" pattern
	assert.Contains(t, path, "module.zip")

	// Verify content
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "PK")
}

func TestDownloadModuleNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	_, err := downloadModule(server.URL+"/notfound.zip", tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestDownloadNPMPackage(t *testing.T) {
	// Create mock npm registry
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "node-red-contrib-test") {
			// Registry metadata
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"name": "node-red-contrib-test",
				"dist-tags": {"latest": "1.0.0"},
				"versions": {
					"1.0.0": {
						"version": "1.0.0",
						"dist": {
							"tarball": "` + "http://" + r.Host + `/tarball.tgz",
							"shasum": "abc123"
						}
					}
				}
			}`))
		} else if strings.Contains(r.URL.Path, "tarball.tgz") {
			// Tarball
			w.Header().Set("Content-Type", "application/gzip")
			w.Write([]byte("\x1f\x8bfake gzip content"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Note: This test would require modifying the registry URL in downloadNPMPackage
	// For a real test, we'd need to inject the server URL
	t.Skip("Requires injectable registry URL - testing parsing functions instead")
}

func TestDownloadGitHubRepoInvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := downloadGitHubRepo("invalid", tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GitHub repo format")
}
