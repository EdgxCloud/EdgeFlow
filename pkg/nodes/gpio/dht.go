// +build linux

package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/node"
)

// DHTConfig configuration for DHT11/DHT22 sensor node
type DHTConfig struct {
	Pin         int     `json:"pin"`          // GPIO pin number (BCM)
	Type        string  `json:"type"`         // "dht11" or "dht22"
	Retries     int     `json:"retries"`      // Number of read retries (default: 3)
	RetryDelay  int     `json:"retry_delay"`  // Delay between retries in ms (default: 2000)
	TempOffset  float64 `json:"temp_offset"`  // Temperature calibration offset
	HumidOffset float64 `json:"humid_offset"` // Humidity calibration offset
}

// DHTExecutor executes DHT sensor readings
type DHTExecutor struct {
	config   DHTConfig
	hal      hal.HAL
	lastRead time.Time
	lastTemp float64
	lastHum  float64
	mu       sync.Mutex
}

// NewDHTExecutor creates a new DHT executor
func NewDHTExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var dhtConfig DHTConfig
	if err := json.Unmarshal(configJSON, &dhtConfig); err != nil {
		return nil, fmt.Errorf("invalid dht config: %w", err)
	}

	// Validate pin
	if dhtConfig.Pin < 0 || dhtConfig.Pin > 27 {
		return nil, fmt.Errorf("invalid GPIO pin number (must be 0-27)")
	}

	// Default type to DHT22
	if dhtConfig.Type == "" {
		dhtConfig.Type = "dht22"
	}
	dhtConfig.Type = strings.ToLower(dhtConfig.Type)

	if dhtConfig.Type != "dht11" && dhtConfig.Type != "dht22" && dhtConfig.Type != "am2302" {
		return nil, fmt.Errorf("invalid DHT type: %s (must be dht11, dht22, or am2302)", dhtConfig.Type)
	}

	// AM2302 is same as DHT22
	if dhtConfig.Type == "am2302" {
		dhtConfig.Type = "dht22"
	}

	// Default retries
	if dhtConfig.Retries <= 0 {
		dhtConfig.Retries = 3
	}

	// Default retry delay (minimum 2 seconds for DHT sensors)
	if dhtConfig.RetryDelay <= 0 {
		dhtConfig.RetryDelay = 2000
	}

	return &DHTExecutor{
		config: dhtConfig,
	}, nil
}

// Init initializes the DHT executor
func (e *DHTExecutor) Init(config map[string]interface{}) error {
	return nil
}

// Execute reads the DHT sensor
func (e *DHTExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get HAL if not initialized
	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h
	}

	// Check if we need to wait before reading again (DHT sensors need 2s between reads)
	minInterval := 2 * time.Second
	if e.config.Type == "dht11" {
		minInterval = 1 * time.Second
	}

	if time.Since(e.lastRead) < minInterval && e.lastRead.Unix() > 0 {
		// Return cached values
		return node.Message{
			Payload: map[string]interface{}{
				"temperature": e.lastTemp,
				"humidity":    e.lastHum,
				"unit":        "C",
				"type":        e.config.Type,
				"pin":         e.config.Pin,
				"cached":      true,
				"timestamp":   time.Now().Unix(),
			},
		}, nil
	}

	// Read DHT sensor with retries
	var temperature, humidity float64
	var err error

	for i := 0; i < e.config.Retries; i++ {
		temperature, humidity, err = e.readDHT()
		if err == nil {
			break
		}

		if i < e.config.Retries-1 {
			select {
			case <-ctx.Done():
				return node.Message{}, ctx.Err()
			case <-time.After(time.Duration(e.config.RetryDelay) * time.Millisecond):
			}
		}
	}

	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read DHT after %d retries: %w", e.config.Retries, err)
	}

	// Apply calibration offsets
	temperature += e.config.TempOffset
	humidity += e.config.HumidOffset

	// Clamp humidity to valid range
	if humidity < 0 {
		humidity = 0
	}
	if humidity > 100 {
		humidity = 100
	}

	// Cache the values
	e.lastRead = time.Now()
	e.lastTemp = temperature
	e.lastHum = humidity

	return node.Message{
		Payload: map[string]interface{}{
			"temperature": temperature,
			"humidity":    humidity,
			"unit":        "C",
			"type":        e.config.Type,
			"pin":         e.config.Pin,
			"cached":      false,
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

// readDHT reads temperature and humidity from DHT sensor
func (e *DHTExecutor) readDHT() (float64, float64, error) {
	// Check for mock HAL first
	if _, ok := e.hal.(*hal.MockHAL); ok {
		return 23.5 + e.config.TempOffset, 55.0 + e.config.HumidOffset, nil
	}

	// Try to use kernel driver if available
	temp, hum, err := e.readFromKernelDriver()
	if err == nil {
		return temp, hum, nil
	}

	// Fall back to bit-banging
	return e.readBitBang()
}

// readFromKernelDriver attempts to read from dht11 kernel module
// The kernel driver provides more accurate timing than userspace
func (e *DHTExecutor) readFromKernelDriver() (float64, float64, error) {
	// Check if iio device exists for this pin
	// The dht11 kernel driver creates /sys/bus/iio/devices/iio:deviceX
	iioDevices, err := os.ReadDir("/sys/bus/iio/devices")
	if err != nil {
		return 0, 0, fmt.Errorf("no IIO devices found: %w", err)
	}

	for _, dev := range iioDevices {
		devPath := fmt.Sprintf("/sys/bus/iio/devices/%s", dev.Name())

		// Check if this is a DHT device
		nameFile := devPath + "/name"
		nameBytes, err := os.ReadFile(nameFile)
		if err != nil {
			continue
		}

		name := strings.TrimSpace(string(nameBytes))
		if name != "dht11" && name != "dht22" {
			continue
		}

		// Read temperature
		tempRaw, err := os.ReadFile(devPath + "/in_temp_input")
		if err != nil {
			continue
		}
		tempVal, err := strconv.ParseFloat(strings.TrimSpace(string(tempRaw)), 64)
		if err != nil {
			continue
		}
		// Kernel driver returns millidegrees
		temperature := tempVal / 1000.0

		// Read humidity
		humRaw, err := os.ReadFile(devPath + "/in_humidityrelative_input")
		if err != nil {
			continue
		}
		humVal, err := strconv.ParseFloat(strings.TrimSpace(string(humRaw)), 64)
		if err != nil {
			continue
		}
		// Kernel driver returns milli-percent
		humidity := humVal / 1000.0

		return temperature, humidity, nil
	}

	return 0, 0, fmt.Errorf("no DHT kernel driver found")
}

// readBitBang reads DHT sensor using bit-banging via GPIO
// This requires precise timing and may have reliability issues
func (e *DHTExecutor) readBitBang() (float64, float64, error) {
	gpio := e.hal.GPIO()
	pin := e.config.Pin

	// DHT Protocol:
	// 1. Host sends start signal: LOW for at least 18ms
	// 2. Host pulls HIGH for 20-40us
	// 3. DHT responds: LOW for 80us, then HIGH for 80us
	// 4. DHT sends 40 bits of data
	//    - Each bit starts with 50us LOW
	//    - Then HIGH: 26-28us = 0, 70us = 1
	// 5. Data format: 2 bytes humidity, 2 bytes temperature, 1 byte checksum

	// Send start signal
	gpio.SetMode(pin, hal.Output)
	gpio.DigitalWrite(pin, false)

	// DHT22 needs 1-10ms, DHT11 needs 18ms minimum
	if e.config.Type == "dht11" {
		time.Sleep(20 * time.Millisecond)
	} else {
		time.Sleep(1 * time.Millisecond)
	}

	// Pull high briefly
	gpio.DigitalWrite(pin, true)
	time.Sleep(30 * time.Microsecond)

	// Switch to input mode
	gpio.SetMode(pin, hal.Input)
	gpio.SetPull(pin, hal.PullUp)

	// Read the response
	data, err := e.readBits(gpio, pin)
	if err != nil {
		return 0, 0, err
	}

	// Parse the data
	return e.parseData(data)
}

// readBits reads 40 bits from the DHT sensor
func (e *DHTExecutor) readBits(gpio hal.GPIOProvider, pin int) ([]byte, error) {
	// Wait for DHT to pull low (response signal)
	timeout := time.Now().Add(100 * time.Millisecond)
	for gpio.DigitalRead(pin) {
		if time.Now().After(timeout) {
			return nil, fmt.Errorf("timeout waiting for DHT response (initial low)")
		}
	}

	// Wait for DHT to pull high
	timeout = time.Now().Add(100 * time.Millisecond)
	for !gpio.DigitalRead(pin) {
		if time.Now().After(timeout) {
			return nil, fmt.Errorf("timeout waiting for DHT response (initial high)")
		}
	}

	// Wait for DHT to pull low again (start of data)
	timeout = time.Now().Add(100 * time.Millisecond)
	for gpio.DigitalRead(pin) {
		if time.Now().After(timeout) {
			return nil, fmt.Errorf("timeout waiting for data start")
		}
	}

	// Read 40 bits
	data := make([]byte, 5)
	for i := 0; i < 40; i++ {
		// Wait for high (end of 50us low pulse)
		timeout = time.Now().Add(100 * time.Millisecond)
		for !gpio.DigitalRead(pin) {
			if time.Now().After(timeout) {
				return nil, fmt.Errorf("timeout at bit %d (waiting for high)", i)
			}
		}

		// Measure high pulse duration
		start := time.Now()
		timeout = time.Now().Add(100 * time.Millisecond)
		for gpio.DigitalRead(pin) {
			if time.Now().After(timeout) {
				return nil, fmt.Errorf("timeout at bit %d (waiting for low)", i)
			}
		}
		duration := time.Since(start)

		// If high pulse > 40us, it's a 1
		byteIdx := i / 8
		bitIdx := uint(7 - (i % 8))
		if duration > 40*time.Microsecond {
			data[byteIdx] |= 1 << bitIdx
		}
	}

	return data, nil
}

// parseData parses the 40-bit data from DHT sensor
func (e *DHTExecutor) parseData(data []byte) (float64, float64, error) {
	if len(data) != 5 {
		return 0, 0, fmt.Errorf("invalid data length: %d", len(data))
	}

	// Verify checksum
	checksum := data[0] + data[1] + data[2] + data[3]
	if checksum != data[4] {
		return 0, 0, fmt.Errorf("checksum mismatch: expected %d, got %d", checksum, data[4])
	}

	var temperature, humidity float64

	if e.config.Type == "dht11" {
		// DHT11: integer values only
		// Byte 0: Humidity integer
		// Byte 1: Humidity decimal (usually 0)
		// Byte 2: Temperature integer
		// Byte 3: Temperature decimal (usually 0)
		humidity = float64(data[0]) + float64(data[1])/10.0
		temperature = float64(data[2]) + float64(data[3])/10.0
	} else {
		// DHT22/AM2302: 16-bit values with 1 decimal place
		// Bytes 0-1: Humidity (16-bit, divide by 10)
		// Bytes 2-3: Temperature (16-bit, divide by 10, MSB is sign)
		humidity = float64(uint16(data[0])<<8|uint16(data[1])) / 10.0

		tempRaw := uint16(data[2]&0x7F)<<8 | uint16(data[3])
		temperature = float64(tempRaw) / 10.0
		if data[2]&0x80 != 0 {
			temperature = -temperature
		}
	}

	// Validate ranges
	if humidity < 0 || humidity > 100 {
		return 0, 0, fmt.Errorf("humidity out of range: %.1f", humidity)
	}

	if e.config.Type == "dht11" {
		if temperature < 0 || temperature > 50 {
			return 0, 0, fmt.Errorf("DHT11 temperature out of range: %.1f", temperature)
		}
	} else {
		if temperature < -40 || temperature > 80 {
			return 0, 0, fmt.Errorf("DHT22 temperature out of range: %.1f", temperature)
		}
	}

	return temperature, humidity, nil
}

// Cleanup releases resources
func (e *DHTExecutor) Cleanup() error {
	return nil
}
