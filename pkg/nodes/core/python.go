package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// PythonNode executes Python code
type PythonNode struct {
	code          string
	pythonPath    string
	timeout       int // seconds
	useVirtualEnv bool
	venvPath      string
}

// NewPythonNode creates a new python node
func NewPythonNode() *PythonNode {
	return &PythonNode{
		code:          "",
		pythonPath:    "python3",
		timeout:       30,
		useVirtualEnv: false,
		venvPath:      "",
	}
}

// Init initializes the python node with configuration
func (n *PythonNode) Init(config map[string]interface{}) error {
	if code, ok := config["code"].(string); ok {
		n.code = code
	}
	if pythonPath, ok := config["pythonPath"].(string); ok && pythonPath != "" {
		n.pythonPath = pythonPath
	}
	if timeout, ok := config["timeout"].(float64); ok {
		n.timeout = int(timeout)
	}
	if useVirtualEnv, ok := config["useVirtualEnv"].(bool); ok {
		n.useVirtualEnv = useVirtualEnv
	}
	if venvPath, ok := config["venvPath"].(string); ok {
		n.venvPath = venvPath
	}
	return nil
}

// Execute runs the Python code
func (n *PythonNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if n.code == "" {
		return msg, fmt.Errorf("python node: no code specified")
	}

	// Determine Python executable
	pythonExe := n.pythonPath
	if n.useVirtualEnv && n.venvPath != "" {
		if runtime.GOOS == "windows" {
			pythonExe = filepath.Join(n.venvPath, "Scripts", "python.exe")
		} else {
			pythonExe = filepath.Join(n.venvPath, "bin", "python")
		}
	}

	// Create wrapper script that handles msg input/output
	wrapperCode := n.buildWrapperScript(msg)

	// Create temp file for the script
	tmpFile, err := os.CreateTemp("", "edgeflow_python_*.py")
	if err != nil {
		return msg, fmt.Errorf("python node: failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(wrapperCode); err != nil {
		tmpFile.Close()
		return msg, fmt.Errorf("python node: failed to write script: %w", err)
	}
	tmpFile.Close()

	// Create context with timeout
	timeout := time.Duration(n.timeout) * time.Second
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Run Python script
	cmd := exec.CommandContext(execCtx, pythonExe, tmpFile.Name())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// Prepare result
	result := make(map[string]interface{})

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result["error"] = "script timed out"
			result["stderr"] = stderr.String()
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			result["error"] = fmt.Sprintf("script exited with code %d", exitErr.ExitCode())
			result["stderr"] = stderr.String()
		} else {
			result["error"] = err.Error()
			result["stderr"] = stderr.String()
		}
		msg.Payload = result
		return msg, nil
	}

	// Try to parse JSON output from stdout
	output := strings.TrimSpace(stdout.String())
	if output != "" {
		// Look for JSON output marker
		lines := strings.Split(output, "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, "__EDGEFLOW_OUTPUT__:") {
				jsonStr := strings.TrimPrefix(line, "__EDGEFLOW_OUTPUT__:")
				if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
					msg.Payload = result
					return msg, nil
				}
			}
		}
		// No JSON output, return raw stdout
		result["stdout"] = output
	}

	if stderr.Len() > 0 {
		result["stderr"] = stderr.String()
	}

	msg.Payload = result
	return msg, nil
}

// buildWrapperScript creates a Python script that wraps user code
func (n *PythonNode) buildWrapperScript(msg node.Message) string {
	// Serialize message payload to JSON
	payloadJSON, _ := json.Marshal(msg.Payload)
	topicJSON, _ := json.Marshal(msg.Topic)

	wrapper := fmt.Sprintf(`
import json
import sys

# Message object
class Msg:
    def __init__(self):
        self.payload = %s
        self.topic = %s

msg = Msg()

# User code
%s

# Output result
try:
    output = {"payload": msg.payload, "topic": msg.topic}
    print("__EDGEFLOW_OUTPUT__:" + json.dumps(output))
except Exception as e:
    print("__EDGEFLOW_OUTPUT__:" + json.dumps({"error": str(e)}))
`, string(payloadJSON), string(topicJSON), n.code)

	return wrapper
}

// Cleanup cleans up resources
func (n *PythonNode) Cleanup() error {
	return nil
}
