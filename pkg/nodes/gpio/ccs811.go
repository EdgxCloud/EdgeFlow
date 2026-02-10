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

// CCS811 I2C registers
const (
	ccs811DefaultAddr = 0x5A
	ccs811AltAddr     = 0x5B

	ccs811RegStatus       = 0x00
	ccs811RegMeasMode     = 0x01
	ccs811RegAlgResultData = 0x02
	ccs811RegRawData      = 0x03
	ccs811RegEnvData      = 0x05
	ccs811RegThresholds   = 0x10
	ccs811RegBaseline     = 0x11
	ccs811RegHWID         = 0x20
	ccs811RegHWVersion    = 0x21
	ccs811RegFWBootVer    = 0x23
	ccs811RegFWAppVer     = 0x24
	ccs811RegErrorID      = 0xE0
	ccs811RegSwReset      = 0xFF
	ccs811BootAppStart    = 0xF4
)

// CCS811 drive modes
const (
	ccs811DriveMode0 = 0x00 // Idle
	ccs811DriveMode1 = 0x10 // 1 second
	ccs811DriveMode2 = 0x20 // 10 seconds
	ccs811DriveMode3 = 0x30 // 60 seconds
	ccs811DriveMode4 = 0x40 // 250ms (raw data only)
)

// CCS811 status bits
const (
	ccs811StatusError     = 0x01
	ccs811StatusDataReady = 0x08
	ccs811StatusAppValid  = 0x10
	ccs811StatusFWMode    = 0x80
)

// CCS811Config holds configuration for CCS811 node
type CCS811Config struct {
	I2CBus       string `json:"i2c_bus"`
	Address      uint16 `json:"address"`
	DriveMode    int    `json:"drive_mode"` // 1, 2, 3, or 4
	EnableInt    bool   `json:"enable_interrupt"`
	PollInterval int    `json:"poll_interval_ms"`
}

// CCS811Executor implements CO2 and TVOC sensor
type CCS811Executor struct {
	config      CCS811Config
	bus         i2c.BusCloser
	dev         i2c.Dev
	mu          sync.Mutex
	hostInited  bool
	initialized bool
	baseline    uint16
}

func (e *CCS811Executor) Init(config map[string]interface{}) error {
	e.config = CCS811Config{
		I2CBus:       "/dev/i2c-1",
		Address:      ccs811DefaultAddr,
		DriveMode:    1,
		EnableInt:    false,
		PollInterval: 1000,
	}

	if config != nil {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := json.Unmarshal(configJSON, &e.config); err != nil {
			return fmt.Errorf("failed to parse CCS811 config: %w", err)
		}
	}

	return nil
}

func (e *CCS811Executor) initHardware() error {
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

	// Verify hardware ID
	hwID, err := e.readByte(ccs811RegHWID)
	if err != nil {
		e.bus.Close()
		return fmt.Errorf("failed to read hardware ID: %w", err)
	}
	if hwID != 0x81 {
		e.bus.Close()
		return fmt.Errorf("invalid hardware ID: expected 0x81, got 0x%02X", hwID)
	}

	// Check if app is valid
	status, err := e.readByte(ccs811RegStatus)
	if err != nil {
		e.bus.Close()
		return fmt.Errorf("failed to read status: %w", err)
	}

	if status&ccs811StatusAppValid == 0 {
		e.bus.Close()
		return fmt.Errorf("no valid application firmware")
	}

	// Start app if in boot mode
	if status&ccs811StatusFWMode == 0 {
		if err := e.writeByte(ccs811BootAppStart); err != nil {
			e.bus.Close()
			return fmt.Errorf("failed to start app: %w", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Configure measurement mode
	if err := e.setDriveMode(e.config.DriveMode); err != nil {
		e.bus.Close()
		return fmt.Errorf("failed to set drive mode: %w", err)
	}

	e.initialized = true
	return nil
}

func (e *CCS811Executor) readByte(reg byte) (byte, error) {
	write := []byte{reg}
	read := make([]byte, 1)
	if err := e.dev.Tx(write, read); err != nil {
		return 0, err
	}
	return read[0], nil
}

func (e *CCS811Executor) readBytes(reg byte, n int) ([]byte, error) {
	write := []byte{reg}
	read := make([]byte, n)
	if err := e.dev.Tx(write, read); err != nil {
		return nil, err
	}
	return read, nil
}

func (e *CCS811Executor) writeByte(data byte) error {
	return e.dev.Tx([]byte{data}, nil)
}

func (e *CCS811Executor) writeRegister(reg byte, data ...byte) error {
	cmd := append([]byte{reg}, data...)
	return e.dev.Tx(cmd, nil)
}

func (e *CCS811Executor) setDriveMode(mode int) error {
	var driveMode byte
	switch mode {
	case 0:
		driveMode = ccs811DriveMode0
	case 1:
		driveMode = ccs811DriveMode1
	case 2:
		driveMode = ccs811DriveMode2
	case 3:
		driveMode = ccs811DriveMode3
	case 4:
		driveMode = ccs811DriveMode4
	default:
		driveMode = ccs811DriveMode1
	}

	// Add interrupt enable if configured
	if e.config.EnableInt {
		driveMode |= 0x08
	}

	return e.writeRegister(ccs811RegMeasMode, driveMode)
}

func (e *CCS811Executor) readAlgResult() (co2 uint16, tvoc uint16, status byte, err error) {
	data, err := e.readBytes(ccs811RegAlgResultData, 5)
	if err != nil {
		return 0, 0, 0, err
	}

	co2 = uint16(data[0])<<8 | uint16(data[1])
	tvoc = uint16(data[2])<<8 | uint16(data[3])
	status = data[4]

	return co2, tvoc, status, nil
}

func (e *CCS811Executor) readRawData() (current uint8, voltage uint16, err error) {
	data, err := e.readBytes(ccs811RegRawData, 2)
	if err != nil {
		return 0, 0, err
	}

	current = data[0] >> 2
	voltage = uint16(data[0]&0x03)<<8 | uint16(data[1])

	return current, voltage, nil
}

func (e *CCS811Executor) setEnvironmentalData(humidity float64, temperature float64) error {
	// Convert to format: humidity in 1/512 %, temperature in 1/512 degrees + 25C offset
	humData := uint16(humidity * 512)
	tempData := uint16((temperature + 25) * 512)

	return e.writeRegister(ccs811RegEnvData,
		byte(humData>>8), byte(humData&0xFF),
		byte(tempData>>8), byte(tempData&0xFF))
}

func (e *CCS811Executor) getBaseline() (uint16, error) {
	data, err := e.readBytes(ccs811RegBaseline, 2)
	if err != nil {
		return 0, err
	}
	return uint16(data[0])<<8 | uint16(data[1]), nil
}

func (e *CCS811Executor) setBaseline(baseline uint16) error {
	return e.writeRegister(ccs811RegBaseline, byte(baseline>>8), byte(baseline&0xFF))
}

func (e *CCS811Executor) getError() (string, error) {
	errID, err := e.readByte(ccs811RegErrorID)
	if err != nil {
		return "", err
	}

	var errors []string
	if errID&0x01 != 0 {
		errors = append(errors, "WRITE_REG_INVALID")
	}
	if errID&0x02 != 0 {
		errors = append(errors, "READ_REG_INVALID")
	}
	if errID&0x04 != 0 {
		errors = append(errors, "MEASMODE_INVALID")
	}
	if errID&0x08 != 0 {
		errors = append(errors, "MAX_RESISTANCE")
	}
	if errID&0x10 != 0 {
		errors = append(errors, "HEATER_FAULT")
	}
	if errID&0x20 != 0 {
		errors = append(errors, "HEATER_SUPPLY")
	}

	if len(errors) == 0 {
		return "NO_ERROR", nil
	}
	return fmt.Sprintf("%v", errors), nil
}

func (e *CCS811Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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
		return e.readData()
	case "read_raw":
		return e.readRaw()
	case "set_environment":
		return e.handleSetEnvironment(msg)
	case "get_baseline":
		return e.handleGetBaseline()
	case "set_baseline":
		return e.handleSetBaseline(msg)
	case "reset":
		return e.handleReset()
	case "info":
		return e.handleInfo()
	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *CCS811Executor) readData() (node.Message, error) {
	// Check if data is ready
	status, err := e.readByte(ccs811RegStatus)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read status: %w", err)
	}

	dataReady := status&ccs811StatusDataReady != 0
	hasError := status&ccs811StatusError != 0

	var errorStr string
	if hasError {
		errorStr, _ = e.getError()
	}

	if !dataReady {
		return node.Message{
			Payload: map[string]interface{}{
				"data_ready": false,
				"error":      errorStr,
				"timestamp":  time.Now().Unix(),
			},
		}, nil
	}

	co2, tvoc, _, err := e.readAlgResult()
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read algorithm result: %w", err)
	}

	// Interpret air quality
	airQuality := "excellent"
	if co2 > 1000 {
		airQuality = "poor"
	} else if co2 > 800 {
		airQuality = "moderate"
	} else if co2 > 600 {
		airQuality = "good"
	}

	return node.Message{
		Payload: map[string]interface{}{
			"eco2":        co2,          // ppm
			"tvoc":        tvoc,         // ppb
			"air_quality": airQuality,
			"data_ready":  true,
			"error":       errorStr,
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

func (e *CCS811Executor) readRaw() (node.Message, error) {
	current, voltage, err := e.readRawData()
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read raw data: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"current_ua": current,
			"voltage":    voltage,
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

func (e *CCS811Executor) handleSetEnvironment(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	humidity := 50.0
	temperature := 25.0

	if h, ok := payload["humidity"].(float64); ok {
		humidity = h
	}
	if t, ok := payload["temperature"].(float64); ok {
		temperature = t
	}

	if err := e.setEnvironmentalData(humidity, temperature); err != nil {
		return node.Message{}, fmt.Errorf("failed to set environmental data: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":      "environment_set",
			"humidity":    humidity,
			"temperature": temperature,
		},
	}, nil
}

func (e *CCS811Executor) handleGetBaseline() (node.Message, error) {
	baseline, err := e.getBaseline()
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to get baseline: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"baseline": baseline,
		},
	}, nil
}

func (e *CCS811Executor) handleSetBaseline(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	baseline, ok := payload["baseline"].(float64)
	if !ok {
		return node.Message{}, fmt.Errorf("baseline value required")
	}

	if err := e.setBaseline(uint16(baseline)); err != nil {
		return node.Message{}, fmt.Errorf("failed to set baseline: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":   "baseline_set",
			"baseline": uint16(baseline),
		},
	}, nil
}

func (e *CCS811Executor) handleReset() (node.Message, error) {
	// Software reset sequence
	resetSeq := []byte{0x11, 0xE5, 0x72, 0x8A}
	if err := e.writeRegister(ccs811RegSwReset, resetSeq...); err != nil {
		return node.Message{}, fmt.Errorf("failed to reset: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Reinitialize
	e.initialized = false
	if err := e.initHardware(); err != nil {
		return node.Message{}, fmt.Errorf("failed to reinitialize after reset: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status": "reset_complete",
		},
	}, nil
}

func (e *CCS811Executor) handleInfo() (node.Message, error) {
	hwID, _ := e.readByte(ccs811RegHWID)
	hwVer, _ := e.readByte(ccs811RegHWVersion)
	fwBootVer, _ := e.readBytes(ccs811RegFWBootVer, 2)
	fwAppVer, _ := e.readBytes(ccs811RegFWAppVer, 2)

	return node.Message{
		Payload: map[string]interface{}{
			"hardware_id":      fmt.Sprintf("0x%02X", hwID),
			"hardware_version": fmt.Sprintf("%d.%d", hwVer>>4, hwVer&0x0F),
			"boot_version":     fmt.Sprintf("%d.%d", fwBootVer[0]>>4, fwBootVer[0]&0x0F),
			"app_version":      fmt.Sprintf("%d.%d", fwAppVer[0]>>4, fwAppVer[0]&0x0F),
		},
	}, nil
}

func (e *CCS811Executor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized && e.bus != nil {
		// Put sensor in idle mode
		e.setDriveMode(0)
		e.bus.Close()
		e.initialized = false
	}
	return nil
}
