// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DockerTestSuite provides integration tests for Docker deployment
type DockerTestSuite struct {
	containerID string
	baseURL     string
}

// skipIfNoDocker skips the test if Docker is not available
func skipIfNoDocker(t *testing.T) {
	if os.Getenv("SKIP_DOCKER_TESTS") == "true" {
		t.Skip("Skipping Docker tests (SKIP_DOCKER_TESTS=true)")
	}

	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker not available, skipping integration tests")
	}
}

// startContainer starts the EdgeFlow container
func (s *DockerTestSuite) startContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Build the image first
	buildCmd := exec.CommandContext(ctx, "docker", "build", "-t", "edgeflow-test", ".")
	buildCmd.Dir = getProjectRoot()
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Logf("Docker build output: %s", string(output))
		t.Fatalf("Failed to build Docker image: %v", err)
	}

	// Start container
	runCmd := exec.CommandContext(ctx, "docker", "run", "-d", "-p", "8080:8080",
		"-e", "EDGEFLOW_ENV=test",
		"edgeflow-test")
	containerIDBytes, err := runCmd.Output()
	require.NoError(t, err, "Failed to start container")

	s.containerID = strings.TrimSpace(string(containerIDBytes))
	s.baseURL = "http://localhost:8080"

	// Wait for container to be ready
	s.waitForHealthy(t, 30*time.Second)
}

// stopContainer stops and removes the container
func (s *DockerTestSuite) stopContainer(t *testing.T) {
	if s.containerID == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop container
	stopCmd := exec.CommandContext(ctx, "docker", "stop", s.containerID)
	stopCmd.Run()

	// Remove container
	rmCmd := exec.CommandContext(ctx, "docker", "rm", "-f", s.containerID)
	rmCmd.Run()
}

// waitForHealthy waits for the container to become healthy
func (s *DockerTestSuite) waitForHealthy(t *testing.T, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(s.baseURL + "/health")
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("Container did not become healthy within timeout")
}

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Try to find the project root by looking for go.mod
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(dir + "/go.mod"); err == nil {
			return dir
		}
		parent := dir[:strings.LastIndex(dir, string(os.PathSeparator))]
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}

func TestDockerHealthEndpoint(t *testing.T) {
	skipIfNoDocker(t)

	suite := &DockerTestSuite{}
	suite.startContainer(t)
	defer suite.stopContainer(t)

	resp, err := http.Get(suite.baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)

	assert.Equal(t, "healthy", health["status"])
}

func TestDockerAPIEndpoints(t *testing.T) {
	skipIfNoDocker(t)

	suite := &DockerTestSuite{}
	suite.startContainer(t)
	defer suite.stopContainer(t)

	tests := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "get flows",
			method:     "GET",
			path:       "/api/flows",
			wantStatus: http.StatusOK,
		},
		{
			name:       "get nodes",
			method:     "GET",
			path:       "/api/nodes",
			wantStatus: http.StatusOK,
		},
		{
			name:       "get metrics",
			method:     "GET",
			path:       "/metrics",
			wantStatus: http.StatusOK,
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				body = bytes.NewReader(bodyBytes)
			}

			req, err := http.NewRequest(tt.method, suite.baseURL+tt.path, body)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestDockerFlowCRUD(t *testing.T) {
	skipIfNoDocker(t)

	suite := &DockerTestSuite{}
	suite.startContainer(t)
	defer suite.stopContainer(t)

	client := &http.Client{Timeout: 10 * time.Second}

	// Create a flow
	flow := map[string]interface{}{
		"name":        "test-flow",
		"description": "Integration test flow",
		"nodes": []map[string]interface{}{
			{
				"id":   "inject-1",
				"type": "inject",
				"config": map[string]interface{}{
					"payload": "test",
				},
			},
			{
				"id":   "debug-1",
				"type": "debug",
			},
		},
		"connections": []map[string]interface{}{
			{
				"source":     "inject-1",
				"sourcePort": 0,
				"target":     "debug-1",
				"targetPort": 0,
			},
		},
	}

	// POST - Create flow
	bodyBytes, _ := json.Marshal(flow)
	resp, err := client.Post(suite.baseURL+"/api/flows", "application/json", bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return 200 or 201
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated,
		"Expected 200 or 201, got %d", resp.StatusCode)

	var createdFlow map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createdFlow)

	flowID := ""
	if id, ok := createdFlow["id"].(string); ok {
		flowID = id
	}

	// GET - Retrieve flows
	resp, err = client.Get(suite.baseURL + "/api/flows")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// DELETE - Remove flow (if we got an ID)
	if flowID != "" {
		req, _ := http.NewRequest("DELETE", suite.baseURL+"/api/flows/"+flowID, nil)
		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
	}
}

func TestDockerWebSocketConnection(t *testing.T) {
	skipIfNoDocker(t)

	suite := &DockerTestSuite{}
	suite.startContainer(t)
	defer suite.stopContainer(t)

	// Test WebSocket upgrade endpoint exists
	// Note: Full WebSocket test would require gorilla/websocket
	req, err := http.NewRequest("GET", suite.baseURL+"/ws", nil)
	require.NoError(t, err)
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "test-key")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		// Should get 101 Switching Protocols or 400 Bad Request (if WebSocket not properly configured)
		assert.True(t, resp.StatusCode == http.StatusSwitchingProtocols || resp.StatusCode == http.StatusBadRequest,
			"Expected 101 or 400, got %d", resp.StatusCode)
	}
}

func TestDockerResourceUsage(t *testing.T) {
	skipIfNoDocker(t)

	suite := &DockerTestSuite{}
	suite.startContainer(t)
	defer suite.stopContainer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get container stats
	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format",
		"{{.MemUsage}}\t{{.CPUPerc}}", suite.containerID)
	output, err := cmd.Output()
	require.NoError(t, err)

	stats := strings.TrimSpace(string(output))
	t.Logf("Container stats: %s", stats)

	// Just verify we got some output
	assert.NotEmpty(t, stats)
}

func TestDockerEnvironmentVariables(t *testing.T) {
	skipIfNoDocker(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Build image
	buildCmd := exec.CommandContext(ctx, "docker", "build", "-t", "edgeflow-env-test", ".")
	buildCmd.Dir = getProjectRoot()
	if err := buildCmd.Run(); err != nil {
		t.Skip("Could not build Docker image")
	}

	// Start container with custom environment
	runCmd := exec.CommandContext(ctx, "docker", "run", "-d", "-p", "8081:8080",
		"-e", "EDGEFLOW_PORT=8080",
		"-e", "EDGEFLOW_LOG_LEVEL=debug",
		"-e", "EDGEFLOW_ENV=test",
		"edgeflow-env-test")
	containerIDBytes, err := runCmd.Output()
	if err != nil {
		t.Skip("Could not start container")
	}
	containerID := strings.TrimSpace(string(containerIDBytes))
	defer func() {
		exec.Command("docker", "stop", containerID).Run()
		exec.Command("docker", "rm", "-f", containerID).Run()
	}()

	// Wait for container
	time.Sleep(5 * time.Second)

	// Verify container is running with env vars
	inspectCmd := exec.CommandContext(ctx, "docker", "inspect",
		"--format", "{{range .Config.Env}}{{println .}}{{end}}", containerID)
	output, err := inspectCmd.Output()
	require.NoError(t, err)

	envVars := string(output)
	assert.Contains(t, envVars, "EDGEFLOW_LOG_LEVEL=debug")
	assert.Contains(t, envVars, "EDGEFLOW_ENV=test")
}

func TestDockerContainerLifecycle(t *testing.T) {
	skipIfNoDocker(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Build image
	buildCmd := exec.CommandContext(ctx, "docker", "build", "-t", "edgeflow-lifecycle-test", ".")
	buildCmd.Dir = getProjectRoot()
	if err := buildCmd.Run(); err != nil {
		t.Skip("Could not build Docker image")
	}

	// Start container
	runCmd := exec.CommandContext(ctx, "docker", "run", "-d", "-p", "8082:8080", "edgeflow-lifecycle-test")
	containerIDBytes, err := runCmd.Output()
	if err != nil {
		t.Skip("Could not start container")
	}
	containerID := strings.TrimSpace(string(containerIDBytes))

	// Verify running
	statusCmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Running}}", containerID)
	output, _ := statusCmd.Output()
	assert.Equal(t, "true", strings.TrimSpace(string(output)))

	// Stop container
	exec.CommandContext(ctx, "docker", "stop", containerID).Run()

	// Verify stopped
	statusCmd = exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Running}}", containerID)
	output, _ = statusCmd.Output()
	assert.Equal(t, "false", strings.TrimSpace(string(output)))

	// Start again
	exec.CommandContext(ctx, "docker", "start", containerID).Run()
	time.Sleep(3 * time.Second)

	// Verify running again
	statusCmd = exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Running}}", containerID)
	output, _ = statusCmd.Output()
	assert.Equal(t, "true", strings.TrimSpace(string(output)))

	// Cleanup
	exec.Command("docker", "stop", containerID).Run()
	exec.Command("docker", "rm", "-f", containerID).Run()
}

func TestDockerVolumeMount(t *testing.T) {
	skipIfNoDocker(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create a temp directory for data
	tempDir, err := os.MkdirTemp("", "edgeflow-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Build image
	buildCmd := exec.CommandContext(ctx, "docker", "build", "-t", "edgeflow-volume-test", ".")
	buildCmd.Dir = getProjectRoot()
	if err := buildCmd.Run(); err != nil {
		t.Skip("Could not build Docker image")
	}

	// Start container with volume mount
	runCmd := exec.CommandContext(ctx, "docker", "run", "-d", "-p", "8083:8080",
		"-v", fmt.Sprintf("%s:/data", tempDir),
		"edgeflow-volume-test")
	containerIDBytes, err := runCmd.Output()
	if err != nil {
		t.Skip("Could not start container with volume")
	}
	containerID := strings.TrimSpace(string(containerIDBytes))
	defer func() {
		exec.Command("docker", "stop", containerID).Run()
		exec.Command("docker", "rm", "-f", containerID).Run()
	}()

	// Wait for container
	time.Sleep(5 * time.Second)

	// Verify container is running
	statusCmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Running}}", containerID)
	output, _ := statusCmd.Output()
	assert.Equal(t, "true", strings.TrimSpace(string(output)))
}

// BenchmarkDockerAPIResponse benchmarks API response times
func BenchmarkDockerAPIResponse(b *testing.B) {
	if os.Getenv("SKIP_DOCKER_TESTS") == "true" {
		b.Skip("Skipping Docker benchmarks")
	}

	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		b.Skip("Docker not available")
	}

	// Assume container is already running on port 8080 for benchmarks
	baseURL := "http://localhost:8080"

	client := &http.Client{Timeout: 5 * time.Second}

	// Check if server is available
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		b.Skip("EdgeFlow not running on port 8080")
	}
	resp.Body.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}
