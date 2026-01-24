# EdgeFlow Build System

EdgeFlow features a modular build system that allows you to compile different versions of the platform optimized for specific IoT boards and use cases.

## Overview

The build system supports three main profiles:

| Profile | Binary Size | RAM Usage | Target Boards | Modules |
|---------|-------------|-----------|---------------|---------|
| **Minimal** | ~10MB | ~50MB | Pi Zero, BeagleBone (512MB RAM) | Core only |
| **Standard** | ~20MB | ~200MB | Pi 3/4, Orange Pi (1GB+ RAM) | Core, Network, GPIO, Database |
| **Full** | ~35MB | ~400MB | Pi 4/5, Jetson Nano (2GB+ RAM) | All modules |

## Quick Start

### Using Make (Recommended)

```bash
# Build with default (standard) profile
make build

# Build for minimal profile
make build PROFILE=minimal

# Build for full profile
make build PROFILE=full

# Build for Raspberry Pi Zero (minimal, ARM64)
make build-pi-minimal

# Build for Raspberry Pi 3/4 (standard, ARM64)
make build-pi-standard

# Build for Raspberry Pi 4/5 (full, ARM64)
make build-pi-full

# Build all profiles for all platforms
make build-all-profiles
```

### Using Build Scripts

**Linux/macOS:**
```bash
# Make script executable
chmod +x scripts/build.sh

# Build minimal for Pi Zero
./scripts/build.sh minimal linux-arm64

# Build standard for Pi 3/4
./scripts/build.sh standard linux-arm64

# Build all profiles for all platforms
./scripts/build.sh all all
```

**Windows (PowerShell):**
```powershell
# Build minimal for Pi Zero
.\scripts\build.ps1 minimal linux-arm64

# Build standard for Pi 3/4
.\scripts\build.ps1 standard linux-arm64

# Build all profiles
.\scripts\build.ps1 all all
```

## Make Targets

### Build Targets

- `make build PROFILE=[minimal|standard|full]` - Build for current platform
- `make build-pi-minimal` - Build minimal profile for Pi Zero (ARM64)
- `make build-pi-standard` - Build standard profile for Pi 3/4 (ARM64)
- `make build-pi-full` - Build full profile for Pi 4/5 (ARM64)
- `make build-pi32-minimal` - Build minimal profile for Pi Zero (ARM32)
- `make build-all-profiles` - Build all profiles for all platforms

### Module Information

- `make modules-list` - List available modules
- `make modules-info` - Show module resource requirements

### Development

- `make run PROFILE=[profile]` - Build and run
- `make dev` - Run with hot reload (using air)
- `make test` - Run tests with current profile
- `make lint` - Run linter
- `make fmt` - Format code

### Frontend

- `make frontend-install` - Install frontend dependencies
- `make frontend-build` - Build frontend for production
- `make frontend-dev` - Run frontend dev server

### Testing

- `make test` - Run unit tests
- `make test-coverage` - Generate coverage report
- `make test-integration` - Run integration tests
- `make bench` - Run benchmarks

### Utilities

- `make clean` - Clean build artifacts
- `make deps` - Download dependencies
- `make install` - Install binary to `/usr/local/bin`
- `make security` - Run security checks

## Build Profiles in Detail

### Minimal Profile

**Target:** Raspberry Pi Zero W/2W, BeagleBone Black/Green

**Features:**
- Core nodes only (inject, debug, function, if, delay)
- No network capabilities
- No database support
- Ultra-lightweight

**Use Cases:**
- Simple sensor reading
- Basic automation
- Learning Node-RED concepts
- Testing on constrained devices

**Build:**
```bash
make build PROFILE=minimal
```

### Standard Profile

**Target:** Raspberry Pi 3B/3B+/4B (2GB), Orange Pi, Banana Pi

**Features:**
- Core nodes
- Network nodes (HTTP, MQTT, WebSocket, TCP/UDP)
- GPIO nodes (I2C, SPI, PWM)
- Database nodes (SQLite, PostgreSQL)

**Use Cases:**
- IoT gateways
- Home automation
- Data logging
- Sensor networks with cloud connectivity

**Build:**
```bash
make build PROFILE=standard
```

### Full Profile

**Target:** Raspberry Pi 4B/5 (4GB+), NVIDIA Jetson Nano, Tinker Board

**Features:**
- All standard features
- Messaging nodes (Telegram, Email, SMS)
- AI/ML nodes (TensorFlow Lite, OpenCV)
- Industrial protocols (Modbus, OPC-UA)
- Advanced scripting

**Use Cases:**
- Advanced home automation
- Industrial IoT
- Edge AI applications
- Computer vision projects
- Complex integrations

**Build:**
```bash
make build PROFILE=full
```

## Cross-Compilation

### Raspberry Pi Targets

```bash
# Pi Zero/3/4/5 (ARM64) - Minimal
GOOS=linux GOARCH=arm64 go build -tags "minimal" \
  -ldflags "-w -s -X main.Version=0.1.0 -X main.Profile=minimal" \
  -trimpath -o bin/edgeflow-minimal-arm64 ./cmd/edgeflow

# Pi 3/4 (ARM64) - Standard
GOOS=linux GOARCH=arm64 go build -tags "standard,network,gpio,database" \
  -ldflags "-w -s -X main.Version=0.1.0 -X main.Profile=standard" \
  -trimpath -o bin/edgeflow-standard-arm64 ./cmd/edgeflow

# Pi 4/5 (ARM64) - Full
GOOS=linux GOARCH=arm64 go build -tags "full,network,gpio,database,messaging,ai,industrial,advanced" \
  -ldflags "-w -X main.Version=0.1.0 -X main.Profile=full" \
  -o bin/edgeflow-full-arm64 ./cmd/edgeflow

# Pi Zero/3 (ARM32)
GOOS=linux GOARCH=arm GOARM=7 go build -tags "minimal" \
  -ldflags "-w -s" -trimpath -o bin/edgeflow-minimal-arm ./cmd/edgeflow
```

### Other Platforms

```bash
# Linux AMD64 (Development)
GOOS=linux GOARCH=amd64 go build -tags "full,network,gpio,database,messaging,ai,industrial,advanced" \
  -o bin/edgeflow-full-amd64 ./cmd/edgeflow

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -tags "full,network,gpio,database,messaging,ai,industrial,advanced" \
  -o bin/edgeflow-full-darwin-arm64 ./cmd/edgeflow

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -tags "standard,network,database" \
  -o bin/edgeflow-standard.exe ./cmd/edgeflow
```

## Build Optimizations

### Size Optimization (Minimal/Standard)

```bash
# Strip debug symbols and reduce binary size
-ldflags "-w -s"

# Remove file paths
-trimpath

# Disable CGO (if possible)
CGO_ENABLED=0

# Further compression with UPX (optional)
upx --best --lzma bin/edgeflow
```

### Performance Optimization (Full)

```bash
# Keep some debug symbols for production debugging
-ldflags "-w"

# Enable optimizations
go build -gcflags="-m -l"
```

## Testing Different Profiles

```bash
# Test minimal profile
make test PROFILE=minimal

# Test standard profile
make test PROFILE=standard

# Test full profile
make test PROFILE=full

# Test specific module
go test -tags "network" ./internal/nodes/network/...
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        profile: [minimal, standard, full]
        platform: [linux-arm64, linux-arm, linux-amd64]

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build
        run: make build PROFILE=${{ matrix.profile }}

      - name: Test
        run: make test PROFILE=${{ matrix.profile }}

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: edgeflow-${{ matrix.profile }}-${{ matrix.platform }}
          path: bin/
```

## Deployment

### Installation on Target Device

```bash
# Download appropriate binary
wget https://github.com/yourusername/edgeflow/releases/download/v0.1.0/edgeflow-standard-linux-arm64.tar.gz

# Extract
tar -xzf edgeflow-standard-linux-arm64.tar.gz

# Make executable
chmod +x edgeflow

# Install systemically (optional)
sudo cp edgeflow /usr/local/bin/

# Run
./edgeflow
```

### Systemd Service

Create `/etc/systemd/system/edgeflow.service`:

```ini
[Unit]
Description=EdgeFlow IoT Platform
After=network.target

[Service]
Type=simple
User=pi
ExecStart=/usr/local/bin/edgeflow
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable edgeflow
sudo systemctl start edgeflow
```

## Troubleshooting

### Build Fails with "undefined: XXX"

**Cause:** Missing build tags for required modules

**Solution:** Include all required tags
```bash
# messaging requires network
make build PROFILE=full
# or
go build -tags "messaging,network" ./cmd/edgeflow
```

### Binary Too Large

**Cause:** Using full profile on constrained device

**Solution:** Use appropriate profile
```bash
# Switch to minimal or standard
make build PROFILE=minimal
```

### Runtime Error: "module not available"

**Cause:** Module not included in build

**Solution:** Rebuild with required module
```bash
make build PROFILE=standard  # includes network, gpio, database
```

## Related Documentation

- [Build Tags Guide](BUILD_TAGS.md) - Detailed build tag documentation
- [IoT Deployment Strategy](IOT_DEPLOYMENT_STRATEGY.md) - Deployment guide
- [Module Development](MODULES.md) - Creating custom modules
- [build.config.yaml](../build.config.yaml) - Build configuration

## Advanced Topics

### Custom Build Profiles

You can create custom builds with specific module combinations:

```bash
# Core + Network + GPIO only (custom lightweight build)
go build -tags "network,gpio" \
  -ldflags "-w -s -X main.Profile=custom" \
  -o bin/edgeflow-custom ./cmd/edgeflow
```

### Conditional Compilation

See [BUILD_TAGS.md](BUILD_TAGS.md) for details on using build tags in your code.

### Module Dependencies

Some modules depend on others:
- `messaging` requires `network`
- `industrial` requires `network`
- `ai` can work standalone but benefits from `network`

Always include dependencies when building custom profiles.
