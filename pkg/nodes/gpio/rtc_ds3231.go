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

// DS3231 Register addresses
const (
	ds3231RegSeconds     = 0x00
	ds3231RegMinutes     = 0x01
	ds3231RegHours       = 0x02
	ds3231RegDay         = 0x03 // Day of week (1-7)
	ds3231RegDate        = 0x04 // Day of month (1-31)
	ds3231RegMonth       = 0x05 // Month (1-12) + Century bit
	ds3231RegYear        = 0x06 // Year (0-99)
	ds3231RegAlarm1Sec   = 0x07
	ds3231RegAlarm1Min   = 0x08
	ds3231RegAlarm1Hour  = 0x09
	ds3231RegAlarm1Day   = 0x0A
	ds3231RegAlarm2Min   = 0x0B
	ds3231RegAlarm2Hour  = 0x0C
	ds3231RegAlarm2Day   = 0x0D
	ds3231RegControl     = 0x0E
	ds3231RegStatus      = 0x0F
	ds3231RegAgingOffset = 0x10
	ds3231RegTempMSB     = 0x11
	ds3231RegTempLSB     = 0x12
)

// DS3231 Control register bits
const (
	ds3231CtrlEOSC   = 0x80 // Enable oscillator (active low)
	ds3231CtrlBBSQW  = 0x40 // Battery-backed square wave enable
	ds3231CtrlCONV   = 0x20 // Convert temperature
	ds3231CtrlRS2    = 0x10 // Rate select 2
	ds3231CtrlRS1    = 0x08 // Rate select 1
	ds3231CtrlINTCN  = 0x04 // Interrupt control
	ds3231CtrlA2IE   = 0x02 // Alarm 2 interrupt enable
	ds3231CtrlA1IE   = 0x01 // Alarm 1 interrupt enable
)

// DS3231 Status register bits
const (
	ds3231StatOSF  = 0x80 // Oscillator stop flag
	ds3231StatEN32 = 0x08 // Enable 32kHz output
	ds3231StatBSY  = 0x04 // Busy (temp conversion)
	ds3231StatA2F  = 0x02 // Alarm 2 flag
	ds3231StatA1F  = 0x01 // Alarm 1 flag
)

// DS3231Config configuration for DS3231 RTC
type DS3231Config struct {
	Bus     string `json:"bus"`     // I2C bus (default: "")
	Address int    `json:"address"` // I2C address (default: 0x68)
}

// DS3231Executor executes DS3231 RTC operations
type DS3231Executor struct {
	config     DS3231Config
	bus        i2c.BusCloser
	dev        i2c.Dev
	mu         sync.Mutex
	hostInited bool
}

// NewDS3231Executor creates a new DS3231 executor
func NewDS3231Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var rtcConfig DS3231Config
	if err := json.Unmarshal(configJSON, &rtcConfig); err != nil {
		return nil, fmt.Errorf("invalid DS3231 config: %w", err)
	}

	// Defaults
	if rtcConfig.Address == 0 {
		rtcConfig.Address = 0x68 // Default DS3231 address
	}

	return &DS3231Executor{
		config: rtcConfig,
	}, nil
}

// Init initializes the DS3231 executor
func (e *DS3231Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles RTC operations
func (e *DS3231Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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

	// Parse command
	if msg.Payload == nil {
		// Default: read time
		return e.readTime()
	}
	payload := msg.Payload

	action, _ := payload["action"].(string)

	switch action {
	case "read", "get", "":
		return e.readTime()

	case "set":
		// Set time from payload or use current system time
		if _, ok := payload["year"]; ok {
			year := int(getFloat(payload, "year", 2024))
			month := int(getFloat(payload, "month", 1))
			day := int(getFloat(payload, "day", 1))
			hour := int(getFloat(payload, "hour", 0))
			minute := int(getFloat(payload, "minute", 0))
			second := int(getFloat(payload, "second", 0))
			return e.setTime(year, month, day, hour, minute, second)
		}
		// Use system time
		now := time.Now()
		return e.setTime(now.Year(), int(now.Month()), now.Day(),
			now.Hour(), now.Minute(), now.Second())

	case "sync":
		// Sync RTC to system time
		now := time.Now()
		return e.setTime(now.Year(), int(now.Month()), now.Day(),
			now.Hour(), now.Minute(), now.Second())

	case "temperature":
		return e.readTemperature()

	case "alarm1_set":
		hour := int(getFloat(payload, "hour", 0))
		minute := int(getFloat(payload, "minute", 0))
		second := int(getFloat(payload, "second", 0))
		mode := payload["mode"].(string)
		return e.setAlarm1(hour, minute, second, mode)

	case "alarm2_set":
		hour := int(getFloat(payload, "hour", 0))
		minute := int(getFloat(payload, "minute", 0))
		mode, _ := payload["mode"].(string)
		return e.setAlarm2(hour, minute, mode)

	case "alarm_clear":
		return e.clearAlarms()

	case "alarm_status":
		return e.getAlarmStatus()

	case "status":
		return e.getStatus()

	case "aging_offset":
		// Set aging offset for calibration (-128 to +127)
		offset := int(getFloat(payload, "offset", 0))
		return e.setAgingOffset(int8(offset))

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// readTime reads the current time from the RTC
func (e *DS3231Executor) readTime() (node.Message, error) {
	// Read all time registers (7 bytes starting from 0x00)
	data := make([]byte, 7)
	if err := e.dev.Tx([]byte{ds3231RegSeconds}, data); err != nil {
		return node.Message{}, fmt.Errorf("failed to read time: %w", err)
	}

	// Parse BCD values
	second := bcdToDec(data[0] & 0x7F)
	minute := bcdToDec(data[1] & 0x7F)

	// Handle 12/24 hour mode
	hour := 0
	is12Hour := data[2]&0x40 != 0
	isPM := false
	if is12Hour {
		hour = bcdToDec(data[2] & 0x1F)
		isPM = data[2]&0x20 != 0
		if isPM && hour != 12 {
			hour += 12
		} else if !isPM && hour == 12 {
			hour = 0
		}
	} else {
		hour = bcdToDec(data[2] & 0x3F)
	}

	dayOfWeek := bcdToDec(data[3] & 0x07)
	date := bcdToDec(data[4] & 0x3F)
	month := bcdToDec(data[5] & 0x1F)
	century := data[5]&0x80 != 0
	year := bcdToDec(data[6])

	// Full year calculation
	fullYear := 2000 + year
	if century {
		fullYear += 100
	}

	// Day of week names
	dayNames := []string{"", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	dayName := ""
	if dayOfWeek >= 1 && dayOfWeek <= 7 {
		dayName = dayNames[dayOfWeek]
	}

	// Create time object
	rtcTime := time.Date(fullYear, time.Month(month), date, hour, minute, second, 0, time.Local)

	return node.Message{
		Payload: map[string]interface{}{
			"year":         fullYear,
			"month":        month,
			"day":          date,
			"hour":         hour,
			"minute":       minute,
			"second":       second,
			"day_of_week":  dayOfWeek,
			"day_name":     dayName,
			"iso8601":      rtcTime.Format(time.RFC3339),
			"unix":         rtcTime.Unix(),
			"formatted":    rtcTime.Format("2006-01-02 15:04:05"),
			"address":      fmt.Sprintf("0x%02X", e.config.Address),
			"sensor":       "DS3231",
			"timestamp":    time.Now().Unix(),
		},
	}, nil
}

// setTime sets the RTC time
func (e *DS3231Executor) setTime(year, month, day, hour, minute, second int) (node.Message, error) {
	// Validate inputs
	if year < 2000 || year > 2199 {
		return node.Message{}, fmt.Errorf("year must be 2000-2199")
	}
	if month < 1 || month > 12 {
		return node.Message{}, fmt.Errorf("month must be 1-12")
	}
	if day < 1 || day > 31 {
		return node.Message{}, fmt.Errorf("day must be 1-31")
	}
	if hour < 0 || hour > 23 {
		return node.Message{}, fmt.Errorf("hour must be 0-23")
	}
	if minute < 0 || minute > 59 {
		return node.Message{}, fmt.Errorf("minute must be 0-59")
	}
	if second < 0 || second > 59 {
		return node.Message{}, fmt.Errorf("second must be 0-59")
	}

	// Calculate day of week
	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
	dayOfWeek := int(t.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7 // Sunday = 7 in DS3231
	}

	// Prepare data
	y := year - 2000
	century := byte(0)
	if y >= 100 {
		century = 0x80
		y -= 100
	}

	data := []byte{
		ds3231RegSeconds,
		decToBcd(second),
		decToBcd(minute),
		decToBcd(hour), // 24-hour mode
		decToBcd(dayOfWeek),
		decToBcd(day),
		decToBcd(month) | century,
		decToBcd(y),
	}

	if _, err := e.dev.Write(data); err != nil {
		return node.Message{}, fmt.Errorf("failed to set time: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "set",
			"year":      year,
			"month":     month,
			"day":       day,
			"hour":      hour,
			"minute":    minute,
			"second":    second,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// readTemperature reads the temperature from DS3231's internal sensor
func (e *DS3231Executor) readTemperature() (node.Message, error) {
	// Trigger temperature conversion
	ctrl, err := e.readReg(ds3231RegControl)
	if err != nil {
		return node.Message{}, err
	}
	if err := e.writeReg(ds3231RegControl, ctrl|ds3231CtrlCONV); err != nil {
		return node.Message{}, err
	}

	// Wait for conversion (max 200ms)
	time.Sleep(200 * time.Millisecond)

	// Read temperature registers
	data := make([]byte, 2)
	if err := e.dev.Tx([]byte{ds3231RegTempMSB}, data); err != nil {
		return node.Message{}, fmt.Errorf("failed to read temperature: %w", err)
	}

	// Temperature is in 2's complement, 0.25°C resolution
	tempMSB := int8(data[0])
	tempLSB := data[1] >> 6 // Only upper 2 bits used

	tempC := float64(tempMSB) + float64(tempLSB)*0.25
	tempF := tempC*9/5 + 32

	return node.Message{
		Payload: map[string]interface{}{
			"temperature_c": tempC,
			"temperature_f": tempF,
			"address":       fmt.Sprintf("0x%02X", e.config.Address),
			"sensor":        "DS3231",
			"timestamp":     time.Now().Unix(),
		},
	}, nil
}

// setAlarm1 sets Alarm 1 (seconds precision)
func (e *DS3231Executor) setAlarm1(hour, minute, second int, mode string) (node.Message, error) {
	// Alarm mask bits: A1M1-A1M4
	// 0000 = Alarm when seconds, minutes, hours, and day/date match
	// 1000 = Alarm when minutes, hours, and day/date match
	// 1100 = Alarm when hours and day/date match
	// 1110 = Alarm when day/date matches
	// 1111 = Alarm once per second

	var a1m1, a1m2, a1m3, a1m4 byte

	switch mode {
	case "per_second":
		a1m1, a1m2, a1m3, a1m4 = 0x80, 0x80, 0x80, 0x80
	case "match_seconds":
		a1m1, a1m2, a1m3, a1m4 = 0x00, 0x80, 0x80, 0x80
	case "match_minutes":
		a1m1, a1m2, a1m3, a1m4 = 0x00, 0x00, 0x80, 0x80
	case "match_hours":
		a1m1, a1m2, a1m3, a1m4 = 0x00, 0x00, 0x00, 0x80
	case "match_date":
		a1m1, a1m2, a1m3, a1m4 = 0x00, 0x00, 0x00, 0x00
	default:
		a1m1, a1m2, a1m3, a1m4 = 0x00, 0x00, 0x00, 0x80 // Default: match hours
	}

	data := []byte{
		ds3231RegAlarm1Sec,
		decToBcd(second) | a1m1,
		decToBcd(minute) | a1m2,
		decToBcd(hour) | a1m3,
		0x01 | a1m4, // Day = 1 (ignored in most modes)
	}

	if _, err := e.dev.Write(data); err != nil {
		return node.Message{}, fmt.Errorf("failed to set alarm 1: %w", err)
	}

	// Enable alarm 1 interrupt
	ctrl, _ := e.readReg(ds3231RegControl)
	if err := e.writeReg(ds3231RegControl, ctrl|ds3231CtrlA1IE|ds3231CtrlINTCN); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "alarm1_set",
			"hour":      hour,
			"minute":    minute,
			"second":    second,
			"mode":      mode,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// setAlarm2 sets Alarm 2 (minute precision)
func (e *DS3231Executor) setAlarm2(hour, minute int, mode string) (node.Message, error) {
	var a2m2, a2m3, a2m4 byte

	switch mode {
	case "per_minute":
		a2m2, a2m3, a2m4 = 0x80, 0x80, 0x80
	case "match_minutes":
		a2m2, a2m3, a2m4 = 0x00, 0x80, 0x80
	case "match_hours":
		a2m2, a2m3, a2m4 = 0x00, 0x00, 0x80
	default:
		a2m2, a2m3, a2m4 = 0x00, 0x00, 0x80
	}

	data := []byte{
		ds3231RegAlarm2Min,
		decToBcd(minute) | a2m2,
		decToBcd(hour) | a2m3,
		0x01 | a2m4,
	}

	if _, err := e.dev.Write(data); err != nil {
		return node.Message{}, fmt.Errorf("failed to set alarm 2: %w", err)
	}

	// Enable alarm 2 interrupt
	ctrl, _ := e.readReg(ds3231RegControl)
	if err := e.writeReg(ds3231RegControl, ctrl|ds3231CtrlA2IE|ds3231CtrlINTCN); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "alarm2_set",
			"hour":      hour,
			"minute":    minute,
			"mode":      mode,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// clearAlarms clears alarm flags
func (e *DS3231Executor) clearAlarms() (node.Message, error) {
	status, err := e.readReg(ds3231RegStatus)
	if err != nil {
		return node.Message{}, err
	}

	// Clear alarm flags
	status &^= (ds3231StatA1F | ds3231StatA2F)
	if err := e.writeReg(ds3231RegStatus, status); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "alarm_clear",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// getAlarmStatus returns alarm flags
func (e *DS3231Executor) getAlarmStatus() (node.Message, error) {
	status, err := e.readReg(ds3231RegStatus)
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"alarm1_triggered": status&ds3231StatA1F != 0,
			"alarm2_triggered": status&ds3231StatA2F != 0,
			"timestamp":        time.Now().Unix(),
		},
	}, nil
}

// getStatus returns RTC status
func (e *DS3231Executor) getStatus() (node.Message, error) {
	status, err := e.readReg(ds3231RegStatus)
	if err != nil {
		return node.Message{}, err
	}
	ctrl, err := e.readReg(ds3231RegControl)
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"oscillator_stopped": status&ds3231StatOSF != 0,
			"busy":               status&ds3231StatBSY != 0,
			"32khz_enabled":      status&ds3231StatEN32 != 0,
			"alarm1_enabled":     ctrl&ds3231CtrlA1IE != 0,
			"alarm2_enabled":     ctrl&ds3231CtrlA2IE != 0,
			"alarm1_triggered":   status&ds3231StatA1F != 0,
			"alarm2_triggered":   status&ds3231StatA2F != 0,
			"address":            fmt.Sprintf("0x%02X", e.config.Address),
			"sensor":             "DS3231",
			"timestamp":          time.Now().Unix(),
		},
	}, nil
}

// setAgingOffset sets the aging offset for calibration
func (e *DS3231Executor) setAgingOffset(offset int8) (node.Message, error) {
	if err := e.writeReg(ds3231RegAgingOffset, byte(offset)); err != nil {
		return node.Message{}, fmt.Errorf("failed to set aging offset: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "aging_offset",
			"offset":    offset,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// readReg reads a single register
func (e *DS3231Executor) readReg(reg byte) (byte, error) {
	data := make([]byte, 1)
	if err := e.dev.Tx([]byte{reg}, data); err != nil {
		return 0, err
	}
	return data[0], nil
}

// writeReg writes a single register
func (e *DS3231Executor) writeReg(reg, value byte) error {
	_, err := e.dev.Write([]byte{reg, value})
	return err
}

// bcdToDec converts BCD to decimal
func bcdToDec(bcd byte) int {
	return int(bcd>>4)*10 + int(bcd&0x0F)
}

// decToBcd converts decimal to BCD
func decToBcd(dec int) byte {
	return byte((dec/10)<<4 | (dec % 10))
}

// Cleanup releases resources
func (e *DS3231Executor) Cleanup() error {
	if e.bus != nil {
		e.bus.Close()
		e.bus = nil
	}
	return nil
}
