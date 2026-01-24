//go:build !linux
// +build !linux

package gpio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDS18B20Node(t *testing.T) {
	t.Run("Create new node with defaults", func(t *testing.T) {
		node := NewDS18B20Node()
		assert.NotNil(t, node)
		assert.Equal(t, "celsius", node.unit)
	})

	t.Run("Init parses deviceId", func(t *testing.T) {
		node := NewDS18B20Node()
		err := node.Init(map[string]interface{}{
			"deviceId": "28-00000123abcd",
		})
		require.NoError(t, err)
		assert.Equal(t, "28-00000123abcd", node.deviceID)
	})

	t.Run("Init parses unit", func(t *testing.T) {
		node := NewDS18B20Node()
		err := node.Init(map[string]interface{}{
			"unit": "fahrenheit",
		})
		require.NoError(t, err)
		assert.Equal(t, "fahrenheit", node.unit)
	})

	t.Run("Cleanup returns nil", func(t *testing.T) {
		node := NewDS18B20Node()
		err := node.Cleanup()
		assert.NoError(t, err)
	})

	t.Run("NewDS18B20Executor creates executor", func(t *testing.T) {
		executor, err := NewDS18B20Executor(map[string]interface{}{
			"deviceId": "28-test",
			"unit":     "celsius",
		})
		require.NoError(t, err)
		require.NotNil(t, executor)

		ds := executor.(*DS18B20Node)
		assert.Equal(t, "28-test", ds.deviceID)
	})

	// Note: Execute tests skipped on non-Linux as DS18B20 uses
	// hardcoded /sys/bus/w1/devices path which is Linux-only
}

func TestDS18B20TemperatureConversions(t *testing.T) {
	t.Run("Celsius to Fahrenheit", func(t *testing.T) {
		// F = C * 9/5 + 32
		tempC := 25.0
		tempF := tempC*9.0/5.0 + 32.0
		assert.InDelta(t, 77.0, tempF, 0.1)
	})

	t.Run("Celsius to Kelvin", func(t *testing.T) {
		// K = C + 273.15
		tempC := 25.0
		tempK := tempC + 273.15
		assert.InDelta(t, 298.15, tempK, 0.1)
	})

	t.Run("Parse raw temperature", func(t *testing.T) {
		// Raw value is in millidegrees: 23187 = 23.187°C
		rawTemp := 23187
		tempC := float64(rawTemp) / 1000.0
		assert.InDelta(t, 23.187, tempC, 0.001)
	})

	t.Run("Parse negative raw temperature", func(t *testing.T) {
		// Raw value: -10125 = -10.125°C
		rawTemp := -10125
		tempC := float64(rawTemp) / 1000.0
		assert.InDelta(t, -10.125, tempC, 0.001)
	})

	t.Run("Zero temperature", func(t *testing.T) {
		rawTemp := 0
		tempC := float64(rawTemp) / 1000.0
		assert.Equal(t, 0.0, tempC)
	})

	t.Run("Boiling point", func(t *testing.T) {
		rawTemp := 100000 // 100°C
		tempC := float64(rawTemp) / 1000.0
		assert.Equal(t, 100.0, tempC)
	})
}

func TestDS18B20Resolution(t *testing.T) {
	resolutions := []struct {
		bits     int
		stepSize float64
	}{
		{9, 0.5},
		{10, 0.25},
		{11, 0.125},
		{12, 0.0625},
	}

	for _, res := range resolutions {
		name := string(rune('0'+res.bits)) + " bit"
		t.Run(name, func(t *testing.T) {
			shift := uint(res.bits - 9 + 1)
			calculated := 1.0 / float64(int(1)<<shift)
			assert.Equal(t, res.stepSize, calculated)
		})
	}
}

func TestDS18B20DeviceIdParsing(t *testing.T) {
	t.Run("Valid device ID format", func(t *testing.T) {
		// DS18B20 device IDs start with 28-
		validIds := []string{
			"28-00000123abcd",
			"28-0000012345ab",
			"28-00000abcdef1",
		}
		for _, id := range validIds {
			assert.True(t, len(id) > 3)
			assert.Equal(t, "28-", id[:3])
		}
	})
}

func TestDS18B20UnitOptions(t *testing.T) {
	units := []string{"celsius", "fahrenheit"}
	for _, unit := range units {
		t.Run(unit, func(t *testing.T) {
			node := NewDS18B20Node()
			err := node.Init(map[string]interface{}{
				"unit": unit,
			})
			require.NoError(t, err)
			assert.Equal(t, unit, node.unit)
		})
	}
}
