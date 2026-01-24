//go:build !linux
// +build !linux

package gpio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAHT20Node(t *testing.T) {
	t.Run("Create new node with defaults", func(t *testing.T) {
		node := NewAHT20Node()
		assert.NotNil(t, node)
		assert.Equal(t, "1", node.i2cBus)
		assert.Equal(t, uint16(0x38), node.address)
	})

	t.Run("Init fails without hardware", func(t *testing.T) {
		node := NewAHT20Node()
		err := node.Init(map[string]interface{}{
			"i2cBus":  "1",
			"address": 0x38,
		})
		// Should fail on non-Linux because no I2C hardware
		assert.Error(t, err)
	})

	t.Run("Parse config with float address", func(t *testing.T) {
		node := NewAHT20Node()
		// Init will fail but config should be parsed
		_ = node.Init(map[string]interface{}{
			"address": float64(0x39),
		})
		assert.Equal(t, uint16(0x39), node.address)
	})

	t.Run("Parse config with int address", func(t *testing.T) {
		node := NewAHT20Node()
		_ = node.Init(map[string]interface{}{
			"address": int(0x40),
		})
		assert.Equal(t, uint16(0x40), node.address)
	})

	t.Run("Parse config with custom bus", func(t *testing.T) {
		node := NewAHT20Node()
		_ = node.Init(map[string]interface{}{
			"i2cBus": "0",
		})
		assert.Equal(t, "0", node.i2cBus)
	})

	t.Run("Cleanup without init returns nil", func(t *testing.T) {
		node := NewAHT20Node()
		err := node.Cleanup()
		assert.NoError(t, err)
	})

	t.Run("NewAHT20Executor fails without hardware", func(t *testing.T) {
		executor, err := NewAHT20Executor(map[string]interface{}{})
		assert.Error(t, err)
		assert.Nil(t, executor)
	})
}
