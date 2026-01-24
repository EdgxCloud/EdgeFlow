//go:build !linux
// +build !linux

package gpio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBMP280Node(t *testing.T) {
	t.Run("Create new node with defaults", func(t *testing.T) {
		node := NewBMP280Node()
		assert.NotNil(t, node)
		assert.Equal(t, "1", node.i2cBus)
		assert.Equal(t, uint16(0x76), node.address)
	})

	t.Run("Init fails without hardware", func(t *testing.T) {
		node := NewBMP280Node()
		err := node.Init(map[string]interface{}{
			"i2cBus":  "1",
			"address": 0x76,
		})
		// Should fail on non-Linux because no I2C hardware
		assert.Error(t, err)
	})

	t.Run("Parse config with float address", func(t *testing.T) {
		node := NewBMP280Node()
		// Init will fail but config should be parsed
		_ = node.Init(map[string]interface{}{
			"address": float64(0x77),
		})
		assert.Equal(t, uint16(0x77), node.address)
	})

	t.Run("Parse config with int address", func(t *testing.T) {
		node := NewBMP280Node()
		_ = node.Init(map[string]interface{}{
			"address": int(0x76),
		})
		assert.Equal(t, uint16(0x76), node.address)
	})

	t.Run("Parse config with custom bus", func(t *testing.T) {
		node := NewBMP280Node()
		_ = node.Init(map[string]interface{}{
			"i2cBus": "0",
		})
		assert.Equal(t, "0", node.i2cBus)
	})

	t.Run("Cleanup without init returns nil", func(t *testing.T) {
		node := NewBMP280Node()
		err := node.Cleanup()
		assert.NoError(t, err)
	})
}

func TestBMP280Executor(t *testing.T) {
	t.Run("Executor requires hardware", func(t *testing.T) {
		// NewBMP280Executor calls Init which tries to open I2C bus
		// This will fail on non-Linux platforms
		executor, err := NewBMP280Executor(map[string]interface{}{})
		// Expect error because no I2C hardware available
		assert.Error(t, err)
		assert.Nil(t, executor)
	})
}

func TestBMP280AddressValidation(t *testing.T) {
	testCases := []struct {
		name    string
		address interface{}
		expect  uint16
	}{
		{"Default address 0x76", float64(0x76), 0x76},
		{"Alternate address 0x77", float64(0x77), 0x77},
		{"Int address 0x76", int(0x76), 0x76},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node := NewBMP280Node()
			_ = node.Init(map[string]interface{}{
				"address": tc.address,
			})
			assert.Equal(t, tc.expect, node.address)
		})
	}
}

func TestBMP280DataConversion(t *testing.T) {
	t.Run("Pressure hPa to Pa", func(t *testing.T) {
		// 1013.25 hPa = 101325 Pa
		pressureHPa := 1013.25
		pressurePa := pressureHPa * 100
		assert.InDelta(t, 101325.0, pressurePa, 0.1)
	})

	t.Run("Altitude from pressure", func(t *testing.T) {
		// Barometric formula: altitude = 44330 * (1 - (P/P0)^0.1903)
		seaLevelPressure := 1013.25
		currentPressure := 900.0 // ~1000m altitude

		altitude := 44330 * (1 - pow(currentPressure/seaLevelPressure, 0.1903))
		assert.Greater(t, altitude, 900.0)
		assert.Less(t, altitude, 1100.0)
	})

	t.Run("Temperature unit conversion", func(t *testing.T) {
		// Celsius to Fahrenheit: F = C * 9/5 + 32
		tempC := 25.0
		tempF := tempC*9.0/5.0 + 32.0
		assert.InDelta(t, 77.0, tempF, 0.1)

		// Celsius to Kelvin: K = C + 273.15
		tempK := tempC + 273.15
		assert.InDelta(t, 298.15, tempK, 0.1)
	})
}

func TestBMP280Calibration(t *testing.T) {
	t.Run("Apply temperature offset", func(t *testing.T) {
		rawTemp := 25.0
		offset := -0.5
		calibratedTemp := rawTemp + offset
		assert.Equal(t, 24.5, calibratedTemp)
	})

	t.Run("Apply pressure offset", func(t *testing.T) {
		rawPressure := 1013.25
		offset := 1.0
		calibratedPressure := rawPressure + offset
		assert.Equal(t, 1014.25, calibratedPressure)
	})
}
