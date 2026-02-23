package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/hal"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"
)

// BME280Config configuration for BME280 sensor node
type BME280Config struct {
	Bus           string  `json:"bus"`            // I2C bus (default: "")
	Address       int     `json:"address"`        // I2C address (default: 0x76)
	Oversampling  string  `json:"oversampling"`   // Oversampling: "1x", "2x", "4x", "8x", "16x" (default: "1x")
	TempOffset    float64 `json:"temp_offset"`    // Temperature calibration offset
	PressOffset   float64 `json:"press_offset"`   // Pressure calibration offset
	HumidOffset   float64 `json:"humid_offset"`   // Humidity calibration offset
	SeaLevelPress float64 `json:"sea_level_hpa"`  // Sea level pressure for altitude calc (default: 1013.25)
}

// BME280Executor executes BME280 sensor readings
type BME280Executor struct {
	config     BME280Config
	hal        hal.HAL
	dev        *bmxx80.Dev
	bus        i2c.BusCloser
	lastRead   time.Time
	lastData   map[string]interface{}
	mu         sync.Mutex
	hostInited bool
}

// NewBME280Executor creates a new BME280 executor
func NewBME280Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var bmeConfig BME280Config
	if err := json.Unmarshal(configJSON, &bmeConfig); err != nil {
		return nil, fmt.Errorf("invalid BME280 config: %w", err)
	}

	// Default address is 0x76 (SDO to GND) or 0x77 (SDO to VCC)
	if bmeConfig.Address == 0 {
		bmeConfig.Address = 0x76
	}

	// Validate address
	if bmeConfig.Address != 0x76 && bmeConfig.Address != 0x77 {
		return nil, fmt.Errorf("invalid I2C address: 0x%02X (must be 0x76 or 0x77)", bmeConfig.Address)
	}

	// Default sea level pressure
	if bmeConfig.SeaLevelPress == 0 {
		bmeConfig.SeaLevelPress = 1013.25
	}

	// Default oversampling
	if bmeConfig.Oversampling == "" {
		bmeConfig.Oversampling = "1x"
	}

	return &BME280Executor{
		config: bmeConfig,
	}, nil
}

// Init initializes the BME280 executor
func (e *BME280Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute reads the BME280 sensor
func (e *BME280Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Initialize hardware if needed
	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init periph host: %w", err)
		}
		e.hostInited = true
	}

	// Open I2C bus if not already open
	if e.bus == nil {
		busName := e.config.Bus
		if busName == "" {
			busName = "" // Let periph.io find the default bus
		}

		bus, err := i2creg.Open(busName)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to open I2C bus: %w", err)
		}
		e.bus = bus
	}

	// Initialize device if not already done
	if e.dev == nil {
		opts := &bmxx80.Opts{
			Temperature: getOversampling(e.config.Oversampling),
			Pressure:    getOversampling(e.config.Oversampling),
			Humidity:    getOversampling(e.config.Oversampling),
		}

		dev, err := bmxx80.NewI2C(e.bus, uint16(e.config.Address), opts)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to init BME280: %w", err)
		}
		e.dev = dev
	}

	// Read sensor
	var env physic.Env
	if err := e.dev.Sense(&env); err != nil {
		return node.Message{}, fmt.Errorf("failed to read BME280: %w", err)
	}

	// Convert values
	temperature := env.Temperature.Celsius() + e.config.TempOffset
	pressure := float64(env.Pressure)/float64(physic.Pascal)/100.0 + e.config.PressOffset // Convert to hPa
	humidity := float64(env.Humidity)/float64(physic.PercentRH) + e.config.HumidOffset

	// Calculate altitude using barometric formula
	// altitude = 44330 * (1 - (pressure / seaLevelPressure) ^ 0.1903)
	altitude := 44330.0 * (1.0 - pow(pressure/e.config.SeaLevelPress, 0.1903))

	// Calculate dew point
	// Using Magnus formula approximation
	dewPoint := calcDewPoint(temperature, humidity)

	// Clamp humidity
	if humidity < 0 {
		humidity = 0
	}
	if humidity > 100 {
		humidity = 100
	}

	return node.Message{
		Payload: map[string]interface{}{
			"temperature":    temperature,
			"pressure":       pressure,
			"humidity":       humidity,
			"altitude":       altitude,
			"dew_point":      dewPoint,
			"temp_unit":      "C",
			"pressure_unit":  "hPa",
			"humidity_unit":  "%",
			"altitude_unit":  "m",
			"address":        fmt.Sprintf("0x%02X", e.config.Address),
			"sensor":         "BME280",
			"timestamp":      time.Now().Unix(),
		},
	}, nil
}

// getOversampling converts string to bmxx80 oversampling constant
func getOversampling(s string) bmxx80.Oversampling {
	switch s {
	case "1x":
		return bmxx80.O1x
	case "2x":
		return bmxx80.O2x
	case "4x":
		return bmxx80.O4x
	case "8x":
		return bmxx80.O8x
	case "16x":
		return bmxx80.O16x
	default:
		return bmxx80.O1x
	}
}

// pow calculates x^y for positive x
func pow(x, y float64) float64 {
	if x <= 0 {
		return 0
	}
	// Use exp(y * ln(x))
	return exp(y * ln(x))
}

// Simple exp approximation
func exp(x float64) float64 {
	// Use Taylor series for small values, otherwise use repeated squaring
	if x < 0 {
		return 1.0 / exp(-x)
	}
	if x > 20 {
		return exp(x/2) * exp(x/2)
	}

	result := 1.0
	term := 1.0
	for i := 1; i < 50; i++ {
		term *= x / float64(i)
		result += term
		if term < 1e-15 {
			break
		}
	}
	return result
}

// Simple ln approximation
func ln(x float64) float64 {
	if x <= 0 {
		return -999999
	}
	if x == 1 {
		return 0
	}

	// Reduce to range [0.5, 2]
	k := 0
	for x > 2 {
		x /= 2
		k++
	}
	for x < 0.5 {
		x *= 2
		k--
	}

	// Taylor series for ln((1+y)/(1-y)) where y = (x-1)/(x+1)
	y := (x - 1) / (x + 1)
	y2 := y * y
	result := 0.0
	term := y
	for i := 1; i < 100; i += 2 {
		result += term / float64(i)
		term *= y2
		if term < 1e-15 {
			break
		}
	}
	return 2*result + float64(k)*0.6931471805599453 // ln(2)
}

// calcDewPoint calculates dew point using Magnus formula
func calcDewPoint(tempC, humidity float64) float64 {
	if humidity <= 0 {
		return tempC
	}

	// Magnus formula constants
	const a = 17.27
	const b = 237.7

	// Calculate gamma
	gamma := (a*tempC)/(b+tempC) + ln(humidity/100.0)

	// Calculate dew point
	dewPoint := (b * gamma) / (a - gamma)

	return dewPoint
}

// Cleanup releases resources
func (e *BME280Executor) Cleanup() error {
	if e.dev != nil {
		e.dev.Halt()
		e.dev = nil
	}
	if e.bus != nil {
		e.bus.Close()
		e.bus = nil
	}
	return nil
}
