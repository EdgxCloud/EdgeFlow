# Build Tags Guide

EdgeFlow uses Go build tags to enable conditional compilation of modules based on target board capabilities.

## Overview

Build tags allow you to compile only the modules you need, reducing binary size and memory footprint for resource-constrained IoT boards.

## Available Build Tags

### Profile Tags

- **`minimal`** - Minimal profile (Pi Zero, BeagleBone, 512MB RAM)
  - Includes: Core nodes only
  - Binary: ~10MB
  - RAM: ~50MB

- **`standard`** - Standard profile (Pi 3/4, 1GB+ RAM)
  - Includes: Core, Network, GPIO, Database
  - Binary: ~20MB
  - RAM: ~200MB

- **`full`** - Full profile (Pi 4/5, Jetson Nano, 2GB+ RAM)
  - Includes: All modules
  - Binary: ~35MB
  - RAM: ~400MB

### Module Tags

- **`network`** - HTTP, WebSocket, MQTT, TCP/UDP nodes
- **`gpio`** - GPIO, I2C, SPI, PWM control
- **`database`** - SQLite, PostgreSQL, MongoDB, InfluxDB
- **`messaging`** - Telegram, Email, SMS, Webhook
- **`ai`** - TensorFlow Lite, OpenCV, Image Recognition
- **`industrial`** - Modbus, OPC-UA, industrial protocols
- **`advanced`** - Custom scripting, advanced transformations

## Using Build Tags

### Command Line

```bash
# Build with minimal profile
make build PROFILE=minimal

# Build with standard profile (default)
make build PROFILE=standard

# Build with full profile
make build PROFILE=full

# Build for specific platform
make build-pi-minimal      # Pi Zero ARM64
make build-pi-standard     # Pi 3/4 ARM64
make build-pi-full         # Pi 4/5 ARM64
```

### Manual Go Build

```bash
# Minimal build
go build -tags "minimal" -o bin/edgeflow ./cmd/edgeflow

# Standard build
go build -tags "standard,network,gpio,database" -o bin/edgeflow ./cmd/edgeflow

# Full build
go build -tags "full,network,gpio,database,messaging,ai,industrial,advanced" -o bin/edgeflow ./cmd/edgeflow

# Custom build (core + network + gpio only)
go build -tags "network,gpio" -o bin/edgeflow ./cmd/edgeflow
```

## File Naming Convention

To use build tags in your code, add build constraints at the top of Go files:

```go
//go:build network
// +build network

package nodes

// This file is only compiled when the 'network' build tag is present
```

### Examples

**File: `internal/nodes/network/http_node.go`**
```go
//go:build network
// +build network

package network

// HTTP node implementation
```

**File: `internal/nodes/gpio/gpio_node.go`**
```go
//go:build gpio
// +build gpio

package gpio

// GPIO node implementation
```

**File: `internal/plugin/network_plugin.go`**
```go
//go:build network
// +build network

package plugin

func init() {
    // Register network plugin only when network tag is present
    GetRegistry().Register(NewNetworkPlugin())
}
```

## Module Registration

Each optional module should self-register only when its build tag is present:

**File: `internal/plugin/modules.go`** (always compiled)
```go
package plugin

// Base registration - always included
```

**File: `internal/plugin/network_module.go`** (conditional)
```go
//go:build network
// +build network

package plugin

func init() {
    // Auto-register network module when compiled with network tag
    GetRegistry().Register(NewNetworkPlugin())
}
```

## Testing with Build Tags

```bash
# Test with specific profile
go test -tags "standard,network,gpio,database" ./...

# Test specific module
go test -tags "network" ./internal/nodes/network/...
```

## Profile Matrix

| Profile  | Tags | Binary Size | RAM Usage | Target Boards |
|----------|------|-------------|-----------|---------------|
| Minimal  | `minimal` | ~10MB | ~50MB | Pi Zero, BeagleBone |
| Standard | `standard,network,gpio,database` | ~20MB | ~200MB | Pi 3/4, Orange Pi |
| Full     | `full,network,gpio,database,messaging,ai,industrial,advanced` | ~35MB | ~400MB | Pi 4/5, Jetson Nano |

## Cross-Compilation Examples

```bash
# Pi Zero (ARM64) - Minimal
GOOS=linux GOARCH=arm64 go build -tags "minimal" \
  -ldflags "-w -s" -trimpath -o edgeflow-minimal-arm64 ./cmd/edgeflow

# Pi 3/4 (ARM64) - Standard
GOOS=linux GOARCH=arm64 go build -tags "standard,network,gpio,database" \
  -ldflags "-w -s" -trimpath -o edgeflow-standard-arm64 ./cmd/edgeflow

# Pi 4/5 (ARM64) - Full
GOOS=linux GOARCH=arm64 go build -tags "full,network,gpio,database,messaging,ai,industrial,advanced" \
  -ldflags "-w" -o edgeflow-full-arm64 ./cmd/edgeflow

# Pi Zero (ARM32) - Minimal
GOOS=linux GOARCH=arm GOARM=7 go build -tags "minimal" \
  -ldflags "-w -s" -trimpath -o edgeflow-minimal-arm ./cmd/edgeflow
```

## Best Practices

1. **Always include core**: Core nodes are required for basic functionality
2. **Check dependencies**: Network module requires core, messaging requires network
3. **Test each profile**: Ensure your code compiles with all profile combinations
4. **Document requirements**: Clearly state which modules your custom nodes depend on
5. **Use `//go:build` directive**: The newer directive format (Go 1.17+)

## Module Dependencies

```
core (required)
├── network
│   ├── messaging
│   └── industrial
├── gpio
├── database
├── ai
└── advanced
```

## Troubleshooting

### Issue: Module not available at runtime

**Solution**: Ensure the build tag was included during compilation
```bash
# Check if module was compiled in
strings bin/edgeflow | grep "NetworkPlugin"
```

### Issue: Binary too large

**Solution**: Use minimal or standard profile, or disable unused modules
```bash
# Check binary size
ls -lh bin/edgeflow
```

### Issue: Missing dependencies

**Solution**: Include all required module tags
```bash
# messaging requires network
go build -tags "messaging,network" ./cmd/edgeflow
```

## Adding New Modules

1. Create module directory: `internal/nodes/mymodule/`
2. Add build tag to all files: `//go:build mymodule`
3. Create plugin registration: `internal/plugin/mymodule_plugin.go`
4. Update `build.config.yaml`
5. Update Makefile build tags
6. Document in this file

## Related Files

- `Makefile` - Build system with profile support
- `build.config.yaml` - Module configuration
- `internal/plugin/` - Plugin registration
- `docs/IOT_DEPLOYMENT_STRATEGY.md` - Deployment guide
