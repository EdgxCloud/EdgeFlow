//go:build !linux

package gpio

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// InterruptNode stub for non-Linux platforms
type InterruptNode struct{}

// NewInterruptExecutor creates a stub interrupt node executor
func NewInterruptExecutor() node.Executor {
	return &InterruptNode{}
}

func (n *InterruptNode) Init(config map[string]interface{}) error {
	return fmt.Errorf("interrupt node requires Linux with GPIO support")
}

func (n *InterruptNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	return node.Message{}, fmt.Errorf("interrupt node requires Linux with GPIO support")
}

func (n *InterruptNode) Cleanup() error {
	return nil
}
