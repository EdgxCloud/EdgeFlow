//go:build !linux

package gpio

import (
	"context"
	"fmt"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// OneWireNode stub for non-Linux platforms
type OneWireNode struct{}

// NewOneWireExecutor creates a stub one-wire node executor
func NewOneWireExecutor() node.Executor {
	return &OneWireNode{}
}

func (n *OneWireNode) Init(config map[string]interface{}) error {
	return fmt.Errorf("one-wire node requires Linux with 1-Wire support")
}

func (n *OneWireNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	return node.Message{}, fmt.Errorf("one-wire node requires Linux with 1-Wire support")
}

func (n *OneWireNode) Cleanup() error {
	return nil
}
