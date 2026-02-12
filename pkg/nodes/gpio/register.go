//go:build linux
// +build linux

package gpio

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterAllNodes registers all GPIO/hardware nodes with the registry using NodeInfo and PropertySchema
func RegisterAllNodes(registry *node.Registry) error {
	// ============================================
	// GPIO CORE NODES (6 nodes)
	// ============================================

	// GPIO Input
	if err := registry.Register(&node.NodeInfo{
		Type:        "gpio-in",
		Name:        "GPIO Input",
		Category:    node.NodeTypeInput,
		Description: "Read digital value from GPIO pin",
		Icon:        "arrow-down-circle",
		Color:       "#22c55e",
		Properties: []node.PropertySchema{
			{Name: "pin", Label: "GPIO Pin", Type: "number", Default: 0, Required: true, Description: "BCM GPIO pin number (0-27)"},
			{Name: "pullMode", Label: "Pull Mode", Type: "select", Default: "none", Options: []string{"none", "up", "down"}, Description: "Internal pull resistor"},
			{Name: "edgeMode", Label: "Edge Detection", Type: "select", Default: "both", Options: []string{"rising", "falling", "both"}, Description: "Edge trigger mode"},
			{Name: "debounce", Label: "Debounce (ms)", Type: "number", Default: 50, Description: "Debounce time in milliseconds"},
			{Name: "pollInterval", Label: "Poll Interval (ms)", Type: "number", Default: 100, Description: "Polling interval for non-interrupt mode"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "GPIO state output"},
		},
		Factory: func() node.Executor { return &GPIOInExecutor{} },
	}); err != nil {
		return err
	}

	// GPIO Output
	if err := registry.Register(&node.NodeInfo{
		Type:        "gpio-out",
		Name:        "GPIO Output",
		Category:    node.NodeTypeOutput,
		Description: "Write digital value to GPIO pin",
		Icon:        "arrow-up-circle",
		Color:       "#ef4444",
		Properties: []node.PropertySchema{
			{Name: "pin", Label: "GPIO Pin", Type: "number", Default: 0, Required: true, Description: "BCM GPIO pin number"},
			{Name: "initialValue", Label: "Initial Value", Type: "boolean", Default: false, Description: "Initial output state"},
			{Name: "invert", Label: "Invert Output", Type: "boolean", Default: false, Description: "Invert output logic (HIGH=0)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Value to write"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Confirmation output"},
		},
		Factory: func() node.Executor { return &GPIOOutExecutor{} },
	}); err != nil {
		return err
	}

	// PWM
	if err := registry.Register(&node.NodeInfo{
		Type:        "pwm",
		Name:        "PWM",
		Category:    node.NodeTypeOutput,
		Description: "Generate PWM signal",
		Icon:        "activity",
		Color:       "#f59e0b",
		Properties: []node.PropertySchema{
			{Name: "pin", Label: "PWM Pin", Type: "number", Default: 12, Required: true, Description: "Hardware PWM pin (12, 13, 18, 19)"},
			{Name: "frequency", Label: "Frequency (Hz)", Type: "number", Default: 1000, Description: "PWM frequency 1-125000000 Hz"},
			{Name: "dutyCycle", Label: "Duty Cycle", Type: "number", Default: 128, Description: "Initial duty cycle 0-255"},
			{Name: "polarity", Label: "Polarity", Type: "select", Default: "normal", Options: []string{"normal", "inverted"}, Description: "PWM polarity"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Duty cycle value"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Confirmation output"},
		},
		Factory: func() node.Executor { return &PWMExecutor{} },
	}); err != nil {
		return err
	}

	// I2C
	if err := registry.Register(&node.NodeInfo{
		Type:        "i2c",
		Name:        "I2C",
		Category:    node.NodeTypeProcessing,
		Description: "Read/write data via I2C bus",
		Icon:        "cpu",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "bus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus device"},
			{Name: "address", Label: "Device Address", Type: "string", Default: "0x00", Required: true, Description: "I2C address (hex)"},
			{Name: "register", Label: "Register", Type: "number", Default: -1, Description: "Register address to read/write (-1 for none)"},
			{Name: "length", Label: "Read Length", Type: "number", Default: 1, Description: "Number of bytes to read"},
			{Name: "mode", Label: "Mode", Type: "select", Default: "read", Options: []string{"read", "write", "readwrite"}, Description: "Operation mode"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to write"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Data read"},
		},
		Factory: func() node.Executor { return &I2CExecutor{} },
	}); err != nil {
		return err
	}

	// SPI
	if err := registry.Register(&node.NodeInfo{
		Type:        "spi",
		Name:        "SPI",
		Category:    node.NodeTypeProcessing,
		Description: "SPI communication",
		Icon:        "cpu",
		Color:       "#06b6d4",
		Properties: []node.PropertySchema{
			{Name: "bus", Label: "SPI Bus", Type: "number", Default: 0, Description: "SPI bus number (0 or 1)"},
			{Name: "device", Label: "Device (CS)", Type: "number", Default: 0, Description: "Chip select (0 or 1)"},
			{Name: "speed", Label: "Speed (Hz)", Type: "number", Default: 1000000, Description: "Clock speed in Hz"},
			{Name: "mode", Label: "SPI Mode", Type: "select", Default: "0", Options: []string{"0", "1", "2", "3"}, Description: "SPI mode (CPOL/CPHA)"},
			{Name: "bitsWord", Label: "Bits Per Word", Type: "number", Default: 8, Description: "Bits per transfer word"},
			{Name: "mode3Wire", Label: "3-Wire Mode", Type: "boolean", Default: false, Description: "Enable 3-wire SPI mode"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to send"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Data received"},
		},
		Factory: func() node.Executor { return &SPIExecutor{} },
	}); err != nil {
		return err
	}

	// Serial
	if err := registry.Register(&node.NodeInfo{
		Type:        "serial",
		Name:        "Serial",
		Category:    node.NodeTypeProcessing,
		Description: "Serial/UART communication",
		Icon:        "radio",
		Color:       "#ec4899",
		Properties: []node.PropertySchema{
			{Name: "port", Label: "Serial Port", Type: "string", Default: "/dev/ttyS0", Required: true, Description: "Serial port device"},
			{Name: "baudRate", Label: "Baud Rate", Type: "select", Default: "9600", Options: []string{"1200", "2400", "4800", "9600", "19200", "38400", "57600", "115200"}, Description: "Baud rate"},
			{Name: "dataBits", Label: "Data Bits", Type: "select", Default: "8", Options: []string{"5", "6", "7", "8"}, Description: "Data bits"},
			{Name: "stopBits", Label: "Stop Bits", Type: "select", Default: "1", Options: []string{"1", "2"}, Description: "Stop bits"},
			{Name: "parity", Label: "Parity", Type: "select", Default: "none", Options: []string{"none", "odd", "even"}, Description: "Parity"},
			{Name: "mode", Label: "Mode", Type: "select", Default: "readwrite", Options: []string{"read", "write", "readwrite"}, Description: "Operation mode"},
			{Name: "timeout", Label: "Timeout (ms)", Type: "number", Default: 1000, Description: "Read timeout"},
			{Name: "delimiter", Label: "Delimiter", Type: "string", Default: "", Description: "Message delimiter"},
			{Name: "bufferSize", Label: "Buffer Size", Type: "number", Default: 1024, Description: "Read buffer size"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to send"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Data received"},
		},
		Factory: func() node.Executor { return &SerialExecutor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// TEMPERATURE/HUMIDITY SENSORS (7 nodes)
	// ============================================

	// DHT Sensor (DHT11/DHT22)
	if err := registry.Register(&node.NodeInfo{
		Type:        "dht",
		Name:        "DHT Sensor",
		Category:    node.NodeTypeInput,
		Description: "Read temperature and humidity from DHT11/DHT22",
		Icon:        "thermometer",
		Color:       "#f97316",
		Properties: []node.PropertySchema{
			{Name: "pin", Label: "Data Pin", Type: "number", Default: 4, Required: true, Description: "GPIO pin connected to DATA"},
			{Name: "type", Label: "Sensor Type", Type: "select", Default: "dht22", Options: []string{"dht11", "dht22", "am2302"}, Description: "DHT sensor model"},
			{Name: "retries", Label: "Retries", Type: "number", Default: 3, Description: "Number of read retries"},
			{Name: "unit", Label: "Temperature Unit", Type: "select", Default: "celsius", Options: []string{"celsius", "fahrenheit"}, Description: "Temperature unit"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Temperature and humidity data"},
		},
		Factory: func() node.Executor { return &DHTExecutor{} },
	}); err != nil {
		return err
	}

	// DS18B20
	if err := registry.Register(&node.NodeInfo{
		Type:        "ds18b20",
		Name:        "DS18B20",
		Category:    node.NodeTypeInput,
		Description: "Digital temperature sensor DS18B20",
		Icon:        "thermometer",
		Color:       "#ef4444",
		Properties: []node.PropertySchema{
			{Name: "deviceId", Label: "Device ID", Type: "string", Default: "", Description: "1-Wire device ID (28-xxxx) or empty for auto"},
			{Name: "unit", Label: "Temperature Unit", Type: "select", Default: "celsius", Options: []string{"celsius", "fahrenheit", "kelvin"}, Description: "Output temperature unit"},
			{Name: "resolution", Label: "Resolution", Type: "select", Default: "12", Options: []string{"9", "10", "11", "12"}, Description: "Resolution bits (9=0.5C, 12=0.0625C)"},
			{Name: "basePath", Label: "1-Wire Path", Type: "string", Default: "/sys/bus/w1/devices", Description: "1-Wire sysfs path"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Temperature data"},
		},
		Factory: func() node.Executor { return &DS18B20Node{} },
	}); err != nil {
		return err
	}

	// BMP280
	if err := registry.Register(&node.NodeInfo{
		Type:        "bmp280",
		Name:        "BMP280",
		Category:    node.NodeTypeInput,
		Description: "Pressure and temperature sensor BMP280",
		Icon:        "gauge",
		Color:       "#3b82f6",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "select", Default: "0x76", Options: []string{"0x76", "0x77"}, Description: "BMP280 I2C address"},
			{Name: "oversampling", Label: "Oversampling", Type: "select", Default: "x16", Options: []string{"x1", "x2", "x4", "x8", "x16"}, Description: "Sampling accuracy"},
			{Name: "mode", Label: "Mode", Type: "select", Default: "normal", Options: []string{"sleep", "forced", "normal"}, Description: "Operating mode"},
			{Name: "standby", Label: "Standby Time", Type: "select", Default: "1000ms", Options: []string{"0.5ms", "62.5ms", "125ms", "250ms", "500ms", "1000ms"}, Description: "Normal mode standby"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Pressure and temperature data"},
		},
		Factory: func() node.Executor { return &BMP280Node{} },
	}); err != nil {
		return err
	}

	// BME280
	if err := registry.Register(&node.NodeInfo{
		Type:        "bme280",
		Name:        "BME280",
		Category:    node.NodeTypeInput,
		Description: "Temperature, humidity, and pressure sensor BME280",
		Icon:        "cloud",
		Color:       "#0ea5e9",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "select", Default: "0x76", Options: []string{"0x76", "0x77"}, Description: "BME280 I2C address"},
			{Name: "temperatureOversampling", Label: "Temp Oversampling", Type: "select", Default: "x2", Options: []string{"off", "x1", "x2", "x4", "x8", "x16"}, Description: "Temperature oversampling"},
			{Name: "pressureOversampling", Label: "Pressure Oversampling", Type: "select", Default: "x16", Options: []string{"off", "x1", "x2", "x4", "x8", "x16"}, Description: "Pressure oversampling"},
			{Name: "humidityOversampling", Label: "Humidity Oversampling", Type: "select", Default: "x1", Options: []string{"off", "x1", "x2", "x4", "x8", "x16"}, Description: "Humidity oversampling"},
			{Name: "mode", Label: "Mode", Type: "select", Default: "normal", Options: []string{"sleep", "forced", "normal"}, Description: "Operating mode"},
			{Name: "iirFilter", Label: "IIR Filter", Type: "select", Default: "x4", Options: []string{"off", "x2", "x4", "x8", "x16"}, Description: "IIR filter coefficient"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Environmental data"},
		},
		Factory: func() node.Executor { return &BME280Executor{} },
	}); err != nil {
		return err
	}

	// BME680
	if err := registry.Register(&node.NodeInfo{
		Type:        "bme680",
		Name:        "BME680",
		Category:    node.NodeTypeInput,
		Description: "Environmental sensor with air quality BME680",
		Icon:        "wind",
		Color:       "#14b8a6",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "select", Default: "0x76", Options: []string{"0x76", "0x77"}, Description: "BME680 address"},
			{Name: "temperatureOversampling", Label: "Temp Oversampling", Type: "select", Default: "x8", Options: []string{"off", "x1", "x2", "x4", "x8", "x16"}, Description: "Temperature oversampling"},
			{Name: "pressureOversampling", Label: "Pressure Oversampling", Type: "select", Default: "x4", Options: []string{"off", "x1", "x2", "x4", "x8", "x16"}, Description: "Pressure oversampling"},
			{Name: "humidityOversampling", Label: "Humidity Oversampling", Type: "select", Default: "x2", Options: []string{"off", "x1", "x2", "x4", "x8", "x16"}, Description: "Humidity oversampling"},
			{Name: "gasHeaterTemp", Label: "Gas Heater Temp (C)", Type: "number", Default: 320, Description: "Gas heater temperature 200-400C"},
			{Name: "gasHeaterDuration", Label: "Gas Heater Duration (ms)", Type: "number", Default: 150, Description: "Gas heater duration 1-4032ms"},
			{Name: "iirFilter", Label: "IIR Filter", Type: "select", Default: "x3", Options: []string{"off", "x1", "x3", "x7", "x15", "x31", "x63", "x127"}, Description: "IIR filter"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Environmental and air quality data"},
		},
		Factory: func() node.Executor { return &BME680Executor{} },
	}); err != nil {
		return err
	}

	// SHT3x
	if err := registry.Register(&node.NodeInfo{
		Type:        "sht3x",
		Name:        "SHT3x",
		Category:    node.NodeTypeInput,
		Description: "Temperature and humidity sensor SHT3x",
		Icon:        "droplet",
		Color:       "#06b6d4",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "select", Default: "0x44", Options: []string{"0x44", "0x45"}, Description: "SHT3x address"},
			{Name: "repeatability", Label: "Repeatability", Type: "select", Default: "high", Options: []string{"low", "medium", "high"}, Description: "Measurement repeatability"},
			{Name: "mode", Label: "Mode", Type: "select", Default: "single", Options: []string{"single", "periodic"}, Description: "Measurement mode"},
			{Name: "heater", Label: "Enable Heater", Type: "boolean", Default: false, Description: "Enable internal heater"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Temperature and humidity data"},
		},
		Factory: func() node.Executor { return &SHT3xNode{} },
	}); err != nil {
		return err
	}

	// AHT20
	if err := registry.Register(&node.NodeInfo{
		Type:        "aht20",
		Name:        "AHT20",
		Category:    node.NodeTypeInput,
		Description: "Temperature and humidity sensor AHT20",
		Icon:        "cloud-drizzle",
		Color:       "#0891b2",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "string", Default: "0x38", Description: "AHT20 I2C address (fixed 0x38)"},
			{Name: "measurementDelay", Label: "Measurement Delay (ms)", Type: "number", Default: 80, Description: "Delay after trigger (75-100ms)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Temperature and humidity data"},
		},
		Factory: func() node.Executor { return &AHT20Node{} },
	}); err != nil {
		return err
	}

	// ============================================
	// LIGHT SENSORS (3 nodes)
	// ============================================

	// BH1750
	if err := registry.Register(&node.NodeInfo{
		Type:        "bh1750",
		Name:        "BH1750",
		Category:    node.NodeTypeInput,
		Description: "Ambient light sensor BH1750",
		Icon:        "sun",
		Color:       "#fbbf24",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "select", Default: "0x23", Options: []string{"0x23", "0x5c"}, Description: "ADDR pin LOW=0x23, HIGH=0x5c"},
			{Name: "mode", Label: "Resolution Mode", Type: "select", Default: "continuous_high", Options: []string{"continuous_high", "continuous_high2", "continuous_low", "onetime_high", "onetime_high2", "onetime_low"}, Description: "Measurement mode"},
			{Name: "measurementTime", Label: "Measurement Time", Type: "number", Default: 69, Description: "MTreg value 31-254 (default 69)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Light intensity data"},
		},
		Factory: func() node.Executor { return &BH1750Executor{} },
	}); err != nil {
		return err
	}

	// TSL2561
	if err := registry.Register(&node.NodeInfo{
		Type:        "tsl2561",
		Name:        "TSL2561",
		Category:    node.NodeTypeInput,
		Description: "Lux sensor TSL2561",
		Icon:        "sun",
		Color:       "#f59e0b",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "select", Default: "0x39", Options: []string{"0x29", "0x39", "0x49"}, Description: "ADDR pin: GND=0x29, Float=0x39, VCC=0x49"},
			{Name: "gain", Label: "Gain", Type: "select", Default: "auto", Options: []string{"1x", "16x", "auto"}, Description: "Sensor gain"},
			{Name: "integrationTime", Label: "Integration Time", Type: "select", Default: "402ms", Options: []string{"13.7ms", "101ms", "402ms"}, Description: "Integration time"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Lux data"},
		},
		Factory: func() node.Executor { return &TSL2561Executor{} },
	}); err != nil {
		return err
	}

	// VEML7700
	if err := registry.Register(&node.NodeInfo{
		Type:        "veml7700",
		Name:        "VEML7700",
		Category:    node.NodeTypeInput,
		Description: "High accuracy ambient light sensor VEML7700",
		Icon:        "sun",
		Color:       "#eab308",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "string", Default: "0x10", Description: "VEML7700 I2C address (fixed 0x10)"},
			{Name: "gain", Label: "ALS Gain", Type: "select", Default: "1", Options: []string{"1", "2", "1/8", "1/4"}, Description: "ALS gain setting"},
			{Name: "integrationTime", Label: "Integration Time", Type: "select", Default: "100ms", Options: []string{"25ms", "50ms", "100ms", "200ms", "400ms", "800ms"}, Description: "ALS integration time"},
			{Name: "persistence", Label: "Persistence", Type: "select", Default: "1", Options: []string{"1", "2", "4", "8"}, Description: "ALS interrupt persistence"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Ambient light data"},
		},
		Factory: func() node.Executor { return &VEML7700Executor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// AIR QUALITY SENSORS (2 nodes)
	// ============================================

	// CCS811
	if err := registry.Register(&node.NodeInfo{
		Type:        "ccs811",
		Name:        "CCS811",
		Category:    node.NodeTypeInput,
		Description: "Air quality sensor CCS811 (eCO2, TVOC)",
		Icon:        "wind",
		Color:       "#84cc16",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "select", Default: "0x5a", Options: []string{"0x5a", "0x5b"}, Description: "ADDR pin LOW=0x5A, HIGH=0x5B"},
			{Name: "driveMode", Label: "Drive Mode", Type: "select", Default: "1", Options: []string{"0", "1", "2", "3", "4"}, Description: "0=idle, 1=1s, 2=10s, 3=60s, 4=250ms"},
			{Name: "wakePin", Label: "WAKE Pin", Type: "number", Default: -1, Description: "GPIO for WAKE pin (-1 for always on)"},
			{Name: "interruptPin", Label: "Interrupt Pin", Type: "number", Default: -1, Description: "GPIO for INT pin (-1 to disable)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "eCO2 and TVOC data"},
		},
		Factory: func() node.Executor { return &CCS811Executor{} },
	}); err != nil {
		return err
	}

	// SGP30
	if err := registry.Register(&node.NodeInfo{
		Type:        "sgp30",
		Name:        "SGP30",
		Category:    node.NodeTypeInput,
		Description: "Air quality sensor SGP30 (eCO2, TVOC)",
		Icon:        "wind",
		Color:       "#22c55e",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "string", Default: "0x58", Description: "SGP30 I2C address (fixed 0x58)"},
			{Name: "baselineInterval", Label: "Baseline Save Interval (h)", Type: "number", Default: 1, Description: "Hours between baseline saves"},
			{Name: "humidity", Label: "Absolute Humidity (g/m3)", Type: "number", Default: 0, Description: "Humidity compensation (0=disabled)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "eCO2 and TVOC data"},
		},
		Factory: func() node.Executor { return &SGP30Executor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// DISTANCE SENSORS (3 nodes)
	// ============================================

	// HC-SR04
	if err := registry.Register(&node.NodeInfo{
		Type:        "hcsr04",
		Name:        "HC-SR04",
		Category:    node.NodeTypeInput,
		Description: "Ultrasonic distance sensor HC-SR04",
		Icon:        "ruler",
		Color:       "#6366f1",
		Properties: []node.PropertySchema{
			{Name: "triggerPin", Label: "Trigger Pin", Type: "number", Default: 23, Required: true, Description: "GPIO pin for TRIG"},
			{Name: "echoPin", Label: "Echo Pin", Type: "number", Default: 24, Required: true, Description: "GPIO pin for ECHO"},
			{Name: "unit", Label: "Distance Unit", Type: "select", Default: "cm", Options: []string{"mm", "cm", "m", "inch"}, Description: "Output unit"},
			{Name: "timeout", Label: "Timeout (ms)", Type: "number", Default: 100, Description: "Measurement timeout"},
			{Name: "samples", Label: "Sample Count", Type: "number", Default: 3, Description: "Number of readings to average"},
			{Name: "temperatureCompensation", Label: "Temperature (C)", Type: "number", Default: 20, Description: "Temperature for speed of sound compensation"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Distance data"},
		},
		Factory: func() node.Executor { return &HCSR04Node{} },
	}); err != nil {
		return err
	}

	// VL53L0X
	if err := registry.Register(&node.NodeInfo{
		Type:        "vl53l0x",
		Name:        "VL53L0X",
		Category:    node.NodeTypeInput,
		Description: "Time-of-Flight distance sensor VL53L0X",
		Icon:        "ruler",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "string", Default: "0x29", Description: "VL53L0X address (default 0x29)"},
			{Name: "mode", Label: "Ranging Mode", Type: "select", Default: "long", Options: []string{"default", "high_speed", "high_accuracy", "long"}, Description: "Ranging profile"},
			{Name: "timingBudget", Label: "Timing Budget (ms)", Type: "number", Default: 33, Description: "Measurement timing budget (20-200ms)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Distance data"},
		},
		Factory: func() node.Executor { return &VL53L0XExecutor{} },
	}); err != nil {
		return err
	}

	// VL53L1X
	if err := registry.Register(&node.NodeInfo{
		Type:        "vl53l1x",
		Name:        "VL53L1X",
		Category:    node.NodeTypeInput,
		Description: "Long range Time-of-Flight sensor VL53L1X",
		Icon:        "ruler",
		Color:       "#a855f7",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "string", Default: "0x29", Description: "VL53L1X address"},
			{Name: "distanceMode", Label: "Distance Mode", Type: "select", Default: "long", Options: []string{"short", "long"}, Description: "Short=1.3m/Long=4m max"},
			{Name: "timingBudget", Label: "Timing Budget (ms)", Type: "select", Default: "100", Options: []string{"15", "20", "33", "50", "100", "200", "500"}, Description: "Timing budget"},
			{Name: "roi", Label: "ROI (Region of Interest)", Type: "string", Default: "16x16", Description: "ROI size WxH (4x4 to 16x16)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Distance data"},
		},
		Factory: func() node.Executor { return &VL53L1XExecutor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// MOTION SENSORS (2 nodes)
	// ============================================

	// PIR
	if err := registry.Register(&node.NodeInfo{
		Type:        "pir",
		Name:        "PIR Motion Sensor",
		Category:    node.NodeTypeInput,
		Description: "PIR motion detection sensor",
		Icon:        "eye",
		Color:       "#f43f5e",
		Properties: []node.PropertySchema{
			{Name: "pin", Label: "Signal Pin", Type: "number", Default: 17, Required: true, Description: "GPIO pin for signal output"},
			{Name: "debounce", Label: "Debounce Time", Type: "string", Default: "50ms", Description: "Debounce duration"},
			{Name: "sensitivity", Label: "Sensitivity", Type: "select", Default: "normal", Options: []string{"low", "normal", "high"}, Description: "Detection sensitivity"},
			{Name: "pullMode", Label: "Pull Mode", Type: "select", Default: "down", Options: []string{"none", "up", "down"}, Description: "Internal pull resistor"},
			{Name: "triggerMode", Label: "Trigger Mode", Type: "select", Default: "rising", Options: []string{"rising", "falling", "both"}, Description: "Edge trigger"},
			{Name: "retriggerTime", Label: "Retrigger Time", Type: "string", Default: "2s", Description: "Minimum time between triggers"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Motion detection output"},
		},
		Factory: func() node.Executor { return &PIRNode{} },
	}); err != nil {
		return err
	}

	// RCWL-0516
	if err := registry.Register(&node.NodeInfo{
		Type:        "rcwl0516",
		Name:        "RCWL-0516",
		Category:    node.NodeTypeInput,
		Description: "Microwave motion sensor RCWL-0516",
		Icon:        "radio",
		Color:       "#ec4899",
		Properties: []node.PropertySchema{
			{Name: "pin", Label: "Signal Pin", Type: "number", Default: 17, Required: true, Description: "GPIO pin for signal"},
			{Name: "activeHigh", Label: "Active High", Type: "boolean", Default: true, Description: "Signal polarity"},
			{Name: "debounceMs", Label: "Debounce (ms)", Type: "number", Default: 50, Description: "Debounce time"},
			{Name: "holdTimeMs", Label: "Hold Time (ms)", Type: "number", Default: 2000, Description: "Motion hold duration"},
			{Name: "enableCallback", Label: "Enable Callback", Type: "boolean", Default: false, Description: "Enable edge callback"},
			{Name: "pollInterval", Label: "Poll Interval (ms)", Type: "number", Default: 100, Description: "Polling interval"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Motion detection data"},
		},
		Factory: func() node.Executor { return &RCWL0516Executor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// GPS/POSITION SENSORS (3 nodes)
	// ============================================

	// GPS
	if err := registry.Register(&node.NodeInfo{
		Type:        "gps",
		Name:        "GPS",
		Category:    node.NodeTypeInput,
		Description: "GPS receiver (NMEA)",
		Icon:        "map-pin",
		Color:       "#10b981",
		Properties: []node.PropertySchema{
			{Name: "port", Label: "Serial Port", Type: "string", Default: "/dev/ttyS0", Required: true, Description: "GPS serial port"},
			{Name: "baudRate", Label: "Baud Rate", Type: "select", Default: "9600", Options: []string{"4800", "9600", "19200", "38400", "57600", "115200"}, Description: "GPS baud rate"},
			{Name: "updateRate", Label: "Update Rate (Hz)", Type: "select", Default: "1", Options: []string{"1", "5", "10"}, Description: "Position update rate"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "GPS position data"},
		},
		Factory: func() node.Executor { return &GPSExecutor{} },
	}); err != nil {
		return err
	}

	// GPS NEO-M8N
	if err := registry.Register(&node.NodeInfo{
		Type:        "gps_neom8n",
		Name:        "GPS NEO-M8N",
		Category:    node.NodeTypeInput,
		Description: "u-blox NEO-M8N GPS module",
		Icon:        "navigation",
		Color:       "#059669",
		Properties: []node.PropertySchema{
			{Name: "port", Label: "Serial Port", Type: "string", Default: "/dev/ttyS0", Required: true, Description: "NEO-M8N serial port"},
			{Name: "baudRate", Label: "Baud Rate", Type: "select", Default: "9600", Options: []string{"9600", "19200", "38400", "57600", "115200"}, Description: "Baud rate"},
			{Name: "dynamicModel", Label: "Dynamic Model", Type: "select", Default: "portable", Options: []string{"portable", "stationary", "pedestrian", "automotive", "sea", "airborne_1g", "airborne_2g", "airborne_4g"}, Description: "Navigation model"},
			{Name: "fixMode", Label: "Fix Mode", Type: "select", Default: "auto", Options: []string{"auto", "2d", "3d"}, Description: "Position fix mode"},
			{Name: "sbas", Label: "SBAS Enabled", Type: "boolean", Default: true, Description: "Enable SBAS augmentation"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "GPS position data"},
		},
		Factory: func() node.Executor { return &NEOM8NExecutor{} },
	}); err != nil {
		return err
	}

	// Compass BN-880
	if err := registry.Register(&node.NodeInfo{
		Type:        "compass_bn880",
		Name:        "BN-880 GPS+Compass",
		Category:    node.NodeTypeInput,
		Description: "BN-880 GPS with QMC5883L compass",
		Icon:        "compass",
		Color:       "#0d9488",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "compassAddress", Label: "Compass Address", Type: "string", Default: "0x0d", Description: "QMC5883L compass address"},
			{Name: "serialPort", Label: "GPS Serial Port", Type: "string", Default: "/dev/ttyS0", Description: "GPS serial port"},
			{Name: "baudRate", Label: "GPS Baud Rate", Type: "select", Default: "9600", Options: []string{"9600", "19200", "38400"}, Description: "GPS baud rate"},
			{Name: "outputRate", Label: "Compass Output Rate", Type: "select", Default: "10Hz", Options: []string{"10Hz", "50Hz", "100Hz", "200Hz"}, Description: "Compass data rate"},
			{Name: "range", Label: "Compass Range", Type: "select", Default: "8G", Options: []string{"2G", "8G"}, Description: "Magnetic field range"},
			{Name: "oversampling", Label: "Oversampling", Type: "select", Default: "512", Options: []string{"64", "128", "256", "512"}, Description: "Oversampling ratio"},
			{Name: "declinationAngle", Label: "Declination Angle", Type: "number", Default: 0, Description: "Magnetic declination in degrees"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "GPS and compass data"},
		},
		Factory: func() node.Executor { return &BN880Executor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// ADC/DAC NODES (5 nodes)
	// ============================================

	// MCP3008
	if err := registry.Register(&node.NodeInfo{
		Type:        "mcp3008",
		Name:        "MCP3008",
		Category:    node.NodeTypeInput,
		Description: "8-channel 10-bit ADC MCP3008",
		Icon:        "bar-chart",
		Color:       "#6366f1",
		Properties: []node.PropertySchema{
			{Name: "spiBus", Label: "SPI Bus", Type: "number", Default: 0, Description: "SPI bus number"},
			{Name: "spiDevice", Label: "SPI Device (CS)", Type: "number", Default: 0, Description: "Chip select"},
			{Name: "channel", Label: "ADC Channel", Type: "select", Default: "0", Options: []string{"0", "1", "2", "3", "4", "5", "6", "7"}, Description: "ADC channel (0-7)"},
			{Name: "vref", Label: "Reference Voltage", Type: "number", Default: 3.3, Description: "Reference voltage (Vref)"},
			{Name: "spiSpeed", Label: "SPI Speed (Hz)", Type: "number", Default: 1000000, Description: "SPI clock speed"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "ADC reading data"},
		},
		Factory: func() node.Executor { return &MCP3008Executor{} },
	}); err != nil {
		return err
	}

	// ADS1015
	if err := registry.Register(&node.NodeInfo{
		Type:        "ads1015",
		Name:        "ADS1015",
		Category:    node.NodeTypeInput,
		Description: "12-bit ADC with PGA ADS1015",
		Icon:        "bar-chart-2",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "select", Default: "0x48", Options: []string{"0x48", "0x49", "0x4a", "0x4b"}, Description: "ADS1015 address"},
			{Name: "channel", Label: "Input Channel", Type: "select", Default: "A0", Options: []string{"A0", "A1", "A2", "A3", "A0-A1", "A0-A3", "A1-A3", "A2-A3"}, Description: "Single-ended or differential"},
			{Name: "gain", Label: "PGA Gain", Type: "select", Default: "1", Options: []string{"2/3", "1", "2", "4", "8", "16"}, Description: "Programmable gain (FSR: 6.144V to 0.256V)"},
			{Name: "dataRate", Label: "Data Rate", Type: "select", Default: "1600", Options: []string{"128", "250", "490", "920", "1600", "2400", "3300"}, Description: "Samples per second"},
			{Name: "mode", Label: "Mode", Type: "select", Default: "single", Options: []string{"single", "continuous"}, Description: "Conversion mode"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "ADC reading data"},
		},
		Factory: func() node.Executor { return &ADS1015Executor{} },
	}); err != nil {
		return err
	}

	// PCF8591
	if err := registry.Register(&node.NodeInfo{
		Type:        "pcf8591",
		Name:        "PCF8591",
		Category:    node.NodeTypeInput,
		Description: "8-bit ADC/DAC PCF8591",
		Icon:        "sliders",
		Color:       "#a855f7",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "select", Default: "0x48", Options: []string{"0x48", "0x49", "0x4a", "0x4b", "0x4c", "0x4d", "0x4e", "0x4f"}, Description: "PCF8591 address (A0-A2 pins)"},
			{Name: "channel", Label: "ADC Channel", Type: "select", Default: "0", Options: []string{"0", "1", "2", "3"}, Description: "Analog input channel"},
			{Name: "inputMode", Label: "Input Mode", Type: "select", Default: "single", Options: []string{"single", "differential", "single_diff", "two_diff"}, Description: "Analog input mode"},
			{Name: "autoIncrement", Label: "Auto Increment", Type: "boolean", Default: false, Description: "Auto-increment channel"},
			{Name: "vref", Label: "Reference Voltage", Type: "number", Default: 3.3, Description: "Reference voltage"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading or DAC value"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "ADC reading data"},
		},
		Factory: func() node.Executor { return &PCF8591Executor{} },
	}); err != nil {
		return err
	}

	// Voltage Monitor
	if err := registry.Register(&node.NodeInfo{
		Type:        "voltage-monitor",
		Name:        "Voltage Monitor",
		Category:    node.NodeTypeInput,
		Description: "Monitor voltage with ADC",
		Icon:        "zap",
		Color:       "#eab308",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "ADC Address", Type: "string", Default: "0x48", Description: "ADC I2C address"},
			{Name: "channel", Label: "ADC Channel", Type: "number", Default: 0, Description: "ADC input channel"},
			{Name: "gain", Label: "PGA Gain", Type: "number", Default: 1, Description: "Programmable gain"},
			{Name: "voltageDivider", Label: "Voltage Divider Ratio", Type: "number", Default: 1, Description: "R1/(R1+R2) ratio"},
			{Name: "threshold", Label: "Alert Threshold", Type: "number", Default: 0, Description: "Alert threshold voltage (0=disabled)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Voltage data"},
		},
		Factory: func() node.Executor { return &VoltageMonitorNode{} },
	}); err != nil {
		return err
	}

	// Current Monitor
	if err := registry.Register(&node.NodeInfo{
		Type:        "current-monitor",
		Name:        "Current Monitor",
		Category:    node.NodeTypeInput,
		Description: "Monitor current with ADC and shunt resistor",
		Icon:        "activity",
		Color:       "#f97316",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "ADC Address", Type: "string", Default: "0x48", Description: "ADC I2C address"},
			{Name: "channel", Label: "ADC Channel", Type: "number", Default: 0, Description: "ADC input channel"},
			{Name: "shuntResistor", Label: "Shunt Resistor (Ohm)", Type: "number", Default: 0.1, Required: true, Description: "Shunt resistor value in ohms"},
			{Name: "maxCurrent", Label: "Max Current (A)", Type: "number", Default: 10, Description: "Maximum expected current"},
			{Name: "averaging", Label: "Averaging Samples", Type: "number", Default: 4, Description: "Number of samples to average"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Current data"},
		},
		Factory: func() node.Executor { return &CurrentMonitorNode{} },
	}); err != nil {
		return err
	}

	// ============================================
	// RTC NODES (3 nodes)
	// ============================================

	// RTC DS3231
	if err := registry.Register(&node.NodeInfo{
		Type:        "rtc_ds3231",
		Name:        "RTC DS3231",
		Category:    node.NodeTypeProcessing,
		Description: "Real-time clock DS3231",
		Icon:        "clock",
		Color:       "#0ea5e9",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "string", Default: "0x68", Description: "DS3231 address (fixed 0x68)"},
			{Name: "use24Hour", Label: "24-Hour Format", Type: "boolean", Default: true, Description: "Use 24-hour time format"},
			{Name: "squareWaveFreq", Label: "Square Wave Freq", Type: "select", Default: "off", Options: []string{"off", "1Hz", "1.024kHz", "4.096kHz", "8.192kHz"}, Description: "SQW output frequency"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Command input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Time data"},
		},
		Factory: func() node.Executor { return &DS3231Executor{} },
	}); err != nil {
		return err
	}

	// RTC DS1307
	if err := registry.Register(&node.NodeInfo{
		Type:        "rtc_ds1307",
		Name:        "RTC DS1307",
		Category:    node.NodeTypeProcessing,
		Description: "Real-time clock DS1307",
		Icon:        "clock",
		Color:       "#3b82f6",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "string", Default: "0x68", Description: "DS1307 address (fixed 0x68)"},
			{Name: "use24Hour", Label: "24-Hour Format", Type: "boolean", Default: true, Description: "Use 24-hour time format"},
			{Name: "squareWaveFreq", Label: "Square Wave Freq", Type: "select", Default: "off", Options: []string{"off", "1Hz", "4.096kHz", "8.192kHz", "32.768kHz"}, Description: "SQW output frequency"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Command input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Time data"},
		},
		Factory: func() node.Executor { return &DS1307Executor{} },
	}); err != nil {
		return err
	}

	// RTC PCF8523
	if err := registry.Register(&node.NodeInfo{
		Type:        "rtc_pcf8523",
		Name:        "RTC PCF8523",
		Category:    node.NodeTypeProcessing,
		Description: "Low-power real-time clock PCF8523",
		Icon:        "clock",
		Color:       "#6366f1",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "string", Default: "0x68", Description: "PCF8523 address (fixed 0x68)"},
			{Name: "use24Hour", Label: "24-Hour Format", Type: "boolean", Default: true, Description: "Use 24-hour time format"},
			{Name: "batteryLowInt", Label: "Battery Low Interrupt", Type: "boolean", Default: false, Description: "Enable battery low interrupt"},
			{Name: "clockoutFreq", Label: "Clockout Freq", Type: "select", Default: "off", Options: []string{"off", "32768hz", "16384hz", "8192hz", "4096hz", "1024hz", "32hz", "1hz"}, Description: "CLKOUT frequency"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Command input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Time data"},
		},
		Factory: func() node.Executor { return &PCF8523Executor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// DISPLAY NODES (2 nodes)
	// ============================================

	// LCD I2C
	if err := registry.Register(&node.NodeInfo{
		Type:        "lcd_i2c",
		Name:        "LCD I2C",
		Category:    node.NodeTypeOutput,
		Description: "I2C LCD display (HD44780 compatible)",
		Icon:        "monitor",
		Color:       "#14b8a6",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "select", Default: "0x27", Options: []string{"0x27", "0x3f"}, Description: "PCF8574 I2C address"},
			{Name: "rows", Label: "Rows", Type: "select", Default: "2", Options: []string{"1", "2", "4"}, Description: "Number of display rows"},
			{Name: "cols", Label: "Columns", Type: "select", Default: "16", Options: []string{"16", "20"}, Description: "Number of display columns"},
			{Name: "backlight", Label: "Backlight", Type: "boolean", Default: true, Description: "Enable backlight"},
			{Name: "cursorBlink", Label: "Cursor Blink", Type: "boolean", Default: false, Description: "Enable cursor blink"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Text to display"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Confirmation"},
		},
		Factory: func() node.Executor { return &LCDI2CExecutor{} },
	}); err != nil {
		return err
	}

	// OLED SSD1306
	if err := registry.Register(&node.NodeInfo{
		Type:        "oled_ssd1306",
		Name:        "OLED SSD1306",
		Category:    node.NodeTypeOutput,
		Description: "OLED display SSD1306",
		Icon:        "tv",
		Color:       "#0d9488",
		Properties: []node.PropertySchema{
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus"},
			{Name: "address", Label: "I2C Address", Type: "select", Default: "0x3c", Options: []string{"0x3c", "0x3d"}, Description: "SSD1306 address"},
			{Name: "width", Label: "Width (pixels)", Type: "select", Default: "128", Options: []string{"128", "64"}, Description: "Display width"},
			{Name: "height", Label: "Height (pixels)", Type: "select", Default: "64", Options: []string{"32", "64"}, Description: "Display height"},
			{Name: "rotation", Label: "Rotation", Type: "select", Default: "0", Options: []string{"0", "90", "180", "270"}, Description: "Display rotation"},
			{Name: "contrast", Label: "Contrast", Type: "number", Default: 255, Description: "Display contrast (0-255)"},
			{Name: "externalVcc", Label: "External VCC", Type: "boolean", Default: false, Description: "External VCC power"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Content to display"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Confirmation"},
		},
		Factory: func() node.Executor { return &OLEDSSD1306Executor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// ACTUATOR NODES (5 nodes)
	// ============================================

	// Relay
	if err := registry.Register(&node.NodeInfo{
		Type:        "relay",
		Name:        "Relay",
		Category:    node.NodeTypeOutput,
		Description: "Relay control",
		Icon:        "toggle-right",
		Color:       "#ef4444",
		Properties: []node.PropertySchema{
			{Name: "pin", Label: "GPIO Pin", Type: "number", Default: 0, Required: true, Description: "Relay control GPIO pin"},
			{Name: "initialState", Label: "Initial State", Type: "boolean", Default: false, Description: "Initial relay state (ON/OFF)"},
			{Name: "activeLow", Label: "Active Low", Type: "boolean", Default: false, Description: "Active low logic (inverted)"},
			{Name: "pulseMode", Label: "Pulse Mode", Type: "boolean", Default: false, Description: "Enable pulse mode"},
			{Name: "pulseDuration", Label: "Pulse Duration (ms)", Type: "number", Default: 100, Description: "Pulse duration in milliseconds"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Control input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "State output"},
		},
		Factory: func() node.Executor { return &RelayExecutor{} },
	}); err != nil {
		return err
	}

	// Servo
	if err := registry.Register(&node.NodeInfo{
		Type:        "servo",
		Name:        "Servo Motor",
		Category:    node.NodeTypeOutput,
		Description: "Servo motor control",
		Icon:        "rotate-cw",
		Color:       "#f59e0b",
		Properties: []node.PropertySchema{
			{Name: "pin", Label: "PWM Pin", Type: "number", Default: 18, Required: true, Description: "Servo PWM GPIO pin"},
			{Name: "minPulse", Label: "Min Pulse (ms)", Type: "number", Default: 1.0, Description: "Minimum pulse width (0 deg)"},
			{Name: "maxPulse", Label: "Max Pulse (ms)", Type: "number", Default: 2.0, Description: "Maximum pulse width (180 deg)"},
			{Name: "frequency", Label: "PWM Frequency (Hz)", Type: "number", Default: 50.0, Description: "PWM frequency (typically 50Hz)"},
			{Name: "minAngle", Label: "Min Angle", Type: "number", Default: 0.0, Description: "Minimum angle in degrees"},
			{Name: "maxAngle", Label: "Max Angle", Type: "number", Default: 180.0, Description: "Maximum angle in degrees"},
			{Name: "startAngle", Label: "Start Angle", Type: "number", Default: 90.0, Description: "Initial angle position"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Angle input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Position output"},
		},
		Factory: func() node.Executor { return &ServoExecutor{} },
	}); err != nil {
		return err
	}

	// Motor L298N
	if err := registry.Register(&node.NodeInfo{
		Type:        "motor_l298n",
		Name:        "Motor L298N",
		Category:    node.NodeTypeOutput,
		Description: "DC motor driver L298N",
		Icon:        "settings",
		Color:       "#dc2626",
		Properties: []node.PropertySchema{
			{Name: "enablePin", Label: "Enable Pin (ENA/ENB)", Type: "number", Default: 12, Required: true, Description: "PWM enable pin"},
			{Name: "in1Pin", Label: "IN1 Pin", Type: "number", Default: 23, Required: true, Description: "Direction control IN1"},
			{Name: "in2Pin", Label: "IN2 Pin", Type: "number", Default: 24, Required: true, Description: "Direction control IN2"},
			{Name: "pwmFrequency", Label: "PWM Frequency (Hz)", Type: "number", Default: 1000, Description: "PWM frequency"},
			{Name: "invertDirection", Label: "Invert Direction", Type: "boolean", Default: false, Description: "Invert rotation direction"},
			{Name: "initialSpeed", Label: "Initial Speed (%)", Type: "number", Default: 0, Description: "Initial motor speed 0-100%"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Speed/direction input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Status output"},
		},
		Factory: func() node.Executor { return &MotorL298NExecutor{} },
	}); err != nil {
		return err
	}

	// Buzzer
	if err := registry.Register(&node.NodeInfo{
		Type:        "buzzer",
		Name:        "Buzzer",
		Category:    node.NodeTypeOutput,
		Description: "Buzzer/piezo control",
		Icon:        "volume-2",
		Color:       "#fbbf24",
		Properties: []node.PropertySchema{
			{Name: "pin", Label: "GPIO Pin", Type: "number", Default: 18, Required: true, Description: "Buzzer GPIO pin"},
			{Name: "type", Label: "Buzzer Type", Type: "select", Default: "active", Options: []string{"active", "passive"}, Description: "Active (on/off) or passive (PWM)"},
			{Name: "frequency", Label: "Frequency (Hz)", Type: "number", Default: 1000, Description: "PWM frequency for passive buzzer"},
			{Name: "activeLow", Label: "Active Low", Type: "boolean", Default: false, Description: "Active low logic"},
			{Name: "duration", Label: "Default Duration (ms)", Type: "number", Default: 100, Description: "Default beep duration"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Beep control input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Status output"},
		},
		Factory: func() node.Executor { return &BuzzerExecutor{} },
	}); err != nil {
		return err
	}

	// WS2812
	if err := registry.Register(&node.NodeInfo{
		Type:        "ws2812",
		Name:        "WS2812 LED Strip",
		Category:    node.NodeTypeOutput,
		Description: "Addressable RGB LED strip WS2812/NeoPixel",
		Icon:        "lightbulb",
		Color:       "#ec4899",
		Properties: []node.PropertySchema{
			{Name: "pin", Label: "Data Pin", Type: "number", Default: 18, Required: true, Description: "WS2812 data GPIO pin (PWM capable)"},
			{Name: "ledCount", Label: "LED Count", Type: "number", Default: 8, Required: true, Description: "Number of LEDs in strip"},
			{Name: "brightness", Label: "Brightness", Type: "number", Default: 255, Description: "Global brightness (0-255)"},
			{Name: "stripType", Label: "Strip Type", Type: "select", Default: "GRB", Options: []string{"RGB", "GRB", "RGBW", "GRBW"}, Description: "LED color order"},
			{Name: "frequency", Label: "PWM Frequency (Hz)", Type: "number", Default: 800000, Description: "Data frequency (800kHz typical)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Color/pattern input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Status output"},
		},
		Factory: func() node.Executor { return &WS2812Executor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// THERMOCOUPLE NODES (2 nodes)
	// ============================================

	// MAX31855
	if err := registry.Register(&node.NodeInfo{
		Type:        "max31855",
		Name:        "MAX31855",
		Category:    node.NodeTypeInput,
		Description: "Thermocouple amplifier MAX31855",
		Icon:        "thermometer",
		Color:       "#dc2626",
		Properties: []node.PropertySchema{
			{Name: "spiBus", Label: "SPI Bus", Type: "number", Default: 0, Description: "SPI bus number"},
			{Name: "spiDevice", Label: "SPI Device (CS)", Type: "number", Default: 0, Description: "Chip select"},
			{Name: "spiSpeed", Label: "SPI Speed (Hz)", Type: "number", Default: 5000000, Description: "SPI clock speed"},
			{Name: "type", Label: "Thermocouple Type", Type: "select", Default: "K", Options: []string{"K", "J", "N", "S", "T", "E", "R"}, Description: "Thermocouple type"},
			{Name: "unit", Label: "Temperature Unit", Type: "select", Default: "celsius", Options: []string{"celsius", "fahrenheit"}, Description: "Output temperature unit"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Temperature data"},
		},
		Factory: func() node.Executor { return &MAX31855Executor{} },
	}); err != nil {
		return err
	}

	// MAX31865
	if err := registry.Register(&node.NodeInfo{
		Type:        "max31865",
		Name:        "MAX31865",
		Category:    node.NodeTypeInput,
		Description: "RTD-to-digital converter MAX31865",
		Icon:        "thermometer",
		Color:       "#b91c1c",
		Properties: []node.PropertySchema{
			{Name: "spiBus", Label: "SPI Bus", Type: "number", Default: 0, Description: "SPI bus number"},
			{Name: "spiDevice", Label: "SPI Device (CS)", Type: "number", Default: 0, Description: "Chip select"},
			{Name: "spiSpeed", Label: "SPI Speed (Hz)", Type: "number", Default: 5000000, Description: "SPI clock speed"},
			{Name: "rtdType", Label: "RTD Type", Type: "select", Default: "PT100", Options: []string{"PT100", "PT1000"}, Description: "RTD sensor type"},
			{Name: "wires", Label: "Wire Configuration", Type: "select", Default: "3", Options: []string{"2", "3", "4"}, Description: "RTD wire configuration"},
			{Name: "refResistor", Label: "Reference Resistor (Ohm)", Type: "number", Default: 430, Description: "Reference resistor value"},
			{Name: "unit", Label: "Temperature Unit", Type: "select", Default: "celsius", Options: []string{"celsius", "fahrenheit"}, Description: "Output unit"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Temperature data"},
		},
		Factory: func() node.Executor { return &MAX31865Executor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// COMMUNICATION NODES (4 nodes)
	// ============================================

	// Modbus
	if err := registry.Register(&node.NodeInfo{
		Type:        "modbus",
		Name:        "Modbus",
		Category:    node.NodeTypeProcessing,
		Description: "Modbus RTU/TCP communication",
		Icon:        "server",
		Color:       "#0284c7",
		Properties: []node.PropertySchema{
			{Name: "mode", Label: "Mode", Type: "select", Default: "rtu", Options: []string{"rtu", "tcp"}, Required: true, Description: "Modbus RTU or TCP"},
			{Name: "port", Label: "Serial Port", Type: "string", Default: "/dev/ttyUSB0", Description: "Serial port (RTU mode)"},
			{Name: "baudRate", Label: "Baud Rate", Type: "select", Default: "9600", Options: []string{"1200", "2400", "4800", "9600", "19200", "38400", "57600", "115200"}, Description: "RTU baud rate"},
			{Name: "parity", Label: "Parity", Type: "select", Default: "none", Options: []string{"none", "even", "odd"}, Description: "RTU parity"},
			{Name: "host", Label: "TCP Host", Type: "string", Default: "127.0.0.1", Description: "TCP host address"},
			{Name: "tcpPort", Label: "TCP Port", Type: "number", Default: 502, Description: "TCP port (502)"},
			{Name: "slaveId", Label: "Slave ID", Type: "number", Default: 1, Description: "Modbus slave ID (1-247)"},
			{Name: "timeout", Label: "Timeout (ms)", Type: "number", Default: 1000, Description: "Response timeout"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Request input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Response output"},
		},
		Factory: func() node.Executor { return &ModbusExecutor{} },
	}); err != nil {
		return err
	}

	// LoRa SX1276
	if err := registry.Register(&node.NodeInfo{
		Type:        "lora_sx1276",
		Name:        "LoRa SX1276",
		Category:    node.NodeTypeProcessing,
		Description: "LoRa transceiver SX1276/SX1278",
		Icon:        "radio",
		Color:       "#7c3aed",
		Properties: []node.PropertySchema{
			{Name: "spiBus", Label: "SPI Bus", Type: "number", Default: 0, Description: "SPI bus number"},
			{Name: "spiDevice", Label: "SPI Device", Type: "number", Default: 0, Description: "Chip select"},
			{Name: "resetPin", Label: "Reset Pin", Type: "number", Default: 25, Description: "GPIO reset pin"},
			{Name: "dio0Pin", Label: "DIO0 Pin", Type: "number", Default: 22, Description: "GPIO DIO0 pin (RX done)"},
			{Name: "frequency", Label: "Frequency (MHz)", Type: "select", Default: "915", Options: []string{"433", "868", "915"}, Description: "Carrier frequency"},
			{Name: "bandwidth", Label: "Bandwidth", Type: "select", Default: "125", Options: []string{"7.8", "10.4", "15.6", "20.8", "31.25", "41.7", "62.5", "125", "250", "500"}, Description: "Signal bandwidth (kHz)"},
			{Name: "spreadingFactor", Label: "Spreading Factor", Type: "select", Default: "7", Options: []string{"6", "7", "8", "9", "10", "11", "12"}, Description: "SF6-SF12"},
			{Name: "codingRate", Label: "Coding Rate", Type: "select", Default: "5", Options: []string{"5", "6", "7", "8"}, Description: "4/5 to 4/8"},
			{Name: "txPower", Label: "TX Power (dBm)", Type: "number", Default: 17, Description: "Transmit power 2-20 dBm"},
			{Name: "syncWord", Label: "Sync Word", Type: "string", Default: "0x12", Description: "Sync word (0x12 private, 0x34 LoRaWAN)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to transmit"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Received data"},
		},
		Factory: func() node.Executor { return &SX1276Executor{} },
	}); err != nil {
		return err
	}

	// NRF24L01
	if err := registry.Register(&node.NodeInfo{
		Type:        "nrf24l01",
		Name:        "NRF24L01",
		Category:    node.NodeTypeProcessing,
		Description: "2.4GHz transceiver NRF24L01",
		Icon:        "wifi",
		Color:       "#2563eb",
		Properties: []node.PropertySchema{
			{Name: "spiBus", Label: "SPI Bus", Type: "number", Default: 0, Description: "SPI bus number"},
			{Name: "spiDevice", Label: "SPI Device", Type: "number", Default: 0, Description: "Chip select"},
			{Name: "cePin", Label: "CE Pin", Type: "number", Default: 22, Required: true, Description: "Chip enable GPIO"},
			{Name: "channel", Label: "RF Channel", Type: "number", Default: 76, Description: "RF channel 0-125 (2400+ch MHz)"},
			{Name: "dataRate", Label: "Data Rate", Type: "select", Default: "1Mbps", Options: []string{"250kbps", "1Mbps", "2Mbps"}, Description: "Air data rate"},
			{Name: "paLevel", Label: "PA Level", Type: "select", Default: "max", Options: []string{"min", "low", "high", "max"}, Description: "Power amplifier level"},
			{Name: "txAddress", Label: "TX Address", Type: "string", Default: "1Node", Description: "5-byte transmit address"},
			{Name: "rxAddress", Label: "RX Address", Type: "string", Default: "2Node", Description: "5-byte receive address"},
			{Name: "autoAck", Label: "Auto ACK", Type: "boolean", Default: true, Description: "Enable auto acknowledgment"},
			{Name: "crcLength", Label: "CRC Length", Type: "select", Default: "16", Options: []string{"disabled", "8", "16"}, Description: "CRC length"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to transmit"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Received data"},
		},
		Factory: func() node.Executor { return &NRF24L01Executor{} },
	}); err != nil {
		return err
	}

	// RF433
	if err := registry.Register(&node.NodeInfo{
		Type:        "rf433",
		Name:        "RF 433MHz",
		Category:    node.NodeTypeProcessing,
		Description: "433MHz RF transmitter/receiver",
		Icon:        "radio",
		Color:       "#059669",
		Properties: []node.PropertySchema{
			{Name: "txPin", Label: "TX Pin", Type: "number", Default: 17, Description: "Transmitter GPIO pin"},
			{Name: "rxPin", Label: "RX Pin", Type: "number", Default: 27, Description: "Receiver GPIO pin"},
			{Name: "protocol", Label: "Protocol", Type: "select", Default: "1", Options: []string{"1", "2", "3", "4", "5"}, Description: "RF protocol"},
			{Name: "pulseLength", Label: "Pulse Length (us)", Type: "number", Default: 350, Description: "Pulse length in microseconds"},
			{Name: "repeatTransmit", Label: "Repeat Transmit", Type: "number", Default: 10, Description: "Number of TX repeats"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Code to transmit"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Received code"},
		},
		Factory: func() node.Executor { return &RF433Executor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// RFID/NFC NODES (2 nodes)
	// ============================================

	// RFID RC522
	if err := registry.Register(&node.NodeInfo{
		Type:        "rfid_rc522",
		Name:        "RFID RC522",
		Category:    node.NodeTypeInput,
		Description: "RFID reader RC522 (13.56MHz)",
		Icon:        "credit-card",
		Color:       "#0891b2",
		Properties: []node.PropertySchema{
			{Name: "spiBus", Label: "SPI Bus", Type: "number", Default: 0, Description: "SPI bus number"},
			{Name: "spiDevice", Label: "SPI Device", Type: "number", Default: 0, Description: "Chip select"},
			{Name: "spiSpeed", Label: "SPI Speed (Hz)", Type: "number", Default: 1000000, Description: "SPI clock speed"},
			{Name: "resetPin", Label: "Reset Pin", Type: "number", Default: 25, Description: "GPIO reset pin (-1 to disable)"},
			{Name: "antennaGain", Label: "Antenna Gain", Type: "select", Default: "48dB", Options: []string{"18dB", "23dB", "33dB", "38dB", "43dB", "48dB"}, Description: "Receiver antenna gain"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger scan"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Tag data"},
		},
		Factory: func() node.Executor { return &RC522Executor{} },
	}); err != nil {
		return err
	}

	// NFC PN532
	if err := registry.Register(&node.NodeInfo{
		Type:        "nfc_pn532",
		Name:        "NFC PN532",
		Category:    node.NodeTypeInput,
		Description: "NFC/RFID reader PN532",
		Icon:        "smartphone",
		Color:       "#0d9488",
		Properties: []node.PropertySchema{
			{Name: "interface", Label: "Interface", Type: "select", Default: "i2c", Options: []string{"i2c", "spi", "uart"}, Required: true, Description: "Communication interface"},
			{Name: "i2cBus", Label: "I2C Bus", Type: "select", Default: "/dev/i2c-1", Options: []string{"/dev/i2c-0", "/dev/i2c-1"}, Description: "I2C bus (I2C mode)"},
			{Name: "i2cAddress", Label: "I2C Address", Type: "string", Default: "0x24", Description: "I2C address (0x24)"},
			{Name: "serialPort", Label: "Serial Port", Type: "string", Default: "/dev/ttyS0", Description: "Serial port (UART mode)"},
			{Name: "spiBus", Label: "SPI Bus", Type: "number", Default: 0, Description: "SPI bus (SPI mode)"},
			{Name: "spiDevice", Label: "SPI Device", Type: "number", Default: 0, Description: "SPI device (SPI mode)"},
			{Name: "irqPin", Label: "IRQ Pin", Type: "number", Default: -1, Description: "IRQ GPIO pin (-1 to disable)"},
			{Name: "resetPin", Label: "Reset Pin", Type: "number", Default: -1, Description: "Reset GPIO pin (-1 to disable)"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger scan or command"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Tag data"},
		},
		Factory: func() node.Executor { return &PN532Executor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// CAN BUS NODE (1 node)
	// ============================================

	// CAN MCP2515
	if err := registry.Register(&node.NodeInfo{
		Type:        "can_mcp2515",
		Name:        "CAN MCP2515",
		Category:    node.NodeTypeProcessing,
		Description: "CAN bus controller MCP2515",
		Icon:        "git-branch",
		Color:       "#ca8a04",
		Properties: []node.PropertySchema{
			{Name: "spiBus", Label: "SPI Bus", Type: "number", Default: 0, Description: "SPI bus number"},
			{Name: "spiDevice", Label: "SPI Device", Type: "number", Default: 0, Description: "Chip select"},
			{Name: "spiSpeed", Label: "SPI Speed (Hz)", Type: "number", Default: 10000000, Description: "SPI clock speed"},
			{Name: "intPin", Label: "Interrupt Pin", Type: "number", Default: 25, Description: "GPIO interrupt pin"},
			{Name: "oscillator", Label: "Crystal (MHz)", Type: "select", Default: "8", Options: []string{"8", "16", "20"}, Description: "Oscillator frequency"},
			{Name: "bitrate", Label: "CAN Bitrate", Type: "select", Default: "500k", Options: []string{"5k", "10k", "20k", "50k", "100k", "125k", "250k", "500k", "1000k"}, Description: "CAN bus bitrate"},
			{Name: "mode", Label: "Mode", Type: "select", Default: "normal", Options: []string{"normal", "loopback", "listen"}, Description: "Operating mode"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "CAN frame to send"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Received CAN frame"},
		},
		Factory: func() node.Executor { return &MCP2515Executor{} },
	}); err != nil {
		return err
	}

	// ============================================
	// CAMERA NODE
	// ============================================

	// Pi Camera (libcamera / rpicam / raspistill)
	if err := registry.Register(&node.NodeInfo{
		Type:        "pi-camera",
		Name:        "Pi Camera",
		Category:    node.NodeTypeInput,
		Description: "Capture photos and video from Raspberry Pi Camera (CSI/USB)",
		Icon:        "camera",
		Color:       "#e11d48",
		Properties: []node.PropertySchema{
			{Name: "mode", Label: "Mode", Type: "select", Default: "photo", Options: []string{"photo", "video", "detect"}, Description: "Capture mode"},
			{Name: "width", Label: "Width", Type: "number", Default: 1920, Description: "Image width in pixels"},
			{Name: "height", Label: "Height", Type: "number", Default: 1080, Description: "Image height in pixels"},
			{Name: "quality", Label: "Quality", Type: "number", Default: 85, Description: "JPEG quality (1-100)"},
			{Name: "format", Label: "Format", Type: "select", Default: "jpeg", Options: []string{"jpeg", "png", "bmp"}, Description: "Image format"},
			{Name: "rotation", Label: "Rotation", Type: "select", Default: "0", Options: []string{"0", "90", "180", "270"}, Description: "Image rotation degrees"},
			{Name: "hflip", Label: "Horizontal Flip", Type: "boolean", Default: false, Description: "Flip image horizontally"},
			{Name: "vflip", Label: "Vertical Flip", Type: "boolean", Default: false, Description: "Flip image vertically"},
			{Name: "exposure", Label: "Exposure", Type: "select", Default: "auto", Options: []string{"auto", "short", "long"}, Description: "Exposure mode"},
			{Name: "awb", Label: "White Balance", Type: "select", Default: "auto", Options: []string{"auto", "daylight", "cloudy", "tungsten", "fluorescent"}, Description: "Auto white balance mode"},
			{Name: "duration", Label: "Video Duration (s)", Type: "number", Default: 5, Description: "Video recording duration in seconds"},
			{Name: "outputDir", Label: "Output Directory", Type: "string", Default: "/tmp/edgeflow-camera", Description: "Directory for captured files"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger capture"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Captured image/video data"},
		},
		Factory: func() node.Executor { return NewCameraExecutor() },
	}); err != nil {
		return err
	}

	// ============================================
	// AUDIO NODE
	// ============================================

	// Audio (ALSA arecord/aplay)
	if err := registry.Register(&node.NodeInfo{
		Type:        "audio",
		Name:        "Audio",
		Category:    node.NodeTypeInput,
		Description: "Record and play audio via ALSA (microphone, speaker, USB audio)",
		Icon:        "mic",
		Color:       "#7c3aed",
		Properties: []node.PropertySchema{
			{Name: "operation", Label: "Operation", Type: "select", Default: "detect", Options: []string{"record", "play", "detect", "volume"}, Description: "Audio operation"},
			{Name: "device", Label: "ALSA Device", Type: "string", Default: "default", Description: "ALSA device name (e.g., hw:0,0, plughw:0,0, default)"},
			{Name: "format", Label: "Format", Type: "select", Default: "wav", Options: []string{"wav", "raw"}, Description: "Audio format"},
			{Name: "sampleRate", Label: "Sample Rate", Type: "select", Default: "44100", Options: []string{"8000", "16000", "22050", "44100", "48000"}, Description: "Sample rate in Hz"},
			{Name: "channels", Label: "Channels", Type: "select", Default: "1", Options: []string{"1", "2"}, Description: "Audio channels (mono/stereo)"},
			{Name: "bitDepth", Label: "Bit Depth", Type: "select", Default: "16", Options: []string{"8", "16", "24", "32"}, Description: "Sample bit depth"},
			{Name: "duration", Label: "Duration (s)", Type: "number", Default: 5, Description: "Recording duration in seconds"},
			{Name: "volume", Label: "Volume (%)", Type: "number", Default: 75, Description: "Playback volume (0-100)"},
			{Name: "outputDir", Label: "Output Directory", Type: "string", Default: "/tmp/edgeflow-audio", Description: "Directory for recordings"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger or audio file path"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Audio data or device info"},
		},
		Factory: func() node.Executor { return NewAudioExecutor() },
	}); err != nil {
		return err
	}

	// ============================================
	// INTERRUPT & 1-WIRE NODES (2 nodes)
	// ============================================

	// GPIO Interrupt
	if err := registry.Register(&node.NodeInfo{
		Type:        "interrupt",
		Name:        "GPIO Interrupt",
		Category:    node.NodeTypeInput,
		Description: "Monitor GPIO pin edge events with interrupt counting",
		Icon:        "zap",
		Color:       "#dc2626",
		Properties: []node.PropertySchema{
			{Name: "pin", Label: "GPIO Pin", Type: "number", Default: 0, Required: true, Description: "BCM GPIO pin number"},
			{Name: "edge", Label: "Edge", Type: "select", Default: "both", Options: []string{"rising", "falling", "both"}, Description: "Edge detection mode"},
			{Name: "debounceMs", Label: "Debounce (ms)", Type: "number", Default: 50, Description: "Debounce time in milliseconds"},
			{Name: "pullMode", Label: "Pull Mode", Type: "select", Default: "none", Options: []string{"none", "up", "down"}, Description: "Internal pull resistor"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Interrupt event (pin, state, count, edge_type, timestamp)"},
		},
		Factory: func() node.Executor { return NewInterruptNode() },
	}); err != nil {
		return err
	}

	// 1-Wire
	if err := registry.Register(&node.NodeInfo{
		Type:        "one-wire",
		Name:        "1-Wire",
		Category:    node.NodeTypeInput,
		Description: "1-Wire bus for DS18B20 temperature sensors and other devices",
		Icon:        "thermometer",
		Color:       "#0891b2",
		Properties: []node.PropertySchema{
			{Name: "busPath", Label: "Bus Path", Type: "string", Default: "/sys/bus/w1/devices/", Description: "1-Wire sysfs bus path"},
			{Name: "deviceId", Label: "Device ID", Type: "string", Default: "", Description: "1-Wire device ID (auto-detect if empty)"},
			{Name: "operation", Label: "Operation", Type: "select", Default: "read_temperature", Options: []string{"scan", "read", "read_temperature"}, Description: "1-Wire operation"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Trigger", Type: "any", Description: "Trigger reading"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "1-Wire device data"},
		},
		Factory: func() node.Executor { return NewOneWireNode() },
	}); err != nil {
		return err
	}

	return nil
}
