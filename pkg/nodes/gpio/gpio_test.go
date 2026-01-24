//go:build !linux
// +build !linux

package gpio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGPIOInExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"pin": 17,
			},
			wantErr: false,
		},
		{
			name: "invalid pin",
			config: map[string]interface{}{
				"pin": -1,
			},
			wantErr: true,
		},
		{
			name: "with pull mode up",
			config: map[string]interface{}{
				"pin":      17,
				"pullMode": "up",
			},
			wantErr: false,
		},
		{
			name: "with pull mode down",
			config: map[string]interface{}{
				"pin":      17,
				"pullMode": "down",
			},
			wantErr: false,
		},
		{
			name: "with edge detection rising",
			config: map[string]interface{}{
				"pin":      17,
				"edgeMode": "rising",
			},
			wantErr: false,
		},
		{
			name: "with edge detection falling",
			config: map[string]interface{}{
				"pin":      17,
				"edgeMode": "falling",
			},
			wantErr: false,
		},
		{
			name: "with edge detection both",
			config: map[string]interface{}{
				"pin":      17,
				"edgeMode": "both",
			},
			wantErr: false,
		},
		{
			name: "with debounce",
			config: map[string]interface{}{
				"pin":      17,
				"debounce": 100,
			},
			wantErr: false,
		},
		{
			name: "with poll interval",
			config: map[string]interface{}{
				"pin":          17,
				"pollInterval": 200,
			},
			wantErr: false,
		},
		{
			name: "with interrupt mode",
			config: map[string]interface{}{
				"pin":           17,
				"edgeMode":      "rising",
				"interruptMode": true,
			},
			wantErr: false,
		},
		{
			name: "with read initial",
			config: map[string]interface{}{
				"pin":         17,
				"readInitial": true,
			},
			wantErr: false,
		},
		{
			name: "with glitch filter",
			config: map[string]interface{}{
				"pin":    17,
				"glitch": 50,
			},
			wantErr: false,
		},
		{
			name: "full config",
			config: map[string]interface{}{
				"pin":           17,
				"pullMode":      "up",
				"edgeMode":      "both",
				"debounce":      50,
				"pollInterval":  100,
				"interruptMode": true,
				"readInitial":   true,
				"glitch":        10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewGPIOInExecutor(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestGPIOInConfig_Defaults(t *testing.T) {
	executor, err := NewGPIOInExecutor(map[string]interface{}{
		"pin": 17,
	})
	require.NoError(t, err)

	gpioExecutor := executor.(*GPIOInExecutor)
	assert.Equal(t, 17, gpioExecutor.config.Pin)
	assert.Equal(t, "none", gpioExecutor.config.PullMode)
	assert.Equal(t, "none", gpioExecutor.config.EdgeMode)
	assert.Equal(t, 50, gpioExecutor.config.Debounce)
	assert.Equal(t, 100, gpioExecutor.config.PollInterval)
}

func TestGPIOInConfig_InterruptModeEnabled(t *testing.T) {
	// Interrupt mode should be enabled when edge detection is used
	executor, err := NewGPIOInExecutor(map[string]interface{}{
		"pin":      17,
		"edgeMode": "rising",
	})
	require.NoError(t, err)

	gpioExecutor := executor.(*GPIOInExecutor)
	assert.True(t, gpioExecutor.config.InterruptMode)
}

func TestGPIOInExecutor_Cleanup(t *testing.T) {
	executor, err := NewGPIOInExecutor(map[string]interface{}{
		"pin": 17,
	})
	require.NoError(t, err)

	// Cleanup should not panic
	err = executor.Cleanup()
	assert.NoError(t, err)
}

func TestNewGPIOOutExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"pin": 18,
			},
			wantErr: false,
		},
		{
			name: "invalid pin",
			config: map[string]interface{}{
				"pin": -1,
			},
			wantErr: true,
		},
		{
			name: "with initial value true",
			config: map[string]interface{}{
				"pin":          18,
				"initialValue": true,
			},
			wantErr: false,
		},
		{
			name: "with invert",
			config: map[string]interface{}{
				"pin":    18,
				"invert": true,
			},
			wantErr: false,
		},
		{
			name: "with persist state",
			config: map[string]interface{}{
				"pin":          18,
				"persistState": true,
			},
			wantErr: false,
		},
		{
			name: "with custom state file",
			config: map[string]interface{}{
				"pin":          18,
				"persistState": true,
				"stateFile":    "/tmp/gpio_states.json",
			},
			wantErr: false,
		},
		{
			name: "with failsafe mode",
			config: map[string]interface{}{
				"pin":           18,
				"failsafeMode":  true,
				"failsafeValue": false,
			},
			wantErr: false,
		},
		{
			name: "with drive strength",
			config: map[string]interface{}{
				"pin":           18,
				"driveStrength": 8,
			},
			wantErr: false,
		},
		{
			name: "full config",
			config: map[string]interface{}{
				"pin":           18,
				"initialValue":  true,
				"invert":        true,
				"persistState":  true,
				"stateFile":     "/tmp/gpio_states.json",
				"failsafeMode":  true,
				"failsafeValue": false,
				"driveStrength": 8,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewGPIOOutExecutor(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestGPIOOutConfig_DefaultStateFile(t *testing.T) {
	executor, err := NewGPIOOutExecutor(map[string]interface{}{
		"pin":          18,
		"persistState": true,
	})
	require.NoError(t, err)

	gpioExecutor := executor.(*GPIOOutExecutor)
	assert.Equal(t, "/var/lib/edgeflow/gpio_states.json", gpioExecutor.config.StateFile)
}

func TestGPIOOutExecutor_Cleanup(t *testing.T) {
	executor, err := NewGPIOOutExecutor(map[string]interface{}{
		"pin": 18,
	})
	require.NoError(t, err)

	// Cleanup should not error without HAL
	err = executor.Cleanup()
	assert.NoError(t, err)
}

func TestNewPWMExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"pin": 12,
			},
			wantErr: false,
		},
		{
			name: "invalid pin",
			config: map[string]interface{}{
				"pin": -1,
			},
			wantErr: true,
		},
		{
			name: "with frequency",
			config: map[string]interface{}{
				"pin":       12,
				"frequency": 2000,
			},
			wantErr: false,
		},
		{
			name: "with duty cycle",
			config: map[string]interface{}{
				"pin":       12,
				"dutyCycle": 128,
			},
			wantErr: false,
		},
		{
			name: "pwm mode",
			config: map[string]interface{}{
				"pin":  12,
				"mode": "pwm",
			},
			wantErr: false,
		},
		{
			name: "servo mode",
			config: map[string]interface{}{
				"pin":  12,
				"mode": "servo",
			},
			wantErr: false,
		},
		{
			name: "led mode",
			config: map[string]interface{}{
				"pin":  12,
				"mode": "led",
			},
			wantErr: false,
		},
		{
			name: "with hardware PWM",
			config: map[string]interface{}{
				"pin":         12,
				"hardwarePwm": true,
				"channel":     0,
			},
			wantErr: false,
		},
		{
			name: "servo mode with servo config",
			config: map[string]interface{}{
				"pin":  12,
				"mode": "servo",
				"servoConfig": map[string]interface{}{
					"minPulse": 0.5,
					"maxPulse": 2.5,
					"minAngle": 0.0,
					"maxAngle": 180.0,
				},
			},
			wantErr: false,
		},
		{
			name: "duty cycle clamped at max",
			config: map[string]interface{}{
				"pin":       12,
				"dutyCycle": 300,
			},
			wantErr: false,
		},
		{
			name: "duty cycle clamped at min",
			config: map[string]interface{}{
				"pin":       12,
				"dutyCycle": -10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewPWMExecutor(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestPWMConfig_Defaults(t *testing.T) {
	executor, err := NewPWMExecutor(map[string]interface{}{
		"pin": 12,
	})
	require.NoError(t, err)

	pwmExecutor := executor.(*PWMExecutor)
	assert.Equal(t, 12, pwmExecutor.config.Pin)
	assert.Equal(t, "pwm", pwmExecutor.config.Mode)
	assert.Equal(t, 1000, pwmExecutor.config.Frequency)
}

func TestPWMConfig_ServoDefaults(t *testing.T) {
	executor, err := NewPWMExecutor(map[string]interface{}{
		"pin":  12,
		"mode": "servo",
	})
	require.NoError(t, err)

	pwmExecutor := executor.(*PWMExecutor)
	assert.Equal(t, "servo", pwmExecutor.config.Mode)
	assert.Equal(t, 50, pwmExecutor.config.Frequency) // 50Hz for servo
	assert.NotNil(t, pwmExecutor.config.ServoConfig)
	assert.Equal(t, 0.5, pwmExecutor.config.ServoConfig.MinPulse)
	assert.Equal(t, 2.5, pwmExecutor.config.ServoConfig.MaxPulse)
	assert.Equal(t, 0.0, pwmExecutor.config.ServoConfig.MinAngle)
	assert.Equal(t, 180.0, pwmExecutor.config.ServoConfig.MaxAngle)
}

func TestPWMConfig_DutyCycleClamping(t *testing.T) {
	// Test max clamping
	executor, err := NewPWMExecutor(map[string]interface{}{
		"pin":       12,
		"dutyCycle": 300,
	})
	require.NoError(t, err)
	pwmExecutor := executor.(*PWMExecutor)
	assert.Equal(t, 255, pwmExecutor.config.DutyCycle)

	// Test min clamping
	executor, err = NewPWMExecutor(map[string]interface{}{
		"pin":       12,
		"dutyCycle": -10,
	})
	require.NoError(t, err)
	pwmExecutor = executor.(*PWMExecutor)
	assert.Equal(t, 0, pwmExecutor.config.DutyCycle)
}

func TestPWMExecutor_Cleanup(t *testing.T) {
	executor, err := NewPWMExecutor(map[string]interface{}{
		"pin": 12,
	})
	require.NoError(t, err)

	// Cleanup should not error without HAL
	err = executor.Cleanup()
	assert.NoError(t, err)
}

func TestNewServoExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"pin": 12,
			},
			wantErr: false,
		},
		{
			name: "invalid pin",
			config: map[string]interface{}{
				"pin": -1,
			},
			wantErr: true,
		},
		{
			name: "with custom pulse range",
			config: map[string]interface{}{
				"pin":      12,
				"minPulse": 0.5,
				"maxPulse": 2.5,
			},
			wantErr: false,
		},
		{
			name: "with custom angle range",
			config: map[string]interface{}{
				"pin":      12,
				"minAngle": -90.0,
				"maxAngle": 90.0,
			},
			wantErr: false,
		},
		{
			name: "with frequency",
			config: map[string]interface{}{
				"pin":       12,
				"frequency": 60,
			},
			wantErr: false,
		},
		{
			name: "with start angle",
			config: map[string]interface{}{
				"pin":        12,
				"startAngle": 90.0,
			},
			wantErr: false,
		},
		{
			name: "full config",
			config: map[string]interface{}{
				"pin":        12,
				"minPulse":   1.0,
				"maxPulse":   2.0,
				"frequency":  50,
				"minAngle":   0,
				"maxAngle":   180,
				"startAngle": 90,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewServoExecutor(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestServoConfig_Defaults(t *testing.T) {
	executor, err := NewServoExecutor(map[string]interface{}{
		"pin": 12,
	})
	require.NoError(t, err)

	servoExecutor := executor.(*ServoExecutor)
	assert.Equal(t, 12, servoExecutor.config.Pin)
	assert.Equal(t, 1.0, servoExecutor.config.MinPulse)
	assert.Equal(t, 2.0, servoExecutor.config.MaxPulse)
	assert.Equal(t, 50.0, servoExecutor.config.Frequency)
	assert.Equal(t, 0.0, servoExecutor.config.MinAngle)
	assert.Equal(t, 180.0, servoExecutor.config.MaxAngle)
}

func TestServoExecutor_Cleanup(t *testing.T) {
	executor, err := NewServoExecutor(map[string]interface{}{
		"pin": 12,
	})
	require.NoError(t, err)

	// Cleanup should not error without HAL
	err = executor.Cleanup()
	assert.NoError(t, err)
}
