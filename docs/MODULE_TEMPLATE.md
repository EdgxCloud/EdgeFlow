# EdgeFlow Module Template

This document explains how to create custom EdgeFlow modules that can be installed via `.tgz` files and run on Raspberry Pi.

## Module Structure

```
my-custom-module/
├── edgeflow.json          # Module manifest (required)
├── LICENSE                 # License file (recommended)
├── README.md              # Documentation (optional)
├── main.go                # Entry point (if Go-based)
├── nodes/                 # Node implementations
│   ├── my_sensor.go
│   └── my_actuator.go
└── bin/                   # Pre-compiled binaries (optional)
    ├── linux-arm64/
    │   └── my-module
    └── linux-arm/
        └── my-module
```

## edgeflow.json Manifest

```json
{
  "name": "my-custom-module",
  "version": "1.0.0",
  "description": "A custom EdgeFlow module for my sensors",
  "author": "Your Name <your@email.com>",
  "license": "MIT",
  "homepage": "https://github.com/yourname/my-custom-module",
  "repository": "https://github.com/yourname/my-custom-module",
  "keywords": ["iot", "sensor", "raspberry-pi"],
  "platform": "raspberry-pi",
  "arch": ["arm64", "arm"],
  "go_version": "1.21",
  "entry_point": "main.go",
  "nodes": [
    {
      "type": "my-temperature-sensor",
      "name": "My Temperature Sensor",
      "category": "sensors",
      "description": "Reads temperature from my custom sensor",
      "icon": "thermometer",
      "color": "#4CAF50",
      "inputs": [
        {
          "name": "trigger",
          "label": "Trigger",
          "type": "any",
          "description": "Trigger a reading"
        }
      ],
      "outputs": [
        {
          "name": "temperature",
          "label": "Temperature",
          "type": "number",
          "description": "Temperature value in Celsius"
        }
      ],
      "properties": [
        {
          "name": "i2cBus",
          "label": "I2C Bus",
          "type": "select",
          "default": "/dev/i2c-1",
          "options": ["/dev/i2c-0", "/dev/i2c-1"],
          "description": "I2C bus to use"
        },
        {
          "name": "address",
          "label": "I2C Address",
          "type": "string",
          "default": "0x48",
          "required": true,
          "description": "Device I2C address"
        },
        {
          "name": "interval",
          "label": "Read Interval (ms)",
          "type": "number",
          "default": 1000,
          "description": "Polling interval in milliseconds"
        }
      ]
    }
  ]
}
```

## Creating the Module Package

### Option 1: Using npm pack (for npm-style packaging)

```bash
cd my-custom-module
npm pack
# Creates: my-custom-module-1.0.0.tgz
```

### Option 2: Using tar directly

```bash
cd /path/to/modules
tar -czvf my-custom-module.tgz my-custom-module/
```

### Option 3: From Go project

```bash
cd my-custom-module
# Build for Raspberry Pi
GOOS=linux GOARCH=arm64 go build -o bin/linux-arm64/my-module .
GOOS=linux GOARCH=arm go build -o bin/linux-arm/my-module .

# Create package
tar -czvf my-custom-module.tgz \
  edgeflow.json \
  LICENSE \
  README.md \
  bin/
```

## Platform Values

| Platform | Description |
|----------|-------------|
| `all` | Any platform |
| `raspberry-pi` | Raspberry Pi (any model) |
| `linux` | Any Linux system |
| `windows` | Windows systems |

## Architecture Values

| Arch | Description |
|------|-------------|
| `arm64` | 64-bit ARM (Pi 4, Pi 5) |
| `arm` | 32-bit ARM (Pi 3, Pi Zero) |
| `amd64` | 64-bit x86 |
| `all` | Any architecture |

## Node Categories

Available categories for nodes:

- `input` - Data input nodes
- `output` - Data output nodes
- `function` - Logic/transformation
- `processing` - Data processing
- `gpio` - GPIO control
- `sensors` - Sensor interfaces
- `actuators` - Actuator control
- `communication` - Communication protocols
- `network` - Network operations
- `database` - Database operations
- `storage` - Storage operations
- `messaging` - Message queues
- `ai` - AI/ML operations
- `industrial` - Industrial protocols
- `dashboard` - Dashboard widgets
- `advanced` - Advanced features

## Property Types

| Type | Description | Example |
|------|-------------|---------|
| `string` | Text input | `"Hello"` |
| `number` | Numeric input | `123` |
| `boolean` | Checkbox | `true` |
| `select` | Dropdown (requires `options`) | `"option1"` |
| `json` | JSON editor | `{"key": "value"}` |
| `password` | Password input | `"secret"` |
| `textarea` | Multi-line text | `"Line1\nLine2"` |

## Installing the Module

1. Go to **Settings > Modules** in the EdgeFlow web interface
2. Click **Upload Module**
3. Select your `.tgz` file
4. Click **Install**

The module will be validated and installed. After installation:
- Nodes appear in the Node Palette under their categories
- Properties are shown in the Node Configuration panel

## Example: Simple Temperature Sensor Node

```go
// nodes/temperature_sensor.go
package nodes

import (
    "context"
    "github.com/edgeflow/edgeflow/internal/node"
)

type TemperatureSensorConfig struct {
    I2CBus   string `json:"i2cBus"`
    Address  string `json:"address"`
    Interval int    `json:"interval"`
}

type TemperatureSensorExecutor struct {
    config TemperatureSensorConfig
}

func NewTemperatureSensorExecutor(config map[string]interface{}) (node.Executor, error) {
    // Parse config and create executor
    return &TemperatureSensorExecutor{}, nil
}

func (e *TemperatureSensorExecutor) Init(config map[string]interface{}) error {
    return nil
}

func (e *TemperatureSensorExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
    // Read temperature from sensor
    temperature := 25.5 // Example value

    return node.Message{
        Payload: map[string]interface{}{
            "temperature": temperature,
            "unit":        "celsius",
        },
    }, nil
}

func (e *TemperatureSensorExecutor) Cleanup() error {
    return nil
}
```

## Deploying to Raspberry Pi

1. Build for Raspberry Pi architecture:
   ```bash
   GOOS=linux GOARCH=arm64 go build -o bin/my-module .
   ```

2. Package the module:
   ```bash
   tar -czvf my-module.tgz edgeflow.json bin/my-module
   ```

3. Upload to EdgeFlow running on Raspberry Pi

4. The module will be automatically loaded and nodes will be available in the flow editor
