package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/hal"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// MAX31865 Register addresses
const (
	max31865RegConfig    = 0x00
	max31865RegRTDMSB    = 0x01
	max31865RegRTDLSB    = 0x02
	max31865RegHighFault = 0x03
	max31865RegLowFault  = 0x05
	max31865RegFaultStat = 0x07
)

// MAX31865 Configuration bits
const (
	max31865CfgBias        = 0x80 // Bias enable
	max31865CfgModeAuto    = 0x40 // Auto conversion mode
	max31865Cfg1Shot       = 0x20 // 1-shot conversion
	max31865Cfg3Wire       = 0x10 // 3-wire RTD
	max31865CfgFaultClear  = 0x02 // Fault status clear
	max31865Cfg50Hz        = 0x01 // 50Hz filter (vs 60Hz)
)

// MAX31865 Fault bits
const (
	max31865FaultHighThresh = 0x80
	max31865FaultLowThresh  = 0x40
	max31865FaultRefInLow   = 0x20
	max31865FaultRefInHigh  = 0x10
	max31865FaultRTDInLow   = 0x08
	max31865FaultOVUV       = 0x04
)

// RTD Constants
const (
	rtdA = 3.9083e-3
	rtdB = -5.775e-7
	rtdC = -4.183e-12 // Only used below 0°C
)

// MAX31865Config configuration for MAX31865 RTD interface
type MAX31865Config struct {
	SPIBus     int     `json:"spi_bus"`     // SPI bus number (default: 0)
	SPIDevice  int     `json:"spi_device"`  // SPI device number (default: 0)
	Speed      int     `json:"speed"`       // SPI speed in Hz (default: 5MHz)
	CSPin      int     `json:"cs_pin"`      // Chip select GPIO pin
	RTDType    string  `json:"rtd_type"`    // RTD type: "PT100", "PT1000" (default: PT100)
	Wires      int     `json:"wires"`       // Wire configuration: 2, 3, or 4 (default: 2)
	RefRes     float64 `json:"ref_res"`     // Reference resistor (default: 430 for PT100, 4300 for PT1000)
	Filter     int     `json:"filter"`      // Filter: 50 or 60 Hz (default: 60)
	TempOffset float64 `json:"temp_offset"` // Temperature calibration offset
}

// MAX31865Executor executes MAX31865 RTD readings
type MAX31865Executor struct {
	config      MAX31865Config
	hal         hal.HAL
	mu          sync.Mutex
	initialized bool
	rtdNominal  float64 // RTD resistance at 0°C
}

// NewMAX31865Executor creates a new MAX31865 executor
func NewMAX31865Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var rtdConfig MAX31865Config
	if err := json.Unmarshal(configJSON, &rtdConfig); err != nil {
		return nil, fmt.Errorf("invalid MAX31865 config: %w", err)
	}

	// Defaults
	if rtdConfig.Speed == 0 {
		rtdConfig.Speed = 5000000 // 5MHz max
	}
	if rtdConfig.RTDType == "" {
		rtdConfig.RTDType = "PT100"
	}
	if rtdConfig.Wires == 0 {
		rtdConfig.Wires = 2
	}
	if rtdConfig.Filter == 0 {
		rtdConfig.Filter = 60
	}

	// Set nominal resistance and reference resistor based on RTD type
	rtdNominal := 100.0
	if rtdConfig.RTDType == "PT1000" {
		rtdNominal = 1000.0
		if rtdConfig.RefRes == 0 {
			rtdConfig.RefRes = 4300.0
		}
	} else {
		if rtdConfig.RefRes == 0 {
			rtdConfig.RefRes = 430.0
		}
	}

	return &MAX31865Executor{
		config:     rtdConfig,
		rtdNominal: rtdNominal,
	}, nil
}

// Init initializes the MAX31865 executor
func (e *MAX31865Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute reads the MAX31865 RTD sensor
func (e *MAX31865Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get HAL
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h
	}

	// Initialize
	if !e.initialized {
		if err := e.initMAX31865(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init MAX31865: %w", err)
		}
		e.initialized = true
	}

	// Parse command
	payload := msg.Payload
	if payload == nil {
		return e.readTemperature()
	}

	action, _ := payload["action"].(string)

	switch action {
	case "read", "":
		return e.readTemperature()

	case "resistance":
		return e.readResistance()

	case "fault":
		return e.readFault()

	case "clear_fault":
		return e.clearFault()

	case "set_thresholds":
		high := getFloat(payload, "high", 32767)
		low := getFloat(payload, "low", 0)
		return e.setThresholds(uint16(high), uint16(low))

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// initMAX31865 initializes the MAX31865
func (e *MAX31865Executor) initMAX31865() error {
	// Setup CS pin
	if e.config.CSPin > 0 {
		gpio := e.hal.GPIO()
		gpio.SetMode(e.config.CSPin, hal.Output)
		gpio.DigitalWrite(e.config.CSPin, true)
	}

	// Build configuration
	config := byte(max31865CfgBias) // Enable bias

	// Wire configuration
	if e.config.Wires == 3 {
		config |= max31865Cfg3Wire
	}

	// Filter selection
	if e.config.Filter == 50 {
		config |= max31865Cfg50Hz
	}

	// Write configuration
	if err := e.writeRegister(max31865RegConfig, config); err != nil {
		return err
	}

	// Wait for bias to stabilize
	time.Sleep(10 * time.Millisecond)

	return nil
}

// readTemperature reads temperature from the RTD
func (e *MAX31865Executor) readTemperature() (node.Message, error) {
	// Start 1-shot conversion
	config, err := e.readRegister(max31865RegConfig)
	if err != nil {
		return node.Message{}, err
	}
	config |= max31865Cfg1Shot
	if err := e.writeRegister(max31865RegConfig, config); err != nil {
		return node.Message{}, err
	}

	// Wait for conversion (typical 52ms for 60Hz filter, 62.5ms for 50Hz)
	time.Sleep(65 * time.Millisecond)

	// Read RTD value
	rtdMSB, err := e.readRegister(max31865RegRTDMSB)
	if err != nil {
		return node.Message{}, err
	}
	rtdLSB, err := e.readRegister(max31865RegRTDLSB)
	if err != nil {
		return node.Message{}, err
	}

	// Check fault bit
	fault := rtdLSB&0x01 != 0
	var faultType string
	if fault {
		faultStat, _ := e.readRegister(max31865RegFaultStat)
		faultType = e.decodeFault(faultStat)
	}

	// Calculate RTD resistance
	rtdRaw := (uint16(rtdMSB) << 8) | uint16(rtdLSB)
	rtdRaw >>= 1 // Remove fault bit

	resistance := float64(rtdRaw) / 32768.0 * e.config.RefRes

	// Calculate temperature using Callendar-Van Dusen equation
	tempC := e.resistanceToTemperature(resistance)
	tempC += e.config.TempOffset

	tempF := tempC*9/5 + 32

	return node.Message{
		Payload: map[string]interface{}{
			"temperature_c": tempC,
			"temperature_f": tempF,
			"resistance":    resistance,
			"rtd_raw":       rtdRaw,
			"fault":         fault,
			"fault_type":    faultType,
			"valid":         !fault,
			"rtd_type":      e.config.RTDType,
			"wires":         e.config.Wires,
			"sensor":        "MAX31865",
			"timestamp":     time.Now().Unix(),
		},
	}, nil
}

// readResistance returns just the RTD resistance
func (e *MAX31865Executor) readResistance() (node.Message, error) {
	// Start 1-shot conversion
	config, _ := e.readRegister(max31865RegConfig)
	config |= max31865Cfg1Shot
	e.writeRegister(max31865RegConfig, config)
	time.Sleep(65 * time.Millisecond)

	// Read RTD value
	rtdMSB, _ := e.readRegister(max31865RegRTDMSB)
	rtdLSB, _ := e.readRegister(max31865RegRTDLSB)

	rtdRaw := (uint16(rtdMSB) << 8) | uint16(rtdLSB)
	rtdRaw >>= 1

	resistance := float64(rtdRaw) / 32768.0 * e.config.RefRes

	return node.Message{
		Payload: map[string]interface{}{
			"resistance": resistance,
			"rtd_raw":    rtdRaw,
			"ref_res":    e.config.RefRes,
			"rtd_type":   e.config.RTDType,
			"sensor":     "MAX31865",
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

// readFault reads the fault status
func (e *MAX31865Executor) readFault() (node.Message, error) {
	faultStat, err := e.readRegister(max31865RegFaultStat)
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"fault":          faultStat != 0,
			"fault_code":     faultStat,
			"fault_type":     e.decodeFault(faultStat),
			"high_threshold": faultStat&max31865FaultHighThresh != 0,
			"low_threshold":  faultStat&max31865FaultLowThresh != 0,
			"refin_low":      faultStat&max31865FaultRefInLow != 0,
			"refin_high":     faultStat&max31865FaultRefInHigh != 0,
			"rtdin_low":      faultStat&max31865FaultRTDInLow != 0,
			"ovuv":           faultStat&max31865FaultOVUV != 0,
			"sensor":         "MAX31865",
			"timestamp":      time.Now().Unix(),
		},
	}, nil
}

// clearFault clears the fault status
func (e *MAX31865Executor) clearFault() (node.Message, error) {
	config, err := e.readRegister(max31865RegConfig)
	if err != nil {
		return node.Message{}, err
	}

	// Set fault clear bit
	config |= max31865CfgFaultClear
	if err := e.writeRegister(max31865RegConfig, config); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "clear_fault",
			"sensor":    "MAX31865",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// setThresholds sets high and low fault thresholds
func (e *MAX31865Executor) setThresholds(high, low uint16) (node.Message, error) {
	// High threshold (15-bit)
	e.writeRegister(max31865RegHighFault, byte(high>>8))
	e.writeRegister(max31865RegHighFault+1, byte(high&0xFF))

	// Low threshold (15-bit)
	e.writeRegister(max31865RegLowFault, byte(low>>8))
	e.writeRegister(max31865RegLowFault+1, byte(low&0xFF))

	return node.Message{
		Payload: map[string]interface{}{
			"action":         "set_thresholds",
			"high_threshold": high,
			"low_threshold":  low,
			"sensor":         "MAX31865",
			"timestamp":      time.Now().Unix(),
		},
	}, nil
}

// resistanceToTemperature converts RTD resistance to temperature using Callendar-Van Dusen equation
func (e *MAX31865Executor) resistanceToTemperature(resistance float64) float64 {
	// Normalized resistance
	Z1 := -rtdA
	Z2 := rtdA*rtdA - 4*rtdB
	Z3 := 4 * rtdB / e.rtdNominal
	Z4 := 2 * rtdB

	temp := Z2 + Z3*resistance
	if temp < 0 {
		// Negative temperature - use polynomial approximation
		// For temperatures below 0°C, the equation is more complex
		rpoly := resistance
		temp = -242.02
		temp += 2.2228 * rpoly
		temp += 2.5859e-3 * rpoly * rpoly
		temp -= 4.8260e-6 * rpoly * rpoly * rpoly
		temp -= 2.8183e-8 * rpoly * rpoly * rpoly * rpoly
		temp += 1.5243e-10 * rpoly * rpoly * rpoly * rpoly * rpoly
	} else {
		temp = math.Sqrt(temp)
		temp = (temp + Z1) / Z4
	}

	return temp
}

// decodeFault returns a human-readable fault description
func (e *MAX31865Executor) decodeFault(faultStat byte) string {
	if faultStat == 0 {
		return "none"
	}

	var faults []string
	if faultStat&max31865FaultHighThresh != 0 {
		faults = append(faults, "high_threshold")
	}
	if faultStat&max31865FaultLowThresh != 0 {
		faults = append(faults, "low_threshold")
	}
	if faultStat&max31865FaultRefInLow != 0 {
		faults = append(faults, "refin_low")
	}
	if faultStat&max31865FaultRefInHigh != 0 {
		faults = append(faults, "refin_high")
	}
	if faultStat&max31865FaultRTDInLow != 0 {
		faults = append(faults, "rtdin_low")
	}
	if faultStat&max31865FaultOVUV != 0 {
		faults = append(faults, "ovuv")
	}

	if len(faults) == 0 {
		return "unknown"
	}
	return faults[0] // Return first fault
}

// readRegister reads a register
func (e *MAX31865Executor) readRegister(reg byte) (byte, error) {
	spi := e.hal.SPI()
	gpio := e.hal.GPIO()

	// Open SPI device
	if err := spi.Open(e.config.SPIBus, e.config.SPIDevice); err != nil {
		return 0, fmt.Errorf("failed to open SPI: %w", err)
	}

	if e.config.CSPin > 0 {
		gpio.DigitalWrite(e.config.CSPin, false)
		time.Sleep(1 * time.Microsecond)
	}

	// Read command: address with bit 7 clear
	writeData := []byte{reg & 0x7F, 0x00}
	data, err := spi.Transfer(writeData)
	if err != nil {
		if e.config.CSPin > 0 {
			gpio.DigitalWrite(e.config.CSPin, true)
		}
		return 0, err
	}

	if e.config.CSPin > 0 {
		gpio.DigitalWrite(e.config.CSPin, true)
	}

	return data[1], nil
}

// writeRegister writes a register
func (e *MAX31865Executor) writeRegister(reg, value byte) error {
	spi := e.hal.SPI()
	gpio := e.hal.GPIO()

	// Open SPI device
	if err := spi.Open(e.config.SPIBus, e.config.SPIDevice); err != nil {
		return fmt.Errorf("failed to open SPI: %w", err)
	}

	if e.config.CSPin > 0 {
		gpio.DigitalWrite(e.config.CSPin, false)
		time.Sleep(1 * time.Microsecond)
	}

	// Write command: address with bit 7 set
	data := []byte{reg | 0x80, value}
	_, err := spi.Transfer(data)

	if e.config.CSPin > 0 {
		gpio.DigitalWrite(e.config.CSPin, true)
	}

	return err
}

// Cleanup releases resources
func (e *MAX31865Executor) Cleanup() error {
	return nil
}
