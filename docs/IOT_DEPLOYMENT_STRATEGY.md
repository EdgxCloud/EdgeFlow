# EdgeFlow IoT Deployment Strategy

## Overview
EdgeFlow is designed to run on resource-constrained IoT boards with **modular, installable features** instead of a monolithic binary. This allows deployment from Raspberry Pi Zero (512MB RAM) to Pi 5 (8GB RAM).

---

## 🎯 Design Philosophy

### Core Principles:
1. **Minimal Core** - Base system ~5-10MB, runs with <50MB RAM
2. **Optional Modules** - Install only what you need
3. **Hot-Pluggable** - Add/remove features without restart
4. **Resource-Aware** - Auto-disable features on low memory
5. **Progressive Enhancement** - More features on more powerful boards

---

## 📦 Modular Architecture

### Core System (Always Installed)
**Binary Size:** ~5-8MB
**RAM Usage:** ~30-50MB
**Features:**
- ✅ Workflow engine (essential)
- ✅ Basic HTTP server
- ✅ SQLite storage
- ✅ WebSocket (minimal)
- ✅ Core nodes (Inject, Debug, Function, If, Delay)
- ✅ JSON/YAML config loader
- ✅ Plugin system

```go
// Core only includes essential components
CORE_NODES = [
    "inject",     // Trigger
    "debug",      // Output
    "function",   // JavaScript
    "if",         // Conditional
    "delay",      // Timing
    "change",     // Data manipulation
    "switch",     // Routing
]
```

### Optional Feature Modules

#### 1. Network Module (~2MB, ~10MB RAM)
```bash
# Install on demand
edgeflow install network

# Or via config
modules:
  network:
    enabled: true
    nodes: [http, mqtt, websocket, tcp, udp]
```

**Includes:**
- HTTP In/Out
- MQTT In/Out
- WebSocket
- TCP/UDP

#### 2. GPIO Module (~1MB, ~5MB RAM)
```bash
edgeflow install gpio
```

**Includes:**
- GPIO In/Out
- PWM
- I2C
- SPI
- HAL abstraction

**Platform-Specific:**
- Only loads on ARM/Raspberry Pi
- Auto-disabled on x86/cloud deployments

#### 3. Database Module (~3MB, ~15MB RAM)
```bash
edgeflow install database
# Or specific drivers
edgeflow install database --drivers=mysql,postgres
```

**Includes:**
- MySQL client
- PostgreSQL client
- MongoDB client
- Redis client

**Smart Loading:**
- Drivers loaded only when used
- Connection pooling disabled on low RAM

#### 4. Messaging Module (~2MB, ~10MB RAM)
```bash
edgeflow install messaging
```

**Includes:**
- Telegram
- Email (SMTP)
- Slack
- Discord
- SMS

#### 5. AI Module (~5MB, ~30MB RAM)
```bash
edgeflow install ai
```

**Includes:**
- OpenAI
- Anthropic Claude
- Ollama (local LLM)
- Text processing

**Resource Check:**
- Requires minimum 256MB RAM
- Auto-warns if insufficient memory

#### 6. Industrial Module (~3MB, ~20MB RAM)
```bash
edgeflow install industrial
```

**Includes:**
- Modbus (RTU/TCP)
- OPC-UA
- KNX
- BACnet

#### 7. Advanced UI Module (~8MB frontend, ~20MB RAM)
```bash
edgeflow install ui-advanced
```

**Includes:**
- Expression editor
- Execution visualization
- Advanced debugging
- Subflow editor
- Template gallery

**Lite Alternative:**
- Basic UI (default): ~2MB, ~10MB RAM
- Only essential editing features

---

## 🔧 Installation Profiles

### Profile 1: Ultra-Minimal (Pi Zero, 512MB RAM)
**Total Size:** ~8MB binary + 2MB UI
**RAM Usage:** ~50MB at idle, ~80MB under load

```yaml
# configs/profile-minimal.yaml
profile: minimal

core:
  enabled: true

modules:
  network:
    enabled: true
    nodes: [http, mqtt]  # Only essential
  gpio:
    enabled: true
  messaging:
    enabled: false
  database:
    enabled: false
  ai:
    enabled: false
  industrial:
    enabled: false

ui:
  mode: lite
  features:
    - basic-editor
    - simple-debug

performance:
  max_flows: 5
  max_nodes_per_flow: 20
  execution_queue_size: 10
  websocket_max_connections: 5
```

**Use Cases:**
- Home automation sensors
- Simple GPIO control
- MQTT data collection

### Profile 2: Standard (Pi 3/4, 1-2GB RAM)
**Total Size:** ~15MB binary + 5MB UI
**RAM Usage:** ~100MB at idle, ~200MB under load

```yaml
# configs/profile-standard.yaml
profile: standard

modules:
  network:
    enabled: true
  gpio:
    enabled: true
  messaging:
    enabled: true
    services: [telegram, email]
  database:
    enabled: true
    drivers: [mysql, redis]
  ai:
    enabled: false
  industrial:
    enabled: true

ui:
  mode: standard
  features:
    - advanced-editor
    - debug-panel
    - execution-tracking

performance:
  max_flows: 20
  max_nodes_per_flow: 100
  execution_queue_size: 50
  websocket_max_connections: 20
```

**Use Cases:**
- Industrial monitoring
- Smart home hub
- Data processing

### Profile 3: Full-Featured (Pi 4/5, 4-8GB RAM)
**Total Size:** ~25MB binary + 10MB UI
**RAM Usage:** ~150MB at idle, ~400MB under load

```yaml
# configs/profile-full.yaml
profile: full

modules:
  network:
    enabled: true
  gpio:
    enabled: true
  messaging:
    enabled: true
  database:
    enabled: true
  ai:
    enabled: true
    models: [openai, anthropic]
  industrial:
    enabled: true

ui:
  mode: advanced
  features:
    - expression-editor
    - execution-visualization
    - subflows
    - templates
    - collaboration

performance:
  max_flows: 100
  max_nodes_per_flow: 500
  execution_queue_size: 200
  websocket_max_connections: 100
```

**Use Cases:**
- Edge AI processing
- Complex automation
- Development environment

---

## 🏗️ Technical Implementation

### 1. Plugin Architecture

```go
// internal/plugin/plugin.go
package plugin

type Plugin interface {
    Name() string
    Version() string
    Load() error
    Unload() error
    Nodes() []NodeDefinition
    RequiredMemory() uint64  // Minimum RAM in bytes
    RequiredDisk() uint64    // Disk space
}

type PluginManager struct {
    plugins map[string]Plugin
    loaded  map[string]bool
}

func (pm *PluginManager) Install(name string) error {
    // Download plugin from registry
    // Verify signature
    // Check system resources
    // Load plugin
}

func (pm *PluginManager) Enable(name string) error {
    plugin := pm.plugins[name]

    // Check if system has enough resources
    if !pm.hasResources(plugin) {
        return errors.New("insufficient system resources")
    }

    return plugin.Load()
}
```

### 2. Dynamic Node Loading

```go
// internal/node/loader.go
package node

type NodeLoader struct {
    registry *Registry
    modules  map[string]bool
}

func (nl *NodeLoader) LoadModule(module string) error {
    switch module {
    case "network":
        return nl.loadNetworkNodes()
    case "gpio":
        return nl.loadGPIONodes()
    case "database":
        return nl.loadDatabaseNodes()
    // ...
    }
}

func (nl *NodeLoader) loadNetworkNodes() error {
    // Only load if module is enabled
    if !nl.modules["network"] {
        return nil
    }

    // Lazy loading - load nodes on first use
    nl.registry.RegisterFactory("http-in", func() Node {
        return &HTTPInNode{}
    })
    // ...
}
```

### 3. Resource Monitoring

```go
// internal/resources/monitor.go
package resources

type ResourceMonitor struct {
    memoryThreshold uint64
    diskThreshold   uint64
}

func (rm *ResourceMonitor) CheckBeforeLoad(plugin Plugin) error {
    stats := rm.getSystemStats()

    if stats.AvailableMemory < plugin.RequiredMemory() {
        return fmt.Errorf("insufficient memory: need %dMB, have %dMB",
            plugin.RequiredMemory()/1024/1024,
            stats.AvailableMemory/1024/1024)
    }

    return nil
}

func (rm *ResourceMonitor) AutoDisableFeatures() {
    stats := rm.getSystemStats()

    // If memory is critically low, disable non-essential features
    if stats.AvailableMemory < 50*1024*1024 { // <50MB
        rm.disableFeature("ai")
        rm.disableFeature("advanced-ui")
        log.Warn("Auto-disabled AI and Advanced UI due to low memory")
    }
}
```

### 4. Build System for Modules

```makefile
# Makefile additions

## build-core: Build minimal core only
build-core:
	@echo "$(GREEN)Building minimal core...$(NC)"
	go build -tags=minimal -ldflags "-w -s" -o $(BUILD_DIR)/edgeflow-core ./cmd/edgeflow

## build-module: Build specific module
build-module:
	@echo "$(GREEN)Building module: $(MODULE)...$(NC)"
	go build -buildmode=plugin -o $(BUILD_DIR)/modules/$(MODULE).so ./pkg/nodes/$(MODULE)

## build-all-modules: Build all modules
build-all-modules:
	$(MAKE) build-module MODULE=network
	$(MAKE) build-module MODULE=gpio
	$(MAKE) build-module MODULE=database
	$(MAKE) build-module MODULE=messaging
	$(MAKE) build-module MODULE=ai
	$(MAKE) build-module MODULE=industrial

## install-profile: Install with specific profile
install-profile:
	@echo "$(GREEN)Installing EdgeFlow $(PROFILE) profile...$(NC)"
	./scripts/install-profile.sh $(PROFILE)
```

### 5. Build Tags for Conditional Compilation

```go
// pkg/nodes/network/http.go
//go:build !minimal

package network

// HTTP nodes only included in non-minimal builds
```

```go
// cmd/edgeflow/main.go
//go:build minimal

package main

// Minimal build - only core nodes
func init() {
    registerCoreNodes()
    // Skip registering optional modules
}
```

---

## 📥 Installation System

### Smart Installer Script

```bash
#!/bin/bash
# scripts/install-profile.sh

detect_board() {
    TOTAL_RAM=$(free -m | awk '/^Mem:/{print $2}')
    CPU_CORES=$(nproc)

    if [ "$TOTAL_RAM" -lt 512 ]; then
        echo "minimal"
    elif [ "$TOTAL_RAM" -lt 2048 ]; then
        echo "standard"
    else
        echo "full"
    fi
}

PROFILE=${1:-$(detect_board)}

echo "Installing EdgeFlow with profile: $PROFILE"

case $PROFILE in
    minimal)
        # Download minimal binary (8MB)
        wget https://edgeflow.io/releases/edgeflow-minimal-arm64

        # Install only core + network + gpio
        ./edgeflow install network --nodes=http,mqtt
        ./edgeflow install gpio
        ;;

    standard)
        # Download standard binary (15MB)
        wget https://edgeflow.io/releases/edgeflow-standard-arm64

        # Install common modules
        ./edgeflow install network
        ./edgeflow install gpio
        ./edgeflow install messaging --services=telegram,email
        ./edgeflow install database --drivers=mysql,redis
        ;;

    full)
        # Download full binary (25MB)
        wget https://edgeflow.io/releases/edgeflow-full-arm64

        # Install all modules
        ./edgeflow install all
        ;;
esac

# Create systemd service with appropriate resource limits
create_service $PROFILE
```

### Package Manager Integration

```bash
# Install via package manager with profile selection
apt-get install edgeflow-minimal    # For Pi Zero
apt-get install edgeflow-standard   # For Pi 3/4
apt-get install edgeflow-full       # For Pi 4/5

# Or interactive
edgeflow-installer
# Detects board, suggests profile, allows customization
```

---

## 🎛️ Runtime Module Management

### CLI Commands

```bash
# List installed modules
edgeflow modules list

# Install module
edgeflow modules install network

# Uninstall module
edgeflow modules uninstall ai

# Enable/disable without uninstalling
edgeflow modules enable database
edgeflow modules disable database

# Check system compatibility
edgeflow modules check ai
# Output: ✓ Compatible (Requires 256MB RAM, System has 1024MB)

# Show resource usage
edgeflow modules stats
# Output:
# network:    enabled, 2.1MB, 8.5MB RAM
# gpio:       enabled, 1.2MB, 4.2MB RAM
# messaging:  enabled, 1.8MB, 9.1MB RAM
# database:   disabled
# ai:         disabled
```

### Web UI Module Manager

```typescript
// web/src/pages/ModuleManager.tsx
interface Module {
    name: string;
    enabled: boolean;
    size: number;          // Disk size
    ramUsage: number;      // RAM usage
    required: boolean;     // Core modules
    compatible: boolean;   // System has resources
    nodes: string[];       // Provided nodes
}

function ModuleManager() {
    return (
        <div>
            <h2>Module Manager</h2>

            {/* System Resources */}
            <ResourceBar
                total={systemRAM}
                used={usedRAM}
                reserved={reservedByModules}
            />

            {/* Module List */}
            {modules.map(mod => (
                <ModuleCard
                    module={mod}
                    onToggle={handleToggle}
                    onInstall={handleInstall}
                    onUninstall={handleUninstall}
                />
            ))}

            {/* Profile Switcher */}
            <ProfileSelector
                current={currentProfile}
                onChange={switchProfile}
            />
        </div>
    );
}
```

---

## 🔍 Node-RED Features Priority for IoT

Based on the modular approach, here's how Node-RED features map to modules:

### Core (Always Available)
- ✅ Basic message passing
- ✅ Flow execution
- ✅ Core nodes (Inject, Debug, Function, If, Delay, Change, Switch)
- ✅ Context storage (memory only)

### Network Module (High Priority for IoT)
- ✅ HTTP In/Out
- ✅ MQTT In/Out
- ✅ WebSocket
- ✅ TCP/UDP
- 🔴 **Add to Module:**
  - Template node (Mustache)
  - JSON/XML/CSV parsers
  - Range/Split/Join nodes

### GPIO Module (Essential for IoT)
- ✅ GPIO In/Out
- ✅ PWM
- ✅ I2C
- 🔴 **Add to Module:**
  - SPI support
  - Serial port
  - 1-Wire sensors
  - Sensor-specific nodes (DHT, BMP, DS18B20)

### Advanced Features (Optional Modules)
- 🔴 Subflows → **Advanced Module**
- 🔴 Link nodes → **Advanced Module**
- 🔴 Expression editor → **UI Advanced Module**
- 🔴 Execution visualization → **UI Advanced Module**
- 🔴 Projects/Git → **Developer Module**
- 🔴 Multi-user → **Collaboration Module**

---

## 📊 Resource Comparison

### Binary Size Breakdown

| Configuration | Binary | Modules | UI | Total | RAM Usage |
|--------------|--------|---------|----|----|-----------|
| **Ultra-Minimal** | 5MB | 3MB | 2MB | **10MB** | **50-80MB** |
| **Standard** | 8MB | 7MB | 5MB | **20MB** | **100-200MB** |
| **Full** | 12MB | 13MB | 10MB | **35MB** | **150-400MB** |

### vs Competitors

| Platform | Minimal Size | Minimal RAM | IoT-Ready | Modular |
|----------|-------------|-------------|-----------|---------|
| **EdgeFlow Minimal** | **10MB** | **50MB** | ✅ | ✅ |
| **EdgeFlow Full** | 35MB | 150MB | ✅ | ✅ |
| Node-RED | 80MB+ | 150MB+ | ✅ | ⚠️ |
| n8n | 200MB+ | 300MB+ | ❌ | ❌ |
| Home Assistant | 500MB+ | 500MB+ | ⚠️ | ❌ |

---

## 🚀 Deployment Recommendations

### Raspberry Pi Zero / Zero W (512MB RAM)
```yaml
profile: minimal
modules: [network, gpio]
ui: lite
max_flows: 5
```

### Raspberry Pi 3B/3B+ (1GB RAM)
```yaml
profile: standard
modules: [network, gpio, messaging, database]
ui: standard
max_flows: 20
```

### Raspberry Pi 4/5 (2-8GB RAM)
```yaml
profile: full
modules: all
ui: advanced
max_flows: 100
```

### ESP32/Arduino (Not Supported Yet)
Future: **EdgeFlow Micro** (500KB binary, 50KB RAM)
- Only core execution engine
- No web UI (config via JSON file)
- Pre-compiled flows
- C++ implementation

---

## 📝 Next Steps

1. **Implement Plugin System** ✅ (Architecture designed)
2. **Create Build Profiles** (minimal, standard, full)
3. **Implement Module Installer** (CLI + Web UI)
4. **Resource Monitor** (Auto-disable on low memory)
5. **Smart Installer Script** (Detect board, suggest profile)
6. **Documentation** (Module development guide)
7. **Registry System** (npm-like module repository)

---

## 🎯 Success Metrics

- ✅ Run on Pi Zero with <100MB RAM
- ✅ Binary size <10MB for minimal build
- ✅ Module installation without restart
- ✅ Auto-detection and profile suggestion
- ✅ Graceful degradation on resource constraints
- ✅ No performance loss vs Node-RED on same hardware

---

**This modular approach makes EdgeFlow the ONLY IoT automation platform that scales from 512MB to 8GB+ with the same codebase!**
