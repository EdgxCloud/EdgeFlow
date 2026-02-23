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

// PCF8523 I2C registers
const (
	pcf8523DefaultAddr = 0x68

	pcf8523RegControl1    = 0x00
	pcf8523RegControl2    = 0x01
	pcf8523RegControl3    = 0x02
	pcf8523RegSeconds     = 0x03
	pcf8523RegMinutes     = 0x04
	pcf8523RegHours       = 0x05
	pcf8523RegDays        = 0x06
	pcf8523RegWeekdays    = 0x07
	pcf8523RegMonths      = 0x08
	pcf8523RegYears       = 0x09
	pcf8523RegMinuteAlarm = 0x0A
	pcf8523RegHourAlarm   = 0x0B
	pcf8523RegDayAlarm    = 0x0C
	pcf8523RegWeekdayAlarm = 0x0D
	pcf8523RegOffset      = 0x0E
	pcf8523RegTmrClkout   = 0x0F
	pcf8523RegTmrAFreq    = 0x10
	pcf8523RegTmrACnt     = 0x11
	pcf8523RegTmrBFreq    = 0x12
	pcf8523RegTmrBCnt     = 0x13
)

// PCF8523Config holds configuration for PCF8523 RTC
type PCF8523Config struct {
	I2CBus       string `json:"i2c_bus"`
	Address      uint16 `json:"address"`
	Use24Hour    bool   `json:"use_24_hour"`
	BatteryLow   bool   `json:"battery_low_int"`
}

// PCF8523Executor implements low-power RTC
type PCF8523Executor struct {
	config      PCF8523Config
	bus         i2c.BusCloser
	dev         i2c.Dev
	mu          sync.Mutex
	hostInited  bool
	initialized bool
}

func (e *PCF8523Executor) Init(config map[string]interface{}) error {
	e.config = PCF8523Config{
		I2CBus:     "/dev/i2c-1",
		Address:    pcf8523DefaultAddr,
		Use24Hour:  true,
		BatteryLow: false,
	}

	if config != nil {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := json.Unmarshal(configJSON, &e.config); err != nil {
			return fmt.Errorf("failed to parse PCF8523 config: %w", err)
		}
	}

	return nil
}

func (e *PCF8523Executor) initHardware() error {
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

	// Initialize RTC
	if err := e.initRTC(); err != nil {
		e.bus.Close()
		return fmt.Errorf("failed to initialize RTC: %w", err)
	}

	e.initialized = true
	return nil
}

func (e *PCF8523Executor) initRTC() error {
	// Read control 3 to check battery status
	ctrl3, err := e.readRegister(pcf8523RegControl3)
	if err != nil {
		return err
	}

	// Configure control 3 for battery switchover
	// Bits 7-5: PM = 000 (battery switchover standard mode)
	// Bit 2: BSF (battery switchover flag)
	// Bit 0: BLF (battery low flag)
	ctrl3 = (ctrl3 & 0x1F) | 0x00 // Standard battery switchover
	if err := e.writeRegister(pcf8523RegControl3, ctrl3); err != nil {
		return err
	}

	return nil
}

func (e *PCF8523Executor) readRegister(reg byte) (byte, error) {
	write := []byte{reg}
	read := make([]byte, 1)
	if err := e.dev.Tx(write, read); err != nil {
		return 0, err
	}
	return read[0], nil
}

func (e *PCF8523Executor) readRegisters(reg byte, length int) ([]byte, error) {
	write := []byte{reg}
	read := make([]byte, length)
	if err := e.dev.Tx(write, read); err != nil {
		return nil, err
	}
	return read, nil
}

func (e *PCF8523Executor) writeRegister(reg, value byte) error {
	return e.dev.Tx([]byte{reg, value}, nil)
}

func (e *PCF8523Executor) writeRegisters(reg byte, values []byte) error {
	cmd := append([]byte{reg}, values...)
	return e.dev.Tx(cmd, nil)
}

func (e *PCF8523Executor) bcdToDecimal(bcd byte) int {
	return int((bcd>>4)*10 + (bcd & 0x0F))
}

func (e *PCF8523Executor) decimalToBCD(dec int) byte {
	return byte((dec/10)<<4 | (dec % 10))
}

func (e *PCF8523Executor) readDateTime() (time.Time, error) {
	data, err := e.readRegisters(pcf8523RegSeconds, 7)
	if err != nil {
		return time.Time{}, err
	}

	// Check oscillator stopped flag
	if data[0]&0x80 != 0 {
		return time.Time{}, fmt.Errorf("oscillator stopped, time invalid")
	}

	seconds := e.bcdToDecimal(data[0] & 0x7F)
	minutes := e.bcdToDecimal(data[1] & 0x7F)
	hours := e.bcdToDecimal(data[2] & 0x3F)
	days := e.bcdToDecimal(data[3] & 0x3F)
	months := e.bcdToDecimal(data[5] & 0x1F)
	years := e.bcdToDecimal(data[6]) + 2000

	return time.Date(years, time.Month(months), days, hours, minutes, seconds, 0, time.UTC), nil
}

func (e *PCF8523Executor) writeDateTime(t time.Time) error {
	data := []byte{
		e.decimalToBCD(t.Second()),
		e.decimalToBCD(t.Minute()),
		e.decimalToBCD(t.Hour()),
		e.decimalToBCD(t.Day()),
		byte(t.Weekday()),
		e.decimalToBCD(int(t.Month())),
		e.decimalToBCD(t.Year() - 2000),
	}

	return e.writeRegisters(pcf8523RegSeconds, data)
}

func (e *PCF8523Executor) getBatteryStatus() (bool, bool, error) {
	ctrl3, err := e.readRegister(pcf8523RegControl3)
	if err != nil {
		return false, false, err
	}

	switchoverOccurred := (ctrl3 & 0x04) != 0 // BSF flag
	batteryLow := (ctrl3 & 0x01) != 0         // BLF flag

	return switchoverOccurred, batteryLow, nil
}

func (e *PCF8523Executor) clearBatteryFlags() error {
	ctrl3, err := e.readRegister(pcf8523RegControl3)
	if err != nil {
		return err
	}

	// Clear BSF and BLF flags
	ctrl3 &= ^byte(0x05)
	return e.writeRegister(pcf8523RegControl3, ctrl3)
}

func (e *PCF8523Executor) setAlarm(minute, hour, day int, enabled bool) error {
	minuteAlarm := byte(0x80) // Disabled by default
	hourAlarm := byte(0x80)
	dayAlarm := byte(0x80)

	if enabled {
		if minute >= 0 && minute < 60 {
			minuteAlarm = e.decimalToBCD(minute) // Enabled when MSB is 0
		}
		if hour >= 0 && hour < 24 {
			hourAlarm = e.decimalToBCD(hour)
		}
		if day >= 1 && day <= 31 {
			dayAlarm = e.decimalToBCD(day)
		}
	}

	if err := e.writeRegister(pcf8523RegMinuteAlarm, minuteAlarm); err != nil {
		return err
	}
	if err := e.writeRegister(pcf8523RegHourAlarm, hourAlarm); err != nil {
		return err
	}
	if err := e.writeRegister(pcf8523RegDayAlarm, dayAlarm); err != nil {
		return err
	}

	// Enable/disable alarm interrupt
	ctrl1, err := e.readRegister(pcf8523RegControl1)
	if err != nil {
		return err
	}

	if enabled {
		ctrl1 |= 0x02 // AIE = 1
	} else {
		ctrl1 &= ^byte(0x02) // AIE = 0
	}

	return e.writeRegister(pcf8523RegControl1, ctrl1)
}

func (e *PCF8523Executor) setClockOutput(frequency string) error {
	var freqVal byte

	switch frequency {
	case "32768hz":
		freqVal = 0x00
	case "16384hz":
		freqVal = 0x01
	case "8192hz":
		freqVal = 0x02
	case "4096hz":
		freqVal = 0x03
	case "1024hz":
		freqVal = 0x04
	case "32hz":
		freqVal = 0x05
	case "1hz":
		freqVal = 0x06
	case "off":
		freqVal = 0x07
	default:
		freqVal = 0x07 // Off
	}

	tmrClkout, err := e.readRegister(pcf8523RegTmrClkout)
	if err != nil {
		return err
	}

	tmrClkout = (tmrClkout & 0xF8) | freqVal
	return e.writeRegister(pcf8523RegTmrClkout, tmrClkout)
}

func (e *PCF8523Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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
		return e.handleRead()
	case "set":
		return e.handleSet(msg)
	case "sync":
		return e.handleSync()
	case "battery_status":
		return e.handleBatteryStatus()
	case "set_alarm":
		return e.handleSetAlarm(msg)
	case "disable_alarm":
		return e.handleDisableAlarm()
	case "set_clockout":
		return e.handleSetClockout(msg)
	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *PCF8523Executor) handleRead() (node.Message, error) {
	dt, err := e.readDateTime()
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read datetime: %w", err)
	}

	switchover, batteryLow, _ := e.getBatteryStatus()

	return node.Message{
		Payload: map[string]interface{}{
			"datetime":            dt.Format(time.RFC3339),
			"year":                dt.Year(),
			"month":               int(dt.Month()),
			"day":                 dt.Day(),
			"weekday":             dt.Weekday().String(),
			"hour":                dt.Hour(),
			"minute":              dt.Minute(),
			"second":              dt.Second(),
			"unix":                dt.Unix(),
			"battery_switchover":  switchover,
			"battery_low":         batteryLow,
		},
	}, nil
}

func (e *PCF8523Executor) handleSet(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	var setTime time.Time

	if dtStr, ok := payload["datetime"].(string); ok {
		var err error
		setTime, err = time.Parse(time.RFC3339, dtStr)
		if err != nil {
			return node.Message{}, fmt.Errorf("invalid datetime format: %w", err)
		}
	} else if unix, ok := payload["unix"].(float64); ok {
		setTime = time.Unix(int64(unix), 0).UTC()
	} else {
		return node.Message{}, fmt.Errorf("datetime or unix timestamp required")
	}

	if err := e.writeDateTime(setTime); err != nil {
		return node.Message{}, fmt.Errorf("failed to set datetime: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":   "time_set",
			"datetime": setTime.Format(time.RFC3339),
		},
	}, nil
}

func (e *PCF8523Executor) handleSync() (node.Message, error) {
	now := time.Now().UTC()

	if err := e.writeDateTime(now); err != nil {
		return node.Message{}, fmt.Errorf("failed to sync datetime: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":   "synced",
			"datetime": now.Format(time.RFC3339),
		},
	}, nil
}

func (e *PCF8523Executor) handleBatteryStatus() (node.Message, error) {
	switchover, batteryLow, err := e.getBatteryStatus()
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read battery status: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"battery_switchover": switchover,
			"battery_low":        batteryLow,
		},
	}, nil
}

func (e *PCF8523Executor) handleSetAlarm(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	minute := -1
	hour := -1
	day := -1

	if m, ok := payload["minute"].(float64); ok {
		minute = int(m)
	}
	if h, ok := payload["hour"].(float64); ok {
		hour = int(h)
	}
	if d, ok := payload["day"].(float64); ok {
		day = int(d)
	}

	if err := e.setAlarm(minute, hour, day, true); err != nil {
		return node.Message{}, fmt.Errorf("failed to set alarm: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status": "alarm_set",
			"minute": minute,
			"hour":   hour,
			"day":    day,
		},
	}, nil
}

func (e *PCF8523Executor) handleDisableAlarm() (node.Message, error) {
	if err := e.setAlarm(0, 0, 0, false); err != nil {
		return node.Message{}, fmt.Errorf("failed to disable alarm: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status": "alarm_disabled",
		},
	}, nil
}

func (e *PCF8523Executor) handleSetClockout(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	frequency := "off"
	if f, ok := payload["frequency"].(string); ok {
		frequency = f
	}

	if err := e.setClockOutput(frequency); err != nil {
		return node.Message{}, fmt.Errorf("failed to set clock output: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":    "clockout_set",
			"frequency": frequency,
		},
	}, nil
}

func (e *PCF8523Executor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized && e.bus != nil {
		e.bus.Close()
		e.initialized = false
	}
	return nil
}
