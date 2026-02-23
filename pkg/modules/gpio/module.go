// Package gpio provides the GPIO module for EdgeFlow
// This module provides GPIO, I2C, SPI, PWM, and sensor nodes for hardware interaction
package gpio

import (
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/EdgxCloud/EdgeFlow/internal/plugin"
	"github.com/EdgxCloud/EdgeFlow/pkg/nodes/gpio"
)

// GPIOModule is the GPIO module that provides hardware interaction nodes
type GPIOModule struct {
	*plugin.BasePlugin
	loaded bool
}

// wrapFactory wraps a config-based factory into a NodeFactory
// The executor's Init method will be called later with the actual config
func wrapFactory(fn func(config map[string]interface{}) (node.Executor, error)) node.NodeFactory {
	return func() node.Executor {
		// Create with empty config - Init will be called later with real config
		exec, err := fn(make(map[string]interface{}))
		if err != nil {
			return nil
		}
		return exec
	}
}

// NewGPIOModule creates a new GPIO module
func NewGPIOModule() *GPIOModule {
	metadata := plugin.Metadata{
		Name:        "gpio",
		Version:     "1.0.0",
		Description: "Hardware interaction nodes - GPIO, I2C, SPI, PWM, sensors",
		Author:      "EdgeFlow Team",
		Category:    plugin.CategoryGPIO,
		License:     "MIT",
		Keywords:    []string{"gpio", "i2c", "spi", "pwm", "sensor", "raspberry-pi", "hardware"},
		MinEdgeFlow: "0.1.0",
		Config: map[string]interface{}{
			"default_debounce": 50,
			"pwm_frequency":    1000,
			"i2c_bus":          1,
			"spi_bus":          0,
		},
	}

	return &GPIOModule{
		BasePlugin: plugin.NewBasePlugin(metadata),
		loaded:     false,
	}
}

// Load loads the GPIO module
func (m *GPIOModule) Load() error {
	m.SetStatus(plugin.StatusLoading)

	registry := node.GetGlobalRegistry()

	// Register all GPIO nodes with their definitions
	nodeInfos := []*node.NodeInfo{
		{
			Type:        "gpio-in",
			Name:        "GPIO In",
			Category:    node.NodeTypeInput,
			Description: "Read from GPIO pin with edge detection",
			Icon:        "arrow-down-circle",
			Color:       "#E91E63",
			Factory:     wrapFactory(gpio.NewGPIOInExecutor),
		},
		{
			Type:        "gpio-out",
			Name:        "GPIO Out",
			Category:    node.NodeTypeOutput,
			Description: "Write to GPIO pin",
			Icon:        "arrow-up-circle",
			Color:       "#E91E63",
			Factory:     wrapFactory(gpio.NewGPIOOutExecutor),
		},
		{
			Type:        "pwm",
			Name:        "PWM",
			Category:    node.NodeTypeOutput,
			Description: "PWM output with frequency and duty cycle control",
			Icon:        "activity",
			Color:       "#FF5722",
			Factory:     wrapFactory(gpio.NewPWMExecutor),
		},
		{
			Type:        "servo",
			Name:        "Servo",
			Category:    node.NodeTypeOutput,
			Description: "Servo motor control (0-180 degrees)",
			Icon:        "rotate-cw",
			Color:       "#FF5722",
			Factory:     wrapFactory(gpio.NewServoExecutor),
		},
		{
			Type:        "i2c",
			Name:        "I2C",
			Category:    node.NodeTypeFunction,
			Description: "I2C bus communication",
			Icon:        "cpu",
			Color:       "#3F51B5",
			Factory:     wrapFactory(gpio.NewI2CExecutor),
		},
		{
			Type:        "spi",
			Name:        "SPI",
			Category:    node.NodeTypeFunction,
			Description: "SPI bus communication",
			Icon:        "cpu",
			Color:       "#3F51B5",
			Factory:     wrapFactory(gpio.NewSPIExecutor),
		},
		{
			Type:        "serial",
			Name:        "Serial",
			Category:    node.NodeTypeFunction,
			Description: "Serial/UART communication",
			Icon:        "usb",
			Color:       "#3F51B5",
			Factory:     wrapFactory(gpio.NewSerialExecutor),
		},
		{
			Type:        "dht",
			Name:        "DHT Sensor",
			Category:    node.NodeTypeInput,
			Description: "DHT11/DHT22 temperature and humidity sensor",
			Icon:        "thermometer",
			Color:       "#00BCD4",
			Factory:     wrapFactory(gpio.NewDHTExecutor),
		},
		{
			Type:        "ds18b20",
			Name:        "DS18B20",
			Category:    node.NodeTypeInput,
			Description: "DS18B20 1-Wire temperature sensor",
			Icon:        "thermometer",
			Color:       "#00BCD4",
			Factory:     wrapFactory(gpio.NewDS18B20Executor),
		},
		{
			Type:        "bmp280",
			Name:        "BMP280",
			Category:    node.NodeTypeInput,
			Description: "BMP280 pressure and temperature sensor (I2C)",
			Icon:        "gauge",
			Color:       "#00BCD4",
			Factory:     wrapFactory(gpio.NewBMP280Executor),
		},
		{
			Type:        "sht3x",
			Name:        "SHT3x",
			Category:    node.NodeTypeInput,
			Description: "SHT3x temperature and humidity sensor (I2C)",
			Icon:        "droplet",
			Color:       "#00BCD4",
			Factory:     wrapFactory(gpio.NewSHT3xExecutor),
		},
		{
			Type:        "aht20",
			Name:        "AHT20",
			Category:    node.NodeTypeInput,
			Description: "AHT20 temperature and humidity sensor (I2C)",
			Icon:        "droplet",
			Color:       "#00BCD4",
			Factory:     wrapFactory(gpio.NewAHT20Executor),
		},
		{
			Type:        "hcsr04",
			Name:        "HC-SR04",
			Category:    node.NodeTypeInput,
			Description: "HC-SR04 ultrasonic distance sensor",
			Icon:        "radar",
			Color:       "#00BCD4",
			Factory:     wrapFactory(gpio.NewHCSR04Executor),
		},
		{
			Type:        "relay",
			Name:        "Relay",
			Category:    node.NodeTypeOutput,
			Description: "Relay control (on/off/toggle/pulse)",
			Icon:        "zap",
			Color:       "#FFC107",
			Factory:     wrapFactory(gpio.NewRelayExecutor),
		},
		{
			Type:        "voltage-monitor",
			Name:        "Voltage Monitor",
			Category:    node.NodeTypeInput,
			Description: "Monitor voltage via ADC",
			Icon:        "battery",
			Color:       "#8BC34A",
			Factory:     wrapFactory(gpio.NewVoltageMonitorExecutor),
		},
		{
			Type:        "current-monitor",
			Name:        "Current Monitor",
			Category:    node.NodeTypeInput,
			Description: "Monitor current via shunt resistor",
			Icon:        "zap",
			Color:       "#8BC34A",
			Factory:     wrapFactory(gpio.NewCurrentMonitorExecutor),
		},
	}

	for _, info := range nodeInfos {
		registry.Register(info)
	}

	m.loaded = true
	m.SetStatus(plugin.StatusLoaded)
	return nil
}

// Unload unloads the GPIO module
func (m *GPIOModule) Unload() error {
	m.SetStatus(plugin.StatusUnloading)
	m.loaded = false
	m.SetStatus(plugin.StatusNotLoaded)
	return nil
}

// IsLoaded returns whether the module is loaded
func (m *GPIOModule) IsLoaded() bool {
	return m.loaded
}

// Nodes returns the node definitions provided by this module
func (m *GPIOModule) Nodes() []plugin.NodeDefinition {
	return []plugin.NodeDefinition{
		// GPIO Nodes
		{
			Type:        "gpio-in",
			Name:        "GPIO In",
			Category:    "gpio",
			Description: "Read from GPIO pin with edge detection",
			Icon:        "arrow-down-circle",
			Color:       "#E91E63",
			Inputs:      0,
			Outputs:     1,
		},
		{
			Type:        "gpio-out",
			Name:        "GPIO Out",
			Category:    "gpio",
			Description: "Write to GPIO pin",
			Icon:        "arrow-up-circle",
			Color:       "#E91E63",
			Inputs:      1,
			Outputs:     0,
		},
		// PWM Nodes
		{
			Type:        "pwm",
			Name:        "PWM",
			Category:    "gpio",
			Description: "PWM output with frequency and duty cycle control",
			Icon:        "activity",
			Color:       "#FF5722",
			Inputs:      1,
			Outputs:     0,
		},
		{
			Type:        "servo",
			Name:        "Servo",
			Category:    "gpio",
			Description: "Servo motor control (0-180 degrees)",
			Icon:        "rotate-cw",
			Color:       "#FF5722",
			Inputs:      1,
			Outputs:     0,
		},
		// Communication Nodes
		{
			Type:        "i2c",
			Name:        "I2C",
			Category:    "gpio",
			Description: "I2C bus communication",
			Icon:        "cpu",
			Color:       "#3F51B5",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "spi",
			Name:        "SPI",
			Category:    "gpio",
			Description: "SPI bus communication",
			Icon:        "cpu",
			Color:       "#3F51B5",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "serial",
			Name:        "Serial",
			Category:    "gpio",
			Description: "Serial/UART communication",
			Icon:        "usb",
			Color:       "#3F51B5",
			Inputs:      1,
			Outputs:     1,
		},
		// Sensor Nodes
		{
			Type:        "dht",
			Name:        "DHT Sensor",
			Category:    "gpio",
			Description: "DHT11/DHT22 temperature and humidity sensor",
			Icon:        "thermometer",
			Color:       "#00BCD4",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "ds18b20",
			Name:        "DS18B20",
			Category:    "gpio",
			Description: "DS18B20 1-Wire temperature sensor",
			Icon:        "thermometer",
			Color:       "#00BCD4",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "bmp280",
			Name:        "BMP280",
			Category:    "gpio",
			Description: "BMP280 pressure and temperature sensor (I2C)",
			Icon:        "gauge",
			Color:       "#00BCD4",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "sht3x",
			Name:        "SHT3x",
			Category:    "gpio",
			Description: "SHT3x temperature and humidity sensor (I2C)",
			Icon:        "droplet",
			Color:       "#00BCD4",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "aht20",
			Name:        "AHT20",
			Category:    "gpio",
			Description: "AHT20 temperature and humidity sensor (I2C)",
			Icon:        "droplet",
			Color:       "#00BCD4",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "hcsr04",
			Name:        "HC-SR04",
			Category:    "gpio",
			Description: "HC-SR04 ultrasonic distance sensor",
			Icon:        "radar",
			Color:       "#00BCD4",
			Inputs:      1,
			Outputs:     1,
		},
		// Actuator Nodes
		{
			Type:        "relay",
			Name:        "Relay",
			Category:    "gpio",
			Description: "Relay control (on/off/toggle/pulse)",
			Icon:        "zap",
			Color:       "#FFC107",
			Inputs:      1,
			Outputs:     1,
		},
		// Monitor Nodes
		{
			Type:        "voltage-monitor",
			Name:        "Voltage Monitor",
			Category:    "gpio",
			Description: "Monitor voltage via ADC",
			Icon:        "battery",
			Color:       "#8BC34A",
			Inputs:      1,
			Outputs:     1,
		},
		{
			Type:        "current-monitor",
			Name:        "Current Monitor",
			Category:    "gpio",
			Description: "Monitor current via shunt resistor",
			Icon:        "zap",
			Color:       "#8BC34A",
			Inputs:      1,
			Outputs:     1,
		},
	}
}

// RequiredMemory returns the memory requirement in bytes
func (m *GPIOModule) RequiredMemory() uint64 {
	return 30 * 1024 * 1024 // 30 MB
}

// RequiredDisk returns the disk requirement in bytes
func (m *GPIOModule) RequiredDisk() uint64 {
	return 8 * 1024 * 1024 // 8 MB
}

// Dependencies returns the list of required plugins
func (m *GPIOModule) Dependencies() []string {
	return []string{"core"}
}

// init registers the GPIO module with the global registry
func init() {
	registry := plugin.GetRegistry()
	module := NewGPIOModule()
	registry.Register(module)
}
