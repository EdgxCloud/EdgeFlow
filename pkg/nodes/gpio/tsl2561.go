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

// TSL2561 Register addresses
const (
	tsl2561RegControl    = 0x00
	tsl2561RegTiming     = 0x01
	tsl2561RegThreshLL   = 0x02
	tsl2561RegThreshLH   = 0x03
	tsl2561RegThreshHL   = 0x04
	tsl2561RegThreshHH   = 0x05
	tsl2561RegInterrupt  = 0x06
	tsl2561RegID         = 0x0A
	tsl2561RegData0Low   = 0x0C
	tsl2561RegData0High  = 0x0D
	tsl2561RegData1Low   = 0x0E
	tsl2561RegData1High  = 0x0F

	tsl2561Command       = 0x80
	tsl2561CommandClear  = 0xC0
	tsl2561CommandWord   = 0x20

	tsl2561ControlPowerOn  = 0x03
	tsl2561ControlPowerOff = 0x00
)

// TSL2561 Integration times
const (
	tsl2561IntegTime13ms  = 0x00
	tsl2561IntegTime101ms = 0x01
	tsl2561IntegTime402ms = 0x02
)

// TSL2561 Gain settings
const (
	tsl2561Gain1x  = 0x00
	tsl2561Gain16x = 0x10
)

// TSL2561Config configuration for TSL2561 light sensor
type TSL2561Config struct {
	Bus        string `json:"bus"`
	Address    int    `json:"address"`     // 0x29, 0x39 (default), or 0x49
	Gain       string `json:"gain"`        // "1x" or "16x"
	IntegTime  string `json:"integ_time"`  // "13ms", "101ms", "402ms"
	AutoGain   bool   `json:"auto_gain"`
}

// TSL2561Executor executes TSL2561 light sensor readings
type TSL2561Executor struct {
	config      TSL2561Config
	bus         i2c.BusCloser
	dev         i2c.Dev
	mu          sync.Mutex
	hostInited  bool
	initialized bool
	gain        byte
	integTime   byte
}

// NewTSL2561Executor creates a new TSL2561 executor
func NewTSL2561Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var lightConfig TSL2561Config
	if err := json.Unmarshal(configJSON, &lightConfig); err != nil {
		return nil, fmt.Errorf("invalid TSL2561 config: %w", err)
	}

	if lightConfig.Address == 0 {
		lightConfig.Address = 0x39
	}
	if lightConfig.Gain == "" {
		lightConfig.Gain = "1x"
	}
	if lightConfig.IntegTime == "" {
		lightConfig.IntegTime = "402ms"
	}

	return &TSL2561Executor{
		config: lightConfig,
	}, nil
}

// Init initializes the TSL2561 executor
func (e *TSL2561Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute reads the TSL2561 sensor
func (e *TSL2561Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init periph host: %w", err)
		}
		e.hostInited = true
	}

	if e.bus == nil {
		bus, err := i2creg.Open(e.config.Bus)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to open I2C bus: %w", err)
		}
		e.bus = bus
		e.dev = i2c.Dev{Bus: e.bus, Addr: uint16(e.config.Address)}
	}

	if !e.initialized {
		if err := e.initTSL2561(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init TSL2561: %w", err)
		}
		e.initialized = true
	}

	if msg.Payload == nil {
		return e.readLight()
	}
	payload := msg.Payload

	action, _ := payload["action"].(string)

	switch action {
	case "read", "":
		return e.readLight()

	case "set_gain":
		gain, _ := payload["gain"].(string)
		e.setGain(gain)
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "set_gain",
				"gain":      gain,
				"sensor":    "TSL2561",
				"timestamp": time.Now().Unix(),
			},
		}, nil

	case "set_integration":
		integ, _ := payload["integration"].(string)
		e.setIntegrationTime(integ)
		return node.Message{
			Payload: map[string]interface{}{
				"action":      "set_integration",
				"integration": integ,
				"sensor":      "TSL2561",
				"timestamp":   time.Now().Unix(),
			},
		}, nil

	case "id":
		return e.getID()

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// initTSL2561 initializes the sensor
func (e *TSL2561Executor) initTSL2561() error {
	// Power on
	if err := e.writeRegister(tsl2561RegControl, tsl2561ControlPowerOn); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	// Set gain
	e.setGain(e.config.Gain)
	e.setIntegrationTime(e.config.IntegTime)

	return nil
}

// readLight reads light values
func (e *TSL2561Executor) readLight() (node.Message, error) {
	// Power on
	e.writeRegister(tsl2561RegControl, tsl2561ControlPowerOn)

	// Wait for integration
	var waitTime time.Duration
	switch e.integTime {
	case tsl2561IntegTime13ms:
		waitTime = 14 * time.Millisecond
	case tsl2561IntegTime101ms:
		waitTime = 102 * time.Millisecond
	default:
		waitTime = 403 * time.Millisecond
	}
	time.Sleep(waitTime)

	// Read channel 0 (visible + IR)
	ch0Low, _ := e.readRegister(tsl2561RegData0Low)
	ch0High, _ := e.readRegister(tsl2561RegData0High)
	ch0 := uint16(ch0High)<<8 | uint16(ch0Low)

	// Read channel 1 (IR only)
	ch1Low, _ := e.readRegister(tsl2561RegData1Low)
	ch1High, _ := e.readRegister(tsl2561RegData1High)
	ch1 := uint16(ch1High)<<8 | uint16(ch1Low)

	// Power off
	e.writeRegister(tsl2561RegControl, tsl2561ControlPowerOff)

	// Calculate lux
	lux := e.calculateLux(ch0, ch1)

	// Auto-gain adjustment
	if e.config.AutoGain {
		if ch0 < 100 && e.gain == tsl2561Gain1x {
			e.setGain("16x")
		} else if ch0 > 50000 && e.gain == tsl2561Gain16x {
			e.setGain("1x")
		}
	}

	// Light level classification
	var level string
	switch {
	case lux < 1:
		level = "dark"
	case lux < 50:
		level = "dim"
	case lux < 200:
		level = "indoor"
	case lux < 1000:
		level = "bright_indoor"
	case lux < 10000:
		level = "overcast"
	case lux < 50000:
		level = "daylight"
	default:
		level = "direct_sunlight"
	}

	return node.Message{
		Payload: map[string]interface{}{
			"lux":         lux,
			"ch0":         ch0,
			"ch1":         ch1,
			"ir":          ch1,
			"visible":     ch0 - ch1,
			"level":       level,
			"gain":        e.config.Gain,
			"integration": e.config.IntegTime,
			"address":     fmt.Sprintf("0x%02X", e.config.Address),
			"sensor":      "TSL2561",
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

// calculateLux calculates lux from raw readings
func (e *TSL2561Executor) calculateLux(ch0, ch1 uint16) float64 {
	// Scale for gain
	var scale float64 = 1.0
	if e.gain == tsl2561Gain1x {
		scale = 16.0
	}

	// Scale for integration time
	switch e.integTime {
	case tsl2561IntegTime13ms:
		scale *= 402.0 / 13.7
	case tsl2561IntegTime101ms:
		scale *= 402.0 / 101.0
	}

	d0 := float64(ch0) * scale
	d1 := float64(ch1) * scale

	if d0 == 0 {
		return 0
	}

	ratio := d1 / d0

	var lux float64
	if ratio <= 0.50 {
		lux = 0.0304*d0 - 0.062*d0*pow(ratio, 1.4)
	} else if ratio <= 0.61 {
		lux = 0.0224*d0 - 0.031*d1
	} else if ratio <= 0.80 {
		lux = 0.0128*d0 - 0.0153*d1
	} else if ratio <= 1.30 {
		lux = 0.00146*d0 - 0.00112*d1
	} else {
		lux = 0
	}

	if lux < 0 {
		lux = 0
	}

	return lux
}

// setGain sets the gain
func (e *TSL2561Executor) setGain(gain string) {
	if gain == "16x" {
		e.gain = tsl2561Gain16x
		e.config.Gain = "16x"
	} else {
		e.gain = tsl2561Gain1x
		e.config.Gain = "1x"
	}
	e.writeRegister(tsl2561RegTiming, e.gain|e.integTime)
}

// setIntegrationTime sets integration time
func (e *TSL2561Executor) setIntegrationTime(integ string) {
	switch integ {
	case "13ms":
		e.integTime = tsl2561IntegTime13ms
		e.config.IntegTime = "13ms"
	case "101ms":
		e.integTime = tsl2561IntegTime101ms
		e.config.IntegTime = "101ms"
	default:
		e.integTime = tsl2561IntegTime402ms
		e.config.IntegTime = "402ms"
	}
	e.writeRegister(tsl2561RegTiming, e.gain|e.integTime)
}

// getID reads device ID
func (e *TSL2561Executor) getID() (node.Message, error) {
	id, err := e.readRegister(tsl2561RegID)
	if err != nil {
		return node.Message{}, err
	}

	partNo := (id >> 4) & 0x0F
	revNo := id & 0x0F

	var partName string
	switch partNo {
	case 0:
		partName = "TSL2560CS"
	case 1:
		partName = "TSL2561CS"
	case 4:
		partName = "TSL2560T/FN/CL"
	case 5:
		partName = "TSL2561T/FN/CL"
	default:
		partName = "unknown"
	}

	return node.Message{
		Payload: map[string]interface{}{
			"id":        id,
			"part_no":   partNo,
			"rev_no":    revNo,
			"part_name": partName,
			"sensor":    "TSL2561",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// readRegister reads a register
func (e *TSL2561Executor) readRegister(reg byte) (byte, error) {
	data := make([]byte, 1)
	if err := e.dev.Tx([]byte{tsl2561Command | reg}, data); err != nil {
		return 0, err
	}
	return data[0], nil
}

// writeRegister writes a register
func (e *TSL2561Executor) writeRegister(reg, value byte) error {
	_, err := e.dev.Write([]byte{tsl2561Command | reg, value})
	return err
}

// Cleanup releases resources
func (e *TSL2561Executor) Cleanup() error {
	if e.initialized {
		e.writeRegister(tsl2561RegControl, tsl2561ControlPowerOff)
	}
	if e.bus != nil {
		e.bus.Close()
		e.bus = nil
	}
	return nil
}
