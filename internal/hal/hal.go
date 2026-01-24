package hal

import (
	"fmt"
	"sync"
)

// PinMode حالت پین
type PinMode int

const (
	Input PinMode = iota
	Output
	PWM
)

// PullMode حالت pull resistor
type PullMode int

const (
	PullNone PullMode = iota
	PullUp
	PullDown
)

// EdgeMode حالت تشخیص لبه
type EdgeMode int

const (
	EdgeNone EdgeMode = iota
	EdgeRising
	EdgeFalling
	EdgeBoth
)

// GPIOProvider رابط GPIO
type GPIOProvider interface {
	// SetMode تنظیم حالت پین
	SetMode(pin int, mode PinMode) error
	// SetPull تنظیم pull resistor
	SetPull(pin int, pull PullMode) error
	// DigitalRead خواندن مقدار دیجیتال
	DigitalRead(pin int) (bool, error)
	// DigitalWrite نوشتن مقدار دیجیتال
	DigitalWrite(pin int, value bool) error
	// PWMWrite نوشتن PWM (0-255)
	PWMWrite(pin int, value int) error
	// SetPWMFrequency تنظیم فرکانس PWM
	SetPWMFrequency(pin int, freq int) error
	// WatchEdge تشخیص تغییر لبه
	WatchEdge(pin int, edge EdgeMode, callback func(pin int, value bool)) error
	// Close بستن GPIO
	Close() error
}

// I2CProvider رابط I2C
type I2CProvider interface {
	// Open باز کردن I2C با آدرس
	Open(address byte) error
	// Read خواندن byte ها
	Read(length int) ([]byte, error)
	// Write نوشتن byte ها
	Write(data []byte) error
	// ReadRegister خواندن از رجیستر
	ReadRegister(register byte, length int) ([]byte, error)
	// WriteRegister نوشتن به رجیستر
	WriteRegister(register byte, data []byte) error
	// Close بستن I2C
	Close() error
}

// SPIProvider رابط SPI
type SPIProvider interface {
	// Open باز کردن SPI
	Open(bus, device int) error
	// Transfer انتقال داده
	Transfer(data []byte) ([]byte, error)
	// SetSpeed تنظیم سرعت
	SetSpeed(speed int) error
	// SetMode تنظیم SPI mode
	SetMode(mode byte) error
	// SetBitsPerWord تنظیم تعداد بیت در هر کلمه
	SetBitsPerWord(bits byte) error
	// Close بستن SPI
	Close() error
}

// SerialProvider رابط Serial
type SerialProvider interface {
	// Open باز کردن Serial port
	Open(port string) error
	// SetBaudRate تنظیم baud rate
	SetBaudRate(baud int) error
	// SetDataBits تنظیم data bits
	SetDataBits(bits int) error
	// SetStopBits تنظیم stop bits
	SetStopBits(bits int) error
	// SetParity تنظیم parity (0=none, 1=odd, 2=even)
	SetParity(parity byte) error
	// Read خواندن
	Read(buffer []byte) (int, error)
	// Write نوشتن
	Write(data []byte) (int, error)
	// Close بستن Serial
	Close() error
}

// I2CBus رابط I2C Bus
type I2CBus interface {
	// Write نوشتن به آدرس
	Write(addr uint16, data []byte) error
	// Read خواندن از آدرس
	Read(addr uint16, data []byte) error
	// WriteRead نوشتن و خواندن
	WriteRead(addr uint16, write []byte, read []byte) error
	// Close بستن باس
	Close() error
}

// SPIDevice رابط SPI Device
type SPIDevice interface {
	// Transfer انتقال داده
	Transfer(data []byte) ([]byte, error)
	// Close بستن device
	Close() error
}

// HAL رابط Hardware Abstraction Layer
type HAL interface {
	// GPIO دریافت GPIO provider
	GPIO() GPIOProvider
	// I2C دریافت I2C provider
	I2C() I2CProvider
	// SPI دریافت SPI provider
	SPI() SPIProvider
	// Serial دریافت Serial provider
	Serial() SerialProvider
	// Info دریافت اطلاعات برد
	Info() BoardInfo
	// Close بستن HAL
	Close() error
}

var (
	globalHAL HAL
	halMu     sync.RWMutex
)

// SetGlobalHAL تنظیم HAL global
func SetGlobalHAL(hal HAL) {
	halMu.Lock()
	defer halMu.Unlock()
	globalHAL = hal
}

// GetGlobalHAL دریافت HAL global
func GetGlobalHAL() (HAL, error) {
	halMu.RLock()
	defer halMu.RUnlock()
	if globalHAL == nil {
		return nil, fmt.Errorf("HAL not initialized")
	}
	return globalHAL, nil
}
