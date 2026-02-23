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

// DS1307 Register addresses
const (
	ds1307RegSeconds = 0x00
	ds1307RegMinutes = 0x01
	ds1307RegHours   = 0x02
	ds1307RegDay     = 0x03 // Day of week (1-7)
	ds1307RegDate    = 0x04 // Day of month (1-31)
	ds1307RegMonth   = 0x05 // Month (1-12)
	ds1307RegYear    = 0x06 // Year (0-99)
	ds1307RegControl = 0x07
	ds1307RegRAM     = 0x08 // RAM starts at 0x08 (56 bytes until 0x3F)
)

// DS1307 Control register bits
const (
	ds1307CtrlOUT  = 0x80 // Output control
	ds1307CtrlSQWE = 0x10 // Square wave enable
	ds1307CtrlRS1  = 0x02 // Rate select bit 1
	ds1307CtrlRS0  = 0x01 // Rate select bit 0
)

// DS1307Config configuration for DS1307 RTC
type DS1307Config struct {
	Bus     string `json:"bus"`     // I2C bus (default: "")
	Address int    `json:"address"` // I2C address (default: 0x68)
}

// DS1307Executor executes DS1307 RTC operations
type DS1307Executor struct {
	config     DS1307Config
	bus        i2c.BusCloser
	dev        i2c.Dev
	mu         sync.Mutex
	hostInited bool
}

// NewDS1307Executor creates a new DS1307 executor
func NewDS1307Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var rtcConfig DS1307Config
	if err := json.Unmarshal(configJSON, &rtcConfig); err != nil {
		return nil, fmt.Errorf("invalid DS1307 config: %w", err)
	}

	// Defaults
	if rtcConfig.Address == 0 {
		rtcConfig.Address = 0x68 // Default DS1307 address
	}

	return &DS1307Executor{
		config: rtcConfig,
	}, nil
}

// Init initializes the DS1307 executor
func (e *DS1307Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles RTC operations
func (e *DS1307Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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
		return e.readTime()
	}
	payload := msg.Payload

	action, _ := payload["action"].(string)

	switch action {
	case "read", "get", "":
		return e.readTime()

	case "set":
		if _, ok := payload["year"]; ok {
			year := int(getFloat(payload, "year", 2024))
			month := int(getFloat(payload, "month", 1))
			day := int(getFloat(payload, "day", 1))
			hour := int(getFloat(payload, "hour", 0))
			minute := int(getFloat(payload, "minute", 0))
			second := int(getFloat(payload, "second", 0))
			return e.setTime(year, month, day, hour, minute, second)
		}
		now := time.Now()
		return e.setTime(now.Year(), int(now.Month()), now.Day(),
			now.Hour(), now.Minute(), now.Second())

	case "sync":
		now := time.Now()
		return e.setTime(now.Year(), int(now.Month()), now.Day(),
			now.Hour(), now.Minute(), now.Second())

	case "start":
		// Enable oscillator
		return e.setOscillator(true)

	case "stop":
		// Disable oscillator (to save battery)
		return e.setOscillator(false)

	case "sqw":
		// Configure square wave output
		enable, _ := payload["enable"].(bool)
		freq := int(getFloat(payload, "frequency", 1))
		return e.setSquareWave(enable, freq)

	case "ram_read":
		// Read from battery-backed RAM
		addr := int(getFloat(payload, "address", 0))
		length := int(getFloat(payload, "length", 1))
		return e.readRAM(addr, length)

	case "ram_write":
		// Write to battery-backed RAM
		addr := int(getFloat(payload, "address", 0))
		var data []byte
		if dataIface, ok := payload["data"].([]interface{}); ok {
			for _, v := range dataIface {
				if b, ok := v.(float64); ok {
					data = append(data, byte(b))
				}
			}
		}
		return e.writeRAM(addr, data)

	case "status":
		return e.getStatus()

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// readTime reads the current time from the RTC
func (e *DS1307Executor) readTime() (node.Message, error) {
	// Read all time registers (7 bytes starting from 0x00)
	data := make([]byte, 7)
	if err := e.dev.Tx([]byte{ds1307RegSeconds}, data); err != nil {
		return node.Message{}, fmt.Errorf("failed to read time: %w", err)
	}

	// Check if oscillator is running (CH bit in seconds register)
	oscillatorRunning := data[0]&0x80 == 0

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
	year := bcdToDec(data[6]) + 2000 // DS1307 uses 00-99 for 2000-2099

	// Day of week names
	dayNames := []string{"", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	dayName := ""
	if dayOfWeek >= 1 && dayOfWeek <= 7 {
		dayName = dayNames[dayOfWeek]
	}

	// Create time object
	rtcTime := time.Date(year, time.Month(month), date, hour, minute, second, 0, time.Local)

	return node.Message{
		Payload: map[string]interface{}{
			"year":               year,
			"month":              month,
			"day":                date,
			"hour":               hour,
			"minute":             minute,
			"second":             second,
			"day_of_week":        dayOfWeek,
			"day_name":           dayName,
			"oscillator_running": oscillatorRunning,
			"iso8601":            rtcTime.Format(time.RFC3339),
			"unix":               rtcTime.Unix(),
			"formatted":          rtcTime.Format("2006-01-02 15:04:05"),
			"address":            fmt.Sprintf("0x%02X", e.config.Address),
			"sensor":             "DS1307",
			"timestamp":          time.Now().Unix(),
		},
	}, nil
}

// setTime sets the RTC time
func (e *DS1307Executor) setTime(year, month, day, hour, minute, second int) (node.Message, error) {
	// Validate inputs
	if year < 2000 || year > 2099 {
		return node.Message{}, fmt.Errorf("year must be 2000-2099")
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
		dayOfWeek = 7 // Sunday = 7
	}

	// Prepare data (24-hour mode, CH bit = 0 to enable oscillator)
	data := []byte{
		ds1307RegSeconds,
		decToBcd(second),      // Seconds (CH = 0)
		decToBcd(minute),      // Minutes
		decToBcd(hour),        // Hours (24-hour mode)
		decToBcd(dayOfWeek),   // Day of week
		decToBcd(day),         // Date
		decToBcd(month),       // Month
		decToBcd(year - 2000), // Year (0-99)
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
			"sensor":    "DS1307",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// setOscillator enables or disables the oscillator
func (e *DS1307Executor) setOscillator(enable bool) (node.Message, error) {
	// Read current seconds register
	data := make([]byte, 1)
	if err := e.dev.Tx([]byte{ds1307RegSeconds}, data); err != nil {
		return node.Message{}, err
	}

	// Modify CH bit
	if enable {
		data[0] &= 0x7F // Clear CH bit to enable oscillator
	} else {
		data[0] |= 0x80 // Set CH bit to disable oscillator
	}

	// Write back
	if _, err := e.dev.Write([]byte{ds1307RegSeconds, data[0]}); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":             "oscillator",
			"oscillator_enabled": enable,
			"sensor":             "DS1307",
			"timestamp":          time.Now().Unix(),
		},
	}, nil
}

// setSquareWave configures the square wave output
func (e *DS1307Executor) setSquareWave(enable bool, frequency int) (node.Message, error) {
	var ctrl byte

	if enable {
		ctrl = ds1307CtrlSQWE

		// Set frequency
		switch frequency {
		case 1: // 1 Hz
			// RS1=0, RS0=0
		case 4096: // 4.096 kHz
			ctrl |= ds1307CtrlRS0
		case 8192: // 8.192 kHz
			ctrl |= ds1307CtrlRS1
		case 32768: // 32.768 kHz
			ctrl |= ds1307CtrlRS1 | ds1307CtrlRS0
		default:
			frequency = 1
		}
	}

	if _, err := e.dev.Write([]byte{ds1307RegControl, ctrl}); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":           "sqw",
			"sqw_enabled":      enable,
			"frequency":        frequency,
			"sensor":           "DS1307",
			"timestamp":        time.Now().Unix(),
		},
	}, nil
}

// readRAM reads from the battery-backed RAM (56 bytes, 0x08-0x3F)
func (e *DS1307Executor) readRAM(addr, length int) (node.Message, error) {
	if addr < 0 || addr > 55 {
		return node.Message{}, fmt.Errorf("RAM address must be 0-55")
	}
	if length < 1 || addr+length > 56 {
		return node.Message{}, fmt.Errorf("invalid length")
	}

	data := make([]byte, length)
	if err := e.dev.Tx([]byte{byte(ds1307RegRAM + addr)}, data); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "ram_read",
			"address":   addr,
			"length":    length,
			"data":      data,
			"sensor":    "DS1307",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// writeRAM writes to the battery-backed RAM
func (e *DS1307Executor) writeRAM(addr int, data []byte) (node.Message, error) {
	if addr < 0 || addr > 55 {
		return node.Message{}, fmt.Errorf("RAM address must be 0-55")
	}
	if len(data) == 0 || addr+len(data) > 56 {
		return node.Message{}, fmt.Errorf("invalid data length")
	}

	writeData := make([]byte, len(data)+1)
	writeData[0] = byte(ds1307RegRAM + addr)
	copy(writeData[1:], data)

	if _, err := e.dev.Write(writeData); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "ram_write",
			"address":   addr,
			"length":    len(data),
			"sensor":    "DS1307",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// getStatus returns the RTC status
func (e *DS1307Executor) getStatus() (node.Message, error) {
	// Read seconds register
	secData := make([]byte, 1)
	if err := e.dev.Tx([]byte{ds1307RegSeconds}, secData); err != nil {
		return node.Message{}, err
	}

	// Read control register
	ctrlData := make([]byte, 1)
	if err := e.dev.Tx([]byte{ds1307RegControl}, ctrlData); err != nil {
		return node.Message{}, err
	}

	oscillatorRunning := secData[0]&0x80 == 0
	sqwEnabled := ctrlData[0]&ds1307CtrlSQWE != 0

	var sqwFreq int
	switch ctrlData[0] & (ds1307CtrlRS1 | ds1307CtrlRS0) {
	case 0:
		sqwFreq = 1
	case ds1307CtrlRS0:
		sqwFreq = 4096
	case ds1307CtrlRS1:
		sqwFreq = 8192
	default:
		sqwFreq = 32768
	}

	return node.Message{
		Payload: map[string]interface{}{
			"oscillator_running": oscillatorRunning,
			"sqw_enabled":        sqwEnabled,
			"sqw_frequency":      sqwFreq,
			"ram_size":           56,
			"address":            fmt.Sprintf("0x%02X", e.config.Address),
			"sensor":             "DS1307",
			"timestamp":          time.Now().Unix(),
		},
	}, nil
}

// Cleanup releases resources
func (e *DS1307Executor) Cleanup() error {
	if e.bus != nil {
		e.bus.Close()
		e.bus = nil
	}
	return nil
}
