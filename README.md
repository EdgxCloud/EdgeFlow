<a id="readme-top"></a>

<p align="center">
  <img src="docs/images/edgeflow-logo.png" alt="EdgeFlow" width="200">
</p>

<h1 align="center">EdgeFlow Platform</h1>

<p align="center">
  <strong>Lightweight, visual IoT automation platform for Raspberry Pi and edge devices</strong>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg" alt="License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white" alt="Go"></a>
  <a href="https://react.dev"><img src="https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=white" alt="React"></a>
  <a href="https://github.com/edgeflow/edgeflow/stargazers"><img src="https://img.shields.io/github/stars/edgeflow/edgeflow?style=flat" alt="Stars"></a>
  <a href="https://github.com/edgeflow/edgeflow/issues"><img src="https://img.shields.io/github/issues/edgeflow/edgeflow" alt="Issues"></a>
</p>

<p align="center">
  <a href="#features">Features</a> &middot;
  <a href="#quick-start">Quick Start</a> &middot;
  <a href="#tech-stack">Tech Stack</a> &middot;
  <a href="#build-profiles">Build Profiles</a> &middot;
  <a href="#contributing">Contributing</a>
</p>

---

<!-- Add your screenshot here -->
<!-- <p align="center"><img src="docs/images/screenshot-editor.png" alt="EdgeFlow Editor" width="800"></p> -->

## About

EdgeFlow is a node-based automation platform that lets you build IoT workflows visually. Connect sensors, process data, trigger actions, and monitor everything from a single dashboard — **no cloud required**.

- **Single binary** — deploy one Go binary + static frontend, no runtime dependencies
- **Built for constrained devices** — runs on a Raspberry Pi Zero with ~50 MB RAM
- **150+ built-in nodes** — GPIO, MQTT, Modbus, databases, AI, and more out of the box
- **Hardware-first** — native GPIO support for Pi 4/5 via Linux character device, not memory-mapped hacks

Think Node-RED, but lighter, faster, and purpose-built for edge devices.

## Features

**Visual Flow Editor**
- Drag-and-drop node canvas with 150+ pre-built node types
- Real-time execution tracing and debug panel
- Sub-flow support for reusable logic
- Built-in terminal, logs, and system monitoring panels

**Hardware Integration**
- Raspberry Pi 4/5 GPIO via Linux character device (`gpiocdev`)
- I2C sensors (BME280, BMP280, BH1750, AHT20, ADS1015, and more)
- PIR motion, HC-SR04 ultrasonic, relay, LED, button
- Software PWM and edge detection

**Industrial & Wireless Protocols**
- MQTT, Modbus TCP/RTU, OPC-UA, BACnet, CAN Bus, PROFINET
- BLE, Zigbee, Z-Wave, LoRa, NFC, RFID, RF433, IR

**Integrations**
- AI/LLM: OpenAI, Anthropic, Ollama
- Databases: SQLite, PostgreSQL, MySQL, MongoDB, Redis, InfluxDB
- Cloud Storage: AWS S3, Google Drive, Dropbox, OneDrive
- Messaging: Email, Telegram, Slack, Discord
- HTTP webhooks and REST APIs

**Monitoring & Debugging**
- CPU, memory, disk, temperature, network speed
- GPIO state visualization
- Execution history with structured logs
- Web-based interactive terminal

## Quick Start

### Prerequisites

- Go 1.24+
- Node.js 18+
- Git

### Run Locally

```bash
# Clone
git clone https://github.com/edgeflow/edgeflow.git
cd edgeflow

# Build & run backend
make run

# In another terminal — build frontend
cd web
npm install
npm run dev
```

Open `http://localhost:5173` (dev) or `http://localhost:8080` (production build).

### Docker

```bash
docker compose up -d
```

### Raspberry Pi

```bash
# Build for ARM64
make build-pi-standard

# Install as service
sudo cp deploy/edgeflow-backend.service /etc/systemd/system/
sudo systemctl enable --now edgeflow-backend
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24, Fiber v2, Zap logging |
| Frontend | React 18, TypeScript, Vite, Tailwind CSS |
| Flow Canvas | React Flow (`@xyflow/react`) |
| Code Editor | Monaco Editor |
| UI Components | Radix UI |
| State | Zustand |
| Hardware | `go-gpiocdev` (Linux character device) |
| Protocols | MQTT (Paho), Modbus, OPC-UA |
| Databases | SQLite, PostgreSQL, MySQL, MongoDB, Redis, InfluxDB |
| Auth | JWT, API key |

## Node Categories

150+ nodes across 10 categories:

| Category | Count | Examples |
|----------|-------|---------|
| Core | 25+ | inject, debug, function, switch, template, delay, trigger, filter |
| GPIO | 15+ | digital in/out, PWM, PIR, HC-SR04, DHT, BMP280, servo |
| Network | 10+ | HTTP request, webhook, TCP, UDP, FTP, CSV/JSON parser |
| Database | 6 | SQLite, PostgreSQL, MySQL, MongoDB, Redis, InfluxDB |
| Messaging | 4 | Email, Telegram, Slack, Discord |
| AI/ML | 3 | OpenAI, Anthropic, Ollama |
| Industrial | 5 | Modbus TCP/RTU, OPC-UA, BACnet, CAN Bus, PROFINET |
| Wireless | 10+ | BLE, Zigbee, Z-Wave, LoRa, NFC, RFID, RF433, IR |
| Dashboard | 12 | Chart, gauge, button, slider, switch, table, form, template |
| Cloud Storage | 5 | AWS S3, Google Drive, Dropbox, OneDrive, SFTP |

## Build Profiles

Three build profiles optimized for different hardware:

| Profile | Target | Binary Size | RAM Usage |
|---------|--------|------------|-----------|
| Minimal | Pi Zero, BeagleBone | ~10 MB | ~50 MB |
| Standard | Pi 3, Orange Pi | ~20 MB | ~200 MB |
| Full | Pi 4/5, Jetson | ~35 MB | ~400 MB |

```bash
make build PROFILE=minimal    # Pi Zero / constrained devices
make build PROFILE=standard   # Pi 3/4 / general use
make build PROFILE=full       # Pi 4/5 / all features
```

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
EDGEFLOW_SERVER_HOST=0.0.0.0
EDGEFLOW_SERVER_PORT=8080
EDGEFLOW_DATABASE_TYPE=sqlite
EDGEFLOW_DATABASE_PATH=./data/edgeflow.db
EDGEFLOW_LOGGER_LEVEL=info
```

<details>
<summary><strong>Project Structure</strong></summary>

```
EdgeFlow Platform
├── cmd/edgeflow/          # Application entry point
├── internal/
│   ├── api/               # REST API handlers (Fiber)
│   ├── engine/            # Flow execution engine & scheduler
│   ├── hal/               # Hardware Abstraction Layer (GPIO, I2C, SPI, Serial)
│   ├── node/              # Node registry & message framework
│   ├── storage/           # File, SQLite, Redis backends
│   ├── websocket/         # Real-time WebSocket hub
│   ├── resources/         # System monitoring (CPU, memory, temp)
│   ├── security/          # JWT & API key auth
│   ├── logger/            # Structured logging (Zap)
│   ├── plugin/            # Plugin system
│   └── subflow/           # Nested flow support
├── pkg/nodes/
│   ├── core/              # Inject, debug, function, switch, template, delay...
│   ├── gpio/              # PIR, HC-SR04, LED, relay, button, sensors
│   ├── network/           # HTTP request/webhook, FTP, CSV/JSON parsers
│   ├── database/          # SQLite, PostgreSQL, MySQL, MongoDB, Redis, InfluxDB
│   ├── messaging/         # Email, Telegram, Slack, Discord
│   ├── ai/                # OpenAI, Anthropic, Ollama
│   ├── industrial/        # Modbus, OPC-UA, BACnet, CAN Bus, PROFINET
│   ├── wireless/          # BLE, Zigbee, Z-Wave, LoRa, NFC, RF433, IR
│   ├── storage/           # S3, Google Drive, Dropbox, OneDrive, SFTP
│   └── dashboard/         # UI widgets (chart, gauge, table, form, button...)
├── web/                   # React + TypeScript frontend (Vite)
├── configs/               # Default configuration
├── deploy/                # Systemd service files
└── docker-compose.yml
```

</details>

<details>
<summary><strong>REST API</strong></summary>

Base URL: `/api/v1`

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/flows` | GET, POST | List / create flows |
| `/flows/:id` | GET, PUT, DELETE | Get / update / delete flow |
| `/flows/:id/start` | POST | Start flow execution |
| `/flows/:id/stop` | POST | Stop flow |
| `/flows/:id/nodes` | GET, POST | List / add nodes |
| `/flows/:id/connections` | GET, POST | List / add connections |
| `/node-types` | GET | List all 150+ node types |
| `/executions` | GET | Execution history |
| `/resources/stats` | GET | CPU, memory, disk stats |
| `/system/info` | GET | System info, temperature, uptime |
| `/modules` | GET | Installed modules |
| `/modules/install` | POST | Install a module |
| `/health` | GET | Health check |

**WebSocket:** `/ws` (logs, debug, execution events), `/ws/terminal` (interactive shell)

</details>

## Roadmap

- [ ] Plugin marketplace
- [ ] Flow versioning and rollback
- [ ] Multi-device orchestration
- [ ] Mobile companion app
- [ ] Visual rule engine for complex conditions

See the [open issues](https://github.com/edgeflow/edgeflow/issues) for a full list of proposed features and known issues.

## Contributing

Contributions are welcome! Whether it's bug reports, feature requests, or pull requests — all contributions help make EdgeFlow better.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Community

- [GitHub Issues](https://github.com/edgeflow/edgeflow/issues) — Bug reports and feature requests
- [GitHub Discussions](https://github.com/edgeflow/edgeflow/discussions) — Questions and ideas

<!-- Uncomment when available:
- [Discord](https://discord.gg/edgeflow) — Chat with the community
-->

## License

Distributed under the Apache License 2.0. See [LICENSE](LICENSE) for more information.

Copyright 2024-2026 EdgeFlow

<p align="right">(<a href="#readme-top">back to top</a>)</p>
