//go:build linux
// +build linux

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

// ADS1015 I2C registers
const (
	ads1015DefaultAddr = 0x48

	ads1015RegConversion = 0x00
	ads1015RegConfig     = 0x01
	ads1015RegLoThresh   = 0x02
	ads1015RegHiThresh   = 0x03
)

// ADS1015 config register bits
const (
	// Operational status
	ads1015ConfigOSSingle = 0x8000 // Start single conversion

	// Input multiplexer
	ads1015ConfigMuxDiff01  = 0x0000 // AIN0 - AIN1
	ads1015ConfigMuxDiff03  = 0x1000 // AIN0 - AIN3
	ads1015ConfigMuxDiff13  = 0x2000 // AIN1 - AIN3
	ads1015ConfigMuxDiff23  = 0x3000 // AIN2 - AIN3
	ads1015ConfigMuxSingle0 = 0x4000 // AIN0
	ads1015ConfigMuxSingle1 = 0x5000 // AIN1
	ads1015ConfigMuxSingle2 = 0x6000 // AIN2
	ads1015ConfigMuxSingle3 = 0x7000 // AIN3

	// Programmable gain amplifier
	ads1015ConfigPGA6144 = 0x0000 // +/-6.144V
	ads1015ConfigPGA4096 = 0x0200 // +/-4.096V
	ads1015ConfigPGA2048 = 0x0400 // +/-2.048V (default)
	ads1015ConfigPGA1024 = 0x0600 // +/-1.024V
	ads1015ConfigPGA512  = 0x0800 // +/-0.512V
	ads1015ConfigPGA256  = 0x0A00 // +/-0.256V

	// Operating mode
	ads1015ConfigModeContinuous = 0x0000
	ads1015ConfigModeSingle     = 0x0100

	// Data rate
	ads1015ConfigDR128  = 0x0000 // 128 SPS
	ads1015ConfigDR250  = 0x0020 // 250 SPS
	ads1015ConfigDR490  = 0x0040 // 490 SPS
	ads1015ConfigDR920  = 0x0060 // 920 SPS
	ads1015ConfigDR1600 = 0x0080 // 1600 SPS (default)
	ads1015ConfigDR2400 = 0x00A0 // 2400 SPS
	ads1015ConfigDR3300 = 0x00C0 // 3300 SPS

	// Comparator mode
	ads1015ConfigCompTraditional = 0x0000
	ads1015ConfigCompWindow      = 0x0010

	// Comparator polarity
	ads1015ConfigCompPolActiveLow  = 0x0000
	ads1015ConfigCompPolActiveHigh = 0x0008

	// Comparator latch
	ads1015ConfigCompNonLatch = 0x0000
	ads1015ConfigCompLatch    = 0x0004

	// Comparator queue
	ads1015ConfigCompQue1    = 0x0000
	ads1015ConfigCompQue2    = 0x0001
	ads1015ConfigCompQue4    = 0x0002
	ads1015ConfigCompQueNone = 0x0003
)

// ADS1015Config holds configuration for ADS1015 ADC
type ADS1015Config struct {
	I2CBus       string  `json:"i2c_bus"`
	Address      uint16  `json:"address"`
	Gain         string  `json:"gain"`      // "6.144", "4.096", "2.048", "1.024", "0.512", "0.256"
	DataRate     int     `json:"data_rate"` // 128, 250, 490, 920, 1600, 2400, 3300
	PollInterval int     `json:"poll_interval_ms"`
	VRef         float64 `json:"vref"` // For percentage calculations
}

// ADS1015Executor implements 12-bit ADC
type ADS1015Executor struct {
	config      ADS1015Config
	bus         i2c.BusCloser
	dev         i2c.Dev
	mu          sync.Mutex
	hostInited  bool
	initialized bool
	gainConfig  uint16
	gainVoltage float64
	drConfig    uint16
}

func (e *ADS1015Executor) Init(config map[string]interface{}) error {
	e.config = ADS1015Config{
		I2CBus:       "/dev/i2c-1",
		Address:      ads1015DefaultAddr,
		Gain:         "4.096",
		DataRate:     1600,
		PollInterval: 100,
		VRef:         3.3,
	}

	if config != nil {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := json.Unmarshal(configJSON, &e.config); err != nil {
			return fmt.Errorf("failed to parse ADS1015 config: %w", err)
		}
	}

	// Parse gain setting
	switch e.config.Gain {
	case "6.144":
		e.gainConfig = ads1015ConfigPGA6144
		e.gainVoltage = 6.144
	case "4.096":
		e.gainConfig = ads1015ConfigPGA4096
		e.gainVoltage = 4.096
	case "2.048":
		e.gainConfig = ads1015ConfigPGA2048
		e.gainVoltage = 2.048
	case "1.024":
		e.gainConfig = ads1015ConfigPGA1024
		e.gainVoltage = 1.024
	case "0.512":
		e.gainConfig = ads1015ConfigPGA512
		e.gainVoltage = 0.512
	case "0.256":
		e.gainConfig = ads1015ConfigPGA256
		e.gainVoltage = 0.256
	default:
		e.gainConfig = ads1015ConfigPGA4096
		e.gainVoltage = 4.096
	}

	// Parse data rate setting
	switch e.config.DataRate {
	case 128:
		e.drConfig = ads1015ConfigDR128
	case 250:
		e.drConfig = ads1015ConfigDR250
	case 490:
		e.drConfig = ads1015ConfigDR490
	case 920:
		e.drConfig = ads1015ConfigDR920
	case 1600:
		e.drConfig = ads1015ConfigDR1600
	case 2400:
		e.drConfig = ads1015ConfigDR2400
	case 3300:
		e.drConfig = ads1015ConfigDR3300
	default:
		e.drConfig = ads1015ConfigDR1600
	}

	return nil
}

func (e *ADS1015Executor) initHardware() error {
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

	e.initialized = true
	return nil
}

func (e *ADS1015Executor) writeConfig(config uint16) error {
	cmd := []byte{ads1015RegConfig, byte(config >> 8), byte(config & 0xFF)}
	return e.dev.Tx(cmd, nil)
}

func (e *ADS1015Executor) readConfig() (uint16, error) {
	write := []byte{ads1015RegConfig}
	read := make([]byte, 2)
	if err := e.dev.Tx(write, read); err != nil {
		return 0, err
	}
	return uint16(read[0])<<8 | uint16(read[1]), nil
}

func (e *ADS1015Executor) readConversion() (int16, error) {
	write := []byte{ads1015RegConversion}
	read := make([]byte, 2)
	if err := e.dev.Tx(write, read); err != nil {
		return 0, err
	}
	// ADS1015 is 12-bit, left-justified in 16-bit register
	return int16(uint16(read[0])<<8|uint16(read[1])) >> 4, nil
}

func (e *ADS1015Executor) readChannel(channel int) (int16, float64, error) {
	var muxConfig uint16

	switch channel {
	case 0:
		muxConfig = ads1015ConfigMuxSingle0
	case 1:
		muxConfig = ads1015ConfigMuxSingle1
	case 2:
		muxConfig = ads1015ConfigMuxSingle2
	case 3:
		muxConfig = ads1015ConfigMuxSingle3
	default:
		return 0, 0, fmt.Errorf("invalid channel: %d", channel)
	}

	// Build config: single-shot, selected channel, selected gain and data rate
	config := ads1015ConfigOSSingle | muxConfig | e.gainConfig |
		ads1015ConfigModeSingle | e.drConfig | ads1015ConfigCompQueNone

	// Start conversion
	if err := e.writeConfig(config); err != nil {
		return 0, 0, fmt.Errorf("failed to start conversion: %w", err)
	}

	// Wait for conversion (depends on data rate)
	time.Sleep(time.Duration(1000000/e.config.DataRate+1) * time.Microsecond)

	// Wait for conversion complete
	for i := 0; i < 100; i++ {
		cfg, err := e.readConfig()
		if err != nil {
			return 0, 0, err
		}
		if cfg&0x8000 != 0 {
			break
		}
		time.Sleep(100 * time.Microsecond)
	}

	// Read result
	raw, err := e.readConversion()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read conversion: %w", err)
	}

	// Convert to voltage (12-bit ADC, range is +/- gainVoltage)
	voltage := float64(raw) * e.gainVoltage / 2048.0

	return raw, voltage, nil
}

func (e *ADS1015Executor) readDifferential(posChannel, negChannel int) (int16, float64, error) {
	var muxConfig uint16

	switch {
	case posChannel == 0 && negChannel == 1:
		muxConfig = ads1015ConfigMuxDiff01
	case posChannel == 0 && negChannel == 3:
		muxConfig = ads1015ConfigMuxDiff03
	case posChannel == 1 && negChannel == 3:
		muxConfig = ads1015ConfigMuxDiff13
	case posChannel == 2 && negChannel == 3:
		muxConfig = ads1015ConfigMuxDiff23
	default:
		return 0, 0, fmt.Errorf("invalid differential pair: AIN%d - AIN%d", posChannel, negChannel)
	}

	config := ads1015ConfigOSSingle | muxConfig | e.gainConfig |
		ads1015ConfigModeSingle | e.drConfig | ads1015ConfigCompQueNone

	if err := e.writeConfig(config); err != nil {
		return 0, 0, err
	}

	time.Sleep(time.Duration(1000000/e.config.DataRate+1) * time.Microsecond)

	for i := 0; i < 100; i++ {
		cfg, err := e.readConfig()
		if err != nil {
			return 0, 0, err
		}
		if cfg&0x8000 != 0 {
			break
		}
		time.Sleep(100 * time.Microsecond)
	}

	raw, err := e.readConversion()
	if err != nil {
		return 0, 0, err
	}

	voltage := float64(raw) * e.gainVoltage / 2048.0

	return raw, voltage, nil
}

func (e *ADS1015Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.initHardware(); err != nil {
		return node.Message{}, err
	}

	action := "read"
	channel := 0
	payload := msg.Payload
	if payload != nil {
		if a, ok := payload["action"].(string); ok {
			action = a
		}
		if c, ok := payload["channel"].(float64); ok {
			channel = int(c)
		}
	}

	switch action {
	case "read":
		return e.handleReadChannel(channel)
	case "read_all":
		return e.handleReadAll()
	case "read_differential":
		return e.handleReadDifferential(msg)
	case "configure":
		return e.handleConfigure(msg)
	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *ADS1015Executor) handleReadChannel(channel int) (node.Message, error) {
	raw, voltage, err := e.readChannel(channel)
	if err != nil {
		return node.Message{}, err
	}

	percentage := (voltage / e.config.VRef) * 100
	if percentage > 100 {
		percentage = 100
	}
	if percentage < 0 {
		percentage = 0
	}

	return node.Message{
		Payload: map[string]interface{}{
			"channel":    channel,
			"raw":        raw,
			"voltage":    voltage,
			"percentage": percentage,
			"gain":       e.config.Gain,
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

func (e *ADS1015Executor) handleReadAll() (node.Message, error) {
	channels := make([]map[string]interface{}, 4)

	for ch := 0; ch < 4; ch++ {
		raw, voltage, err := e.readChannel(ch)
		if err != nil {
			channels[ch] = map[string]interface{}{
				"channel": ch,
				"error":   err.Error(),
			}
			continue
		}

		percentage := (voltage / e.config.VRef) * 100
		if percentage > 100 {
			percentage = 100
		}
		if percentage < 0 {
			percentage = 0
		}

		channels[ch] = map[string]interface{}{
			"channel":    ch,
			"raw":        raw,
			"voltage":    voltage,
			"percentage": percentage,
		}
	}

	return node.Message{
		Payload: map[string]interface{}{
			"channels":  channels,
			"gain":      e.config.Gain,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

func (e *ADS1015Executor) handleReadDifferential(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("payload is nil")
	}

	posChannel := 0
	negChannel := 1

	if pos, ok := payload["positive"].(float64); ok {
		posChannel = int(pos)
	}
	if neg, ok := payload["negative"].(float64); ok {
		negChannel = int(neg)
	}

	raw, voltage, err := e.readDifferential(posChannel, negChannel)
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"positive":  posChannel,
			"negative":  negChannel,
			"raw":       raw,
			"voltage":   voltage,
			"gain":      e.config.Gain,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

func (e *ADS1015Executor) handleConfigure(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("payload is nil")
	}

	if gain, ok := payload["gain"].(string); ok {
		e.config.Gain = gain
		switch gain {
		case "6.144":
			e.gainConfig = ads1015ConfigPGA6144
			e.gainVoltage = 6.144
		case "4.096":
			e.gainConfig = ads1015ConfigPGA4096
			e.gainVoltage = 4.096
		case "2.048":
			e.gainConfig = ads1015ConfigPGA2048
			e.gainVoltage = 2.048
		case "1.024":
			e.gainConfig = ads1015ConfigPGA1024
			e.gainVoltage = 1.024
		case "0.512":
			e.gainConfig = ads1015ConfigPGA512
			e.gainVoltage = 0.512
		case "0.256":
			e.gainConfig = ads1015ConfigPGA256
			e.gainVoltage = 0.256
		}
	}

	if dr, ok := payload["data_rate"].(float64); ok {
		e.config.DataRate = int(dr)
		switch int(dr) {
		case 128:
			e.drConfig = ads1015ConfigDR128
		case 250:
			e.drConfig = ads1015ConfigDR250
		case 490:
			e.drConfig = ads1015ConfigDR490
		case 920:
			e.drConfig = ads1015ConfigDR920
		case 1600:
			e.drConfig = ads1015ConfigDR1600
		case 2400:
			e.drConfig = ads1015ConfigDR2400
		case 3300:
			e.drConfig = ads1015ConfigDR3300
		}
	}

	if vref, ok := payload["vref"].(float64); ok {
		e.config.VRef = vref
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":    "configured",
			"gain":      e.config.Gain,
			"data_rate": e.config.DataRate,
			"vref":      e.config.VRef,
		},
	}, nil
}

func (e *ADS1015Executor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized && e.bus != nil {
		e.bus.Close()
		e.initialized = false
	}
	return nil
}
