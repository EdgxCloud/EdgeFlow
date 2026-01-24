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

// Musical note frequencies (Hz)
var noteFrequencies = map[string]int{
	"C0": 16, "C#0": 17, "D0": 18, "D#0": 19, "E0": 21, "F0": 22, "F#0": 23, "G0": 25, "G#0": 26, "A0": 28, "A#0": 29, "B0": 31,
	"C1": 33, "C#1": 35, "D1": 37, "D#1": 39, "E1": 41, "F1": 44, "F#1": 46, "G1": 49, "G#1": 52, "A1": 55, "A#1": 58, "B1": 62,
	"C2": 65, "C#2": 69, "D2": 73, "D#2": 78, "E2": 82, "F2": 87, "F#2": 93, "G2": 98, "G#2": 104, "A2": 110, "A#2": 117, "B2": 123,
	"C3": 131, "C#3": 139, "D3": 147, "D#3": 156, "E3": 165, "F3": 175, "F#3": 185, "G3": 196, "G#3": 208, "A3": 220, "A#3": 233, "B3": 247,
	"C4": 262, "C#4": 277, "D4": 294, "D#4": 311, "E4": 330, "F4": 349, "F#4": 370, "G4": 392, "G#4": 415, "A4": 440, "A#4": 466, "B4": 494,
	"C5": 523, "C#5": 554, "D5": 587, "D#5": 622, "E5": 659, "F5": 698, "F#5": 740, "G5": 784, "G#5": 831, "A5": 880, "A#5": 932, "B5": 988,
	"C6": 1047, "C#6": 1109, "D6": 1175, "D#6": 1245, "E6": 1319, "F6": 1397, "F#6": 1480, "G6": 1568, "G#6": 1661, "A6": 1760, "A#6": 1865, "B6": 1976,
	"C7": 2093, "C#7": 2217, "D7": 2349, "D#7": 2489, "E7": 2637, "F7": 2794, "F#7": 2960, "G7": 3136, "G#7": 3322, "A7": 3520, "A#7": 3729, "B7": 3951,
	"C8": 4186,
}

// BuzzerConfig configuration for buzzer node
type BuzzerConfig struct {
	Pin       int  `json:"pin"`        // GPIO pin
	ActiveLow bool `json:"active_low"` // Active low buzzer (default: false)
	PWM       bool `json:"pwm"`        // Use PWM for passive buzzer (default: true)
}

// BuzzerExecutor controls buzzers (active or passive)
type BuzzerExecutor struct {
	config      BuzzerConfig
	hal         hal.HAL
	mu          sync.Mutex
	initialized bool
	playing     bool
	stopChan    chan struct{}
}

// NewBuzzerExecutor creates a new buzzer executor
func NewBuzzerExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var buzConfig BuzzerConfig
	if err := json.Unmarshal(configJSON, &buzConfig); err != nil {
		return nil, fmt.Errorf("invalid buzzer config: %w", err)
	}

	if buzConfig.Pin == 0 {
		return nil, fmt.Errorf("pin is required")
	}

	return &BuzzerExecutor{
		config:   buzConfig,
		stopChan: make(chan struct{}),
	}, nil
}

// Init initializes the buzzer executor
func (e *BuzzerExecutor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles buzzer commands
func (e *BuzzerExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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

	// Initialize pin
	if !e.initialized {
		gpio := e.hal.GPIO()
		gpio.SetMode(e.config.Pin, hal.Output)
		gpio.DigitalWrite(e.config.Pin, e.config.ActiveLow)
		e.initialized = true
	}

	// Parse command
	payload := msg.Payload
	if payload == nil {
		payload = make(map[string]interface{})
	}

	action, _ := payload["action"].(string)

	switch action {
	case "on":
		// Turn on (for active buzzer)
		e.buzzerOn()

	case "off":
		// Turn off
		e.buzzerOff()

	case "beep":
		// Single beep
		duration := int(getFloat(payload, "duration", 100))
		e.beep(time.Duration(duration) * time.Millisecond)

	case "tone":
		// Play a tone (for passive buzzer with PWM)
		frequency := int(getFloat(payload, "frequency", 440))
		duration := int(getFloat(payload, "duration", 500))
		e.playTone(frequency, time.Duration(duration)*time.Millisecond)

	case "note":
		// Play a musical note
		note, _ := payload["note"].(string)
		duration := int(getFloat(payload, "duration", 500))
		if freq, ok := noteFrequencies[note]; ok {
			e.playTone(freq, time.Duration(duration)*time.Millisecond)
		} else {
			return node.Message{}, fmt.Errorf("unknown note: %s", note)
		}

	case "melody":
		// Play a melody (array of notes)
		notes, ok := payload["notes"].([]interface{})
		if !ok {
			return node.Message{}, fmt.Errorf("notes array required")
		}
		tempo := int(getFloat(payload, "tempo", 120))
		go e.playMelody(ctx, notes, tempo)

	case "pattern":
		// Play a beep pattern
		pattern, _ := payload["pattern"].(string)
		e.playPattern(ctx, pattern)

	case "stop":
		// Stop any playing sound
		e.stop()

	case "alert":
		// Alert sound (multiple quick beeps)
		count := int(getFloat(payload, "count", 3))
		go e.alert(ctx, count)

	case "siren":
		// Siren effect
		duration := int(getFloat(payload, "duration", 2000))
		go e.siren(ctx, time.Duration(duration)*time.Millisecond)

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    action,
			"pin":       e.config.Pin,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// buzzerOn turns on the buzzer
func (e *BuzzerExecutor) buzzerOn() {
	gpio := e.hal.GPIO()
	gpio.DigitalWrite(e.config.Pin, !e.config.ActiveLow)
}

// buzzerOff turns off the buzzer
func (e *BuzzerExecutor) buzzerOff() {
	gpio := e.hal.GPIO()
	gpio.DigitalWrite(e.config.Pin, e.config.ActiveLow)
	if e.config.PWM {
		gpio.PWMWrite(e.config.Pin, 0)
	}
}

// beep makes a single beep
func (e *BuzzerExecutor) beep(duration time.Duration) {
	e.buzzerOn()
	time.Sleep(duration)
	e.buzzerOff()
}

// playTone plays a tone at the specified frequency (for passive buzzer)
func (e *BuzzerExecutor) playTone(frequency int, duration time.Duration) {
	if !e.config.PWM {
		// For active buzzer, just beep
		e.beep(duration)
		return
	}

	gpio := e.hal.GPIO()
	gpio.SetMode(e.config.Pin, hal.Output)
	gpio.SetPWMFrequency(e.config.Pin, frequency)
	gpio.PWMWrite(e.config.Pin, 128) // 50% duty cycle

	time.Sleep(duration)

	gpio.PWMWrite(e.config.Pin, 0)
}

// playMelody plays a sequence of notes
func (e *BuzzerExecutor) playMelody(ctx context.Context, notes []interface{}, tempo int) {
	e.playing = true
	beatDuration := time.Minute / time.Duration(tempo)

	for _, n := range notes {
		select {
		case <-ctx.Done():
			e.buzzerOff()
			return
		case <-e.stopChan:
			e.buzzerOff()
			return
		default:
		}

		noteMap, ok := n.(map[string]interface{})
		if !ok {
			continue
		}

		note, _ := noteMap["note"].(string)
		beats := getFloat(noteMap, "beats", 1)

		duration := time.Duration(float64(beatDuration) * beats)

		if note == "R" || note == "rest" {
			// Rest
			time.Sleep(duration)
		} else if freq, ok := noteFrequencies[note]; ok {
			e.playTone(freq, duration*9/10) // Leave small gap between notes
			time.Sleep(duration / 10)
		}
	}

	e.playing = false
}

// playPattern plays a morse-code style pattern
// . = short beep, - = long beep, space = pause
func (e *BuzzerExecutor) playPattern(ctx context.Context, pattern string) {
	shortDuration := 100 * time.Millisecond
	longDuration := 300 * time.Millisecond
	pauseDuration := 100 * time.Millisecond

	for _, c := range pattern {
		select {
		case <-ctx.Done():
			return
		default:
		}

		switch c {
		case '.':
			e.beep(shortDuration)
			time.Sleep(pauseDuration)
		case '-':
			e.beep(longDuration)
			time.Sleep(pauseDuration)
		case ' ':
			time.Sleep(pauseDuration * 2)
		}
	}
}

// alert plays multiple quick beeps
func (e *BuzzerExecutor) alert(ctx context.Context, count int) {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}
		e.beep(100 * time.Millisecond)
		time.Sleep(100 * time.Millisecond)
	}
}

// siren plays a siren effect
func (e *BuzzerExecutor) siren(ctx context.Context, duration time.Duration) {
	if !e.config.PWM {
		// For active buzzer, just beep on/off
		endTime := time.Now().Add(duration)
		for time.Now().Before(endTime) {
			select {
			case <-ctx.Done():
				e.buzzerOff()
				return
			default:
			}
			e.beep(100 * time.Millisecond)
			time.Sleep(100 * time.Millisecond)
		}
		return
	}

	// For passive buzzer, sweep frequency
	gpio := e.hal.GPIO()
	endTime := time.Now().Add(duration)

	for time.Now().Before(endTime) {
		select {
		case <-ctx.Done():
			gpio.PWMWrite(e.config.Pin, 0)
			return
		default:
		}

		// Sweep up
		for freq := 400; freq < 800; freq += 20 {
			gpio.SetPWMFrequency(e.config.Pin, freq)
			gpio.PWMWrite(e.config.Pin, 128)
			time.Sleep(10 * time.Millisecond)
		}
		// Sweep down
		for freq := 800; freq > 400; freq -= 20 {
			gpio.SetPWMFrequency(e.config.Pin, freq)
			gpio.PWMWrite(e.config.Pin, 128)
			time.Sleep(10 * time.Millisecond)
		}
	}

	gpio.PWMWrite(e.config.Pin, 0)
}

// stop stops any playing sound
func (e *BuzzerExecutor) stop() {
	if e.playing {
		select {
		case e.stopChan <- struct{}{}:
		default:
		}
	}
	e.buzzerOff()
	e.playing = false
}

// Cleanup releases resources
func (e *BuzzerExecutor) Cleanup() error {
	e.stop()
	return nil
}
