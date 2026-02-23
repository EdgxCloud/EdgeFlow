package gpio

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EdgxCloud/EdgeFlow/internal/hal"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// PWMConfig PWM node configuration
type PWMConfig struct {
	Pin         int    `json:"pin"`         // pin number
	Frequency   int    `json:"frequency"`   // frequency (Hz)
	DutyCycle   int    `json:"dutyCycle"`   // Duty cycle (0-255)
	Mode        string `json:"mode"`        // Mode: pwm, servo, led
	HardwarePWM bool   `json:"hardwarePwm"` // Use hardware PWM if available
	Channel     int    `json:"channel"`     // Hardware PWM channel (0-1 on most Pi models)
	ServoConfig *PWMServoConfig `json:"servoConfig"` // Servo mode configuration
}

// PWMServoConfig servo mode configuration for PWM
type PWMServoConfig struct {
	MinPulse   float64 `json:"minPulse"`   // Minimum pulse width in ms (default 0.5)
	MaxPulse   float64 `json:"maxPulse"`   // Maximum pulse width in ms (default 2.5)
	MinAngle   float64 `json:"minAngle"`   // Minimum angle (default 0)
	MaxAngle   float64 `json:"maxAngle"`   // Maximum angle (default 180)
}

// PWMExecutor PWM node executor
type PWMExecutor struct {
	config PWMConfig
	hal    hal.HAL
}

// NewPWMExecutor create PWMExecutor
func NewPWMExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var pwmConfig PWMConfig
	if err := json.Unmarshal(configJSON, &pwmConfig); err != nil {
		return nil, fmt.Errorf("invalid pwm config: %w", err)
	}

	// Validate
	if pwmConfig.Pin < 0 {
		return nil, fmt.Errorf("invalid pin number")
	}

	// Default values
	if pwmConfig.Mode == "" {
		pwmConfig.Mode = "pwm"
	}
	if pwmConfig.Frequency == 0 {
		if pwmConfig.Mode == "servo" {
			pwmConfig.Frequency = 50 // 50Hz for servo
		} else {
			pwmConfig.Frequency = 1000 // 1kHz for general PWM
		}
	}
	if pwmConfig.DutyCycle < 0 {
		pwmConfig.DutyCycle = 0
	}
	if pwmConfig.DutyCycle > 255 {
		pwmConfig.DutyCycle = 255
	}

	// Default servo config
	if pwmConfig.Mode == "servo" && pwmConfig.ServoConfig == nil {
		pwmConfig.ServoConfig = &PWMServoConfig{
			MinPulse: 0.5,
			MaxPulse: 2.5,
			MinAngle: 0,
			MaxAngle: 180,
		}
	}

	return &PWMExecutor{
		config: pwmConfig,
	}, nil
}

// Init initializes the PWM executor with config
func (e *PWMExecutor) Init(config map[string]interface{}) error {
	// Config is already parsed in NewPWMExecutor
	return nil
}

// Execute execute node
func (e *PWMExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get HAL if not initialized
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h

		// Setup PWM
		if err := e.setup(); err != nil {
			return node.Message{}, fmt.Errorf("failed to setup PWM: %w", err)
		}
	}

	// Handle servo mode
	if e.config.Mode == "servo" && e.config.ServoConfig != nil {
		return e.executeServoMode(ctx, msg)
	}

	// Get duty cycle from message (0-100% or 0-255)
	dutyCycle := e.config.DutyCycle
	if msg.Payload != nil {
		if v, ok := msg.Payload["dutyCycle"].(float64); ok {
			// If value is 0-100, convert to 0-255
			if v <= 100 {
				dutyCycle = int(v * 2.55)
			} else {
				dutyCycle = int(v)
			}
		} else if v, ok := msg.Payload["value"].(float64); ok {
			if v <= 100 {
				dutyCycle = int(v * 2.55)
			} else {
				dutyCycle = int(v)
			}
		} else if v, ok := msg.Payload["payload"].(float64); ok {
			if v <= 100 {
				dutyCycle = int(v * 2.55)
			} else {
				dutyCycle = int(v)
			}
		}

		// Get frequency if provided
		if v, ok := msg.Payload["frequency"].(float64); ok {
			e.config.Frequency = int(v)
			gpio := e.hal.GPIO()
			gpio.SetPWMFrequency(e.config.Pin, e.config.Frequency)
		}
	}

	// Clamp duty cycle
	if dutyCycle < 0 {
		dutyCycle = 0
	}
	if dutyCycle > 255 {
		dutyCycle = 255
	}

	// Write PWM
	gpio := e.hal.GPIO()
	if err := gpio.PWMWrite(e.config.Pin, dutyCycle); err != nil {
		return node.Message{}, fmt.Errorf("failed to write PWM: %w", err)
	}

	// Return message
	return node.Message{
		Payload: map[string]interface{}{
			"pin":       e.config.Pin,
			"dutyCycle": dutyCycle,
			"percent":   float64(dutyCycle) / 2.55,
			"frequency": e.config.Frequency,
		},
	}, nil
}

// setup initialize PWM
func (e *PWMExecutor) setup() error {
	gpio := e.hal.GPIO()

	// Set mode to PWM
	if err := gpio.SetMode(e.config.Pin, hal.PWM); err != nil {
		return err
	}

	// Set frequency
	if err := gpio.SetPWMFrequency(e.config.Pin, e.config.Frequency); err != nil {
		return err
	}

	// Set initial duty cycle
	if err := gpio.PWMWrite(e.config.Pin, e.config.DutyCycle); err != nil {
		return err
	}

	return nil
}

// executeServoMode executes the PWM node in servo mode
func (e *PWMExecutor) executeServoMode(ctx context.Context, msg node.Message) (node.Message, error) {
	servoConf := e.config.ServoConfig

	// Get angle from message
	var angle float64
	if msg.Payload != nil {
		if v, ok := msg.Payload["angle"].(float64); ok {
			angle = v
		} else if v, ok := msg.Payload["value"].(float64); ok {
			angle = v
		} else if v, ok := msg.Payload["payload"].(float64); ok {
			angle = v
		}
	}

	// Clamp angle to valid range
	if angle < servoConf.MinAngle {
		angle = servoConf.MinAngle
	}
	if angle > servoConf.MaxAngle {
		angle = servoConf.MaxAngle
	}

	// Convert angle to pulse width
	angleRange := servoConf.MaxAngle - servoConf.MinAngle
	pulseRange := servoConf.MaxPulse - servoConf.MinPulse
	pulseWidth := servoConf.MinPulse + ((angle - servoConf.MinAngle) / angleRange * pulseRange)

	// Convert pulse width to duty cycle (0-255)
	// Period at 50Hz = 20ms
	periodMs := 1000.0 / float64(e.config.Frequency)
	dutyCycle := int((pulseWidth / periodMs) * 255.0)

	// Clamp duty cycle
	if dutyCycle < 0 {
		dutyCycle = 0
	}
	if dutyCycle > 255 {
		dutyCycle = 255
	}

	// Write PWM
	gpio := e.hal.GPIO()
	if err := gpio.PWMWrite(e.config.Pin, dutyCycle); err != nil {
		return node.Message{}, fmt.Errorf("failed to write servo PWM: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"pin":        e.config.Pin,
			"mode":       "servo",
			"angle":      angle,
			"pulseWidth": pulseWidth,
			"dutyCycle":  dutyCycle,
		},
	}, nil
}

// Cleanup cleanup resources
func (e *PWMExecutor) Cleanup() error {
	// Set PWM to 0
	if e.hal != nil {
		gpio := e.hal.GPIO()
		gpio.PWMWrite(e.config.Pin, 0)
	}
	return nil
}

// ServoConfig Servo node configuration
type ServoConfig struct {
	Pin        int     `json:"pin"`        // pin number
	MinPulse   float64 `json:"minPulse"`   // minimum pulse width (ms)
	MaxPulse   float64 `json:"maxPulse"`   // maximum pulse width (ms)
	Frequency  float64 `json:"frequency"`  // PWM frequency (Hz)
	MinAngle   float64 `json:"minAngle"`   // minimum angle
	MaxAngle   float64 `json:"maxAngle"`   // maximum angle
	StartAngle float64 `json:"startAngle"` // start angle
}

// ServoExecutor Servo node executor
type ServoExecutor struct {
	config      ServoConfig
	hal         hal.HAL
	initialized bool
}

// NewServoExecutor create ServoExecutor
func NewServoExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var servoConfig ServoConfig
	if err := json.Unmarshal(configJSON, &servoConfig); err != nil {
		return nil, fmt.Errorf("invalid servo config: %w", err)
	}

	// Defaults
	if servoConfig.MinPulse == 0 {
		servoConfig.MinPulse = 1.0 // 1ms
	}
	if servoConfig.MaxPulse == 0 {
		servoConfig.MaxPulse = 2.0 // 2ms
	}
	if servoConfig.Frequency == 0 {
		servoConfig.Frequency = 50 // 50Hz standard for servos
	}
	if servoConfig.MaxAngle == 0 {
		servoConfig.MaxAngle = 180
	}

	// Validate
	if servoConfig.Pin < 0 {
		return nil, fmt.Errorf("invalid pin number")
	}

	return &ServoExecutor{
		config: servoConfig,
	}, nil
}

// Init initializes the Servo executor with config
func (e *ServoExecutor) Init(config map[string]interface{}) error {
	// Config is already parsed in NewServoExecutor
	return nil
}

// Execute execute node
func (e *ServoExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get HAL if not initialized
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h

		if err := e.setup(); err != nil {
			return node.Message{}, fmt.Errorf("failed to setup servo: %w", err)
		}
		e.initialized = true
	}

	// Get angle from message
	var angle float64

	if msg.Payload != nil {
		if v, ok := msg.Payload["angle"].(float64); ok {
			angle = v
		} else if v, ok := msg.Payload["value"].(float64); ok {
			angle = v
		} else if v, ok := msg.Payload["payload"].(float64); ok {
			angle = v
		}
	}

	// Set angle
	if err := e.setAngle(angle); err != nil {
		return node.Message{}, fmt.Errorf("failed to set servo angle: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"pin":   e.config.Pin,
			"angle": angle,
		},
	}, nil
}

// setup initialize Servo
func (e *ServoExecutor) setup() error {
	gpio := e.hal.GPIO()

	// Set mode to PWM
	if err := gpio.SetMode(e.config.Pin, hal.PWM); err != nil {
		return err
	}

	// Set frequency (50Hz for servos)
	if err := gpio.SetPWMFrequency(e.config.Pin, int(e.config.Frequency)); err != nil {
		return err
	}

	// Set initial angle
	return e.setAngle(e.config.StartAngle)
}

// setAngle set servo angle
func (e *ServoExecutor) setAngle(angle float64) error {
	// Clamp angle
	if angle < e.config.MinAngle {
		angle = e.config.MinAngle
	}
	if angle > e.config.MaxAngle {
		angle = e.config.MaxAngle
	}

	// Convert angle to pulse width
	angleRange := e.config.MaxAngle - e.config.MinAngle
	pulseRange := e.config.MaxPulse - e.config.MinPulse
	pulseWidth := e.config.MinPulse + ((angle - e.config.MinAngle) / angleRange * pulseRange)

	// Convert pulse width to duty cycle (0-255)
	// duty cycle = (pulse_width_ms / period_ms) * 255
	periodMs := 1000.0 / e.config.Frequency
	dutyCycle := int((pulseWidth / periodMs) * 255.0)

	// Clamp duty cycle
	if dutyCycle < 0 {
		dutyCycle = 0
	}
	if dutyCycle > 255 {
		dutyCycle = 255
	}

	gpio := e.hal.GPIO()
	return gpio.PWMWrite(e.config.Pin, dutyCycle)
}

// Cleanup cleanup resources
func (e *ServoExecutor) Cleanup() error {
	if e.hal != nil {
		// Return to start angle
		e.setAngle(e.config.StartAngle)
	}
	return nil
}
