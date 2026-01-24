//go:build integration
// +build integration

package api

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests that make real HTTP requests
// Run with: go test -tags=integration ./internal/api/...

func TestDownloadNPMPackageIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Download a small, well-known Node-RED package
	path, err := downloadNPMPackage("node-red-node-ping@0.3.3", tmpDir)
	require.NoError(t, err)
	defer os.Remove(path)

	// Verify file was downloaded
	assert.FileExists(t, path)

	// Check file is not empty
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))

	t.Logf("Downloaded package to: %s (size: %d bytes)", path, info.Size())
}

func TestDownloadNPMPackageLatestVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Download latest version
	path, err := downloadNPMPackage("node-red-node-random", tmpDir)
	require.NoError(t, err)
	defer os.Remove(path)

	assert.FileExists(t, path)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))

	t.Logf("Downloaded latest version to: %s (size: %d bytes)", path, info.Size())
}

func TestDownloadNPMPackageNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to download non-existent package
	_, err := downloadNPMPackage("this-package-definitely-does-not-exist-12345", tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDownloadGitHubRepoIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Download a small public repository
	path, err := downloadGitHubRepo("node-red/node-red-nodegen@master", tmpDir)
	if err != nil {
		// Skip if rate limited
		if strings.Contains(err.Error(), "rate limit") {
			t.Skip("GitHub rate limit exceeded")
		}
		t.Fatalf("Download failed: %v", err)
	}
	defer os.Remove(path)

	assert.FileExists(t, path)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))

	t.Logf("Downloaded repo to: %s (size: %d bytes)", path, info.Size())
}

func TestDownloadGitHubRepoNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := downloadGitHubRepo("this-owner-does-not-exist/this-repo-does-not-exist", tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDownloadModuleFromURL(t *testing.T) {
	tmpDir := t.TempDir()

	// Download from npm tarball URL directly
	url := "https://registry.npmjs.org/node-red-node-random/-/node-red-node-random-0.4.1.tgz"
	path, err := downloadModule(url, tmpDir)
	require.NoError(t, err)
	defer os.Remove(path)

	assert.FileExists(t, path)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))

	t.Logf("Downloaded from URL to: %s (size: %d bytes)", path, info.Size())
}
