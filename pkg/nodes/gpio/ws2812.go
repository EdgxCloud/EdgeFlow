// +build linux

package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/node"
)

// WS2812Config configuration for WS2812 LED strip
type WS2812Config struct {
	Pin        int    `json:"pin"`         // GPIO pin (PWM capable recommended)
	NumLEDs    int    `json:"num_leds"`    // Number of LEDs in strip
	Brightness int    `json:"brightness"`  // Global brightness 0-255 (default: 255)
	Order      string `json:"order"`       // Color order: "GRB", "RGB", "BGR" (default: "GRB")
	Frequency  int    `json:"frequency"`   // PWM frequency (default: 800000)
	Invert     bool   `json:"invert"`      // Invert signal (for level shifters)
	Channel    int    `json:"channel"`     // DMA channel (default: 10)
}

// Color represents an RGB color
type Color struct {
	R, G, B uint8
}

// WS2812Executor controls WS2812/NeoPixel LED strips
type WS2812Executor struct {
	config      WS2812Config
	hal         hal.HAL
	leds        []Color
	mu          sync.Mutex
	initialized bool
}

// NewWS2812Executor creates a new WS2812 executor
func NewWS2812Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var wsConfig WS2812Config
	if err := json.Unmarshal(configJSON, &wsConfig); err != nil {
		return nil, fmt.Errorf("invalid WS2812 config: %w", err)
	}

	// Validate
	if wsConfig.Pin == 0 {
		return nil, fmt.Errorf("pin is required")
	}
	if wsConfig.NumLEDs == 0 {
		wsConfig.NumLEDs = 1
	}
	if wsConfig.NumLEDs > 1000 {
		return nil, fmt.Errorf("too many LEDs: max 1000")
	}

	// Defaults
	if wsConfig.Brightness == 0 {
		wsConfig.Brightness = 255
	}
	if wsConfig.Order == "" {
		wsConfig.Order = "GRB"
	}
	if wsConfig.Frequency == 0 {
		wsConfig.Frequency = 800000
	}
	if wsConfig.Channel == 0 {
		wsConfig.Channel = 10
	}

	return &WS2812Executor{
		config: wsConfig,
		leds:   make([]Color, wsConfig.NumLEDs),
	}, nil
}

// Init initializes the WS2812 executor
func (e *WS2812Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles LED strip commands
func (e *WS2812Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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

	// Parse command
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	action, _ := payload["action"].(string)

	switch action {
	case "set":
		// Set single LED or all LEDs
		index := int(getFloat(payload, "index", -1))
		r := uint8(getFloat(payload, "r", 0))
		g := uint8(getFloat(payload, "g", 0))
		b := uint8(getFloat(payload, "b", 0))

		if index >= 0 && index < e.config.NumLEDs {
			e.leds[index] = Color{R: r, G: g, B: b}
		} else if index < 0 {
			// Set all LEDs
			for i := range e.leds {
				e.leds[i] = Color{R: r, G: g, B: b}
			}
		}
		e.render()

	case "set_hex":
		// Set LED by hex color
		index := int(getFloat(payload, "index", -1))
		hex, _ := payload["color"].(string)
		color := parseHexColor(hex)

		if index >= 0 && index < e.config.NumLEDs {
			e.leds[index] = color
		} else if index < 0 {
			for i := range e.leds {
				e.leds[i] = color
			}
		}
		e.render()

	case "fill":
		// Fill range of LEDs
		start := int(getFloat(payload, "start", 0))
		end := int(getFloat(payload, "end", float64(e.config.NumLEDs)))
		r := uint8(getFloat(payload, "r", 0))
		g := uint8(getFloat(payload, "g", 0))
		b := uint8(getFloat(payload, "b", 0))

		for i := start; i < end && i < e.config.NumLEDs; i++ {
			if i >= 0 {
				e.leds[i] = Color{R: r, G: g, B: b}
			}
		}
		e.render()

	case "rainbow":
		// Rainbow effect
		offset := int(getFloat(payload, "offset", 0))
		e.rainbow(offset)
		e.render()

	case "chase":
		// Chase effect (single LED moving)
		color := Color{
			R: uint8(getFloat(payload, "r", 255)),
			G: uint8(getFloat(payload, "g", 0)),
			B: uint8(getFloat(payload, "b", 0)),
		}
		position := int(getFloat(payload, "position", 0)) % e.config.NumLEDs
		bgColor := Color{R: 0, G: 0, B: 0}

		for i := range e.leds {
			if i == position {
				e.leds[i] = color
			} else {
				e.leds[i] = bgColor
			}
		}
		e.render()

	case "gradient":
		// Gradient between two colors
		r1 := uint8(getFloat(payload, "r1", 0))
		g1 := uint8(getFloat(payload, "g1", 0))
		b1 := uint8(getFloat(payload, "b1", 0))
		r2 := uint8(getFloat(payload, "r2", 255))
		g2 := uint8(getFloat(payload, "g2", 255))
		b2 := uint8(getFloat(payload, "b2", 255))

		e.gradient(Color{R: r1, G: g1, B: b1}, Color{R: r2, G: g2, B: b2})
		e.render()

	case "brightness":
		// Set global brightness
		brightness := int(getFloat(payload, "brightness", 255))
		if brightness < 0 {
			brightness = 0
		}
		if brightness > 255 {
			brightness = 255
		}
		e.config.Brightness = brightness
		e.render()

	case "clear", "off":
		// Turn off all LEDs
		for i := range e.leds {
			e.leds[i] = Color{R: 0, G: 0, B: 0}
		}
		e.render()

	case "get":
		// Get current LED states
		colors := make([]map[string]interface{}, e.config.NumLEDs)
		for i, led := range e.leds {
			colors[i] = map[string]interface{}{
				"index": i,
				"r":     led.R,
				"g":     led.G,
				"b":     led.B,
				"hex":   fmt.Sprintf("#%02X%02X%02X", led.R, led.G, led.B),
			}
		}
		return node.Message{
			Payload: map[string]interface{}{
				"leds":       colors,
				"num_leds":   e.config.NumLEDs,
				"brightness": e.config.Brightness,
				"timestamp":  time.Now().Unix(),
			},
		}, nil

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":     action,
			"num_leds":   e.config.NumLEDs,
			"brightness": e.config.Brightness,
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

// render sends the LED data to the strip
// Note: This is a simplified implementation. Full WS2812 support requires
// precise timing (800KHz) which typically needs:
// - SPI at specific frequency
// - PWM with DMA
// - PIO (on RP2040)
// For production, use libraries like rpi_ws281x
func (e *WS2812Executor) render() {
	// This implementation uses a software approach which may not be reliable
	// for large strips. For production, use ws2811 library with DMA.

	gpio := e.hal.GPIO()
	pin := e.config.Pin

	// Set pin to output
	gpio.SetMode(pin, hal.Output)

	// Apply brightness scaling
	brightness := float64(e.config.Brightness) / 255.0

	// Send data for each LED
	for _, led := range e.leds {
		r := uint8(float64(led.R) * brightness)
		g := uint8(float64(led.G) * brightness)
		b := uint8(float64(led.B) * brightness)

		var data [3]byte
		switch e.config.Order {
		case "RGB":
			data = [3]byte{r, g, b}
		case "BGR":
			data = [3]byte{b, g, r}
		case "GRB":
			fallthrough
		default:
			data = [3]byte{g, r, b}
		}

		// Send 24 bits
		for _, byteVal := range data {
			for bit := 7; bit >= 0; bit-- {
				if byteVal&(1<<bit) != 0 {
					// Send '1': ~700ns high, ~600ns low
					gpio.DigitalWrite(pin, true)
					nopDelay(55) // ~700ns at 1GHz
					gpio.DigitalWrite(pin, false)
					nopDelay(45) // ~600ns
				} else {
					// Send '0': ~350ns high, ~800ns low
					gpio.DigitalWrite(pin, true)
					nopDelay(28) // ~350ns
					gpio.DigitalWrite(pin, false)
					nopDelay(65) // ~800ns
				}
			}
		}
	}

	// Reset: low for >50us
	gpio.DigitalWrite(pin, false)
	time.Sleep(60 * time.Microsecond)
}

// nopDelay performs a busy-wait delay (very approximate)
func nopDelay(cycles int) {
	for i := 0; i < cycles; i++ {
		// NOP equivalent
	}
}

// rainbow generates a rainbow pattern
func (e *WS2812Executor) rainbow(offset int) {
	for i := range e.leds {
		hue := (i*256/e.config.NumLEDs + offset) % 256
		e.leds[i] = hueToRGB(hue)
	}
}

// gradient fills LEDs with a gradient between two colors
func (e *WS2812Executor) gradient(c1, c2 Color) {
	for i := range e.leds {
		t := float64(i) / float64(e.config.NumLEDs-1)
		e.leds[i] = Color{
			R: uint8(float64(c1.R)*(1-t) + float64(c2.R)*t),
			G: uint8(float64(c1.G)*(1-t) + float64(c2.G)*t),
			B: uint8(float64(c1.B)*(1-t) + float64(c2.B)*t),
		}
	}
}

// hueToRGB converts hue (0-255) to RGB
func hueToRGB(hue int) Color {
	h := hue % 256
	region := h / 43
	remainder := (h % 43) * 6

	var r, g, b uint8
	switch region {
	case 0:
		r, g, b = 255, uint8(remainder), 0
	case 1:
		r, g, b = uint8(255-remainder), 255, 0
	case 2:
		r, g, b = 0, 255, uint8(remainder)
	case 3:
		r, g, b = 0, uint8(255-remainder), 255
	case 4:
		r, g, b = uint8(remainder), 0, 255
	default:
		r, g, b = 255, 0, uint8(255-remainder)
	}
	return Color{R: r, G: g, B: b}
}

// parseHexColor parses a hex color string
func parseHexColor(hex string) Color {
	if len(hex) == 0 {
		return Color{}
	}
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return Color{}
	}

	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return Color{R: r, G: g, B: b}
}

// Cleanup releases resources
func (e *WS2812Executor) Cleanup() error {
	// Turn off all LEDs
	for i := range e.leds {
		e.leds[i] = Color{R: 0, G: 0, B: 0}
	}
	if e.hal != nil {
		e.render()
	}
	return nil
}
