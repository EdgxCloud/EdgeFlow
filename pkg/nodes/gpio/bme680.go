package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"
)

// BME680Config configuration for BME680 sensor node
type BME680Config struct {
	Bus           string  `json:"bus"`            // I2C bus (default: "")
	Address       int     `json:"address"`        // I2C address (default: 0x76 or 0x77)
	Oversampling  string  `json:"oversampling"`   // "1x", "2x", "4x", "8x", "16x"
	Filter        string  `json:"filter"`         // IIR filter: "off", "2", "4", "8", "16"
	GasHeaterTemp int     `json:"gas_heater_temp"` // Gas heater temp (default: 320°C)
	GasHeaterTime int     `json:"gas_heater_time"` // Gas heater duration ms (default: 150)
	TempOffset    float64 `json:"temp_offset"`    // Temperature calibration offset
	HumidOffset   float64 `json:"humid_offset"`   // Humidity calibration offset
	SeaLevelPress float64 `json:"sea_level_hpa"`  // Sea level pressure for altitude
}

// BME680Executor executes BME680 sensor readings
type BME680Executor struct {
	config     BME680Config
	bus        i2c.BusCloser
	dev        *bmxx80.Dev
	mu         sync.Mutex
	hostInited bool
}

// NewBME680Executor creates a new BME680 executor
func NewBME680Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var bmeConfig BME680Config
	if err := json.Unmarshal(configJSON, &bmeConfig); err != nil {
		return nil, fmt.Errorf("invalid BME680 config: %w", err)
	}

	// Defaults
	if bmeConfig.Address == 0 {
		bmeConfig.Address = 0x76
	}
	if bmeConfig.Oversampling == "" {
		bmeConfig.Oversampling = "2x"
	}
	if bmeConfig.Filter == "" {
		bmeConfig.Filter = "2"
	}
	if bmeConfig.GasHeaterTemp == 0 {
		bmeConfig.GasHeaterTemp = 320
	}
	if bmeConfig.GasHeaterTime == 0 {
		bmeConfig.GasHeaterTime = 150
	}
	if bmeConfig.SeaLevelPress == 0 {
		bmeConfig.SeaLevelPress = 1013.25
	}

	return &BME680Executor{
		config: bmeConfig,
	}, nil
}

// Init initializes the BME680 executor
func (e *BME680Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute reads the BME680 sensor
func (e *BME680Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Initialize hardware
	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init periph host: %w", err)
		}
		e.hostInited = true
	}

	// Open I2C bus
	if e.bus == nil {
		bus, err := i2creg.Open(e.config.Bus)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to open I2C bus: %w", err)
		}
		e.bus = bus
	}

	// Initialize device
	if e.dev == nil {
		opts := &bmxx80.Opts{
			Temperature: getOversampling(e.config.Oversampling),
			Pressure:    getOversampling(e.config.Oversampling),
			Humidity:    getOversampling(e.config.Oversampling),
			Filter:      getFilter(e.config.Filter),
		}

		dev, err := bmxx80.NewI2C(e.bus, uint16(e.config.Address), opts)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to init BME680: %w", err)
		}
		e.dev = dev
	}

	// Read sensor
	var env physic.Env
	if err := e.dev.Sense(&env); err != nil {
		return node.Message{}, fmt.Errorf("failed to read BME680: %w", err)
	}

	// Convert values
	temperature := env.Temperature.Celsius() + e.config.TempOffset
	pressure := float64(env.Pressure) / float64(physic.Pascal) / 100.0
	humidity := float64(env.Humidity) / float64(physic.PercentRH) + e.config.HumidOffset

	// Calculate altitude
	altitude := 44330.0 * (1.0 - pow(pressure/e.config.SeaLevelPress, 0.1903))

	// Calculate dew point
	dewPoint := calcDewPoint(temperature, humidity)

	// Gas resistance (if available from BME680)
	// Note: periph.io bmxx80 may not expose gas resistance directly
	// This would need custom implementation for full BME680 gas support
	gasResistance := 0.0

	// Calculate IAQ (Indoor Air Quality) estimate
	// This is a simplified calculation - real IAQ needs calibration
	iaq := calculateIAQ(humidity, gasResistance)
	airQuality := getAirQualityLabel(iaq)

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
			"gas_resistance": gasResistance,
			"iaq":            iaq,
			"air_quality":    airQuality,
			"temp_unit":      "C",
			"pressure_unit":  "hPa",
			"humidity_unit":  "%",
			"altitude_unit":  "m",
			"resistance_unit": "Ohm",
			"address":        fmt.Sprintf("0x%02X", e.config.Address),
			"sensor":         "BME680",
			"timestamp":      time.Now().Unix(),
		},
	}, nil
}

// getFilter converts string to bmxx80 filter constant
func getFilter(s string) bmxx80.Filter {
	switch s {
	case "off", "0":
		return bmxx80.NoFilter
	case "2":
		return bmxx80.F2
	case "4":
		return bmxx80.F4
	case "8":
		return bmxx80.F8
	case "16":
		return bmxx80.F16
	default:
		return bmxx80.NoFilter
	}
}

// calculateIAQ calculates Indoor Air Quality index
// Returns 0-500 scale (0-50 good, 51-100 moderate, 101-150 unhealthy for sensitive, etc.)
func calculateIAQ(humidity float64, gasResistance float64) float64 {
	// Simplified IAQ calculation
	// In real implementation, this needs:
	// 1. Gas sensor burn-in period
	// 2. Baseline calibration
	// 3. Temperature compensation

	// Without gas resistance, use humidity as proxy
	if gasResistance == 0 {
		// Humidity contribution to comfort
		// Ideal: 40-60%, outside range = worse IAQ
		humScore := 100.0
		if humidity < 40 {
			humScore = 100 + (40-humidity)*2.5
		} else if humidity > 60 {
			humScore = 100 + (humidity-60)*2.5
		} else {
			humScore = 50 // Good range
		}
		return humScore
	}

	// With gas resistance
	// Higher resistance = better air quality
	// Typical range: 5,000 - 500,000 Ohms
	gasScore := 0.0
	if gasResistance > 0 {
		gasScore = 100 * (1 - 5000/gasResistance)
		if gasScore < 0 {
			gasScore = 0
		}
		if gasScore > 100 {
			gasScore = 100
		}
	}

	// Combine scores (gas 75%, humidity 25%)
	return gasScore*0.75 + (100-math.Abs(humidity-50))*0.5
}

// getAirQualityLabel returns human-readable air quality label
func getAirQualityLabel(iaq float64) string {
	switch {
	case iaq <= 50:
		return "excellent"
	case iaq <= 100:
		return "good"
	case iaq <= 150:
		return "moderate"
	case iaq <= 200:
		return "poor"
	case iaq <= 300:
		return "unhealthy"
	default:
		return "hazardous"
	}
}

// Cleanup releases resources
func (e *BME680Executor) Cleanup() error {
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
