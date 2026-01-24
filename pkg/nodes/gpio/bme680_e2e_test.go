//go:build !linux
// +build !linux

package gpio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBME680Executor(t *testing.T) {
	t.Run("Create executor with defaults", func(t *testing.T) {
		executor, err := NewBME680Executor(map[string]interface{}{})
		require.NoError(t, err)
		require.NotNil(t, executor)

		bme := executor.(*BME680Executor)
		assert.Equal(t, 0x76, bme.config.Address)
		assert.Equal(t, "2x", bme.config.Oversampling)
		assert.Equal(t, "2", bme.config.Filter)
		assert.Equal(t, 320, bme.config.GasHeaterTemp)
		assert.Equal(t, 150, bme.config.GasHeaterTime)
		assert.Equal(t, 1013.25, bme.config.SeaLevelPress)
	})

	t.Run("Create executor with custom address 0x77", func(t *testing.T) {
		executor, err := NewBME680Executor(map[string]interface{}{
			"address": 0x77,
		})
		require.NoError(t, err)

		bme := executor.(*BME680Executor)
		assert.Equal(t, 0x77, bme.config.Address)
	})

	t.Run("Parse oversampling config", func(t *testing.T) {
		executor, err := NewBME680Executor(map[string]interface{}{
			"oversampling": "16x",
		})
		require.NoError(t, err)

		bme := executor.(*BME680Executor)
		assert.Equal(t, "16x", bme.config.Oversampling)
	})

	t.Run("Parse filter config", func(t *testing.T) {
		executor, err := NewBME680Executor(map[string]interface{}{
			"filter": "16",
		})
		require.NoError(t, err)

		bme := executor.(*BME680Executor)
		assert.Equal(t, "16", bme.config.Filter)
	})

	t.Run("Parse gas heater config", func(t *testing.T) {
		executor, err := NewBME680Executor(map[string]interface{}{
			"gas_heater_temp": 350,
			"gas_heater_time": 200,
		})
		require.NoError(t, err)

		bme := executor.(*BME680Executor)
		assert.Equal(t, 350, bme.config.GasHeaterTemp)
		assert.Equal(t, 200, bme.config.GasHeaterTime)
	})

	t.Run("Parse calibration offsets", func(t *testing.T) {
		executor, err := NewBME680Executor(map[string]interface{}{
			"temp_offset":  -0.5,
			"humid_offset": 2.0,
		})
		require.NoError(t, err)

		bme := executor.(*BME680Executor)
		assert.Equal(t, -0.5, bme.config.TempOffset)
		assert.Equal(t, 2.0, bme.config.HumidOffset)
	})

	t.Run("Parse custom sea level pressure", func(t *testing.T) {
		executor, err := NewBME680Executor(map[string]interface{}{
			"sea_level_hpa": 1025.0,
		})
		require.NoError(t, err)

		bme := executor.(*BME680Executor)
		assert.Equal(t, 1025.0, bme.config.SeaLevelPress)
	})

	t.Run("Init returns nil", func(t *testing.T) {
		executor, err := NewBME680Executor(map[string]interface{}{})
		require.NoError(t, err)

		err = executor.Init(map[string]interface{}{})
		assert.NoError(t, err)
	})

	t.Run("Cleanup without device returns nil", func(t *testing.T) {
		executor, err := NewBME680Executor(map[string]interface{}{})
		require.NoError(t, err)

		err = executor.Cleanup()
		assert.NoError(t, err)
	})
}

func TestBME680FilterConversion(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"off", "off"},
		{"0", "0"},
		{"2", "2"},
		{"4", "4"},
		{"8", "8"},
		{"16", "16"},
		{"", "2"}, // Default
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			executor, err := NewBME680Executor(map[string]interface{}{
				"filter": tc.input,
			})
			require.NoError(t, err)

			bme := executor.(*BME680Executor)
			expected := tc.expected
			if tc.input == "" {
				expected = "2"
			}
			assert.Equal(t, expected, bme.config.Filter)
		})
	}
}

func TestIAQCalculation(t *testing.T) {
	t.Run("Calculate IAQ with optimal humidity", func(t *testing.T) {
		// Humidity in optimal range (40-60%)
		iaq := calculateIAQ(50.0, 0)
		// Should return good score (around 50)
		assert.InDelta(t, 50.0, iaq, 5.0)
	})

	t.Run("Calculate IAQ with low humidity", func(t *testing.T) {
		// Humidity below optimal (< 40%)
		iaq := calculateIAQ(30.0, 0)
		// Should return worse score
		assert.Greater(t, iaq, 50.0)
	})

	t.Run("Calculate IAQ with high humidity", func(t *testing.T) {
		// Humidity above optimal (> 60%)
		iaq := calculateIAQ(75.0, 0)
		// Should return worse score
		assert.Greater(t, iaq, 50.0)
	})

	t.Run("Calculate IAQ with gas resistance", func(t *testing.T) {
		// Good gas resistance (high value = good air)
		iaq := calculateIAQ(50.0, 100000)
		// Should give reasonable score (algorithm is simplified)
		assert.Greater(t, iaq, 0.0)
		assert.Less(t, iaq, 200.0)
	})
}

func TestAirQualityLabel(t *testing.T) {
	testCases := []struct {
		iaq      float64
		expected string
	}{
		{25.0, "excellent"},
		{50.0, "excellent"},
		{75.0, "good"},
		{100.0, "good"},
		{125.0, "moderate"},
		{175.0, "poor"},
		{250.0, "unhealthy"},
		{400.0, "hazardous"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			label := getAirQualityLabel(tc.iaq)
			assert.Equal(t, tc.expected, label)
		})
	}
}
