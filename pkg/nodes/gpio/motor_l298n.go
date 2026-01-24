package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/node"
)

// MotorL298NConfig configuration for L298N motor driver
type MotorL298NConfig struct {
	// Motor A pins
	ENA int `json:"ena"` // Enable A (PWM)
	IN1 int `json:"in1"` // Input 1
	IN2 int `json:"in2"` // Input 2

	// Motor B pins (optional)
	ENB int `json:"enb"` // Enable B (PWM)
	IN3 int `json:"in3"` // Input 3
	IN4 int `json:"in4"` // Input 4

	// PWM settings
	PWMFrequency int `json:"pwm_frequency"` // PWM frequency in Hz (default: 1000)
	MaxSpeed     int `json:"max_speed"`     // Max speed 0-255 (default: 255)
}

// MotorL298NExecutor controls DC motors via L298N driver
type MotorL298NExecutor struct {
	config     MotorL298NConfig
	hal        hal.HAL
	mu         sync.Mutex
	initialized bool
	speedA     int
	speedB     int
	dirA       string // "forward", "backward", "stop"
	dirB       string
}

// NewMotorL298NExecutor creates a new L298N motor executor
func NewMotorL298NExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var motorConfig MotorL298NConfig
	if err := json.Unmarshal(configJSON, &motorConfig); err != nil {
		return nil, fmt.Errorf("invalid motor config: %w", err)
	}

	// Validate required pins for Motor A
	if motorConfig.ENA == 0 || motorConfig.IN1 == 0 || motorConfig.IN2 == 0 {
		return nil, fmt.Errorf("Motor A requires ENA, IN1, and IN2 pins")
	}

	// Defaults
	if motorConfig.PWMFrequency == 0 {
		motorConfig.PWMFrequency = 1000
	}
	if motorConfig.MaxSpeed == 0 {
		motorConfig.MaxSpeed = 255
	}

	return &MotorL298NExecutor{
		config: motorConfig,
		dirA:   "stop",
		dirB:   "stop",
	}, nil
}

// Init initializes the motor executor
func (e *MotorL298NExecutor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles motor control commands
func (e *MotorL298NExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get HAL if not initialized
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h
	}

	// Initialize pins if not done
	if !e.initialized {
		if err := e.initPins(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init motor pins: %w", err)
		}
		e.initialized = true
	}

	// Parse command from message
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	action, _ := payload["action"].(string)
	motor, _ := payload["motor"].(string)
	if motor == "" {
		motor = "A"
	}

	speed := int(getFloat(payload, "speed", 0))
	if speed > e.config.MaxSpeed {
		speed = e.config.MaxSpeed
	}
	if speed < 0 {
		speed = 0
	}

	switch action {
	case "forward":
		if motor == "A" || motor == "both" {
			e.setMotorA("forward", speed)
		}
		if motor == "B" || motor == "both" {
			e.setMotorB("forward", speed)
		}

	case "backward":
		if motor == "A" || motor == "both" {
			e.setMotorA("backward", speed)
		}
		if motor == "B" || motor == "both" {
			e.setMotorB("backward", speed)
		}

	case "stop":
		if motor == "A" || motor == "both" {
			e.setMotorA("stop", 0)
		}
		if motor == "B" || motor == "both" {
			e.setMotorB("stop", 0)
		}

	case "brake":
		if motor == "A" || motor == "both" {
			e.brakeMotorA()
		}
		if motor == "B" || motor == "both" {
			e.brakeMotorB()
		}

	case "speed":
		// Just update speed without changing direction
		if motor == "A" || motor == "both" {
			e.setSpeed("A", speed)
		}
		if motor == "B" || motor == "both" {
			e.setSpeed("B", speed)
		}

	case "turn_left":
		// Motor A backward, Motor B forward
		turnSpeed := int(getFloat(payload, "speed", 128))
		e.setMotorA("backward", turnSpeed)
		e.setMotorB("forward", turnSpeed)

	case "turn_right":
		// Motor A forward, Motor B backward
		turnSpeed := int(getFloat(payload, "speed", 128))
		e.setMotorA("forward", turnSpeed)
		e.setMotorB("backward", turnSpeed)

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":      action,
			"motor":       motor,
			"motor_a":     map[string]interface{}{"direction": e.dirA, "speed": e.speedA},
			"motor_b":     map[string]interface{}{"direction": e.dirB, "speed": e.speedB},
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

// initPins initializes GPIO pins
func (e *MotorL298NExecutor) initPins() error {
	gpio := e.hal.GPIO()

	// Motor A pins
	gpio.SetMode(e.config.ENA, hal.Output)
	gpio.SetMode(e.config.IN1, hal.Output)
	gpio.SetMode(e.config.IN2, hal.Output)

	// Set PWM frequency
	gpio.SetPWMFrequency(e.config.ENA, e.config.PWMFrequency)

	// Motor B pins (if configured)
	if e.config.ENB != 0 && e.config.IN3 != 0 && e.config.IN4 != 0 {
		gpio.SetMode(e.config.ENB, hal.Output)
		gpio.SetMode(e.config.IN3, hal.Output)
		gpio.SetMode(e.config.IN4, hal.Output)
		gpio.SetPWMFrequency(e.config.ENB, e.config.PWMFrequency)
	}

	// Start with motors stopped
	e.setMotorA("stop", 0)
	e.setMotorB("stop", 0)

	return nil
}

// setMotorA sets Motor A direction and speed
func (e *MotorL298NExecutor) setMotorA(direction string, speed int) {
	gpio := e.hal.GPIO()

	switch direction {
	case "forward":
		gpio.DigitalWrite(e.config.IN1, true)
		gpio.DigitalWrite(e.config.IN2, false)
	case "backward":
		gpio.DigitalWrite(e.config.IN1, false)
		gpio.DigitalWrite(e.config.IN2, true)
	case "stop":
		gpio.DigitalWrite(e.config.IN1, false)
		gpio.DigitalWrite(e.config.IN2, false)
		speed = 0
	}

	gpio.PWMWrite(e.config.ENA, speed)
	e.dirA = direction
	e.speedA = speed
}

// setMotorB sets Motor B direction and speed
func (e *MotorL298NExecutor) setMotorB(direction string, speed int) {
	if e.config.ENB == 0 {
		return // Motor B not configured
	}

	gpio := e.hal.GPIO()

	switch direction {
	case "forward":
		gpio.DigitalWrite(e.config.IN3, true)
		gpio.DigitalWrite(e.config.IN4, false)
	case "backward":
		gpio.DigitalWrite(e.config.IN3, false)
		gpio.DigitalWrite(e.config.IN4, true)
	case "stop":
		gpio.DigitalWrite(e.config.IN3, false)
		gpio.DigitalWrite(e.config.IN4, false)
		speed = 0
	}

	gpio.PWMWrite(e.config.ENB, speed)
	e.dirB = direction
	e.speedB = speed
}

// brakeMotorA applies brake to Motor A
func (e *MotorL298NExecutor) brakeMotorA() {
	gpio := e.hal.GPIO()
	// Both inputs HIGH = brake
	gpio.DigitalWrite(e.config.IN1, true)
	gpio.DigitalWrite(e.config.IN2, true)
	gpio.PWMWrite(e.config.ENA, 255)
	e.dirA = "brake"
	e.speedA = 0
}

// brakeMotorB applies brake to Motor B
func (e *MotorL298NExecutor) brakeMotorB() {
	if e.config.ENB == 0 {
		return
	}

	gpio := e.hal.GPIO()
	gpio.DigitalWrite(e.config.IN3, true)
	gpio.DigitalWrite(e.config.IN4, true)
	gpio.PWMWrite(e.config.ENB, 255)
	e.dirB = "brake"
	e.speedB = 0
}

// setSpeed sets speed without changing direction
func (e *MotorL298NExecutor) setSpeed(motor string, speed int) {
	gpio := e.hal.GPIO()

	if motor == "A" {
		gpio.PWMWrite(e.config.ENA, speed)
		e.speedA = speed
	} else if motor == "B" && e.config.ENB != 0 {
		gpio.PWMWrite(e.config.ENB, speed)
		e.speedB = speed
	}
}

// Cleanup releases resources and stops motors
func (e *MotorL298NExecutor) Cleanup() error {
	if e.hal != nil && e.initialized {
		// Stop both motors
		e.setMotorA("stop", 0)
		e.setMotorB("stop", 0)
	}
	e.initialized = false
	return nil
}

// StepperMotorConfig configuration for stepper motor
type StepperMotorConfig struct {
	Step      int    `json:"step"`       // Step pin
	Direction int    `json:"direction"`  // Direction pin
	Enable    int    `json:"enable"`     // Enable pin (optional)
	MS1       int    `json:"ms1"`        // Microstep pin 1 (optional)
	MS2       int    `json:"ms2"`        // Microstep pin 2 (optional)
	MS3       int    `json:"ms3"`        // Microstep pin 3 (optional)
	StepsPerRev int  `json:"steps_per_rev"` // Steps per revolution (default: 200)
	Microsteps int   `json:"microsteps"` // Microstepping (1, 2, 4, 8, 16, 32)
}

// StepperMotorExecutor controls stepper motors via A4988/DRV8825
type StepperMotorExecutor struct {
	config      StepperMotorConfig
	hal         hal.HAL
	mu          sync.Mutex
	initialized bool
	position    int64 // Current position in steps
	enabled     bool
}

// NewStepperMotorExecutor creates a new stepper motor executor
func NewStepperMotorExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var stepperConfig StepperMotorConfig
	if err := json.Unmarshal(configJSON, &stepperConfig); err != nil {
		return nil, fmt.Errorf("invalid stepper config: %w", err)
	}

	// Validate required pins
	if stepperConfig.Step == 0 || stepperConfig.Direction == 0 {
		return nil, fmt.Errorf("stepper requires step and direction pins")
	}

	// Defaults
	if stepperConfig.StepsPerRev == 0 {
		stepperConfig.StepsPerRev = 200 // Common NEMA 17
	}
	if stepperConfig.Microsteps == 0 {
		stepperConfig.Microsteps = 1
	}

	return &StepperMotorExecutor{
		config:  stepperConfig,
		enabled: true,
	}, nil
}

// Init initializes the stepper executor
func (e *StepperMotorExecutor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles stepper motor commands
func (e *StepperMotorExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get HAL if not initialized
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h
	}

	// Initialize pins if not done
	if !e.initialized {
		if err := e.initPins(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init stepper pins: %w", err)
		}
		e.initialized = true
	}

	// Parse command
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	action, _ := payload["action"].(string)

	switch action {
	case "step":
		steps := int(getFloat(payload, "steps", 1))
		delayUs := int(getFloat(payload, "delay_us", 1000))
		if err := e.step(ctx, steps, delayUs); err != nil {
			return node.Message{}, err
		}

	case "rotate":
		degrees := getFloat(payload, "degrees", 0)
		rpm := getFloat(payload, "rpm", 60)
		if err := e.rotate(ctx, degrees, rpm); err != nil {
			return node.Message{}, err
		}

	case "home":
		e.position = 0

	case "enable":
		e.setEnable(true)

	case "disable":
		e.setEnable(false)

	case "set_microsteps":
		microsteps := int(getFloat(payload, "microsteps", 1))
		e.setMicrosteps(microsteps)

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":     action,
			"position":   e.position,
			"enabled":    e.enabled,
			"microsteps": e.config.Microsteps,
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

// initPins initializes stepper pins
func (e *StepperMotorExecutor) initPins() error {
	gpio := e.hal.GPIO()

	gpio.SetMode(e.config.Step, hal.Output)
	gpio.SetMode(e.config.Direction, hal.Output)

	if e.config.Enable != 0 {
		gpio.SetMode(e.config.Enable, hal.Output)
		gpio.DigitalWrite(e.config.Enable, false) // Enable (active low)
	}

	// Microstep pins
	if e.config.MS1 != 0 {
		gpio.SetMode(e.config.MS1, hal.Output)
	}
	if e.config.MS2 != 0 {
		gpio.SetMode(e.config.MS2, hal.Output)
	}
	if e.config.MS3 != 0 {
		gpio.SetMode(e.config.MS3, hal.Output)
	}

	return nil
}

// step performs a number of steps
func (e *StepperMotorExecutor) step(ctx context.Context, steps int, delayUs int) error {
	gpio := e.hal.GPIO()

	// Set direction
	if steps >= 0 {
		gpio.DigitalWrite(e.config.Direction, true)
	} else {
		gpio.DigitalWrite(e.config.Direction, false)
		steps = -steps
	}

	delay := time.Duration(delayUs) * time.Microsecond

	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		gpio.DigitalWrite(e.config.Step, true)
		time.Sleep(delay / 2)
		gpio.DigitalWrite(e.config.Step, false)
		time.Sleep(delay / 2)

		if steps >= 0 {
			e.position++
		} else {
			e.position--
		}
	}

	return nil
}

// rotate rotates by degrees at given RPM
func (e *StepperMotorExecutor) rotate(ctx context.Context, degrees float64, rpm float64) error {
	// Calculate steps
	totalSteps := e.config.StepsPerRev * e.config.Microsteps
	steps := int(degrees / 360.0 * float64(totalSteps))

	// Calculate delay from RPM
	// RPM = (60 * 1000000) / (steps_per_rev * microsteps * delay_us)
	delayUs := int(60.0 * 1000000.0 / (float64(totalSteps) * rpm))
	if delayUs < 1 {
		delayUs = 1
	}

	return e.step(ctx, steps, delayUs)
}

// setEnable enables or disables the stepper
func (e *StepperMotorExecutor) setEnable(enable bool) {
	if e.config.Enable != 0 {
		gpio := e.hal.GPIO()
		gpio.DigitalWrite(e.config.Enable, !enable) // Active low
		e.enabled = enable
	}
}

// setMicrosteps configures microstepping
func (e *StepperMotorExecutor) setMicrosteps(microsteps int) {
	gpio := e.hal.GPIO()
	e.config.Microsteps = microsteps

	// A4988/DRV8825 microstep table
	// MS1 MS2 MS3 -> Microsteps
	//  0   0   0  -> Full step
	//  1   0   0  -> Half step
	//  0   1   0  -> Quarter step
	//  1   1   0  -> Eighth step
	//  1   1   1  -> Sixteenth step

	ms1, ms2, ms3 := false, false, false
	switch microsteps {
	case 2:
		ms1 = true
	case 4:
		ms2 = true
	case 8:
		ms1, ms2 = true, true
	case 16, 32:
		ms1, ms2, ms3 = true, true, true
	}

	if e.config.MS1 != 0 {
		gpio.DigitalWrite(e.config.MS1, ms1)
	}
	if e.config.MS2 != 0 {
		gpio.DigitalWrite(e.config.MS2, ms2)
	}
	if e.config.MS3 != 0 {
		gpio.DigitalWrite(e.config.MS3, ms3)
	}
}

// Cleanup releases resources
func (e *StepperMotorExecutor) Cleanup() error {
	if e.initialized && e.hal != nil {
		e.setEnable(false)
	}
	e.initialized = false
	return nil
}
