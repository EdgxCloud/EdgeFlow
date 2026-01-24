// Stub implementation of go-rpio for non-Linux platforms
package rpio

// Pin represents a GPIO pin (stub)
type Pin byte

// State represents pin state (stub)
type State byte

// Mode represents pin mode (stub)
type Mode byte

// Pull represents pull up/down state (stub)
type Pull byte

const (
	Input Mode = iota
	Output
)

const (
	Low State = iota
	High
)

const (
	PullOff Pull = iota
	PullDown
	PullUp
)

// Open is a stub
func Open() error {
	return nil
}

// Close is a stub
func Close() error {
	return nil
}

// Pin methods (stubs)
func (pin Pin) Input() {}
func (pin Pin) Output() {}
func (pin Pin) High() {}
func (pin Pin) Low() {}
func (pin Pin) Mode(mode Mode) {}
func (pin Pin) Write(state State) {}
func (pin Pin) Read() State { return Low }
func (pin Pin) Pull(pull Pull) {}
func (pin Pin) PullUp() {}
func (pin Pin) PullDown() {}
func (pin Pin) PullOff() {}
