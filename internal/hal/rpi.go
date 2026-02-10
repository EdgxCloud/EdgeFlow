package hal

import (
	"fmt"
	"sync"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

// RaspberryPiHAL implements the HAL interface for Raspberry Pi boards.
// GPIO uses the Linux character device interface (go-gpiocdev) which supports
// both Pi 4 (gpiochip0/BCM2711) and Pi 5 (gpiochip4/RP1).
// I2C and SPI use periph.io.
type RaspberryPiHAL struct {
	mu         sync.Mutex
	gpio       GPIOProvider
	i2c        *RpiI2CProvider
	spi        *RpiSPIProvider
	serial     *RpiSerialProvider
	boardInfo  BoardInfo
	i2cBuses   map[string]i2c.BusCloser
	spiDevices map[string]spi.PortCloser
}

// NewRaspberryPiHAL creates and initializes the HAL for Raspberry Pi.
// It detects the board model, selects the correct GPIO chip, and
// initializes periph.io for I2C/SPI.
func NewRaspberryPiHAL() (*RaspberryPiHAL, error) {
	// Initialize periph.io for I2C/SPI support
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize periph.io: %w", err)
	}

	// Detect board model and determine GPIO chip
	boardInfo := BoardInfo{
		Name:     "Unknown Board",
		NumGPIO:  26,
		GPIOChip: "gpiochip0",
	}
	if detected, err := DetectBoard(); err == nil {
		boardInfo = *detected
	}

	// Create GPIO provider using character device interface
	gpioProvider, err := NewGpiocdevGPIO(boardInfo.GPIOChip)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GPIO (%s): %w", boardInfo.GPIOChip, err)
	}

	h := &RaspberryPiHAL{
		gpio:       gpioProvider,
		boardInfo:  boardInfo,
		i2cBuses:   make(map[string]i2c.BusCloser),
		spiDevices: make(map[string]spi.PortCloser),
	}

	h.i2c = &RpiI2CProvider{hal: h}
	h.spi = &RpiSPIProvider{hal: h}
	h.serial = &RpiSerialProvider{}

	return h, nil
}

func (h *RaspberryPiHAL) GPIO() GPIOProvider   { return h.gpio }
func (h *RaspberryPiHAL) I2C() I2CProvider     { return h.i2c }
func (h *RaspberryPiHAL) SPI() SPIProvider     { return h.spi }
func (h *RaspberryPiHAL) Serial() SerialProvider { return h.serial }
func (h *RaspberryPiHAL) Info() BoardInfo       { return h.boardInfo }

func (h *RaspberryPiHAL) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.gpio != nil {
		h.gpio.Close()
	}

	for _, bus := range h.i2cBuses {
		bus.Close()
	}

	for _, dev := range h.spiDevices {
		dev.Close()
	}

	return nil
}

// I2COpen opens an I2C bus by name (e.g. "1" for /dev/i2c-1).
func (h *RaspberryPiHAL) I2COpen(bus string) (I2CBus, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if existingBus, ok := h.i2cBuses[bus]; ok {
		return &I2CBusWrapper{bus: existingBus}, nil
	}

	i2cBus, err := i2creg.Open(bus)
	if err != nil {
		return nil, fmt.Errorf("failed to open I2C bus %s: %w", bus, err)
	}

	h.i2cBuses[bus] = i2cBus
	return &I2CBusWrapper{bus: i2cBus}, nil
}

// SPIOpen opens an SPI device.
func (h *RaspberryPiHAL) SPIOpen(bus int, device int) (SPIDevice, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := fmt.Sprintf("%d-%d", bus, device)

	if existingDev, ok := h.spiDevices[key]; ok {
		conn, err := existingDev.Connect(physic.MegaHertz, spi.Mode0, 8)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SPI device: %w", err)
		}
		return &SPIDeviceWrapper{dev: conn}, nil
	}

	spiPort, err := spireg.Open(fmt.Sprintf("SPI%d.%d", bus, device))
	if err != nil {
		return nil, fmt.Errorf("failed to open SPI device: %w", err)
	}

	h.spiDevices[key] = spiPort
	conn, err := spiPort.Connect(physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		spiPort.Close()
		return nil, fmt.Errorf("failed to connect to SPI device: %w", err)
	}
	return &SPIDeviceWrapper{dev: conn}, nil
}

// --- I2C Bus Wrapper (periph.io) ---

type I2CBusWrapper struct {
	bus i2c.Bus
}

func (w *I2CBusWrapper) Write(addr uint16, data []byte) error {
	return w.bus.Tx(addr, data, nil)
}

func (w *I2CBusWrapper) Read(addr uint16, data []byte) error {
	return w.bus.Tx(addr, nil, data)
}

func (w *I2CBusWrapper) WriteRead(addr uint16, write []byte, read []byte) error {
	return w.bus.Tx(addr, write, read)
}

func (w *I2CBusWrapper) Close() error {
	return nil
}

// --- SPI Device Wrapper (periph.io) ---

type SPIDeviceWrapper struct {
	dev spi.Conn
}

func (w *SPIDeviceWrapper) Transfer(data []byte) ([]byte, error) {
	read := make([]byte, len(data))
	if err := w.dev.Tx(data, read); err != nil {
		return nil, err
	}
	return read, nil
}

func (w *SPIDeviceWrapper) Close() error {
	return nil
}

// --- I2C Provider (implements I2CProvider interface) ---

type RpiI2CProvider struct {
	hal     *RaspberryPiHAL
	address byte
	bus     I2CBus
}

func (p *RpiI2CProvider) Open(address byte) error {
	p.address = address
	bus, err := p.hal.I2COpen("1") // Default I2C bus
	if err != nil {
		return err
	}
	p.bus = bus
	return nil
}

func (p *RpiI2CProvider) Read(length int) ([]byte, error) {
	if p.bus == nil {
		return nil, fmt.Errorf("I2C bus not opened")
	}
	data := make([]byte, length)
	err := p.bus.Read(uint16(p.address), data)
	return data, err
}

func (p *RpiI2CProvider) Write(data []byte) error {
	if p.bus == nil {
		return fmt.Errorf("I2C bus not opened")
	}
	return p.bus.Write(uint16(p.address), data)
}

func (p *RpiI2CProvider) ReadRegister(register byte, length int) ([]byte, error) {
	if p.bus == nil {
		return nil, fmt.Errorf("I2C bus not opened")
	}
	read := make([]byte, length)
	err := p.bus.WriteRead(uint16(p.address), []byte{register}, read)
	return read, err
}

func (p *RpiI2CProvider) WriteRegister(register byte, data []byte) error {
	if p.bus == nil {
		return fmt.Errorf("I2C bus not opened")
	}
	writeData := append([]byte{register}, data...)
	return p.bus.Write(uint16(p.address), writeData)
}

func (p *RpiI2CProvider) Close() error {
	if p.bus != nil {
		return p.bus.Close()
	}
	return nil
}

// --- SPI Provider (implements SPIProvider interface) ---

type RpiSPIProvider struct {
	hal    *RaspberryPiHAL
	device SPIDevice
}

func (p *RpiSPIProvider) Open(bus, device int) error {
	dev, err := p.hal.SPIOpen(bus, device)
	if err != nil {
		return err
	}
	p.device = dev
	return nil
}

func (p *RpiSPIProvider) Transfer(data []byte) ([]byte, error) {
	if p.device == nil {
		return nil, fmt.Errorf("SPI device not opened")
	}
	return p.device.Transfer(data)
}

func (p *RpiSPIProvider) SetSpeed(speed int) error {
	return nil // Speed is set at connection time via periph.io
}

func (p *RpiSPIProvider) SetMode(mode byte) error {
	return nil // Mode is set at connection time via periph.io
}

func (p *RpiSPIProvider) SetBitsPerWord(bits byte) error {
	return nil // Bits per word is set at connection time via periph.io
}

func (p *RpiSPIProvider) Close() error {
	if p.device != nil {
		return p.device.Close()
	}
	return nil
}

// --- Serial Provider (basic stub - full serial uses go.bug.st/serial) ---

type RpiSerialProvider struct {
	port string
}

func (p *RpiSerialProvider) Open(port string) error {
	p.port = port
	return nil
}

func (p *RpiSerialProvider) SetBaudRate(baud int) error  { return nil }
func (p *RpiSerialProvider) SetDataBits(bits int) error  { return nil }
func (p *RpiSerialProvider) SetStopBits(bits int) error  { return nil }
func (p *RpiSerialProvider) SetParity(parity byte) error { return nil }

func (p *RpiSerialProvider) Read(buffer []byte) (int, error) {
	return 0, fmt.Errorf("serial: use go.bug.st/serial for full serial support")
}

func (p *RpiSerialProvider) Write(data []byte) (int, error) {
	return 0, fmt.Errorf("serial: use go.bug.st/serial for full serial support")
}

func (p *RpiSerialProvider) Close() error { return nil }
