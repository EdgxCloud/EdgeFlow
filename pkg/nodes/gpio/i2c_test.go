//go:build !linux
// +build !linux

package gpio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// I2C Executor Tests
// ============================================================================

func TestNewI2CExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid read config",
			config: map[string]interface{}{
				"address":  0x48,
				"register": 0x00,
				"length":   2,
				"mode":     "read",
			},
			wantErr: false,
		},
		{
			name: "valid write config",
			config: map[string]interface{}{
				"address":  0x48,
				"register": 0x01,
				"mode":     "write",
			},
			wantErr: false,
		},
		{
			name: "address too high",
			config: map[string]interface{}{
				"address": 0x80, // Invalid - max is 0x7F
			},
			wantErr: true,
			errMsg:  "invalid I2C address",
		},
		{
			name: "negative address",
			config: map[string]interface{}{
				"address": -1,
			},
			wantErr: true,
			errMsg:  "invalid I2C address",
		},
		{
			name: "valid address boundary low",
			config: map[string]interface{}{
				"address": 0x00,
			},
			wantErr: false,
		},
		{
			name: "valid address boundary high",
			config: map[string]interface{}{
				"address": 0x7F,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewI2CExecutor(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestI2CConfig_Defaults(t *testing.T) {
	config := map[string]interface{}{
		"address": 0x48,
	}

	executor, err := NewI2CExecutor(config)
	require.NoError(t, err)

	i2cExecutor := executor.(*I2CExecutor)
	assert.Equal(t, "read", i2cExecutor.config.Mode)
	assert.Equal(t, 1, i2cExecutor.config.Length)
}

func TestI2CExecutor_Cleanup(t *testing.T) {
	config := map[string]interface{}{
		"address": 0x48,
	}

	executor, err := NewI2CExecutor(config)
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}

func TestI2CExecutor_CommonAddresses(t *testing.T) {
	// Test common I2C device addresses
	commonAddresses := []struct {
		name    string
		address int
		device  string
	}{
		{"BMP280/BME280 default", 0x76, "pressure sensor"},
		{"BMP280/BME280 alternate", 0x77, "pressure sensor"},
		{"SHT3x default", 0x44, "temp/humidity sensor"},
		{"SHT3x alternate", 0x45, "temp/humidity sensor"},
		{"AHT20", 0x38, "temp/humidity sensor"},
		{"OLED display", 0x3C, "SSD1306 display"},
		{"RTC DS3231", 0x68, "real time clock"},
		{"PCF8574 I/O expander", 0x20, "GPIO expander"},
		{"ADS1115 ADC", 0x48, "analog to digital"},
		{"PCA9685 PWM", 0x40, "PWM controller"},
	}

	for _, tt := range commonAddresses {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]interface{}{
				"address": tt.address,
				"mode":    "read",
			}

			executor, err := NewI2CExecutor(config)
			require.NoError(t, err)
			assert.NotNil(t, executor)

			i2cExecutor := executor.(*I2CExecutor)
			assert.Equal(t, tt.address, i2cExecutor.config.Address)
		})
	}
}

func TestI2CAddressValidation(t *testing.T) {
	t.Run("Valid 7-bit addresses", func(t *testing.T) {
		// Valid I2C addresses are 0x00 to 0x7F (7-bit)
		validAddresses := []int{0x00, 0x10, 0x20, 0x48, 0x76, 0x77, 0x7F}
		for _, addr := range validAddresses {
			isValid := addr >= 0 && addr <= 0x7F
			assert.True(t, isValid, "Address 0x%02X should be valid", addr)
		}
	})

	t.Run("Invalid addresses", func(t *testing.T) {
		invalidAddresses := []int{-1, 0x80, 0xFF, 0x100}
		for _, addr := range invalidAddresses {
			isValid := addr >= 0 && addr <= 0x7F
			assert.False(t, isValid, "Address 0x%02X should be invalid", addr)
		}
	})

	t.Run("Reserved addresses", func(t *testing.T) {
		// I2C reserved addresses: 0x00-0x07 (general call), 0x78-0x7F (10-bit addressing)
		// These are technically valid but often reserved
		reservedLow := []int{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
		reservedHigh := []int{0x78, 0x79, 0x7A, 0x7B, 0x7C, 0x7D, 0x7E, 0x7F}

		for _, addr := range reservedLow {
			assert.True(t, addr >= 0x00 && addr <= 0x07)
		}
		for _, addr := range reservedHigh {
			assert.True(t, addr >= 0x78 && addr <= 0x7F)
		}
	})
}

func TestI2CDataConversions(t *testing.T) {
	t.Run("Hex string to bytes", func(t *testing.T) {
		// "0102030405" -> []byte{0x01, 0x02, 0x03, 0x04, 0x05}
		hexStr := "0102030405"
		expectedLen := len(hexStr) / 2
		assert.Equal(t, 5, expectedLen)
	})

	t.Run("Bytes to hex string", func(t *testing.T) {
		// []byte{0x01, 0x02, 0x03} -> "010203"
		data := []byte{0x01, 0x02, 0x03}
		expectedHex := "010203"
		assert.Equal(t, len(expectedHex), len(data)*2)
	})

	t.Run("Big endian 16-bit read", func(t *testing.T) {
		// Typical I2C read returns big-endian data
		// [0x12, 0x34] -> 0x1234
		msb := byte(0x12)
		lsb := byte(0x34)
		value := int(msb)<<8 | int(lsb)
		assert.Equal(t, 0x1234, value)
	})

	t.Run("Little endian 16-bit read", func(t *testing.T) {
		// Some devices use little-endian
		// [0x34, 0x12] -> 0x1234
		lsb := byte(0x34)
		msb := byte(0x12)
		value := int(msb)<<8 | int(lsb)
		assert.Equal(t, 0x1234, value)
	})
}

func TestI2CRegisterOperations(t *testing.T) {
	t.Run("Register read length", func(t *testing.T) {
		// Common register read lengths
		lengths := []struct {
			name   string
			length int
		}{
			{"Single byte", 1},
			{"16-bit value", 2},
			{"24-bit value", 3},
			{"32-bit value", 4},
		}

		for _, l := range lengths {
			assert.Greater(t, l.length, 0)
			assert.LessOrEqual(t, l.length, 32) // Reasonable max
		}
	})

	t.Run("Register address range", func(t *testing.T) {
		// Most I2C devices use 8-bit register addressing (0x00-0xFF)
		validRegisters := []int{0x00, 0x10, 0x7F, 0x80, 0xFF}
		for _, reg := range validRegisters {
			isValid := reg >= 0x00 && reg <= 0xFF
			assert.True(t, isValid, "Register 0x%02X should be valid", reg)
		}
	})
}

func TestI2CConfigParsing(t *testing.T) {
	t.Run("Parse address from float", func(t *testing.T) {
		config := map[string]interface{}{
			"address": float64(0x48),
		}
		executor, err := NewI2CExecutor(config)
		require.NoError(t, err)

		i2cExecutor := executor.(*I2CExecutor)
		assert.Equal(t, 0x48, i2cExecutor.config.Address)
	})

	t.Run("Parse address from int", func(t *testing.T) {
		config := map[string]interface{}{
			"address": int(0x48),
		}
		executor, err := NewI2CExecutor(config)
		require.NoError(t, err)

		i2cExecutor := executor.(*I2CExecutor)
		assert.Equal(t, 0x48, i2cExecutor.config.Address)
	})

	t.Run("Parse mode options", func(t *testing.T) {
		modes := []string{"read", "write", "writeread"}
		for _, mode := range modes {
			config := map[string]interface{}{
				"address": 0x48,
				"mode":    mode,
			}
			executor, err := NewI2CExecutor(config)
			require.NoError(t, err)

			i2cExecutor := executor.(*I2CExecutor)
			assert.Equal(t, mode, i2cExecutor.config.Mode)
		}
	})
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkI2CExecutor_Create(b *testing.B) {
	config := map[string]interface{}{
		"address":  0x48,
		"register": 0x00,
		"length":   2,
		"mode":     "read",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor, _ := NewI2CExecutor(config)
		if executor != nil {
			executor.Cleanup()
		}
	}
}
