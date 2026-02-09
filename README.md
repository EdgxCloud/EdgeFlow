# EdgeFlow

<div align="center">

![EdgeFlow Logo](docs/images/logo.png)

**Lightweight Visual Automation Platform for Edge & IoT**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Raspberry%20Pi-C51A4A?logo=raspberrypi)](https://www.raspberrypi.org/)

</div>

---

## What is EdgeFlow?

EdgeFlow is a visual automation platform designed to run on edge devices like Raspberry Pi. It combines the simplicity of n8n with the hardware capabilities of Node-RED, built in Go for superior performance and minimal resource usage.

### Key Features

| Feature | Description |
|---------|-------------|
| **Lightweight** | Runs on Pi Zero with 512MB RAM |
| **Native GPIO** | Direct hardware access (GPIO, I2C, SPI, PWM) |
| **Visual Editor** | Drag & drop workflow designer |
| **Fast** | Written in Go, sub-second startup |
| **100+ Nodes** | HTTP, MQTT, Telegram, GPIO, AI, Database, Industrial |
| **Modular** | Install only what you need (minimal / standard / full) |
| **Multilingual** | English and Farsi UI support |

---

## Comparison

| Feature | EdgeFlow Minimal | EdgeFlow Full | Node-RED | n8n | Home Assistant |
|---------|-----------------|---------------|----------|-----|----------------|
| Binary Size | **10MB** | 35MB | 80MB+ | 200MB+ | 500MB+ |
| Memory (Idle) | **50MB** | 150MB | 150MB | 300MB | 500MB |
| Memory (Load) | **80MB** | 400MB | 250MB+ | 500MB+ | 1GB+ |
| Startup Time | **<1s** | <1s | ~5s | ~10s | ~30s |
| Pi Zero Support | **Yes** | No | Partial | No | No |
| Modular Install | **Yes** | Yes | No | No | No |
| Native GPIO | Yes | Yes | Yes | No | No |
| Visual Flow Editor | Yes | Yes | Yes | Yes | No |
| AI/LLM Nodes | No | Yes | Partial | Yes | Partial |
| Hot Module Load | **Yes** | Yes | No | No | No |
| Auto Resource Scaling | **Yes** | Yes | No | No | No |

---

## Quick Start

### One-Command Install (Raspberry Pi OS)

The install script handles everything automatically: system dependencies, Go 1.24, Node.js 20, cloning, building, and systemd service setup.

```bash
# Standard install (recommended for Pi 3/4/5)
curl -fsSL http://192.168.1.63/f.hosseini/edgeflow/raw/master/scripts/install-raspberry.sh | sudo bash

# Or with wget
wget -qO- http://192.168.1.63/f.hosseini/edgeflow/raw/master/scripts/install-raspberry.sh | sudo bash

# Choose a profile
curl -fsSL http://192.168.1.63/f.hosseini/edgeflow/raw/master/scripts/install-raspberry.sh | sudo bash -s -- --profile minimal   # Pi Zero (512MB)
curl -fsSL http://192.168.1.63/f.hosseini/edgeflow/raw/master/scripts/install-raspberry.sh | sudo bash -s -- --profile full      # Pi 4/5 (2GB+)

# Install via HTTP clone (no SSH key required)
curl -fsSL http://192.168.1.63/f.hosseini/edgeflow/raw/master/scripts/install-raspberry.sh | sudo bash -s -- --http
```

The installer will:
1. Install system dependencies (git, build-essential, gcc, sqlite3)
2. Install Go 1.24 and Node.js 20
3. Generate an SSH key (if needed)
4. Clone the repository
5. Build the backend and frontend
6. Create and enable a systemd service

### Manual Installation (Step by Step)

```bash
# 1. Update system
sudo apt update && sudo apt upgrade -y

# 2. Install dependencies
sudo apt install -y git build-essential gcc make curl wget sqlite3

# 3. Install Go 1.24
wget https://go.dev/dl/go1.24.0.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-arm64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 4. Install Node.js 20
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# 5. Generate SSH key (if needed)
ssh-keygen -t ed25519 -C "pi@raspberrypi"
cat ~/.ssh/id_ed25519.pub
# Add the public key to your Git server at http://192.168.1.63

# 6. Clone the repository
git clone git@192.168.1.63:f.hosseini/edgeflow.git ~/edgeflow
cd ~/edgeflow

# 7. Build the backend
go mod download
make build PROFILE=standard

# 8. Build the frontend
cd web && npm install && npm run build && cd ..

# 9. Run
./bin/edgeflow
```

### Docker

```bash
docker run -d -p 8080:8080 -v /sys:/sys --privileged edgeflow/edgeflow
```

### Access the Web UI

Once running, open your browser:

```
http://localhost:8080
http://<raspberry-pi-ip>:8080
```

---

## Installation Profiles

EdgeFlow supports three build profiles optimized for different hardware:

| Profile | Binary Size | RAM Usage | Target Devices | Modules |
|---------|------------|-----------|----------------|---------|
| `minimal` | ~10MB | ~50MB | Pi Zero, BeagleBone | Core |
| `standard` | ~20MB | ~200MB | Pi 3/4, Orange Pi | Core, Network, GPIO, Database |
| `full` | ~35MB | ~400MB | Pi 4/5, Jetson Nano | All modules |

```bash
# Build with a specific profile
make build PROFILE=minimal
make build PROFILE=standard   # default
make build PROFILE=full

# Cross-compile for Raspberry Pi
make build-pi-minimal          # Pi Zero (ARM64)
make build-pi-standard         # Pi 3/4 (ARM64)
make build-pi-full             # Pi 4/5 (ARM64)
```

---

## Service Management

After installation, EdgeFlow runs as a systemd service:

```bash
sudo systemctl status edgeflow      # Check status
sudo journalctl -u edgeflow -f      # View logs
sudo systemctl restart edgeflow     # Restart
sudo systemctl stop edgeflow        # Stop
sudo systemctl start edgeflow       # Start
```

---

## Node Modules

EdgeFlow includes 100+ nodes organized into categories:

| Module | Nodes | Description |
|--------|-------|-------------|
| **Core** | inject, debug, function, switch, delay, change, template | Essential flow control |
| **Network** | HTTP, MQTT, WebSocket, TCP, UDP | Network communication |
| **GPIO** | GPIO In/Out, PWM, I2C, SPI, sensors (DHT, DS18B20, BMP280) | Hardware control |
| **Database** | SQLite, MySQL, PostgreSQL, MongoDB, Redis, InfluxDB | Data storage |
| **Messaging** | Telegram, Email, Slack, Discord, SMS, Webhooks | Notifications |
| **AI** | OpenAI, Anthropic, Ollama, TensorFlow Lite | AI/ML integration |
| **Industrial** | Modbus TCP/RTU, OPC-UA | Industrial protocols |
| **Dashboard** | Gauges, Charts, Buttons, Text, Sliders | Real-time UI widgets |

---

## Project Structure

```
edgeflow/
├── cmd/edgeflow/              # Application entry point
├── internal/
│   ├── engine/                # Core workflow engine
│   ├── api/                   # HTTP API server (Fiber)
│   ├── node/                  # Node registry & execution
│   ├── storage/               # SQLite storage layer
│   ├── websocket/             # Real-time WebSocket hub
│   ├── config/                # Configuration management
│   ├── security/              # JWT auth, API keys, encryption
│   ├── health/                # Health check endpoints
│   ├── metrics/               # Prometheus metrics
│   ├── hal/                   # Hardware abstraction layer
│   └── plugin/                # Plugin system
├── pkg/nodes/
│   ├── core/                  # Core nodes
│   ├── network/               # HTTP, MQTT, WebSocket, TCP/UDP
│   ├── gpio/                  # GPIO, I2C, SPI, sensors
│   ├── database/              # SQL and NoSQL nodes
│   ├── messaging/             # Telegram, Email, Slack, Discord
│   ├── ai/                    # OpenAI, Anthropic, Ollama
│   ├── industrial/            # Modbus, OPC-UA
│   ├── dashboard/             # Dashboard widgets
│   └── parser/                # HTML/XML/JSON parsing
├── web/                       # React + TypeScript frontend
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── configs/                   # YAML configuration profiles
├── scripts/                   # Build & install scripts
├── Makefile                   # Build system
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

---

## Development

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.24+ | Backend |
| Node.js | 18+ | Frontend |
| Make | any | Build system |
| Docker | any | Optional |

### Development Setup

```bash
# Clone
git clone git@192.168.1.63:f.hosseini/edgeflow.git
cd edgeflow

# Install Go dependencies
go mod download

# Install frontend dependencies
cd web && npm install && cd ..

# Run in development mode (hot reload)
make dev

# Or run frontend dev server separately
make frontend-dev    # http://localhost:3000
```

### Build Commands

```bash
make build                     # Build standard profile
make build PROFILE=full        # Build full profile
make build-all-profiles        # Build all profiles (ARM64)
make build-all-platforms       # Build for all OS/arch combinations
make test                      # Run tests
make lint                      # Run linter
make frontend-build            # Build frontend for production
make clean                     # Remove build artifacts
make help                      # Show all available targets
```

---

## Configuration

EdgeFlow is configured via YAML files in the `configs/` directory:

```yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  driver: sqlite
  path: ./data/edgeflow.db

logging:
  level: info          # debug, info, warn, error
  format: json

security:
  enabled: false       # Enable in production
  jwt_secret: "change-me-in-production"

hardware:
  gpio_enabled: false  # Auto-detected on Raspberry Pi
  i2c_enabled: false
  spi_enabled: false

scheduler:
  enabled: true
  timezone: "UTC"
```

Environment variables override config values:

```bash
EDGEFLOW_SERVER_PORT=8080
EDGEFLOW_SERVER_HOST=0.0.0.0
EDGEFLOW_LOGGING_LEVEL=info
```

---

## Roadmap

### v0.1.0 - Core MVP (Complete)
- [x] Core workflow engine
- [x] 29 nodes (Core, Network, Hardware, Integration)
- [x] Web UI with React + TypeScript
- [x] REST API
- [x] WebSocket real-time updates
- [x] SQLite storage

### v0.2.0 - Hardware Support (Complete)
- [x] GPIO support (In/Out, PWM)
- [x] I2C support
- [x] Hardware abstraction layer
- [x] Sensor nodes (DHT, DS18B20)
- [x] Actuator nodes (Relay)

### v0.3.0 - Integrations (Complete)
- [x] Messaging nodes (Telegram, Email, Slack, Discord)
- [x] Database nodes (MySQL, PostgreSQL, MongoDB, Redis)
- [x] AI nodes (OpenAI, Anthropic, Ollama)
- [x] Network nodes (HTTP, MQTT, WebSocket, TCP, UDP)

### v0.4.0 - Production Ready (80% Complete)
- [x] JWT Authentication
- [x] API Key management
- [x] Credential encryption
- [x] Prometheus metrics
- [x] Health checks
- [x] Docker deployment
- [x] Modular architecture (IoT-optimized)
- [x] Resource monitoring & auto-scaling
- [x] Installation profiles (minimal/standard/full)
- [x] API documentation
- [ ] Test suite (in progress)
- [ ] Raspberry Pi OS image

### v1.0.0 - Release (Planned)
- [ ] 50+ additional nodes
- [ ] Multi-user support
- [ ] Execution history & debugging
- [ ] Video tutorials

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

```bash
# Fork and clone
git clone git@192.168.1.63:YOUR_USERNAME/edgeflow.git

# Create a feature branch
git checkout -b feature/my-feature

# Make changes, commit, push
git commit -m "Add my feature"
git push origin feature/my-feature

# Open a Merge Request
```

---

## License

This project is licensed under [Apache 2.0](LICENSE).

---

## Acknowledgments

- [Go](https://golang.org) - Programming language
- [Fiber](https://gofiber.io) - HTTP framework
- [React Flow / XYFlow](https://reactflow.dev) - Flow editor library
- [periph.io](https://periph.io) - Hardware abstraction for Go
- [Vite](https://vitejs.dev) - Frontend build tool

---

<div align="center">

Made with care in Iran

[Website](https://edgeflow.io) | [Documentation](https://docs.edgeflow.io) | [Discord](https://discord.gg/edgeflow)

</div>
