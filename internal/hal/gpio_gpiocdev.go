//go:build linux
// +build linux

package hal

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// GpiocdevGPIO implements GPIOProvider using the Linux GPIO character device
// interface via go-gpiocdev. This works on both Pi 4 (gpiochip0) and
// Pi 5 (gpiochip4 / RP1 southbridge).
type GpiocdevGPIO struct {
	mu       sync.Mutex
	chipName string
	lines    map[int]*gpiocdev.Line
	pinModes map[int]PinMode
	pinPulls map[int]PullMode
	pwm      map[int]*SoftPWM
	watchers map[int]context.CancelFunc
}

// SoftPWM implements software-based PWM using a goroutine that toggles
// the output pin. Hardware PWM is not available through the character device.
type SoftPWM struct {
	mu        sync.Mutex
	line      *gpiocdev.Line
	frequency int // Hz
	dutyCycle int // 0-255
	cancel    context.CancelFunc
	running   bool
}

// NewGpiocdevGPIO creates a new GPIO provider for the given chip name.
func NewGpiocdevGPIO(chipName string) (*GpiocdevGPIO, error) {
	// Verify the chip exists by briefly opening and closing it
	c, err := gpiocdev.NewChip(chipName)
	if err != nil {
		return nil, fmt.Errorf("failed to open GPIO chip %s: %w", chipName, err)
	}
	c.Close()

	return &GpiocdevGPIO{
		chipName: chipName,
		lines:    make(map[int]*gpiocdev.Line),
		pinModes: make(map[int]PinMode),
		pinPulls: make(map[int]PullMode),
		pwm:      make(map[int]*SoftPWM),
		watchers: make(map[int]context.CancelFunc),
	}, nil
}

func (g *GpiocdevGPIO) SetMode(pin int, mode PinMode) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Close existing line if any
	if err := g.closeLineLocked(pin); err != nil {
		return err
	}

	var opts []gpiocdev.LineReqOption

	// Apply stored pull mode if set
	if pull, ok := g.pinPulls[pin]; ok {
		opts = append(opts, pullOption(pull))
	}

	switch mode {
	case Input:
		opts = append([]gpiocdev.LineReqOption{gpiocdev.AsInput}, opts...)
		line, err := gpiocdev.RequestLine(g.chipName, pin, opts...)
		if err != nil {
			return fmt.Errorf("failed to request pin %d as input: %w", pin, err)
		}
		g.lines[pin] = line

	case Output:
		opts = append([]gpiocdev.LineReqOption{gpiocdev.AsOutput(0)}, opts...)
		line, err := gpiocdev.RequestLine(g.chipName, pin, opts...)
		if err != nil {
			return fmt.Errorf("failed to request pin %d as output: %w", pin, err)
		}
		g.lines[pin] = line

	case PWM:
		opts = append([]gpiocdev.LineReqOption{gpiocdev.AsOutput(0)}, opts...)
		line, err := gpiocdev.RequestLine(g.chipName, pin, opts...)
		if err != nil {
			return fmt.Errorf("failed to request pin %d for PWM: %w", pin, err)
		}
		g.lines[pin] = line

		// Initialize software PWM
		ctx, cancel := context.WithCancel(context.Background())
		sp := &SoftPWM{
			line:      line,
			frequency: 1000,
			dutyCycle: 0,
			cancel:    cancel,
			running:   true,
		}
		g.pwm[pin] = sp
		go sp.run(ctx)

	default:
		return fmt.Errorf("unsupported pin mode: %v", mode)
	}

	g.pinModes[pin] = mode
	return nil
}

func (g *GpiocdevGPIO) SetPull(pin int, pull PullMode) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.pinPulls[pin] = pull

	// If the line is already open, close and re-request with new pull setting
	_, ok := g.lines[pin]
	if !ok {
		return nil // Pull will be applied when SetMode is called
	}

	mode, modeOk := g.pinModes[pin]
	if !modeOk {
		return nil
	}

	// Close existing line
	if err := g.closeLineLocked(pin); err != nil {
		return fmt.Errorf("failed to close pin %d for pull reconfigure: %w", pin, err)
	}

	// Re-request with updated pull
	var opts []gpiocdev.LineReqOption
	opts = append(opts, pullOption(pull))

	switch mode {
	case Input:
		opts = append([]gpiocdev.LineReqOption{gpiocdev.AsInput}, opts...)
	case Output, PWM:
		opts = append([]gpiocdev.LineReqOption{gpiocdev.AsOutput(0)}, opts...)
	}

	line, err := gpiocdev.RequestLine(g.chipName, pin, opts...)
	if err != nil {
		return fmt.Errorf("failed to re-request pin %d with pull %v: %w", pin, pull, err)
	}
	g.lines[pin] = line

	return nil
}

func (g *GpiocdevGPIO) DigitalRead(pin int) (bool, error) {
	g.mu.Lock()
	line, ok := g.lines[pin]
	g.mu.Unlock()

	if !ok {
		return false, fmt.Errorf("pin %d not initialized", pin)
	}

	val, err := line.Value()
	if err != nil {
		return false, fmt.Errorf("failed to read pin %d: %w", pin, err)
	}
	return val != 0, nil
}

func (g *GpiocdevGPIO) DigitalWrite(pin int, value bool) error {
	g.mu.Lock()
	line, ok := g.lines[pin]
	g.mu.Unlock()

	if !ok {
		return fmt.Errorf("pin %d not initialized", pin)
	}

	v := 0
	if value {
		v = 1
	}
	if err := line.SetValue(v); err != nil {
		return fmt.Errorf("failed to write pin %d: %w", pin, err)
	}
	return nil
}

func (g *GpiocdevGPIO) PWMWrite(pin int, dutyCycle int) error {
	g.mu.Lock()
	sp, ok := g.pwm[pin]
	g.mu.Unlock()

	if !ok {
		return fmt.Errorf("pin %d not configured for PWM", pin)
	}

	if dutyCycle < 0 {
		dutyCycle = 0
	}
	if dutyCycle > 255 {
		dutyCycle = 255
	}

	sp.mu.Lock()
	sp.dutyCycle = dutyCycle
	sp.mu.Unlock()
	return nil
}

func (g *GpiocdevGPIO) SetPWMFrequency(pin int, freq int) error {
	g.mu.Lock()
	sp, ok := g.pwm[pin]
	g.mu.Unlock()

	if !ok {
		return fmt.Errorf("pin %d not configured for PWM", pin)
	}
	if freq <= 0 {
		return fmt.Errorf("frequency must be positive, got %d", freq)
	}

	sp.mu.Lock()
	sp.frequency = freq
	sp.mu.Unlock()
	return nil
}

func (g *GpiocdevGPIO) WatchEdge(pin int, edge EdgeMode, callback func(pin int, value bool)) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Cancel existing watcher if any
	if cancel, ok := g.watchers[pin]; ok {
		cancel()
		delete(g.watchers, pin)
	}

	// Close existing line
	if err := g.closeLineLocked(pin); err != nil {
		return err
	}

	if edge == EdgeNone {
		// Just re-request as input without edge detection
		line, err := gpiocdev.RequestLine(g.chipName, pin, gpiocdev.AsInput)
		if err != nil {
			return fmt.Errorf("failed to request pin %d as input: %w", pin, err)
		}
		g.lines[pin] = line
		g.pinModes[pin] = Input
		return nil
	}

	// Create event handler
	pinNum := pin // capture for closure
	handler := func(evt gpiocdev.LineEvent) {
		val := evt.Type == gpiocdev.LineEventRisingEdge
		callback(pinNum, val)
	}

	opts := []gpiocdev.LineReqOption{
		gpiocdev.WithEventHandler(handler),
	}

	// Apply pull mode if set
	if pull, ok := g.pinPulls[pin]; ok {
		opts = append(opts, pullOption(pull))
	}

	switch edge {
	case EdgeRising:
		opts = append(opts, gpiocdev.WithRisingEdge)
	case EdgeFalling:
		opts = append(opts, gpiocdev.WithFallingEdge)
	case EdgeBoth:
		opts = append(opts, gpiocdev.WithBothEdges)
	}

	line, err := gpiocdev.RequestLine(g.chipName, pin, opts...)
	if err != nil {
		return fmt.Errorf("failed to watch edge on pin %d: %w", pin, err)
	}
	g.lines[pin] = line
	g.pinModes[pin] = Input

	// Store a cancel function to stop the watcher
	_, cancel := context.WithCancel(context.Background())
	g.watchers[pin] = cancel

	return nil
}

// ActivePins returns a map of currently configured pins and their modes
func (g *GpiocdevGPIO) ActivePins() map[int]PinMode {
	g.mu.Lock()
	defer g.mu.Unlock()
	result := make(map[int]PinMode, len(g.pinModes))
	for pin, mode := range g.pinModes {
		result[pin] = mode
	}
	return result
}

func (g *GpiocdevGPIO) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Stop all PWM goroutines
	for pin, sp := range g.pwm {
		sp.cancel()
		delete(g.pwm, pin)
	}

	// Cancel all watchers
	for pin, cancel := range g.watchers {
		cancel()
		delete(g.watchers, pin)
	}

	// Close all lines
	for pin, line := range g.lines {
		line.Close()
		delete(g.lines, pin)
	}

	return nil
}

// closeLineLocked closes the line for the given pin. Must be called with g.mu held.
func (g *GpiocdevGPIO) closeLineLocked(pin int) error {
	// Stop PWM if running
	if sp, ok := g.pwm[pin]; ok {
		sp.cancel()
		delete(g.pwm, pin)
	}

	// Cancel watcher if active
	if cancel, ok := g.watchers[pin]; ok {
		cancel()
		delete(g.watchers, pin)
	}

	// Close the line
	if line, ok := g.lines[pin]; ok {
		line.Close()
		delete(g.lines, pin)
	}

	delete(g.pinModes, pin)
	return nil
}

// pullOption converts a PullMode to a gpiocdev line request option.
func pullOption(pull PullMode) gpiocdev.LineReqOption {
	switch pull {
	case PullUp:
		return gpiocdev.WithPullUp
	case PullDown:
		return gpiocdev.WithPullDown
	default:
		return gpiocdev.WithBiasDisabled
	}
}

// run is the software PWM goroutine loop.
func (sp *SoftPWM) run(ctx context.Context) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for {
		select {
		case <-ctx.Done():
			// Ensure pin is low on exit
			sp.line.SetValue(0)
			return
		default:
		}

		sp.mu.Lock()
		duty := sp.dutyCycle
		freq := sp.frequency
		sp.mu.Unlock()

		if freq <= 0 {
			freq = 1000
		}

		periodUs := int64(1000000) / int64(freq)

		if duty <= 0 {
			// Fully off - keep pin low, sleep one period
			sp.line.SetValue(0)
			sleepMicroseconds(ctx, periodUs)
			continue
		}
		if duty >= 255 {
			// Fully on - keep pin high, sleep one period
			sp.line.SetValue(1)
			sleepMicroseconds(ctx, periodUs)
			continue
		}

		onUs := periodUs * int64(duty) / 255
		offUs := periodUs - onUs

		sp.line.SetValue(1)
		sleepMicroseconds(ctx, onUs)

		sp.line.SetValue(0)
		sleepMicroseconds(ctx, offUs)
	}
}

// sleepMicroseconds sleeps for the given duration, checking for context cancellation.
func sleepMicroseconds(ctx context.Context, us int64) {
	if us <= 0 {
		return
	}
	t := time.NewTimer(time.Duration(us) * time.Microsecond)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
