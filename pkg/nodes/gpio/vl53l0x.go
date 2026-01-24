package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// VL53L0X Register addresses
const (
	vl53l0xRegSysRangeStart         = 0x00
	vl53l0xRegResultInterruptStatus = 0x13
	vl53l0xRegResultRangeStatus     = 0x14
	vl53l0xRegModelID               = 0xC0
	vl53l0xRegRevisionID            = 0xC2

	// Model ID expected value
	vl53l0xExpectedModelID = 0xEE
)

// VL53L0XConfig configuration for VL53L0X ToF sensor
type VL53L0XConfig struct {
	Bus          string `json:"bus"`           // I2C bus (default: "")
	Address      int    `json:"address"`       // I2C address (default: 0x29)
	Mode         string `json:"mode"`          // Ranging mode: "single", "continuous"
	TimingBudget int    `json:"timing_budget"` // Measurement timing budget in ms (default: 33)
	MaxRangeMM   int    `json:"max_range_mm"`  // Maximum valid range in mm (default: 2000)
}

// VL53L0XExecutor executes VL53L0X ToF sensor readings
type VL53L0XExecutor struct {
	config      VL53L0XConfig
	bus         i2c.BusCloser
	dev         i2c.Dev
	mu          sync.Mutex
	hostInited  bool
	initialized bool
	continuous  bool
	stopChan    chan struct{}
}

// NewVL53L0XExecutor creates a new VL53L0X executor
func NewVL53L0XExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var vlConfig VL53L0XConfig
	if err := json.Unmarshal(configJSON, &vlConfig); err != nil {
		return nil, fmt.Errorf("invalid VL53L0X config: %w", err)
	}

	// Defaults
	if vlConfig.Address == 0 {
		vlConfig.Address = 0x29 // Default VL53L0X address
	}
	if vlConfig.Mode == "" {
		vlConfig.Mode = "single"
	}
	if vlConfig.TimingBudget == 0 {
		vlConfig.TimingBudget = 33 // 33ms default
	}
	if vlConfig.MaxRangeMM == 0 {
		vlConfig.MaxRangeMM = 2000 // 2 meters max
	}

	return &VL53L0XExecutor{
		config:   vlConfig,
		stopChan: make(chan struct{}),
	}, nil
}

// Init initializes the VL53L0X executor
func (e *VL53L0XExecutor) Init(config map[string]interface{}) error {
	return nil
}

// Execute reads the VL53L0X sensor
func (e *VL53L0XExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Initialize hardware
	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init periph host: %w", err)
		}
		e.hostInited = true
	}

	// Open I2C bus
	if e.bus == nil {
		bus, err := i2creg.Open(e.config.Bus)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to open I2C bus: %w", err)
		}
		e.bus = bus
		e.dev = i2c.Dev{Bus: e.bus, Addr: uint16(e.config.Address)}
	}

	// Initialize sensor
	if !e.initialized {
		if err := e.initSensor(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init VL53L0X: %w", err)
		}
		e.initialized = true
	}

	// Parse command
	if msg.Payload == nil {
		// Default: take a single measurement
		return e.singleMeasurement()
	}
	payload := msg.Payload

	action, _ := payload["action"].(string)

	switch action {
	case "read", "measure", "":
		return e.singleMeasurement()

	case "continuous_start":
		period := int(getFloat(payload, "period_ms", 100))
		go e.continuousMeasurement(ctx, period)
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "continuous_start",
				"period_ms": period,
				"timestamp": time.Now().Unix(),
			},
		}, nil

	case "continuous_stop":
		e.stopContinuous()
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "continuous_stop",
				"timestamp": time.Now().Unix(),
			},
		}, nil

	case "calibrate":
		offset, err := e.calibrateOffset()
		if err != nil {
			return node.Message{}, err
		}
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "calibrate",
				"offset_mm": offset,
				"timestamp": time.Now().Unix(),
			},
		}, nil

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// initSensor initializes the VL53L0X sensor
func (e *VL53L0XExecutor) initSensor() error {
	// Check model ID
	modelID, err := e.readReg8(vl53l0xRegModelID)
	if err != nil {
		return fmt.Errorf("failed to read model ID: %w", err)
	}
	if modelID != vl53l0xExpectedModelID {
		return fmt.Errorf("unexpected model ID: 0x%02X (expected 0x%02X)", modelID, vl53l0xExpectedModelID)
	}

	// Basic initialization sequence (simplified)
	// VL53L0X requires complex initialization - for full functionality use ST VL53L0X API

	// Set I2C standard mode
	if err := e.writeReg8(0x88, 0x00); err != nil {
		return err
	}

	// Power on sequence
	if err := e.writeReg8(0x80, 0x01); err != nil {
		return err
	}
	if err := e.writeReg8(0xFF, 0x01); err != nil {
		return err
	}
	if err := e.writeReg8(0x00, 0x00); err != nil {
		return err
	}

	// Load tuning settings
	if err := e.writeReg8(0xFF, 0x06); err != nil {
		return err
	}
	val, _ := e.readReg8(0x83)
	if err := e.writeReg8(0x83, val|0x04); err != nil {
		return err
	}
	if err := e.writeReg8(0xFF, 0x07); err != nil {
		return err
	}
	if err := e.writeReg8(0x81, 0x01); err != nil {
		return err
	}
	if err := e.writeReg8(0x80, 0x01); err != nil {
		return err
	}

	// Set single ranging mode
	if err := e.writeReg8(0xFF, 0x00); err != nil {
		return err
	}
	if err := e.writeReg8(0x09, 0x00); err != nil {
		return err
	}
	if err := e.writeReg8(0x10, 0x00); err != nil {
		return err
	}
	if err := e.writeReg8(0x11, 0x00); err != nil {
		return err
	}

	// Complete init
	if err := e.writeReg8(0xFF, 0x01); err != nil {
		return err
	}
	if err := e.writeReg8(0x00, 0x01); err != nil {
		return err
	}
	if err := e.writeReg8(0xFF, 0x00); err != nil {
		return err
	}
	if err := e.writeReg8(0x80, 0x00); err != nil {
		return err
	}

	return nil
}

// singleMeasurement performs a single distance measurement
func (e *VL53L0XExecutor) singleMeasurement() (node.Message, error) {
	// Start single measurement
	if err := e.writeReg8(vl53l0xRegSysRangeStart, 0x01); err != nil {
		return node.Message{}, fmt.Errorf("failed to start measurement: %w", err)
	}

	// Wait for measurement complete
	timeout := time.After(time.Duration(e.config.TimingBudget*2) * time.Millisecond)
	for {
		select {
		case <-timeout:
			return node.Message{}, fmt.Errorf("measurement timeout")
		default:
			status, _ := e.readReg8(vl53l0xRegResultInterruptStatus)
			if status&0x07 != 0 {
				goto measurementDone
			}
			time.Sleep(1 * time.Millisecond)
		}
	}

measurementDone:
	// Read range result
	rangeStatus, err := e.readReg8(vl53l0xRegResultRangeStatus)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read range status: %w", err)
	}

	// Check for errors
	deviceRangeStatus := (rangeStatus >> 3) & 0x0F
	var rangeError string
	switch deviceRangeStatus {
	case 0:
		rangeError = "none"
	case 1:
		rangeError = "sigma_fail"
	case 2:
		rangeError = "signal_fail"
	case 3:
		rangeError = "min_range_fail"
	case 4:
		rangeError = "phase_fail"
	case 5:
		rangeError = "hardware_fail"
	default:
		rangeError = "unknown"
	}

	// Read distance (2 bytes at offset 10 from result register base)
	distData := make([]byte, 2)
	if err := e.dev.Tx([]byte{vl53l0xRegResultRangeStatus + 10}, distData); err != nil {
		return node.Message{}, fmt.Errorf("failed to read distance: %w", err)
	}
	distanceMM := int(distData[0])<<8 | int(distData[1])

	// Clear interrupt
	e.writeReg8(vl53l0xRegResultInterruptStatus, 0x01)

	// Validate range
	valid := deviceRangeStatus == 0 && distanceMM > 0 && distanceMM < e.config.MaxRangeMM

	// Convert to other units
	distanceCM := float64(distanceMM) / 10.0
	distanceM := float64(distanceMM) / 1000.0
	distanceInch := float64(distanceMM) / 25.4

	return node.Message{
		Payload: map[string]interface{}{
			"distance_mm":   distanceMM,
			"distance_cm":   distanceCM,
			"distance_m":    distanceM,
			"distance_inch": distanceInch,
			"valid":         valid,
			"error":         rangeError,
			"address":       fmt.Sprintf("0x%02X", e.config.Address),
			"sensor":        "VL53L0X",
			"timestamp":     time.Now().Unix(),
		},
	}, nil
}

// continuousMeasurement runs continuous measurements
func (e *VL53L0XExecutor) continuousMeasurement(ctx context.Context, periodMs int) {
	e.continuous = true
	ticker := time.NewTicker(time.Duration(periodMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.continuous = false
			return
		case <-e.stopChan:
			e.continuous = false
			return
		case <-ticker.C:
			e.mu.Lock()
			_, _ = e.singleMeasurement()
			e.mu.Unlock()
		}
	}
}

// stopContinuous stops continuous measurement
func (e *VL53L0XExecutor) stopContinuous() {
	if e.continuous {
		select {
		case e.stopChan <- struct{}{}:
		default:
		}
	}
}

// calibrateOffset performs offset calibration
func (e *VL53L0XExecutor) calibrateOffset() (int, error) {
	const numSamples = 10
	var total int

	for i := 0; i < numSamples; i++ {
		msg, err := e.singleMeasurement()
		if err != nil {
			return 0, err
		}
		total += msg.Payload["distance_mm"].(int)
		time.Sleep(50 * time.Millisecond)
	}

	return total / numSamples, nil
}

// readReg8 reads a single byte from register
func (e *VL53L0XExecutor) readReg8(reg byte) (byte, error) {
	data := make([]byte, 1)
	if err := e.dev.Tx([]byte{reg}, data); err != nil {
		return 0, err
	}
	return data[0], nil
}

// writeReg8 writes a single byte to register
func (e *VL53L0XExecutor) writeReg8(reg, value byte) error {
	_, err := e.dev.Write([]byte{reg, value})
	return err
}

// Cleanup releases resources
func (e *VL53L0XExecutor) Cleanup() error {
	e.stopContinuous()
	if e.bus != nil {
		e.bus.Close()
		e.bus = nil
	}
	e.initialized = false
	return nil
}
