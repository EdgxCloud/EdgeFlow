package hal

import (
	"fmt"
	"sync"
)

// MockHAL پیاده‌سازی mock برای تست
type MockHAL struct {
	gpio   *MockGPIO
	i2c    *MockI2C
	spi    *MockSPI
	serial *MockSerial
	info   BoardInfo
}

// NewMockHAL ایجاد MockHAL
func NewMockHAL() *MockHAL {
	return &MockHAL{
		gpio:   &MockGPIO{pins: make(map[int]*MockPin)},
		i2c:    &MockI2C{},
		spi:    &MockSPI{},
		serial: &MockSerial{},
		info: BoardInfo{
			Model:    BoardUnknown,
			Name:     "Mock Board",
			HasWiFi:  false,
			HasBT:    false,
			NumGPIO:  40,
			NumPWM:   4,
			NumI2C:   2,
			NumSPI:   2,
			CPUCores: 4,
			RAMSize:  1024,
		},
	}
}

func (m *MockHAL) GPIO() GPIOProvider   { return m.gpio }
func (m *MockHAL) I2C() I2CProvider     { return m.i2c }
func (m *MockHAL) SPI() SPIProvider     { return m.spi }
func (m *MockHAL) Serial() SerialProvider { return m.serial }
func (m *MockHAL) Info() BoardInfo      { return m.info }
func (m *MockHAL) Close() error         { return nil }

// MockPin پین mock
type MockPin struct {
	mode  PinMode
	pull  PullMode
	value bool
	pwm   int
	freq  int
}

// MockGPIO GPIO mock
type MockGPIO struct {
	pins map[int]*MockPin
	mu   sync.RWMutex
}

func (g *MockGPIO) SetMode(pin int, mode PinMode) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.pins[pin] == nil {
		g.pins[pin] = &MockPin{}
	}
	g.pins[pin].mode = mode
	return nil
}

func (g *MockGPIO) SetPull(pin int, pull PullMode) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.pins[pin] == nil {
		g.pins[pin] = &MockPin{}
	}
	g.pins[pin].pull = pull
	return nil
}

func (g *MockGPIO) DigitalRead(pin int) (bool, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.pins[pin] == nil {
		return false, fmt.Errorf("pin %d not initialized", pin)
	}
	return g.pins[pin].value, nil
}

func (g *MockGPIO) DigitalWrite(pin int, value bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.pins[pin] == nil {
		g.pins[pin] = &MockPin{}
	}
	g.pins[pin].value = value
	return nil
}

func (g *MockGPIO) PWMWrite(pin int, value int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.pins[pin] == nil {
		g.pins[pin] = &MockPin{}
	}
	if value < 0 || value > 255 {
		return fmt.Errorf("PWM value must be 0-255")
	}
	g.pins[pin].pwm = value
	return nil
}

func (g *MockGPIO) SetPWMFrequency(pin int, freq int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.pins[pin] == nil {
		g.pins[pin] = &MockPin{}
	}
	g.pins[pin].freq = freq
	return nil
}

func (g *MockGPIO) WatchEdge(pin int, edge EdgeMode, callback func(pin int, value bool)) error {
	// Mock implementation - در واقعیت باید event loop داشته باشد
	return nil
}

func (g *MockGPIO) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.pins = make(map[int]*MockPin)
	return nil
}

// MockI2C I2C mock
type MockI2C struct {
	address byte
	data    []byte
	mu      sync.RWMutex
}

func (i *MockI2C) Open(address byte) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.address = address
	return nil
}

func (i *MockI2C) Read(length int) ([]byte, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return make([]byte, length), nil
}

func (i *MockI2C) Write(data []byte) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.data = data
	return nil
}

func (i *MockI2C) ReadRegister(register byte, length int) ([]byte, error) {
	return make([]byte, length), nil
}

func (i *MockI2C) WriteRegister(register byte, data []byte) error {
	return nil
}

func (i *MockI2C) Close() error {
	return nil
}

// MockSPI SPI mock
type MockSPI struct {
	mu          sync.RWMutex
	speed       int
	mode        byte
	bitsPerWord byte
}

func (s *MockSPI) Open(bus, device int) error {
	return nil
}

func (s *MockSPI) Transfer(data []byte) ([]byte, error) {
	return data, nil
}

func (s *MockSPI) SetSpeed(speed int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.speed = speed
	return nil
}

func (s *MockSPI) SetMode(mode byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
	return nil
}

func (s *MockSPI) SetBitsPerWord(bits byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bitsPerWord = bits
	return nil
}

func (s *MockSPI) Close() error {
	return nil
}

// MockSerial Serial mock
type MockSerial struct {
	mu       sync.RWMutex
	port     string
	baudRate int
	dataBits int
	stopBits int
	parity   byte
}

func (s *MockSerial) Open(port string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.port = port
	return nil
}

func (s *MockSerial) SetBaudRate(baud int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.baudRate = baud
	return nil
}

func (s *MockSerial) SetDataBits(bits int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dataBits = bits
	return nil
}

func (s *MockSerial) SetStopBits(bits int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopBits = bits
	return nil
}

func (s *MockSerial) SetParity(parity byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.parity = parity
	return nil
}

func (s *MockSerial) Read(buffer []byte) (int, error) {
	return 0, nil
}

func (s *MockSerial) Write(data []byte) (int, error) {
	return len(data), nil
}

func (s *MockSerial) Close() error {
	return nil
}
