package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/edgeflow/edgeflow/internal/node"
)

// ColorPickerConfig configuration for color picker widget
type ColorPickerConfig struct {
	Format       string `json:"format"`       // hex, rgb, hsl
	DefaultColor string `json:"defaultColor"` // Default color value
	Label        string `json:"label"`        // Widget label
	Group        string `json:"group"`        // Dashboard group
	Tab          string `json:"tab"`          // Dashboard tab
	ShowAlpha    bool   `json:"showAlpha"`    // Show alpha channel
}

// ColorPickerExecutor implements the color picker dashboard widget
type ColorPickerExecutor struct {
	config       ColorPickerConfig
	currentColor string
}

// NewColorPickerExecutor creates a new color picker executor
func NewColorPickerExecutor() node.Executor {
	return &ColorPickerExecutor{}
}

// Init initializes the color picker with configuration
func (e *ColorPickerExecutor) Init(config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal color picker config: %w", err)
	}
	if err := json.Unmarshal(data, &e.config); err != nil {
		return fmt.Errorf("failed to unmarshal color picker config: %w", err)
	}

	if e.config.Format == "" {
		e.config.Format = "hex"
	}
	if e.config.DefaultColor == "" {
		e.config.DefaultColor = "#ffffff"
	}
	e.currentColor = e.config.DefaultColor
	return nil
}

// Execute handles color picker input and produces color output in all formats
func (e *ColorPickerExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	var r, g, b uint8

	rVal, hasR := msg.Payload["r"]
	gVal, hasG := msg.Payload["g"]
	bVal, hasB := msg.Payload["b"]

	if hasR && hasG && hasB {
		r = colorToUint8(rVal)
		g = colorToUint8(gVal)
		b = colorToUint8(bVal)
		e.currentColor = cpRGBToHex(r, g, b)
	} else if colorVal, ok := msg.Payload["color"]; ok {
		colorStr, ok := colorVal.(string)
		if !ok {
			return node.Message{}, fmt.Errorf("color value must be a string")
		}
		parsed, err := cpParseHexColor(colorStr)
		if err != nil {
			return node.Message{}, fmt.Errorf("invalid color value %q: %w", colorStr, err)
		}
		r, g, b = parsed[0], parsed[1], parsed[2]
		e.currentColor = cpRGBToHex(r, g, b)
	} else {
		parsed, err := cpParseHexColor(e.currentColor)
		if err != nil {
			return node.Message{}, fmt.Errorf("invalid current color %q: %w", e.currentColor, err)
		}
		r, g, b = parsed[0], parsed[1], parsed[2]
	}

	h, s, l := cpRGBToHSL(r, g, b)
	hexColor := cpRGBToHex(r, g, b)

	return node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"color":  hexColor,
			"hex":    hexColor,
			"r":      int(r),
			"g":      int(g),
			"b":      int(b),
			"h":      h,
			"s":      s,
			"l":      l,
			"format": e.config.Format,
			"label":  e.config.Label,
		},
	}, nil
}

// Cleanup releases resources
func (e *ColorPickerExecutor) Cleanup() error {
	return nil
}

func cpParseHexColor(hex string) ([3]uint8, error) {
	hex = strings.TrimSpace(hex)
	hex = strings.TrimPrefix(hex, "#")
	hex = strings.ToLower(hex)
	var r, g, b uint8
	switch len(hex) {
	case 3:
		_, err := fmt.Sscanf(hex, "%1x%1x%1x", &r, &g, &b)
		if err != nil {
			return [3]uint8{}, fmt.Errorf("invalid short hex color: %w", err)
		}
		r = r*16 + r
		g = g*16 + g
		b = b*16 + b
	case 6:
		_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return [3]uint8{}, fmt.Errorf("invalid hex color: %w", err)
		}
	default:
		return [3]uint8{}, fmt.Errorf("hex color must be 3 or 6 characters, got %d", len(hex))
	}
	return [3]uint8{r, g, b}, nil
}

func cpRGBToHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func cpRGBToHSL(r, g, b uint8) (int, int, int) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0
	maxVal := math.Max(rf, math.Max(gf, bf))
	minVal := math.Min(rf, math.Min(gf, bf))
	delta := maxVal - minVal
	l := (maxVal + minVal) / 2.0
	if delta == 0 {
		return 0, 0, int(math.Round(l * 100))
	}
	var s float64
	if l < 0.5 {
		s = delta / (maxVal + minVal)
	} else {
		s = delta / (2.0 - maxVal - minVal)
	}
	var h float64
	switch maxVal {
	case rf:
		h = (gf - bf) / delta
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/delta + 2
	case bf:
		h = (rf-gf)/delta + 4
	}
	h *= 60
	return int(math.Round(h)), int(math.Round(s * 100)), int(math.Round(l * 100))
}

func colorToUint8(val interface{}) uint8 {
	switch v := val.(type) {
	case float64:
		return uint8(math.Max(0, math.Min(255, math.Round(v))))
	case int:
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	case int64:
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	default:
		return 0
	}
}
