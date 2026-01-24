//go:build !linux
// +build !linux

package gpio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBME280Executor(t *testing.T) {
	t.Run("Create executor with defaults", func(t *testing.T) {
		executor, err := NewBME280Executor(map[string]interface{}{})
		require.NoError(t, err)
		require.NotNil(t, executor)

		bme := executor.(*BME280Executor)
		assert.Equal(t, 0x76, bme.config.Address)
		assert.Equal(t, "1x", bme.config.Oversampling)
		assert.Equal(t, 1013.25, bme.config.SeaLevelPress)
	})

	t.Run("Create executor with custom address 0x77", func(t *testing.T) {
		executor, err := NewBME280Executor(map[string]interface{}{
			"address": 0x77,
		})
		require.NoError(t, err)

		bme := executor.(*BME280Executor)
		assert.Equal(t, 0x77, bme.config.Address)
	})

	t.Run("Reject invalid address", func(t *testing.T) {
		_, err := NewBME280Executor(map[string]interface{}{
			"address": 0x50, // Invalid for BME280
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid I2C address")
	})

	t.Run("Parse oversampling config", func(t *testing.T) {
		executor, err := NewBME280Executor(map[string]interface{}{
			"oversampling": "16x",
		})
		require.NoError(t, err)

		bme := executor.(*BME280Executor)
		assert.Equal(t, "16x", bme.config.Oversampling)
	})

	t.Run("Parse calibration offsets", func(t *testing.T) {
		executor, err := NewBME280Executor(map[string]interface{}{
			"temp_offset":  -0.5,
			"press_offset": 1.0,
			"humid_offset": 2.5,
		})
		require.NoError(t, err)

		bme := executor.(*BME280Executor)
		assert.Equal(t, -0.5, bme.config.TempOffset)
		assert.Equal(t, 1.0, bme.config.PressOffset)
		assert.Equal(t, 2.5, bme.config.HumidOffset)
	})

	t.Run("Init returns nil", func(t *testing.T) {
		executor, err := NewBME280Executor(map[string]interface{}{})
		require.NoError(t, err)

		err = executor.Init(map[string]interface{}{})
		assert.NoError(t, err)
	})

	t.Run("Cleanup without device returns nil", func(t *testing.T) {
		executor, err := NewBME280Executor(map[string]interface{}{})
		require.NoError(t, err)

		err = executor.Cleanup()
		assert.NoError(t, err)
	})

	t.Run("Custom sea level pressure", func(t *testing.T) {
		executor, err := NewBME280Executor(map[string]interface{}{
			"sea_level_hpa": 1020.0,
		})
		require.NoError(t, err)

		bme := executor.(*BME280Executor)
		assert.Equal(t, 1020.0, bme.config.SeaLevelPress)
	})
}

func TestBME280OversamplingConversion(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1x", "1x"},
		{"2x", "2x"},
		{"4x", "4x"},
		{"8x", "8x"},
		{"16x", "16x"},
		{"", "1x"}, // Default
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			executor, err := NewBME280Executor(map[string]interface{}{
				"oversampling": tc.input,
			})
			require.NoError(t, err)

			bme := executor.(*BME280Executor)
			expected := tc.expected
			if tc.input == "" {
				expected = "1x"
			}
			assert.Equal(t, expected, bme.config.Oversampling)
		})
	}
}

func TestDewPointCalculation(t *testing.T) {
	t.Run("Calculate dew point at 25C and 50% humidity", func(t *testing.T) {
		dp := calcDewPoint(25.0, 50.0)
		// Dew point should be around 13-14°C
		assert.InDelta(t, 13.8, dp, 1.0)
	})

	t.Run("Calculate dew point at 20C and 80% humidity", func(t *testing.T) {
		dp := calcDewPoint(20.0, 80.0)
		// Dew point should be around 16-17°C
		assert.InDelta(t, 16.4, dp, 1.0)
	})

	t.Run("Handle zero humidity", func(t *testing.T) {
		dp := calcDewPoint(25.0, 0.0)
		// Should return temperature when humidity is 0
		assert.Equal(t, 25.0, dp)
	})
}
