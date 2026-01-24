# Changelog

All notable changes to EdgeFlow will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Module installation support for `.tgz` archive files
- EdgeFlow native module format with `edgeflow.json` manifest
- Comprehensive license analysis and attribution tracking for installed modules
- Archive extraction with proper nested directory handling
- Node Palette UI with accordion-style categories (single-open, all closed by default)

### Fixed
- Archive extraction cleanup issue causing "Source file not found" errors during module validation
- Module installation workflow now maintains extracted files throughout the entire installation process

### Changed
- Module manager now extracts archives once at the start of installation instead of multiple times
- Archive cleanup deferred until after all validation and copying is complete

## [0.1.0] - 2026-01-24

### Added
- Initial release of EdgeFlow IoT Edge Platform
- Flow-based visual programming interface
- Node-RED compatibility layer
- GPIO and hardware integration for Raspberry Pi
- Module system for extensibility
- Built-in storage and messaging nodes
- Industrial protocol support (Modbus TCP/RTU, OPC-UA)
- Wireless connectivity (BLE, Zigbee, Z-Wave)
- Dashboard widgets (14 types)
- Core nodes for logic, timing, and data processing
- Database integration (SQLite, PostgreSQL, MySQL, MongoDB, Redis, InfluxDB)
- Network protocols (HTTP, MQTT, WebSocket, TCP, UDP)
- AI/LLM integration (OpenAI, Anthropic, Ollama)
- Subflow system for reusable workflow components
