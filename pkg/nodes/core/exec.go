package core

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// ExecNode executes shell commands
type ExecNode struct {
	command       string
	appendPayload bool
	useSpawn      bool
	timeout       int // seconds
}

// NewExecNode creates a new exec node
func NewExecNode() *ExecNode {
	return &ExecNode{
		command:       "",
		appendPayload: false,
		useSpawn:      false,
		timeout:       10,
	}
}

// Init initializes the exec node with configuration
func (n *ExecNode) Init(config map[string]interface{}) error {
	if command, ok := config["command"].(string); ok {
		n.command = command
	}
	if appendPayload, ok := config["appendPayload"].(bool); ok {
		n.appendPayload = appendPayload
	}
	if useSpawn, ok := config["useSpawn"].(bool); ok {
		n.useSpawn = useSpawn
	}
	if timeout, ok := config["timeout"].(float64); ok {
		n.timeout = int(timeout)
	}
	return nil
}

// Execute runs the shell command
func (n *ExecNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if n.command == "" {
		return msg, fmt.Errorf("exec node: no command specified")
	}

	// Build the command
	cmdStr := n.command
	if n.appendPayload && msg.Payload != nil {
		// Append payload to command
		if val, ok := msg.Payload["value"]; ok {
			cmdStr = fmt.Sprintf("%s %v", cmdStr, val)
		} else {
			// Try to stringify the payload
			cmdStr = fmt.Sprintf("%s %v", cmdStr, msg.Payload)
		}
	}

	// Create context with timeout
	timeout := time.Duration(n.timeout) * time.Second
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Prepare command based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd", "/C", cmdStr)
	} else {
		cmd = exec.CommandContext(execCtx, "sh", "-c", cmdStr)
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	// Prepare output
	result := map[string]interface{}{
		"stdout": strings.TrimSpace(stdout.String()),
		"stderr": strings.TrimSpace(stderr.String()),
	}

	if err != nil {
		// Check if it was a timeout
		if execCtx.Err() == context.DeadlineExceeded {
			result["error"] = "command timed out"
			result["code"] = -1
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			result["error"] = err.Error()
			result["code"] = exitErr.ExitCode()
		} else {
			result["error"] = err.Error()
			result["code"] = -1
		}
	} else {
		result["code"] = 0
	}

	msg.Payload = result
	return msg, nil
}

// Cleanup cleans up resources
func (n *ExecNode) Cleanup() error {
	return nil
}
