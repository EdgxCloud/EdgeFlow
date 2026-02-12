//go:build !linux
// +build !linux

package hal

import "fmt"

// GpiocdevGPIO is a stub for non-Linux platforms.
type GpiocdevGPIO struct {
	chipName string
}

func NewGpiocdevGPIO(chipName string) (*GpiocdevGPIO, error) {
	return &GpiocdevGPIO{chipName: chipName}, nil
}

func (g *GpiocdevGPIO) SetMode(pin int, mode PinMode) error {
	return fmt.Errorf("GPIO not supported on this platform")
}

func (g *GpiocdevGPIO) SetPull(pin int, pull PullMode) error {
	return fmt.Errorf("GPIO not supported on this platform")
}

func (g *GpiocdevGPIO) DigitalRead(pin int) (bool, error) {
	return false, fmt.Errorf("GPIO not supported on this platform")
}

func (g *GpiocdevGPIO) DigitalWrite(pin int, value bool) error {
	return fmt.Errorf("GPIO not supported on this platform")
}

func (g *GpiocdevGPIO) PWMWrite(pin int, value int) error {
	return fmt.Errorf("GPIO not supported on this platform")
}

func (g *GpiocdevGPIO) SetPWMFrequency(pin int, freq int) error {
	return fmt.Errorf("GPIO not supported on this platform")
}

func (g *GpiocdevGPIO) WatchEdge(pin int, edge EdgeMode, callback func(pin int, value bool)) error {
	return fmt.Errorf("GPIO not supported on this platform")
}

func (g *GpiocdevGPIO) ActivePins() map[int]PinMode {
	return nil
}

func (g *GpiocdevGPIO) Close() error {
	return nil
}
