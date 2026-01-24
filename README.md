# EdgeFlow

<div align="center">

![EdgeFlow Logo](docs/images/logo.png)

**پلتفرم اتوماسیون سبک برای Edge و IoT**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Raspberry%20Pi-C51A4A?logo=raspberrypi)](https://www.raspberrypi.org/)

[English](README.md) | [فارسی](README.fa.md)

</div>

---

## 🎯 EdgeFlow چیست؟

EdgeFlow یک پلتفرم اتوماسیون visual است که برای اجرا روی دستگاه‌های edge مثل Raspberry Pi طراحی شده. ترکیبی از سادگی n8n و قدرت سخت‌افزاری Node-RED، با performance بسیار بهتر.

### ✨ ویژگی‌های کلیدی

| ویژگی | توضیح |
|--------|-------|
| 🪶 **سبک** | اجرا روی Pi Zero با 512MB RAM |
| 🔌 **GPIO Native** | دسترسی مستقیم به سخت‌افزار |
| 🎨 **Visual Editor** | طراحی workflow با drag & drop |
| ⚡ **سریع** | نوشته شده با Go، نه Node.js |
| 🔄 **۱۰۰+ نود** | HTTP, MQTT, Telegram, GPIO, AI و... |
| 🌐 **فارسی** | اولین پلتفرم با UI فارسی |

---

## 📊 مقایسه با رقبا

| ویژگی | EdgeFlow Minimal | EdgeFlow Full | Node-RED | n8n | Home Assistant |
|--------|-----------------|---------------|----------|-----|----------------|
| Binary Size | **10MB** | 35MB | 80MB+ | 200MB+ | 500MB+ |
| Memory Usage (Idle) | **50MB** | 150MB | 150MB | 300MB | 500MB |
| Memory Usage (Load) | **80MB** | 400MB | 250MB+ | 500MB+ | 1GB+ |
| Startup Time | **<1s** | <1s | ~5s | ~10s | ~30s |
| Pi Zero Compatible | **✅** | ❌ | ⚠️ | ❌ | ❌ |
| Modular Install | **✅** | ✅ | ❌ | ❌ | ❌ |
| GPIO Native | ✅ | ✅ | ✅ | ❌ | ❌ |
| Visual Flow | ✅ | ✅ | ✅ | ✅ | ❌ |
| Business Automation | ✅ | ✅ | ⚠️ | ✅ | ❌ |
| AI/LLM Nodes | ❌ | ✅ | ⚠️ | ✅ | ⚠️ |
| Hot Module Load | **✅** | ✅ | ❌ | ❌ | ❌ |
| Resource Auto-Scale | **✅** | ✅ | ❌ | ❌ | ❌ |
| فارسی | ✅ | ✅ | ❌ | ❌ | ❌ |

---

## 🚀 شروع سریع

### نصب روی Raspberry Pi

```bash
# One-line install
curl -fsSL https://edgeflow.io/install.sh | bash

# یا با Docker
docker run -d -p 8080:8080 -v /sys:/sys --privileged edgeflow/edgeflow
```

### نصب از Source

```bash
# Clone
git clone https://github.com/edgeflow/edgeflow.git
cd edgeflow

# Build
make build

# Run
./bin/edgeflow
```

### دسترسی به UI

```
http://localhost:8080
یا
http://<raspberry-pi-ip>:8080
```

---

## 📁 ساختار پروژه

```
edgeflow/
├── cmd/
│   └── edgeflow/
│       └── main.go              # Entry point
├── internal/
│   ├── engine/
│   │   ├── engine.go            # Core workflow engine
│   │   ├── executor.go          # Node executor
│   │   ├── scheduler.go         # Cron & triggers
│   │   └── context.go           # Execution context
│   ├── node/
│   │   ├── registry.go          # Node registry
│   │   ├── base.go              # Base node interface
│   │   └── loader.go            # Dynamic node loader
│   ├── api/
│   │   ├── server.go            # HTTP server
│   │   ├── routes.go            # API routes
│   │   ├── handlers/            # Request handlers
│   │   └── middleware/          # Auth, logging, etc.
│   ├── websocket/
│   │   └── hub.go               # WebSocket for real-time
│   ├── storage/
│   │   ├── sqlite.go            # SQLite adapter
│   │   ├── models.go            # Data models
│   │   └── migrations/          # DB migrations
│   └── config/
│       └── config.go            # Configuration
├── pkg/
│   └── nodes/
│       ├── core/                # Core nodes (if, loop, delay)
│       ├── network/             # HTTP, MQTT, WebSocket
│       ├── gpio/                # GPIO, I2C, SPI
│       ├── messaging/           # Telegram, Email, SMS
│       ├── database/            # MySQL, PostgreSQL, MongoDB
│       ├── ai/                  # OpenAI, Claude, Ollama
│       └── industrial/          # Modbus, OPC-UA, KNX
├── web/                         # React frontend
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── stores/
│   │   └── api/
│   ├── package.json
│   └── vite.config.ts
├── configs/
│   └── default.yaml             # Default config
├── scripts/
│   ├── build.sh
│   ├── install.sh
│   └── cross-compile.sh
├── docs/
│   ├── getting-started.md
│   ├── api-reference.md
│   └── node-development.md
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

---

## 🛠️ توسعه

### پیش‌نیازها

| ابزار | نسخه | نصب |
|-------|------|-----|
| Go | 1.21+ | `brew install go` یا [golang.org](https://golang.org) |
| Node.js | 18+ | برای frontend |
| Make | - | معمولاً نصب است |
| Docker | - | اختیاری |

### Setup محیط توسعه

```bash
# Clone
git clone https://github.com/edgeflow/edgeflow.git
cd edgeflow

# Install Go dependencies
go mod download

# Install frontend dependencies
cd web && npm install && cd ..

# Run in development mode
make dev
```

### Build Commands

```bash
# Build for current platform
make build

# Build for Raspberry Pi (ARM64)
make build-pi

# Build for all platforms
make build-all

# Run tests
make test

# Run linter
make lint

# Build Docker image
make docker
```

---

## 📚 مستندات

| سند | توضیح |
|-----|-------|
| [Getting Started](docs/getting-started.md) | شروع سریع |
| [Installation](docs/installation.md) | نصب کامل |
| [Configuration](docs/configuration.md) | تنظیمات |
| [IoT Deployment Strategy](docs/IOT_DEPLOYMENT_STRATEGY.md) | ⭐ استراتژی نصب ماژولار برای IoT |
| [Node-RED Feature Checklist](docs/NODE_RED_FEATURE_CHECKLIST.md) | چک‌لیست ویژگی‌های Node-RED |
| [API Reference](docs/API.md) | مستندات کامل API |
| [Node Development](docs/node-development.md) | ساخت نود جدید |
| [Contributing](CONTRIBUTING.md) | مشارکت |

---

## 🗺️ Roadmap

### ✅ v0.1.0 - Core MVP (تکمیل شده)
- [x] Core engine با workflow management
- [x] 29 نود (Core, Network, Hardware, Integration)
- [x] Web UI با React + TypeScript
- [x] REST API کامل
- [x] WebSocket real-time updates
- [x] Storage layer (SQLite)

### ✅ v0.2.0 - Hardware (تکمیل شده)
- [x] GPIO support (In/Out, PWM)
- [x] I2C support
- [x] HAL abstraction layer
- [x] Mock HAL for development
- [x] Sensor nodes (DHT)
- [x] Actuator nodes (Relay)

### ✅ v0.3.0 - Integration (تکمیل شده)
- [x] Messaging nodes (Telegram, Email, Slack, Discord)
- [x] Database nodes (MySQL, PostgreSQL, MongoDB, Redis)
- [x] AI nodes (OpenAI, Anthropic, Ollama)
- [x] Network nodes (HTTP, MQTT, WebSocket, TCP, UDP)

### 🟡 v0.4.0 - Production Ready (80% تکمیل)
- [x] JWT Authentication
- [x] API Key management
- [x] Credential encryption
- [x] Prometheus metrics
- [x] Health checks
- [x] Docker deployment
- [x] Modular architecture (IoT-optimized)
- [x] Resource monitoring & auto-disable
- [x] Installation profiles (minimal/standard/full)
- [x] Complete API documentation
- [ ] Test suite (در حال توسعه)
- [ ] Raspberry Pi OS image

### v1.0.0 - Release (برنامه‌ریزی شده)
- [ ] 50+ additional nodes
- [ ] Multi-user support
- [ ] Execution history
- [ ] Debugging tools
- [ ] One-line installer
- [ ] Video tutorials

---

## 🤝 مشارکت

مشارکت شما خوش‌آمد است! لطفاً [CONTRIBUTING.md](CONTRIBUTING.md) را مطالعه کنید.

```bash
# Fork و Clone
git clone https://github.com/YOUR_USERNAME/edgeflow.git

# Create branch
git checkout -b feature/amazing-feature

# Commit
git commit -m "Add amazing feature"

# Push
git push origin feature/amazing-feature

# Create Pull Request
```

---

## 📄 لایسنس

این پروژه تحت لایسنس [Apache 2.0](LICENSE) منتشر شده است.

---

## 🙏 تشکر

- [Go](https://golang.org) - زبان برنامه‌نویسی
- [React Flow](https://reactflow.dev) - کتابخانه flow editor
- [Fiber](https://gofiber.io) - Web framework
- [periph.io](https://periph.io) - Hardware abstraction

---

<div align="center">

**ساخته شده با ❤️ در ایران**

[Website](https://edgeflow.io) · [Documentation](https://docs.edgeflow.io) · [Discord](https://discord.gg/edgeflow)

</div>
