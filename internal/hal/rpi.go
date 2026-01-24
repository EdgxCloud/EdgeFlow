package hal

import (
	"fmt"
	"sync"

	"github.com/stianeikeland/go-rpio/v4"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

type RaspberryPiHAL struct {
	mu         sync.Mutex
	pins       map[int]rpio.Pin
	pwmPins    map[int]*PWMPin
	i2cBuses   map[string]i2c.BusCloser
	spiDevices map[string]spi.PortCloser
}

type PWMPin struct {
	pin       rpio.Pin
	frequency int
	dutyCycle int
}

func NewRaspberryPiHAL() (*RaspberryPiHAL, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize periph.io: %w", err)
	}

	return &RaspberryPiHAL{
		pins:       make(map[int]rpio.Pin),
		pwmPins:    make(map[int]*PWMPin),
		i2cBuses:   make(map[string]i2c.BusCloser),
		spiDevices: make(map[string]spi.PortCloser),
	}, nil
}

func (h *RaspberryPiHAL) InitGPIO() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := rpio.Open(); err != nil {
		return fmt.Errorf("failed to open GPIO: %w", err)
	}

	return nil
}

func (h *RaspberryPiHAL) SetPinMode(pin int, mode PinMode) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	p := rpio.Pin(pin)
	h.pins[pin] = p

	switch mode {
	case Input:
		p.Input()
	case Output:
		p.Output()
	case PWM:
		p.Output() // Set pin to output mode for software PWM
		h.pwmPins[pin] = &PWMPin{
			pin:       p,
			frequency: 1000,
			dutyCycle: 0,
		}
	default:
		return fmt.Errorf("unsupported pin mode: %v", mode)
	}

	return nil
}

func (h *RaspberryPiHAL) DigitalWrite(pin int, value bool) error {
	h.mu.Lock()
	p, ok := h.pins[pin]
	h.mu.Unlock()

	if !ok {
		return fmt.Errorf("pin %d not initialized", pin)
	}

	if value {
		p.High()
	} else {
		p.Low()
	}

	return nil
}

func (h *RaspberryPiHAL) DigitalRead(pin int) (bool, error) {
	h.mu.Lock()
	p, ok := h.pins[pin]
	h.mu.Unlock()

	if !ok {
		return false, fmt.Errorf("pin %d not initialized", pin)
	}

	return p.Read() == rpio.High, nil
}

func (h *RaspberryPiHAL) PWMWrite(pin int, dutyCycle int) error {
	h.mu.Lock()
	pwm, ok := h.pwmPins[pin]
	h.mu.Unlock()

	if !ok {
		return fmt.Errorf("pin %d not configured for PWM", pin)
	}

	pwm.dutyCycle = dutyCycle
	// go-rpio uses Write for duty cycle in PWM mode
	pwm.pin.Write(rpio.State(dutyCycle & 0xFF))

	return nil
}

func (h *RaspberryPiHAL) PWMSetFrequency(pin int, frequency int) error {
	h.mu.Lock()
	pwm, ok := h.pwmPins[pin]
	h.mu.Unlock()

	if !ok {
		return fmt.Errorf("pin %d not configured for PWM", pin)
	}

	pwm.frequency = frequency
	// Note: go-rpio v4 doesn't directly expose PWM frequency setting
	// This is a placeholder - actual implementation depends on hardware PWM
	return nil
}

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

func (h *RaspberryPiHAL) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, bus := range h.i2cBuses {
		bus.Close()
	}

	for _, dev := range h.spiDevices {
		dev.Close()
	}

	return rpio.Close()
}

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
