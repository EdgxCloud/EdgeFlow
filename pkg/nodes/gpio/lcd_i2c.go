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

// LCD I2C commands (HD44780 via PCF8574)
const (
	lcdBacklight   = 0x08
	lcdNoBacklight = 0x00
	lcdEnable      = 0x04
	lcdRw          = 0x02 // Read/Write (0=Write)
	lcdRs          = 0x01 // Register Select (0=Command, 1=Data)

	// Commands
	lcdClearDisplay   = 0x01
	lcdReturnHome     = 0x02
	lcdEntryModeSet   = 0x04
	lcdDisplayControl = 0x08
	lcdCursorShift    = 0x10
	lcdFunctionSet    = 0x20
	lcdSetCGRAMAddr   = 0x40
	lcdSetDDRAMAddr   = 0x80

	// Entry mode flags
	lcdEntryRight          = 0x00
	lcdEntryLeft           = 0x02
	lcdEntryShiftIncrement = 0x01
	lcdEntryShiftDecrement = 0x00

	// Display control flags
	lcdDisplayOn  = 0x04
	lcdDisplayOff = 0x00
	lcdCursorOn   = 0x02
	lcdCursorOff  = 0x00
	lcdBlinkOn    = 0x01
	lcdBlinkOff   = 0x00

	// Function set flags
	lcd8BitMode = 0x10
	lcd4BitMode = 0x00
	lcd2Line    = 0x08
	lcd1Line    = 0x00
	lcd5x10Dots = 0x04
	lcd5x8Dots  = 0x00
)

// LCDI2CConfig configuration for I2C LCD display
type LCDI2CConfig struct {
	Bus       string `json:"bus"`       // I2C bus (default: "")
	Address   int    `json:"address"`   // I2C address (default: 0x27)
	Cols      int    `json:"cols"`      // Number of columns (default: 16)
	Rows      int    `json:"rows"`      // Number of rows (default: 2)
	Backlight bool   `json:"backlight"` // Backlight on (default: true)
}

// LCDI2CExecutor executes LCD display operations
type LCDI2CExecutor struct {
	config       LCDI2CConfig
	bus          i2c.BusCloser
	dev          i2c.Dev
	backlight    byte
	displayCtrl  byte
	displayMode  byte
	mu           sync.Mutex
	hostInited   bool
	initialized  bool
}

// NewLCDI2CExecutor creates a new LCD I2C executor
func NewLCDI2CExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var lcdConfig LCDI2CConfig
	if err := json.Unmarshal(configJSON, &lcdConfig); err != nil {
		return nil, fmt.Errorf("invalid LCD config: %w", err)
	}

	// Defaults
	if lcdConfig.Address == 0 {
		lcdConfig.Address = 0x27 // Common address for PCF8574
	}
	if lcdConfig.Cols == 0 {
		lcdConfig.Cols = 16
	}
	if lcdConfig.Rows == 0 {
		lcdConfig.Rows = 2
	}

	// Default backlight on
	backlight := byte(lcdBacklight)
	if !lcdConfig.Backlight {
		backlight = lcdNoBacklight
	}

	return &LCDI2CExecutor{
		config:    lcdConfig,
		backlight: backlight,
	}, nil
}

// Init initializes the LCD executor
func (e *LCDI2CExecutor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles LCD operations
func (e *LCDI2CExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Initialize hardware if needed
	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init periph host: %w", err)
		}
		e.hostInited = true
	}

	// Open I2C bus if not already open
	if e.bus == nil {
		bus, err := i2creg.Open(e.config.Bus)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to open I2C bus: %w", err)
		}
		e.bus = bus
		e.dev = i2c.Dev{Bus: e.bus, Addr: uint16(e.config.Address)}
	}

	// Initialize LCD if not done
	if !e.initialized {
		if err := e.initLCD(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init LCD: %w", err)
		}
		e.initialized = true
	}

	// Parse command from message
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	// If payload has a "text" field as string, we can also write it directly
	if text, ok := payload["text"].(string); ok && payload["action"] == nil {
		if err := e.writeString(text); err != nil {
			return node.Message{}, err
		}
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "write",
				"text":      text,
				"timestamp": time.Now().Unix(),
			},
		}, nil
	}

	// Handle different commands
	action, _ := payload["action"].(string)

	switch action {
	case "clear":
		if err := e.clear(); err != nil {
			return node.Message{}, err
		}

	case "home":
		if err := e.home(); err != nil {
			return node.Message{}, err
		}

	case "write":
		text, _ := payload["text"].(string)
		row := int(getFloat(payload, "row", 0))
		col := int(getFloat(payload, "col", 0))

		if row > 0 || col > 0 {
			if err := e.setCursor(row, col); err != nil {
				return node.Message{}, err
			}
		}
		if err := e.writeString(text); err != nil {
			return node.Message{}, err
		}

	case "cursor":
		row := int(getFloat(payload, "row", 0))
		col := int(getFloat(payload, "col", 0))
		if err := e.setCursor(row, col); err != nil {
			return node.Message{}, err
		}

	case "backlight":
		on, _ := payload["on"].(bool)
		e.setBacklight(on)

	case "display":
		on, _ := payload["on"].(bool)
		if on {
			e.displayOn()
		} else {
			e.displayOff()
		}

	case "cursor_show":
		show, _ := payload["show"].(bool)
		blink, _ := payload["blink"].(bool)
		e.cursorControl(show, blink)

	default:
		// Default: write text
		if text, ok := payload["text"].(string); ok {
			if err := e.writeString(text); err != nil {
				return node.Message{}, err
			}
		}
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    action,
			"cols":      e.config.Cols,
			"rows":      e.config.Rows,
			"address":   fmt.Sprintf("0x%02X", e.config.Address),
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// getFloat safely gets a float64 from a map
func getFloat(m map[string]interface{}, key string, def float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return def
}

// initLCD initializes the HD44780 LCD in 4-bit mode
func (e *LCDI2CExecutor) initLCD() error {
	// Wait for LCD to power up
	time.Sleep(50 * time.Millisecond)

	// Initialize in 4-bit mode
	// See HD44780 datasheet for initialization sequence
	e.write4Bits(0x03 << 4)
	time.Sleep(5 * time.Millisecond)
	e.write4Bits(0x03 << 4)
	time.Sleep(5 * time.Millisecond)
	e.write4Bits(0x03 << 4)
	time.Sleep(200 * time.Microsecond)
	e.write4Bits(0x02 << 4) // Set 4-bit mode

	// Function set: 4-bit, 2-line, 5x8 dots
	lines := byte(lcd2Line)
	if e.config.Rows == 1 {
		lines = lcd1Line
	}
	e.command(lcdFunctionSet | lcd4BitMode | lines | lcd5x8Dots)

	// Display control: display on, cursor off, blink off
	e.displayCtrl = lcdDisplayOn | lcdCursorOff | lcdBlinkOff
	e.command(lcdDisplayControl | e.displayCtrl)

	// Clear display
	e.clear()

	// Entry mode: increment, no shift
	e.displayMode = lcdEntryLeft | lcdEntryShiftDecrement
	e.command(lcdEntryModeSet | e.displayMode)

	return nil
}

// command sends a command to the LCD
func (e *LCDI2CExecutor) command(cmd byte) error {
	return e.send(cmd, 0)
}

// write sends data to the LCD
func (e *LCDI2CExecutor) write(data byte) error {
	return e.send(data, lcdRs)
}

// send sends a byte to the LCD
func (e *LCDI2CExecutor) send(data byte, mode byte) error {
	high := (data & 0xF0) | mode | e.backlight
	low := ((data << 4) & 0xF0) | mode | e.backlight

	e.write4Bits(high)
	e.write4Bits(low)
	return nil
}

// write4Bits writes 4 bits to the LCD with enable pulse
func (e *LCDI2CExecutor) write4Bits(data byte) {
	e.expanderWrite(data)
	e.pulseEnable(data)
}

// expanderWrite writes to the I2C expander
func (e *LCDI2CExecutor) expanderWrite(data byte) {
	e.dev.Write([]byte{data | e.backlight})
}

// pulseEnable sends enable pulse
func (e *LCDI2CExecutor) pulseEnable(data byte) {
	e.expanderWrite(data | lcdEnable)
	time.Sleep(1 * time.Microsecond)
	e.expanderWrite(data & ^byte(lcdEnable))
	time.Sleep(50 * time.Microsecond)
}

// clear clears the display
func (e *LCDI2CExecutor) clear() error {
	e.command(lcdClearDisplay)
	time.Sleep(2 * time.Millisecond)
	return nil
}

// home returns cursor to home position
func (e *LCDI2CExecutor) home() error {
	e.command(lcdReturnHome)
	time.Sleep(2 * time.Millisecond)
	return nil
}

// setCursor sets cursor position
func (e *LCDI2CExecutor) setCursor(row, col int) error {
	// Row offsets for different LCD sizes
	rowOffsets := []byte{0x00, 0x40, 0x14, 0x54}
	if row >= len(rowOffsets) {
		row = len(rowOffsets) - 1
	}
	if col >= e.config.Cols {
		col = e.config.Cols - 1
	}

	e.command(lcdSetDDRAMAddr | (rowOffsets[row] + byte(col)))
	return nil
}

// writeString writes a string to the LCD
func (e *LCDI2CExecutor) writeString(text string) error {
	for _, c := range text {
		e.write(byte(c))
	}
	return nil
}

// setBacklight turns backlight on or off
func (e *LCDI2CExecutor) setBacklight(on bool) {
	if on {
		e.backlight = lcdBacklight
	} else {
		e.backlight = lcdNoBacklight
	}
	e.expanderWrite(0)
}

// displayOn turns display on
func (e *LCDI2CExecutor) displayOn() {
	e.displayCtrl |= lcdDisplayOn
	e.command(lcdDisplayControl | e.displayCtrl)
}

// displayOff turns display off
func (e *LCDI2CExecutor) displayOff() {
	e.displayCtrl &= ^byte(lcdDisplayOn)
	e.command(lcdDisplayControl | e.displayCtrl)
}

// cursorControl controls cursor visibility and blink
func (e *LCDI2CExecutor) cursorControl(show, blink bool) {
	if show {
		e.displayCtrl |= lcdCursorOn
	} else {
		e.displayCtrl &= ^byte(lcdCursorOn)
	}
	if blink {
		e.displayCtrl |= lcdBlinkOn
	} else {
		e.displayCtrl &= ^byte(lcdBlinkOn)
	}
	e.command(lcdDisplayControl | e.displayCtrl)
}

// Cleanup releases resources
func (e *LCDI2CExecutor) Cleanup() error {
	if e.bus != nil {
		e.bus.Close()
		e.bus = nil
	}
	e.initialized = false
	return nil
}
