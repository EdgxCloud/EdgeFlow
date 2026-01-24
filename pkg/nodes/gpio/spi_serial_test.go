//go:build !linux
// +build !linux

package gpio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// SPI Tests
// ============================================================================

func TestNewSPIExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"bus":    0,
				"device": 0,
				"speed":  1000000,
			},
			wantErr: false,
		},
		{
			name: "minimal config with defaults",
			config: map[string]interface{}{
				"bus":    0,
				"device": 0,
			},
			wantErr: false,
		},
		{
			name: "full config",
			config: map[string]interface{}{
				"bus":    1,
				"device": 1,
				"mode":   3,
				"bits":   8,
				"speed":  2000000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewSPIExecutor(tt.config)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, executor)
			}
		})
	}
}

func TestSPIConfig_Defaults(t *testing.T) {
	config := map[string]interface{}{
		"bus":    0,
		"device": 0,
	}

	executor, err := NewSPIExecutor(config)
	require.NoError(t, err)

	spiExecutor := executor.(*SPIExecutor)
	assert.Equal(t, 0, spiExecutor.config.Mode)
	assert.Equal(t, 1000000, spiExecutor.config.Speed)
}

func TestSPIExecutor_Cleanup(t *testing.T) {
	config := map[string]interface{}{
		"bus":    0,
		"device": 0,
	}

	executor, err := NewSPIExecutor(config)
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}

func TestSPIConfigParsing(t *testing.T) {
	t.Run("Parse bus from int", func(t *testing.T) {
		config := map[string]interface{}{
			"bus":    int(1),
			"device": 0,
		}
		executor, err := NewSPIExecutor(config)
		require.NoError(t, err)

		spiExecutor := executor.(*SPIExecutor)
		assert.Equal(t, 1, spiExecutor.config.Bus)
	})

	t.Run("Parse bus from float", func(t *testing.T) {
		config := map[string]interface{}{
			"bus":    float64(1),
			"device": 0,
		}
		executor, err := NewSPIExecutor(config)
		require.NoError(t, err)

		spiExecutor := executor.(*SPIExecutor)
		assert.Equal(t, 1, spiExecutor.config.Bus)
	})

	t.Run("Parse speed from int", func(t *testing.T) {
		config := map[string]interface{}{
			"bus":    0,
			"device": 0,
			"speed":  int(2000000),
		}
		executor, err := NewSPIExecutor(config)
		require.NoError(t, err)

		spiExecutor := executor.(*SPIExecutor)
		assert.Equal(t, 2000000, spiExecutor.config.Speed)
	})

	t.Run("Parse SPI mode", func(t *testing.T) {
		modes := []int{0, 1, 2, 3}
		for _, mode := range modes {
			config := map[string]interface{}{
				"bus":    0,
				"device": 0,
				"mode":   mode,
			}
			executor, err := NewSPIExecutor(config)
			require.NoError(t, err)

			spiExecutor := executor.(*SPIExecutor)
			assert.Equal(t, mode, spiExecutor.config.Mode)
		}
	})
}

func TestSPIModes(t *testing.T) {
	t.Run("SPI Mode 0 (CPOL=0, CPHA=0)", func(t *testing.T) {
		// Clock idle low, data sampled on rising edge
		mode := 0
		cpol := (mode >> 1) & 1
		cpha := mode & 1
		assert.Equal(t, 0, cpol)
		assert.Equal(t, 0, cpha)
	})

	t.Run("SPI Mode 1 (CPOL=0, CPHA=1)", func(t *testing.T) {
		// Clock idle low, data sampled on falling edge
		mode := 1
		cpol := (mode >> 1) & 1
		cpha := mode & 1
		assert.Equal(t, 0, cpol)
		assert.Equal(t, 1, cpha)
	})

	t.Run("SPI Mode 2 (CPOL=1, CPHA=0)", func(t *testing.T) {
		// Clock idle high, data sampled on falling edge
		mode := 2
		cpol := (mode >> 1) & 1
		cpha := mode & 1
		assert.Equal(t, 1, cpol)
		assert.Equal(t, 0, cpha)
	})

	t.Run("SPI Mode 3 (CPOL=1, CPHA=1)", func(t *testing.T) {
		// Clock idle high, data sampled on rising edge
		mode := 3
		cpol := (mode >> 1) & 1
		cpha := mode & 1
		assert.Equal(t, 1, cpol)
		assert.Equal(t, 1, cpha)
	})
}

// ============================================================================
// Serial Tests
// ============================================================================

func TestNewSerialExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"port":     "/dev/ttyUSB0",
				"baudRate": 115200,
			},
			wantErr: false,
		},
		{
			name: "minimal config",
			config: map[string]interface{}{
				"port": "/dev/ttyUSB0",
			},
			wantErr: false,
		},
		{
			name:    "missing port",
			config:  map[string]interface{}{},
			wantErr: true,
			errMsg:  "port is required",
		},
		{
			name: "full config",
			config: map[string]interface{}{
				"port":     "/dev/ttyUSB0",
				"baudRate": 9600,
				"dataBits": 8,
				"stopBits": 1,
				"parity":   "none",
				"mode":     "readwrite",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewSerialExecutor(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, executor)
			}
		})
	}
}

func TestSerialConfig_Defaults(t *testing.T) {
	config := map[string]interface{}{
		"port": "/dev/ttyUSB0",
	}

	executor, err := NewSerialExecutor(config)
	require.NoError(t, err)

	serialExecutor := executor.(*SerialExecutor)
	assert.Equal(t, 9600, serialExecutor.config.BaudRate)
	assert.Equal(t, 8, serialExecutor.config.DataBits)
	assert.Equal(t, 1, serialExecutor.config.StopBits)
	assert.Equal(t, "none", serialExecutor.config.Parity)
	assert.Equal(t, "readwrite", serialExecutor.config.Mode)
}

func TestSerialExecutor_ConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		baudRate int
	}{
		{"standard 9600", "/dev/ttyUSB0", 9600},
		{"standard 115200", "/dev/ttyUSB0", 115200},
		{"standard 57600", "/dev/ttyUSB0", 57600},
		{"COM port Windows", "COM1", 9600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]interface{}{
				"port":     tt.port,
				"baudRate": tt.baudRate,
			}

			executor, err := NewSerialExecutor(config)
			require.NoError(t, err)
			assert.NotNil(t, executor)

			serialExecutor := executor.(*SerialExecutor)
			assert.Equal(t, tt.port, serialExecutor.config.Port)
			assert.Equal(t, tt.baudRate, serialExecutor.config.BaudRate)
		})
	}
}

func TestSerialExecutor_Cleanup(t *testing.T) {
	config := map[string]interface{}{
		"port": "/dev/ttyUSB0",
	}

	executor, err := NewSerialExecutor(config)
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}

func TestSerialBaudRates(t *testing.T) {
	standardBaudRates := []int{
		1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200, 230400, 460800, 921600,
	}

	for _, baud := range standardBaudRates {
		t.Run(string(rune(baud)), func(t *testing.T) {
			config := map[string]interface{}{
				"port":     "/dev/ttyUSB0",
				"baudRate": baud,
			}

			executor, err := NewSerialExecutor(config)
			require.NoError(t, err)

			serialExecutor := executor.(*SerialExecutor)
			assert.Equal(t, baud, serialExecutor.config.BaudRate)
		})
	}
}

func TestSerialParityOptions(t *testing.T) {
	parityOptions := []string{"none", "odd", "even", "mark", "space"}

	for _, parity := range parityOptions {
		t.Run(parity, func(t *testing.T) {
			config := map[string]interface{}{
				"port":   "/dev/ttyUSB0",
				"parity": parity,
			}

			executor, err := NewSerialExecutor(config)
			require.NoError(t, err)

			serialExecutor := executor.(*SerialExecutor)
			assert.Equal(t, parity, serialExecutor.config.Parity)
		})
	}
}

func TestSerialDataBits(t *testing.T) {
	dataBitsOptions := []int{5, 6, 7, 8}

	for _, bits := range dataBitsOptions {
		t.Run(string(rune('0'+bits)), func(t *testing.T) {
			config := map[string]interface{}{
				"port":     "/dev/ttyUSB0",
				"dataBits": bits,
			}

			executor, err := NewSerialExecutor(config)
			require.NoError(t, err)

			serialExecutor := executor.(*SerialExecutor)
			assert.Equal(t, bits, serialExecutor.config.DataBits)
		})
	}
}

func TestSerialStopBits(t *testing.T) {
	stopBitsOptions := []int{1, 2}

	for _, bits := range stopBitsOptions {
		t.Run(string(rune('0'+bits)), func(t *testing.T) {
			config := map[string]interface{}{
				"port":     "/dev/ttyUSB0",
				"stopBits": bits,
			}

			executor, err := NewSerialExecutor(config)
			require.NoError(t, err)

			serialExecutor := executor.(*SerialExecutor)
			assert.Equal(t, bits, serialExecutor.config.StopBits)
		})
	}
}
