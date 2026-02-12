package hal

import (
	"fmt"
	"sync"
)

// PinMode pin mode
type PinMode int

const (
	Input PinMode = iota
	Output
	PWM
)

// PullMode pull resistor mode
type PullMode int

const (
	PullNone PullMode = iota
	PullUp
	PullDown
)

// EdgeMode edge detection mode
type EdgeMode int

const (
	EdgeNone EdgeMode = iota
	EdgeRising
	EdgeFalling
	EdgeBoth
)

// GPIOProvider GPIO interface
type GPIOProvider interface {
	// SetMode set pin mode
	SetMode(pin int, mode PinMode) error
	// SetPull set pull resistor
	SetPull(pin int, pull PullMode) error
	// DigitalRead digital read
	DigitalRead(pin int) (bool, error)
	// DigitalWrite digital write
	DigitalWrite(pin int, value bool) error
	// PWMWrite write PWM value (0-255)
	PWMWrite(pin int, value int) error
	// SetPWMFrequency set PWM frequency
	SetPWMFrequency(pin int, freq int) error
	// WatchEdge watch for edge changes
	WatchEdge(pin int, edge EdgeMode, callback func(pin int, value bool)) error
	// ActivePins returns a map of currently configured pins and their modes
	ActivePins() map[int]PinMode
	// Close close GPIO
	Close() error
}

// I2CProvider I2C interface
type I2CProvider interface {
	// Open open I2C with address
	Open(address byte) error
	// Read read bytes
	Read(length int) ([]byte, error)
	// Write write bytes
	Write(data []byte) error
	// ReadRegister read from register
	ReadRegister(register byte, length int) ([]byte, error)
	// WriteRegister write to register
	WriteRegister(register byte, data []byte) error
	// Close close I2C
	Close() error
}

// SPIProvider SPI interface
type SPIProvider interface {
	// Open open SPI
	Open(bus, device int) error
	// Transfer transfer data
	Transfer(data []byte) ([]byte, error)
	// SetSpeed set speed
	SetSpeed(speed int) error
	// SetMode set SPI mode
	SetMode(mode byte) error
	// SetBitsPerWord set bits per word
	SetBitsPerWord(bits byte) error
	// Close close SPI
	Close() error
}

// SerialProvider Serial interface
type SerialProvider interface {
	// Open open Serial port
	Open(port string) error
	// SetBaudRate set baud rate
	SetBaudRate(baud int) error
	// SetDataBits set data bits
	SetDataBits(bits int) error
	// SetStopBits set stop bits
	SetStopBits(bits int) error
	// SetParity set parity (0=none, 1=odd, 2=even)
	SetParity(parity byte) error
	// Read read
	Read(buffer []byte) (int, error)
	// Write write
	Write(data []byte) (int, error)
	// Close close Serial
	Close() error
}

// I2CBus I2C Bus interface
type I2CBus interface {
	// Write write to address
	Write(addr uint16, data []byte) error
	// Read read from address
	Read(addr uint16, data []byte) error
	// WriteRead write and read
	WriteRead(addr uint16, write []byte, read []byte) error
	// Close close bus
	Close() error
}

// SPIDevice SPI Device interface
type SPIDevice interface {
	// Transfer transfer data
	Transfer(data []byte) ([]byte, error)
	// Close close device
	Close() error
}

// HAL Hardware Abstraction Layer interface
type HAL interface {
	// GPIO get GPIO provider
	GPIO() GPIOProvider
	// I2C get I2C provider
	I2C() I2CProvider
	// SPI get SPI provider
	SPI() SPIProvider
	// Serial get Serial provider
	Serial() SerialProvider
	// Info get board information
	Info() BoardInfo
	// Close close HAL
	Close() error
}

var (
	globalHAL HAL
	halMu     sync.RWMutex
)

// SetGlobalHAL set global HAL
func SetGlobalHAL(hal HAL) {
	halMu.Lock()
	defer halMu.Unlock()
	globalHAL = hal
}

// GetGlobalHAL get global HAL
func GetGlobalHAL() (HAL, error) {
	halMu.RLock()
	defer halMu.RUnlock()
	if globalHAL == nil {
		return nil, fmt.Errorf("HAL not initialized")
	}
	return globalHAL, nil
}
