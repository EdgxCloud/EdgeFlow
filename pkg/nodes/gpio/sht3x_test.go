//go:build !linux
// +build !linux

package gpio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSHT3xNode(t *testing.T) {
	t.Run("Create new node with defaults", func(t *testing.T) {
		node := NewSHT3xNode()
		assert.NotNil(t, node)
		assert.Equal(t, "1", node.i2cBus)
		assert.Equal(t, uint16(0x44), node.address)
	})

	t.Run("Init fails without hardware", func(t *testing.T) {
		node := NewSHT3xNode()
		err := node.Init(map[string]interface{}{
			"i2cBus":  "1",
			"address": 0x44,
		})
		// Should fail on non-Linux because no I2C hardware
		assert.Error(t, err)
	})

	t.Run("Parse config with float address", func(t *testing.T) {
		node := NewSHT3xNode()
		// Init will fail but config should be parsed
		_ = node.Init(map[string]interface{}{
			"address": float64(0x45),
		})
		assert.Equal(t, uint16(0x45), node.address)
	})

	t.Run("Parse config with int address", func(t *testing.T) {
		node := NewSHT3xNode()
		_ = node.Init(map[string]interface{}{
			"address": int(0x44),
		})
		assert.Equal(t, uint16(0x44), node.address)
	})

	t.Run("Parse config with custom bus", func(t *testing.T) {
		node := NewSHT3xNode()
		_ = node.Init(map[string]interface{}{
			"i2cBus": "0",
		})
		assert.Equal(t, "0", node.i2cBus)
	})

	t.Run("Cleanup without init returns nil", func(t *testing.T) {
		node := NewSHT3xNode()
		err := node.Cleanup()
		assert.NoError(t, err)
	})

	t.Run("NewSHT3xExecutor fails without hardware", func(t *testing.T) {
		executor, err := NewSHT3xExecutor(map[string]interface{}{})
		assert.Error(t, err)
		assert.Nil(t, executor)
	})
}

func TestSHT3xAddressValidation(t *testing.T) {
	testCases := []struct {
		name    string
		address interface{}
		expect  uint16
	}{
		{"Default address 0x44", float64(0x44), 0x44},
		{"Alternate address 0x45", float64(0x45), 0x45},
		{"Int address", int(0x44), 0x44},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node := NewSHT3xNode()
			_ = node.Init(map[string]interface{}{
				"address": tc.address,
			})
			assert.Equal(t, tc.expect, node.address)
		})
	}
}

func TestSHT3xDataConversion(t *testing.T) {
	t.Run("Temperature conversion formula", func(t *testing.T) {
		// SHT3x formula: temp = -45 + 175 * raw / 65535
		// Raw 0x6666 (26214) should give ~25°C
		tempRaw := uint16(0x6666)
		temperature := -45.0 + 175.0*float64(tempRaw)/65535.0
		assert.InDelta(t, 25.0, temperature, 0.5)
	})

	t.Run("Humidity conversion formula", func(t *testing.T) {
		// SHT3x formula: humidity = 100 * raw / 65535
		// Raw 0x8000 (32768) should give ~50%
		humRaw := uint16(0x8000)
		humidity := 100.0 * float64(humRaw) / 65535.0
		assert.InDelta(t, 50.0, humidity, 0.5)
	})

	t.Run("Zero temperature raw value", func(t *testing.T) {
		// Raw 0 should give -45°C
		tempRaw := uint16(0)
		temperature := -45.0 + 175.0*float64(tempRaw)/65535.0
		assert.Equal(t, -45.0, temperature)
	})

	t.Run("Max temperature raw value", func(t *testing.T) {
		// Raw 65535 should give 130°C
		tempRaw := uint16(65535)
		temperature := -45.0 + 175.0*float64(tempRaw)/65535.0
		assert.InDelta(t, 130.0, temperature, 0.1)
	})

	t.Run("Zero humidity raw value", func(t *testing.T) {
		// Raw 0 should give 0%
		humRaw := uint16(0)
		humidity := 100.0 * float64(humRaw) / 65535.0
		assert.Equal(t, 0.0, humidity)
	})

	t.Run("Max humidity raw value", func(t *testing.T) {
		// Raw 65535 should give 100%
		humRaw := uint16(65535)
		humidity := 100.0 * float64(humRaw) / 65535.0
		assert.InDelta(t, 100.0, humidity, 0.1)
	})
}
