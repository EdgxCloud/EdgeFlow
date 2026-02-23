package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// VL53L1X Register addresses
const (
	vl53l1xRegSoftReset                    = 0x0000
	vl53l1xRegI2CSlaveDevAddr              = 0x0001
	vl53l1xRegModelID                      = 0x010F
	vl53l1xRegModuleType                   = 0x0110
	vl53l1xRegResultRange                  = 0x0089
	vl53l1xRegResultStatus                 = 0x0089
	vl53l1xRegSystemInterruptClear         = 0x0086
	vl53l1xRegSystemStart                  = 0x0087
	vl53l1xRegTimingBudget                 = 0x0060
	vl53l1xRegDistanceMode                 = 0x004E
	vl53l1xRegROICenter                    = 0x007F
	vl53l1xRegROIXY                        = 0x0081

	// Expected values
	vl53l1xModelID = 0xEA
	vl53l1xModuleType = 0xCC
)

// VL53L1X Distance modes
const (
	vl53l1xDistanceShort = 1 // Up to 1.3m
	vl53l1xDistanceLong  = 2 // Up to 4m
)

// VL53L1XConfig configuration for VL53L1X long-range ToF sensor
type VL53L1XConfig struct {
	Bus          string `json:"bus"`           // I2C bus (default: "")
	Address      int    `json:"address"`       // I2C address (default: 0x29)
	DistanceMode string `json:"distance_mode"` // "short" (1.3m) or "long" (4m)
	TimingBudget int    `json:"timing_budget"` // Timing budget in ms (20, 50, 100, 200, 500)
	MaxRangeMM   int    `json:"max_range_mm"`  // Maximum valid range in mm
}

// VL53L1XExecutor executes VL53L1X ToF sensor readings
type VL53L1XExecutor struct {
	config      VL53L1XConfig
	bus         i2c.BusCloser
	dev         i2c.Dev
	mu          sync.Mutex
	hostInited  bool
	initialized bool
	continuous  bool
	stopChan    chan struct{}
}

// NewVL53L1XExecutor creates a new VL53L1X executor
func NewVL53L1XExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var vlConfig VL53L1XConfig
	if err := json.Unmarshal(configJSON, &vlConfig); err != nil {
		return nil, fmt.Errorf("invalid VL53L1X config: %w", err)
	}

	// Defaults
	if vlConfig.Address == 0 {
		vlConfig.Address = 0x29
	}
	if vlConfig.DistanceMode == "" {
		vlConfig.DistanceMode = "long"
	}
	if vlConfig.TimingBudget == 0 {
		vlConfig.TimingBudget = 50
	}
	if vlConfig.MaxRangeMM == 0 {
		if vlConfig.DistanceMode == "short" {
			vlConfig.MaxRangeMM = 1300
		} else {
			vlConfig.MaxRangeMM = 4000
		}
	}

	return &VL53L1XExecutor{
		config:   vlConfig,
		stopChan: make(chan struct{}),
	}, nil
}

// Init initializes the VL53L1X executor
func (e *VL53L1XExecutor) Init(config map[string]interface{}) error {
	return nil
}

// Execute reads the VL53L1X sensor
func (e *VL53L1XExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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
			return node.Message{}, fmt.Errorf("failed to init VL53L1X: %w", err)
		}
		e.initialized = true
	}

	// Parse command
	if msg.Payload == nil {
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

	case "set_distance_mode":
		mode, _ := payload["mode"].(string)
		if err := e.setDistanceMode(mode); err != nil {
			return node.Message{}, err
		}
		return node.Message{
			Payload: map[string]interface{}{
				"action":        "set_distance_mode",
				"distance_mode": mode,
				"timestamp":     time.Now().Unix(),
			},
		}, nil

	case "set_timing_budget":
		budget := int(getFloat(payload, "budget_ms", 50))
		if err := e.setTimingBudget(budget); err != nil {
			return node.Message{}, err
		}
		return node.Message{
			Payload: map[string]interface{}{
				"action":        "set_timing_budget",
				"timing_budget": budget,
				"timestamp":     time.Now().Unix(),
			},
		}, nil

	case "set_roi":
		// Set Region of Interest
		x := int(getFloat(payload, "x", 16))
		y := int(getFloat(payload, "y", 16))
		center := int(getFloat(payload, "center", 199))
		if err := e.setROI(x, y, byte(center)); err != nil {
			return node.Message{}, err
		}
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "set_roi",
				"x":         x,
				"y":         y,
				"center":    center,
				"timestamp": time.Now().Unix(),
			},
		}, nil

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// initSensor initializes the VL53L1X sensor
func (e *VL53L1XExecutor) initSensor() error {
	// Check model ID
	modelID, err := e.readReg8(vl53l1xRegModelID)
	if err != nil {
		return fmt.Errorf("failed to read model ID: %w", err)
	}
	if modelID != vl53l1xModelID {
		return fmt.Errorf("unexpected model ID: 0x%02X (expected 0x%02X)", modelID, vl53l1xModelID)
	}

	// Software reset
	e.writeReg8(vl53l1xRegSoftReset, 0x00)
	time.Sleep(1 * time.Millisecond)
	e.writeReg8(vl53l1xRegSoftReset, 0x01)
	time.Sleep(100 * time.Millisecond)

	// Wait for boot
	for i := 0; i < 100; i++ {
		val, _ := e.readReg8(0x0061)
		if val&0x01 != 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Load default configuration
	// VL53L1X requires loading extensive default config from ST
	// This is a simplified initialization
	defaultConfig := []struct {
		addr  uint16
		value byte
	}{
		// Essential configuration registers
		{0x002D, 0x05}, // GPIO function
		{0x002E, 0x00}, // GPIO function
		{0x0030, 0x40}, // Algo config
		{0x0031, 0x40}, // Algo config
		{0x0036, 0x00}, // Algo config
		{0x0037, 0x00}, // Algo config
	}

	for _, cfg := range defaultConfig {
		if err := e.writeReg8(cfg.addr, cfg.value); err != nil {
			return err
		}
	}

	// Set distance mode
	if err := e.setDistanceMode(e.config.DistanceMode); err != nil {
		return err
	}

	// Set timing budget
	if err := e.setTimingBudget(e.config.TimingBudget); err != nil {
		return err
	}

	return nil
}

// singleMeasurement performs a single distance measurement
func (e *VL53L1XExecutor) singleMeasurement() (node.Message, error) {
	// Clear interrupt
	e.writeReg8(vl53l1xRegSystemInterruptClear, 0x01)

	// Start measurement
	e.writeReg8(vl53l1xRegSystemStart, 0x40)

	// Wait for data ready
	timeout := time.After(time.Duration(e.config.TimingBudget*2+50) * time.Millisecond)
	for {
		select {
		case <-timeout:
			return node.Message{}, fmt.Errorf("measurement timeout")
		default:
			status, _ := e.readReg8(0x0031)
			if status&0x01 != 0 {
				goto dataReady
			}
			time.Sleep(5 * time.Millisecond)
		}
	}

dataReady:
	// Read result
	result := make([]byte, 17)
	if err := e.readRegMulti(vl53l1xRegResultRange, result); err != nil {
		return node.Message{}, fmt.Errorf("failed to read result: %w", err)
	}

	// Parse results
	rangeStatus := result[0] & 0x1F
	streamCount := result[2]
	distanceMM := int(result[13])<<8 | int(result[14])
	signalRate := (uint32(result[15])<<8 | uint32(result[16])) * 16 // Fixed point 16.16
	ambientRate := (uint32(result[7])<<8 | uint32(result[8])) * 16

	// Clear interrupt and stop
	e.writeReg8(vl53l1xRegSystemInterruptClear, 0x01)
	e.writeReg8(vl53l1xRegSystemStart, 0x00)

	// Decode range status
	var statusStr string
	valid := false
	switch rangeStatus {
	case 0, 4, 7:
		statusStr = "valid"
		valid = true
	case 1:
		statusStr = "sigma_fail"
	case 2:
		statusStr = "signal_fail"
	case 3:
		statusStr = "out_of_bounds_fail"
	case 5:
		statusStr = "hardware_fail"
	case 6:
		statusStr = "wrap_around_fail"
	case 8:
		statusStr = "insufficient_ambient"
	case 14:
		statusStr = "range_invalid"
	default:
		statusStr = "unknown"
	}

	// Validate range
	if distanceMM > e.config.MaxRangeMM || distanceMM <= 0 {
		valid = false
	}

	// Convert to other units
	distanceCM := float64(distanceMM) / 10.0
	distanceM := float64(distanceMM) / 1000.0
	distanceInch := float64(distanceMM) / 25.4

	return node.Message{
		Payload: map[string]interface{}{
			"distance_mm":     distanceMM,
			"distance_cm":     distanceCM,
			"distance_m":      distanceM,
			"distance_inch":   distanceInch,
			"valid":           valid,
			"status":          statusStr,
			"status_code":     rangeStatus,
			"signal_rate":     float64(signalRate) / 65536.0,
			"ambient_rate":    float64(ambientRate) / 65536.0,
			"stream_count":    streamCount,
			"distance_mode":   e.config.DistanceMode,
			"max_range_mm":    e.config.MaxRangeMM,
			"address":         fmt.Sprintf("0x%02X", e.config.Address),
			"sensor":          "VL53L1X",
			"timestamp":       time.Now().Unix(),
		},
	}, nil
}

// continuousMeasurement runs continuous measurements
func (e *VL53L1XExecutor) continuousMeasurement(ctx context.Context, periodMs int) {
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
func (e *VL53L1XExecutor) stopContinuous() {
	if e.continuous {
		select {
		case e.stopChan <- struct{}{}:
		default:
		}
	}
}

// setDistanceMode sets short or long distance mode
func (e *VL53L1XExecutor) setDistanceMode(mode string) error {
	e.config.DistanceMode = mode

	if mode == "short" {
		// Short distance mode (up to 1.3m)
		e.writeReg8(0x004E, 0x00)
		e.writeReg8(0x004F, 0x00)
		e.writeReg8(0x0050, 0x00)
		e.writeReg8(0x0051, 0x00)
		e.config.MaxRangeMM = 1300
	} else {
		// Long distance mode (up to 4m)
		e.writeReg8(0x004E, 0x00)
		e.writeReg8(0x004F, 0x00)
		e.writeReg8(0x0050, 0x00)
		e.writeReg8(0x0051, 0x00)
		e.config.MaxRangeMM = 4000
	}

	return nil
}

// setTimingBudget sets the timing budget in ms
func (e *VL53L1XExecutor) setTimingBudget(budgetMs int) error {
	// Valid values: 20, 50, 100, 200, 500 ms
	// The timing budget affects accuracy and max range
	e.config.TimingBudget = budgetMs

	// Write timing budget
	// The actual register values depend on distance mode
	// This is simplified - full implementation requires more registers
	budgetUs := uint32(budgetMs * 1000)
	e.writeReg32(0x0060, budgetUs)

	return nil
}

// setROI sets the Region of Interest
func (e *VL53L1XExecutor) setROI(x, y int, center byte) error {
	if x < 4 || x > 16 || y < 4 || y > 16 {
		return fmt.Errorf("ROI x and y must be 4-16")
	}

	// Set ROI center
	e.writeReg8(vl53l1xRegROICenter, center)

	// Set ROI size (x | y << 4)
	roiXY := byte(x) | byte(y)<<4
	e.writeReg8(vl53l1xRegROIXY, roiXY)

	return nil
}

// readReg8 reads a single byte from a 16-bit register address
func (e *VL53L1XExecutor) readReg8(reg uint16) (byte, error) {
	data := make([]byte, 1)
	if err := e.dev.Tx([]byte{byte(reg >> 8), byte(reg & 0xFF)}, data); err != nil {
		return 0, err
	}
	return data[0], nil
}

// writeReg8 writes a single byte to a 16-bit register address
func (e *VL53L1XExecutor) writeReg8(reg uint16, value byte) error {
	_, err := e.dev.Write([]byte{byte(reg >> 8), byte(reg & 0xFF), value})
	return err
}

// writeReg32 writes a 32-bit value to a 16-bit register address
func (e *VL53L1XExecutor) writeReg32(reg uint16, value uint32) error {
	_, err := e.dev.Write([]byte{
		byte(reg >> 8), byte(reg & 0xFF),
		byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value),
	})
	return err
}

// readRegMulti reads multiple bytes starting from a 16-bit register address
func (e *VL53L1XExecutor) readRegMulti(reg uint16, data []byte) error {
	return e.dev.Tx([]byte{byte(reg >> 8), byte(reg & 0xFF)}, data)
}

// Cleanup releases resources
func (e *VL53L1XExecutor) Cleanup() error {
	e.stopContinuous()
	if e.bus != nil {
		e.bus.Close()
		e.bus = nil
	}
	e.initialized = false
	return nil
}
