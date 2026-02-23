package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/host/v3"
)

// OLEDSSD1306Config configuration for SSD1306 OLED display
type OLEDSSD1306Config struct {
	Bus      string `json:"bus"`      // I2C bus (default: "")
	Address  int    `json:"address"`  // I2C address (default: 0x3C)
	Width    int    `json:"width"`    // Display width (default: 128)
	Height   int    `json:"height"`   // Display height (default: 64)
	Rotation int    `json:"rotation"` // Rotation: 0, 90, 180, 270 (default: 0)
}

// OLEDSSD1306Executor executes OLED display operations
type OLEDSSD1306Executor struct {
	config      OLEDSSD1306Config
	bus         i2c.BusCloser
	dev         *ssd1306.Dev
	mu          sync.Mutex
	hostInited  bool
	initialized bool
	buffer      *image.RGBA
}

// NewOLEDSSD1306Executor creates a new OLED SSD1306 executor
func NewOLEDSSD1306Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var oledConfig OLEDSSD1306Config
	if err := json.Unmarshal(configJSON, &oledConfig); err != nil {
		return nil, fmt.Errorf("invalid OLED config: %w", err)
	}

	// Defaults
	if oledConfig.Address == 0 {
		oledConfig.Address = 0x3C // Common address for SSD1306
	}
	if oledConfig.Width == 0 {
		oledConfig.Width = 128
	}
	if oledConfig.Height == 0 {
		oledConfig.Height = 64
	}

	return &OLEDSSD1306Executor{
		config: oledConfig,
	}, nil
}

// Init initializes the OLED executor
func (e *OLEDSSD1306Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles OLED display operations
func (e *OLEDSSD1306Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Initialize hardware if needed
	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init periph host: %w", err)
		}
		e.hostInited = true
	}

	// Open I2C bus and initialize display if not done
	if !e.initialized {
		bus, err := i2creg.Open(e.config.Bus)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to open I2C bus: %w", err)
		}
		e.bus = bus

		opts := &ssd1306.Opts{
			W:       e.config.Width,
			H:       e.config.Height,
			Rotated: e.config.Rotation == 180,
		}

		dev, err := ssd1306.NewI2C(e.bus, opts)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to init SSD1306: %w", err)
		}
		e.dev = dev
		e.buffer = image.NewRGBA(image.Rect(0, 0, e.config.Width, e.config.Height))
		e.initialized = true
	}

	// Parse command from message
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	// Check if there's a text field without action - display it directly
	if text, ok := payload["text"].(string); ok && payload["action"] == nil {
		if err := e.drawText(text, 0, 0); err != nil {
			return node.Message{}, err
		}
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "text",
				"text":      text,
				"timestamp": time.Now().Unix(),
			},
		}, nil
	}

	action, _ := payload["action"].(string)

	switch action {
	case "clear":
		e.clear()

	case "text":
		text, _ := payload["text"].(string)
		x := int(getFloat(payload, "x", 0))
		y := int(getFloat(payload, "y", 0))
		if err := e.drawText(text, x, y); err != nil {
			return node.Message{}, err
		}

	case "pixel":
		x := int(getFloat(payload, "x", 0))
		y := int(getFloat(payload, "y", 0))
		on, _ := payload["on"].(bool)
		e.setPixel(x, y, on)

	case "line":
		x1 := int(getFloat(payload, "x1", 0))
		y1 := int(getFloat(payload, "y1", 0))
		x2 := int(getFloat(payload, "x2", 0))
		y2 := int(getFloat(payload, "y2", 0))
		e.drawLine(x1, y1, x2, y2)

	case "rect":
		x := int(getFloat(payload, "x", 0))
		y := int(getFloat(payload, "y", 0))
		w := int(getFloat(payload, "w", 10))
		h := int(getFloat(payload, "h", 10))
		fill, _ := payload["fill"].(bool)
		e.drawRect(x, y, w, h, fill)

	case "circle":
		cx := int(getFloat(payload, "cx", 0))
		cy := int(getFloat(payload, "cy", 0))
		r := int(getFloat(payload, "r", 5))
		fill, _ := payload["fill"].(bool)
		e.drawCircle(cx, cy, r, fill)

	case "invert":
		invert, _ := payload["invert"].(bool)
		e.dev.Invert(invert)

	case "display":
		on, _ := payload["on"].(bool)
		if !on {
			e.dev.Halt()
		}
		// Note: Resume/SetDisplayOn not available in ssd1306.Dev
		// Display is always on after init, Halt turns it off

	case "update":
		// Force display update
		e.update()

	default:
		// Default: display text
		if text, ok := payload["text"].(string); ok {
			if err := e.drawText(text, 0, 0); err != nil {
				return node.Message{}, err
			}
		}
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    action,
			"width":     e.config.Width,
			"height":    e.config.Height,
			"address":   fmt.Sprintf("0x%02X", e.config.Address),
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// clear clears the display
func (e *OLEDSSD1306Executor) clear() {
	for y := 0; y < e.config.Height; y++ {
		for x := 0; x < e.config.Width; x++ {
			e.buffer.Set(x, y, color.Black)
		}
	}
	e.update()
}

// setPixel sets a single pixel
func (e *OLEDSSD1306Executor) setPixel(x, y int, on bool) {
	if x < 0 || x >= e.config.Width || y < 0 || y >= e.config.Height {
		return
	}
	if on {
		e.buffer.Set(x, y, color.White)
	} else {
		e.buffer.Set(x, y, color.Black)
	}
	e.update()
}

// drawLine draws a line using Bresenham's algorithm
func (e *OLEDSSD1306Executor) drawLine(x1, y1, x2, y2 int) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	sy := 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		if x1 >= 0 && x1 < e.config.Width && y1 >= 0 && y1 < e.config.Height {
			e.buffer.Set(x1, y1, color.White)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
	e.update()
}

// drawRect draws a rectangle
func (e *OLEDSSD1306Executor) drawRect(x, y, w, h int, fill bool) {
	if fill {
		for py := y; py < y+h && py < e.config.Height; py++ {
			for px := x; px < x+w && px < e.config.Width; px++ {
				if px >= 0 && py >= 0 {
					e.buffer.Set(px, py, color.White)
				}
			}
		}
	} else {
		// Top and bottom
		for px := x; px < x+w && px < e.config.Width; px++ {
			if px >= 0 {
				if y >= 0 && y < e.config.Height {
					e.buffer.Set(px, y, color.White)
				}
				if y+h-1 >= 0 && y+h-1 < e.config.Height {
					e.buffer.Set(px, y+h-1, color.White)
				}
			}
		}
		// Left and right
		for py := y; py < y+h && py < e.config.Height; py++ {
			if py >= 0 {
				if x >= 0 && x < e.config.Width {
					e.buffer.Set(x, py, color.White)
				}
				if x+w-1 >= 0 && x+w-1 < e.config.Width {
					e.buffer.Set(x+w-1, py, color.White)
				}
			}
		}
	}
	e.update()
}

// drawCircle draws a circle using midpoint algorithm
func (e *OLEDSSD1306Executor) drawCircle(cx, cy, r int, fill bool) {
	x := r
	y := 0
	err := 0

	for x >= y {
		if fill {
			e.drawHLine(cx-x, cx+x, cy+y)
			e.drawHLine(cx-x, cx+x, cy-y)
			e.drawHLine(cx-y, cx+y, cy+x)
			e.drawHLine(cx-y, cx+y, cy-x)
		} else {
			e.setPixelDirect(cx+x, cy+y)
			e.setPixelDirect(cx-x, cy+y)
			e.setPixelDirect(cx+x, cy-y)
			e.setPixelDirect(cx-x, cy-y)
			e.setPixelDirect(cx+y, cy+x)
			e.setPixelDirect(cx-y, cy+x)
			e.setPixelDirect(cx+y, cy-x)
			e.setPixelDirect(cx-y, cy-x)
		}
		y++
		err += 1 + 2*y
		if 2*(err-x)+1 > 0 {
			x--
			err += 1 - 2*x
		}
	}
	e.update()
}

// drawHLine draws a horizontal line
func (e *OLEDSSD1306Executor) drawHLine(x1, x2, y int) {
	if y < 0 || y >= e.config.Height {
		return
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if x >= 0 && x < e.config.Width {
			e.buffer.Set(x, y, color.White)
		}
	}
}

// setPixelDirect sets a pixel without updating display
func (e *OLEDSSD1306Executor) setPixelDirect(x, y int) {
	if x >= 0 && x < e.config.Width && y >= 0 && y < e.config.Height {
		e.buffer.Set(x, y, color.White)
	}
}

// drawText draws text on the display (simple 5x7 font)
func (e *OLEDSSD1306Executor) drawText(text string, x, y int) error {
	// Simple 5x7 font - just the basic printable ASCII characters
	// Each character is 5 pixels wide + 1 pixel spacing
	charWidth := 6
	_ = 8 // charHeight for reference (5x7 font + 1 pixel spacing)

	for i, c := range text {
		if c >= 32 && c <= 126 {
			e.drawChar(c, x+i*charWidth, y)
		}
	}
	e.update()
	return nil
}

// drawChar draws a single character using a simple 5x7 font
func (e *OLEDSSD1306Executor) drawChar(c rune, x, y int) {
	// Simple font data for printable ASCII (space to ~)
	// This is a minimal 5x7 font for basic text display
	font := getBasicFont()

	idx := int(c) - 32
	if idx < 0 || idx >= len(font) {
		return
	}

	charData := font[idx]
	for col := 0; col < 5; col++ {
		bits := charData[col]
		for row := 0; row < 7; row++ {
			if bits&(1<<row) != 0 {
				px := x + col
				py := y + row
				if px >= 0 && px < e.config.Width && py >= 0 && py < e.config.Height {
					e.buffer.Set(px, py, color.White)
				}
			}
		}
	}
}

// update sends the buffer to the display
func (e *OLEDSSD1306Executor) update() {
	if e.dev != nil {
		e.dev.Draw(e.buffer.Bounds(), e.buffer, image.Point{})
	}
}

// abs returns absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// getBasicFont returns a simple 5x7 bitmap font
func getBasicFont() [][5]byte {
	return [][5]byte{
		{0x00, 0x00, 0x00, 0x00, 0x00}, // Space
		{0x00, 0x00, 0x5F, 0x00, 0x00}, // !
		{0x00, 0x07, 0x00, 0x07, 0x00}, // "
		{0x14, 0x7F, 0x14, 0x7F, 0x14}, // #
		{0x24, 0x2A, 0x7F, 0x2A, 0x12}, // $
		{0x23, 0x13, 0x08, 0x64, 0x62}, // %
		{0x36, 0x49, 0x55, 0x22, 0x50}, // &
		{0x00, 0x05, 0x03, 0x00, 0x00}, // '
		{0x00, 0x1C, 0x22, 0x41, 0x00}, // (
		{0x00, 0x41, 0x22, 0x1C, 0x00}, // )
		{0x08, 0x2A, 0x1C, 0x2A, 0x08}, // *
		{0x08, 0x08, 0x3E, 0x08, 0x08}, // +
		{0x00, 0x50, 0x30, 0x00, 0x00}, // ,
		{0x08, 0x08, 0x08, 0x08, 0x08}, // -
		{0x00, 0x60, 0x60, 0x00, 0x00}, // .
		{0x20, 0x10, 0x08, 0x04, 0x02}, // /
		{0x3E, 0x51, 0x49, 0x45, 0x3E}, // 0
		{0x00, 0x42, 0x7F, 0x40, 0x00}, // 1
		{0x42, 0x61, 0x51, 0x49, 0x46}, // 2
		{0x21, 0x41, 0x45, 0x4B, 0x31}, // 3
		{0x18, 0x14, 0x12, 0x7F, 0x10}, // 4
		{0x27, 0x45, 0x45, 0x45, 0x39}, // 5
		{0x3C, 0x4A, 0x49, 0x49, 0x30}, // 6
		{0x01, 0x71, 0x09, 0x05, 0x03}, // 7
		{0x36, 0x49, 0x49, 0x49, 0x36}, // 8
		{0x06, 0x49, 0x49, 0x29, 0x1E}, // 9
		{0x00, 0x36, 0x36, 0x00, 0x00}, // :
		{0x00, 0x56, 0x36, 0x00, 0x00}, // ;
		{0x00, 0x08, 0x14, 0x22, 0x41}, // <
		{0x14, 0x14, 0x14, 0x14, 0x14}, // =
		{0x41, 0x22, 0x14, 0x08, 0x00}, // >
		{0x02, 0x01, 0x51, 0x09, 0x06}, // ?
		{0x32, 0x49, 0x79, 0x41, 0x3E}, // @
		{0x7E, 0x11, 0x11, 0x11, 0x7E}, // A
		{0x7F, 0x49, 0x49, 0x49, 0x36}, // B
		{0x3E, 0x41, 0x41, 0x41, 0x22}, // C
		{0x7F, 0x41, 0x41, 0x22, 0x1C}, // D
		{0x7F, 0x49, 0x49, 0x49, 0x41}, // E
		{0x7F, 0x09, 0x09, 0x01, 0x01}, // F
		{0x3E, 0x41, 0x41, 0x51, 0x32}, // G
		{0x7F, 0x08, 0x08, 0x08, 0x7F}, // H
		{0x00, 0x41, 0x7F, 0x41, 0x00}, // I
		{0x20, 0x40, 0x41, 0x3F, 0x01}, // J
		{0x7F, 0x08, 0x14, 0x22, 0x41}, // K
		{0x7F, 0x40, 0x40, 0x40, 0x40}, // L
		{0x7F, 0x02, 0x04, 0x02, 0x7F}, // M
		{0x7F, 0x04, 0x08, 0x10, 0x7F}, // N
		{0x3E, 0x41, 0x41, 0x41, 0x3E}, // O
		{0x7F, 0x09, 0x09, 0x09, 0x06}, // P
		{0x3E, 0x41, 0x51, 0x21, 0x5E}, // Q
		{0x7F, 0x09, 0x19, 0x29, 0x46}, // R
		{0x46, 0x49, 0x49, 0x49, 0x31}, // S
		{0x01, 0x01, 0x7F, 0x01, 0x01}, // T
		{0x3F, 0x40, 0x40, 0x40, 0x3F}, // U
		{0x1F, 0x20, 0x40, 0x20, 0x1F}, // V
		{0x7F, 0x20, 0x18, 0x20, 0x7F}, // W
		{0x63, 0x14, 0x08, 0x14, 0x63}, // X
		{0x03, 0x04, 0x78, 0x04, 0x03}, // Y
		{0x61, 0x51, 0x49, 0x45, 0x43}, // Z
		{0x00, 0x00, 0x7F, 0x41, 0x41}, // [
		{0x02, 0x04, 0x08, 0x10, 0x20}, // backslash
		{0x41, 0x41, 0x7F, 0x00, 0x00}, // ]
		{0x04, 0x02, 0x01, 0x02, 0x04}, // ^
		{0x40, 0x40, 0x40, 0x40, 0x40}, // _
		{0x00, 0x01, 0x02, 0x04, 0x00}, // `
		{0x20, 0x54, 0x54, 0x54, 0x78}, // a
		{0x7F, 0x48, 0x44, 0x44, 0x38}, // b
		{0x38, 0x44, 0x44, 0x44, 0x20}, // c
		{0x38, 0x44, 0x44, 0x48, 0x7F}, // d
		{0x38, 0x54, 0x54, 0x54, 0x18}, // e
		{0x08, 0x7E, 0x09, 0x01, 0x02}, // f
		{0x08, 0x14, 0x54, 0x54, 0x3C}, // g
		{0x7F, 0x08, 0x04, 0x04, 0x78}, // h
		{0x00, 0x44, 0x7D, 0x40, 0x00}, // i
		{0x20, 0x40, 0x44, 0x3D, 0x00}, // j
		{0x00, 0x7F, 0x10, 0x28, 0x44}, // k
		{0x00, 0x41, 0x7F, 0x40, 0x00}, // l
		{0x7C, 0x04, 0x18, 0x04, 0x78}, // m
		{0x7C, 0x08, 0x04, 0x04, 0x78}, // n
		{0x38, 0x44, 0x44, 0x44, 0x38}, // o
		{0x7C, 0x14, 0x14, 0x14, 0x08}, // p
		{0x08, 0x14, 0x14, 0x18, 0x7C}, // q
		{0x7C, 0x08, 0x04, 0x04, 0x08}, // r
		{0x48, 0x54, 0x54, 0x54, 0x20}, // s
		{0x04, 0x3F, 0x44, 0x40, 0x20}, // t
		{0x3C, 0x40, 0x40, 0x20, 0x7C}, // u
		{0x1C, 0x20, 0x40, 0x20, 0x1C}, // v
		{0x3C, 0x40, 0x30, 0x40, 0x3C}, // w
		{0x44, 0x28, 0x10, 0x28, 0x44}, // x
		{0x0C, 0x50, 0x50, 0x50, 0x3C}, // y
		{0x44, 0x64, 0x54, 0x4C, 0x44}, // z
		{0x00, 0x08, 0x36, 0x41, 0x00}, // {
		{0x00, 0x00, 0x7F, 0x00, 0x00}, // |
		{0x00, 0x41, 0x36, 0x08, 0x00}, // }
		{0x08, 0x08, 0x2A, 0x1C, 0x08}, // ~
	}
}

// Cleanup releases resources
func (e *OLEDSSD1306Executor) Cleanup() error {
	if e.dev != nil {
		e.dev.Halt()
		e.dev = nil
	}
	if e.bus != nil {
		e.bus.Close()
		e.bus = nil
	}
	e.initialized = false
	return nil
}
