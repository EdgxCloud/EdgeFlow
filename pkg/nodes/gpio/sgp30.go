//go:build linux
// +build linux

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

// SGP30 I2C commands
const (
	sgp30DefaultAddr = 0x58

	// Commands (2 bytes each)
	sgp30CmdInitAirQuality     = 0x2003
	sgp30CmdMeasureAirQuality  = 0x2008
	sgp30CmdGetBaseline        = 0x2015
	sgp30CmdSetBaseline        = 0x201E
	sgp30CmdSetHumidity        = 0x2061
	sgp30CmdMeasureTest        = 0x2032
	sgp30CmdGetFeatureSet      = 0x202F
	sgp30CmdMeasureRawSignals  = 0x2050
	sgp30CmdGetSerialID        = 0x3682
)

// SGP30Config holds configuration for SGP30 node
type SGP30Config struct {
	I2CBus       string `json:"i2c_bus"`
	Address      uint16 `json:"address"`
	PollInterval int    `json:"poll_interval_ms"`
}

// SGP30Executor implements air quality sensor
type SGP30Executor struct {
	config       SGP30Config
	bus          i2c.BusCloser
	dev          i2c.Dev
	mu           sync.Mutex
	hostInited   bool
	initialized  bool
	baselineCO2  uint16
	baselineTVOC uint16
}

func (e *SGP30Executor) Init(config map[string]interface{}) error {
	e.config = SGP30Config{
		I2CBus:       "/dev/i2c-1",
		Address:      sgp30DefaultAddr,
		PollInterval: 1000,
	}

	if config != nil {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := json.Unmarshal(configJSON, &e.config); err != nil {
			return fmt.Errorf("failed to parse SGP30 config: %w", err)
		}
	}

	return nil
}

func (e *SGP30Executor) initHardware() error {
	if e.initialized {
		return nil
	}

	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return fmt.Errorf("failed to initialize periph host: %w", err)
		}
		e.hostInited = true
	}

	bus, err := i2creg.Open(e.config.I2CBus)
	if err != nil {
		return fmt.Errorf("failed to open I2C bus %s: %w", e.config.I2CBus, err)
	}
	e.bus = bus
	e.dev = i2c.Dev{Bus: bus, Addr: e.config.Address}

	// Initialize air quality measurement
	if err := e.initAirQuality(); err != nil {
		e.bus.Close()
		return fmt.Errorf("failed to initialize air quality: %w", err)
	}

	e.initialized = true
	return nil
}

func (e *SGP30Executor) sendCommand(cmd uint16) error {
	cmdBytes := []byte{byte(cmd >> 8), byte(cmd & 0xFF)}
	return e.dev.Tx(cmdBytes, nil)
}

func (e *SGP30Executor) sendCommandWithData(cmd uint16, data []byte) error {
	cmdBytes := []byte{byte(cmd >> 8), byte(cmd & 0xFF)}
	cmdBytes = append(cmdBytes, data...)
	return e.dev.Tx(cmdBytes, nil)
}

func (e *SGP30Executor) readResponse(length int) ([]byte, error) {
	data := make([]byte, length)
	if err := e.dev.Tx(nil, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (e *SGP30Executor) sendCommandAndRead(cmd uint16, delay time.Duration, responseLen int) ([]byte, error) {
	if err := e.sendCommand(cmd); err != nil {
		return nil, err
	}
	time.Sleep(delay)
	return e.readResponse(responseLen)
}

// CRC8 calculation for SGP30
func (e *SGP30Executor) crc8(data []byte) byte {
	crc := byte(0xFF)
	for _, b := range data {
		crc ^= b
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ 0x31
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

func (e *SGP30Executor) verifyCRC(data []byte, expectedCRC byte) bool {
	return e.crc8(data) == expectedCRC
}

func (e *SGP30Executor) initAirQuality() error {
	return e.sendCommand(sgp30CmdInitAirQuality)
}

func (e *SGP30Executor) measureAirQuality() (co2 uint16, tvoc uint16, err error) {
	data, err := e.sendCommandAndRead(sgp30CmdMeasureAirQuality, 12*time.Millisecond, 6)
	if err != nil {
		return 0, 0, err
	}

	// Verify CRCs
	if !e.verifyCRC(data[0:2], data[2]) {
		return 0, 0, fmt.Errorf("CO2 CRC mismatch")
	}
	if !e.verifyCRC(data[3:5], data[5]) {
		return 0, 0, fmt.Errorf("TVOC CRC mismatch")
	}

	co2 = uint16(data[0])<<8 | uint16(data[1])
	tvoc = uint16(data[3])<<8 | uint16(data[4])

	return co2, tvoc, nil
}

func (e *SGP30Executor) measureRawSignals() (h2 uint16, ethanol uint16, err error) {
	data, err := e.sendCommandAndRead(sgp30CmdMeasureRawSignals, 25*time.Millisecond, 6)
	if err != nil {
		return 0, 0, err
	}

	// Verify CRCs
	if !e.verifyCRC(data[0:2], data[2]) {
		return 0, 0, fmt.Errorf("H2 CRC mismatch")
	}
	if !e.verifyCRC(data[3:5], data[5]) {
		return 0, 0, fmt.Errorf("Ethanol CRC mismatch")
	}

	h2 = uint16(data[0])<<8 | uint16(data[1])
	ethanol = uint16(data[3])<<8 | uint16(data[4])

	return h2, ethanol, nil
}

func (e *SGP30Executor) getBaseline() (co2Baseline uint16, tvocBaseline uint16, err error) {
	data, err := e.sendCommandAndRead(sgp30CmdGetBaseline, 10*time.Millisecond, 6)
	if err != nil {
		return 0, 0, err
	}

	// Verify CRCs
	if !e.verifyCRC(data[0:2], data[2]) {
		return 0, 0, fmt.Errorf("CO2 baseline CRC mismatch")
	}
	if !e.verifyCRC(data[3:5], data[5]) {
		return 0, 0, fmt.Errorf("TVOC baseline CRC mismatch")
	}

	co2Baseline = uint16(data[0])<<8 | uint16(data[1])
	tvocBaseline = uint16(data[3])<<8 | uint16(data[4])

	return co2Baseline, tvocBaseline, nil
}

func (e *SGP30Executor) setBaseline(co2Baseline, tvocBaseline uint16) error {
	data := []byte{
		byte(tvocBaseline >> 8), byte(tvocBaseline & 0xFF), 0,
		byte(co2Baseline >> 8), byte(co2Baseline & 0xFF), 0,
	}
	// Calculate CRCs
	data[2] = e.crc8(data[0:2])
	data[5] = e.crc8(data[3:5])

	return e.sendCommandWithData(sgp30CmdSetBaseline, data)
}

func (e *SGP30Executor) setHumidity(humidity float64) error {
	// Convert humidity to fixed-point 8.8 format
	// humidity is in g/m³
	humVal := uint16(humidity * 256)
	data := []byte{byte(humVal >> 8), byte(humVal & 0xFF), 0}
	data[2] = e.crc8(data[0:2])

	return e.sendCommandWithData(sgp30CmdSetHumidity, data)
}

func (e *SGP30Executor) getFeatureSet() (uint16, error) {
	data, err := e.sendCommandAndRead(sgp30CmdGetFeatureSet, 10*time.Millisecond, 3)
	if err != nil {
		return 0, err
	}

	if !e.verifyCRC(data[0:2], data[2]) {
		return 0, fmt.Errorf("feature set CRC mismatch")
	}

	return uint16(data[0])<<8 | uint16(data[1]), nil
}

func (e *SGP30Executor) getSerialID() ([]byte, error) {
	data, err := e.sendCommandAndRead(sgp30CmdGetSerialID, 10*time.Millisecond, 9)
	if err != nil {
		return nil, err
	}

	// Verify CRCs for each word
	for i := 0; i < 3; i++ {
		if !e.verifyCRC(data[i*3:i*3+2], data[i*3+2]) {
			return nil, fmt.Errorf("serial ID CRC mismatch at word %d", i)
		}
	}

	serial := []byte{data[0], data[1], data[3], data[4], data[6], data[7]}
	return serial, nil
}

func (e *SGP30Executor) selfTest() (bool, error) {
	data, err := e.sendCommandAndRead(sgp30CmdMeasureTest, 220*time.Millisecond, 3)
	if err != nil {
		return false, err
	}

	if !e.verifyCRC(data[0:2], data[2]) {
		return false, fmt.Errorf("self test CRC mismatch")
	}

	result := uint16(data[0])<<8 | uint16(data[1])
	return result == 0xD400, nil
}

func (e *SGP30Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.initHardware(); err != nil {
		return node.Message{}, err
	}

	action := "read"
	if payload := msg.Payload; payload != nil {
		if a, ok := payload["action"].(string); ok {
			action = a
		}
	}

	switch action {
	case "read":
		return e.readAirQuality()
	case "read_raw":
		return e.readRaw()
	case "get_baseline":
		return e.handleGetBaseline()
	case "set_baseline":
		return e.handleSetBaseline(msg)
	case "set_humidity":
		return e.handleSetHumidity(msg)
	case "self_test":
		return e.handleSelfTest()
	case "info":
		return e.handleInfo()
	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *SGP30Executor) readAirQuality() (node.Message, error) {
	co2, tvoc, err := e.measureAirQuality()
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to measure air quality: %w", err)
	}

	// Interpret air quality
	airQuality := "excellent"
	if co2 > 1500 {
		airQuality = "unhealthy"
	} else if co2 > 1000 {
		airQuality = "poor"
	} else if co2 > 800 {
		airQuality = "moderate"
	} else if co2 > 600 {
		airQuality = "good"
	}

	tvocLevel := "excellent"
	if tvoc > 2200 {
		tvocLevel = "unhealthy"
	} else if tvoc > 660 {
		tvocLevel = "poor"
	} else if tvoc > 220 {
		tvocLevel = "moderate"
	} else if tvoc > 65 {
		tvocLevel = "good"
	}

	return node.Message{
		Payload: map[string]interface{}{
			"eco2":        co2,  // ppm
			"tvoc":        tvoc, // ppb
			"air_quality": airQuality,
			"tvoc_level":  tvocLevel,
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

func (e *SGP30Executor) readRaw() (node.Message, error) {
	h2, ethanol, err := e.measureRawSignals()
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read raw signals: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"h2_raw":      h2,
			"ethanol_raw": ethanol,
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

func (e *SGP30Executor) handleGetBaseline() (node.Message, error) {
	co2Baseline, tvocBaseline, err := e.getBaseline()
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to get baseline: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"co2_baseline":  co2Baseline,
			"tvoc_baseline": tvocBaseline,
		},
	}, nil
}

func (e *SGP30Executor) handleSetBaseline(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	co2Baseline, ok := payload["co2_baseline"].(float64)
	if !ok {
		return node.Message{}, fmt.Errorf("co2_baseline required")
	}

	tvocBaseline, ok := payload["tvoc_baseline"].(float64)
	if !ok {
		return node.Message{}, fmt.Errorf("tvoc_baseline required")
	}

	if err := e.setBaseline(uint16(co2Baseline), uint16(tvocBaseline)); err != nil {
		return node.Message{}, fmt.Errorf("failed to set baseline: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":        "baseline_set",
			"co2_baseline":  uint16(co2Baseline),
			"tvoc_baseline": uint16(tvocBaseline),
		},
	}, nil
}

func (e *SGP30Executor) handleSetHumidity(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	humidity, ok := payload["humidity"].(float64)
	if !ok {
		return node.Message{}, fmt.Errorf("humidity required (g/m³)")
	}

	if err := e.setHumidity(humidity); err != nil {
		return node.Message{}, fmt.Errorf("failed to set humidity: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":   "humidity_set",
			"humidity": humidity,
		},
	}, nil
}

func (e *SGP30Executor) handleSelfTest() (node.Message, error) {
	passed, err := e.selfTest()
	if err != nil {
		return node.Message{}, fmt.Errorf("self test failed: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"self_test_passed": passed,
		},
	}, nil
}

func (e *SGP30Executor) handleInfo() (node.Message, error) {
	serial, err := e.getSerialID()
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to get serial ID: %w", err)
	}

	featureSet, err := e.getFeatureSet()
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to get feature set: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"serial_id":   fmt.Sprintf("%02X%02X%02X%02X%02X%02X", serial[0], serial[1], serial[2], serial[3], serial[4], serial[5]),
			"feature_set": fmt.Sprintf("0x%04X", featureSet),
			"product":     featureSet >> 12,
			"version":     featureSet & 0xFF,
		},
	}, nil
}

func (e *SGP30Executor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized && e.bus != nil {
		e.bus.Close()
		e.initialized = false
	}
	return nil
}
