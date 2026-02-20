#!/bin/bash

# ============================================================================
# EdgeFlow - One-Command Installer for Raspberry Pi OS
#
# Usage:
#   wget -qO- http://192.168.1.63/f.hosseini/edgeflow/raw/master/scripts/install-raspberry.sh | sudo bash
#   # or
#   curl -fsSL http://192.168.1.63/f.hosseini/edgeflow/raw/master/scripts/install-raspberry.sh | sudo bash
#   # or with options:
#   curl -fsSL http://192.168.1.63/f.hosseini/edgeflow/raw/master/scripts/install-raspberry.sh | sudo bash -s -- --profile full
#
# Supports: Raspberry Pi 3/4/5, Raspberry Pi OS (64-bit & 32-bit)
# ============================================================================

set -e

# Configuration
GIT_REPO="${GIT_REPO:-git@192.168.1.63:f.hosseini/edgeflow.git}"
GIT_HTTP_REPO="${GIT_HTTP_REPO:-http://192.168.1.63/f.hosseini/edgeflow.git}"
INSTALL_DIR="/opt/edgeflow"
GO_VERSION="1.24.0"
NODE_VERSION="20"
PROFILE="${PROFILE:-standard}"
SERVICE_USER="edgeflow"
SERVICE_NAME="edgeflow"
TOTAL_STEPS=9

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# ============================================================================
# Helper Functions
# ============================================================================

print_banner() {
    echo ""
    echo -e "${CYAN}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║                                                               ║${NC}"
    echo -e "${CYAN}║   ${BOLD}EdgeFlow${NC}${CYAN} - Raspberry Pi Installer                         ║${NC}"
    echo -e "${CYAN}║   پلتفرم اتوماسیون سبک برای Edge و IoT                        ║${NC}"
    echo -e "${CYAN}║                                                               ║${NC}"
    echo -e "${CYAN}║   Profile: ${BOLD}${PROFILE}${NC}${CYAN}                                               ║${NC}"
    echo -e "${CYAN}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_step() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  [$1/$TOTAL_STEPS]${NC} ${BOLD}$2${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

ok()   { echo -e "  ${GREEN}✓${NC} $1"; }
err()  { echo -e "  ${RED}✗${NC} $1"; }
info() { echo -e "  ${YELLOW}ℹ${NC} $1"; }
warn() { echo -e "  ${YELLOW}⚠${NC} $1"; }

# ============================================================================
# Parse Arguments
# ============================================================================

SKIP_SSH=false
SKIP_FRONTEND=false
USE_HTTP=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --profile)    PROFILE="$2"; shift 2 ;;
        --repo)       GIT_REPO="$2"; shift 2 ;;
        --http)       USE_HTTP=true; shift ;;
        --skip-ssh)   SKIP_SSH=true; shift ;;
        --skip-frontend) SKIP_FRONTEND=true; shift ;;
        --help|-h)
            echo "EdgeFlow Raspberry Pi Installer"
            echo ""
            echo "Usage:"
            echo "  sudo bash install-raspberry.sh [OPTIONS]"
            echo "  curl -fsSL <url>/install-raspberry.sh | sudo bash -s -- [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --profile PROFILE   minimal | standard | full  (default: standard)"
            echo "  --http              Clone via HTTP instead of SSH (no key needed)"
            echo "  --skip-ssh          Skip SSH key generation"
            echo "  --skip-frontend     Skip frontend build"
            echo "  --help              Show this help"
            echo ""
            echo "Profiles:"
            echo "  minimal    Pi Zero / BeagleBone   (10MB binary, ~50MB RAM)"
            echo "  standard   Pi 3/4 / Orange Pi     (20MB binary, ~200MB RAM)"
            echo "  full       Pi 4/5 / Jetson Nano   (35MB binary, ~400MB RAM)"
            echo ""
            echo "Examples:"
            echo "  sudo bash install-raspberry.sh"
            echo "  sudo bash install-raspberry.sh --profile full"
            echo "  sudo bash install-raspberry.sh --http --profile minimal"
            exit 0
            ;;
        *) err "Unknown option: $1"; exit 1 ;;
    esac
done

# Validate
if [[ "$PROFILE" != "minimal" && "$PROFILE" != "standard" && "$PROFILE" != "full" ]]; then
    err "Invalid profile: $PROFILE (use: minimal, standard, full)"
    exit 1
fi

# ============================================================================
# Pre-flight Checks
# ============================================================================

check_root() {
    if [[ $EUID -ne 0 ]]; then
        err "This script must be run as root"
        echo ""
        echo -e "  Run: ${YELLOW}sudo bash $0${NC}"
        echo -e "  Or:  ${YELLOW}curl -fsSL <url> | sudo bash${NC}"
        exit 1
    fi
}

detect_platform() {
    ARCH=$(uname -m)
    OS=$(uname -s)

    if [[ "$OS" != "Linux" ]]; then
        err "This script only supports Linux (Raspberry Pi OS)"
        exit 1
    fi

    case "$ARCH" in
        aarch64|arm64) GO_ARCH="arm64";  ok "Architecture: ARM64 (64-bit)" ;;
        armv7l)        GO_ARCH="armv6l"; ok "Architecture: ARMv7 (32-bit)" ;;
        armv6l)        GO_ARCH="armv6l"; ok "Architecture: ARMv6 (32-bit)" ;;
        x86_64)        GO_ARCH="amd64";  ok "Architecture: x86_64" ;;
        *)             err "Unsupported architecture: $ARCH"; exit 1 ;;
    esac

    # Detect Pi model
    if [[ -f /proc/device-tree/model ]]; then
        PI_MODEL=$(tr -d '\0' < /proc/device-tree/model 2>/dev/null || echo "Unknown")
        ok "Device: $PI_MODEL"
    fi

    # Check RAM and suggest profile
    TOTAL_MEM=$(free -m | awk '/^Mem:/{print $2}')
    ok "Total RAM: ${TOTAL_MEM}MB"

    if   [[ $TOTAL_MEM -lt 600  ]]; then SUGGESTED="minimal"
    elif [[ $TOTAL_MEM -lt 1500 ]]; then SUGGESTED="standard"
    else                                  SUGGESTED="full"
    fi
    ok "Build profile: ${BOLD}$PROFILE${NC} (suggested for this device: $SUGGESTED)"
}

# ============================================================================
# Step 1: System Dependencies
# ============================================================================

install_system_deps() {
    print_step 1 "Installing system dependencies"

    apt-get update -qq
    apt-get install -y -qq \
        git build-essential gcc make curl wget \
        sqlite3 ca-certificates gnupg openssh-client \
        > /dev/null 2>&1

    ok "System packages installed"
}

# ============================================================================
# Step 2: Go
# ============================================================================

install_go() {
    print_step 2 "Installing Go $GO_VERSION"

    # Check existing Go
    if command -v go &>/dev/null; then
        CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
        MINOR=$(echo "$CURRENT_GO" | cut -d. -f2)
        if [[ "$MINOR" -ge 24 ]]; then
            ok "Go $CURRENT_GO already installed (sufficient)"
            return
        fi
        warn "Go $CURRENT_GO too old, upgrading..."
    fi

    GO_TAR="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    info "Downloading $GO_TAR ..."
    wget -q "https://go.dev/dl/${GO_TAR}" -O "/tmp/$GO_TAR"

    rm -rf /usr/local/go
    tar -C /usr/local -xzf "/tmp/$GO_TAR"
    rm -f "/tmp/$GO_TAR"

    # Make Go available system-wide
    cat > /etc/profile.d/golang.sh <<'EOF'
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
EOF
    export PATH=$PATH:/usr/local/go/bin

    if /usr/local/go/bin/go version &>/dev/null; then
        ok "Go $GO_VERSION installed"
    else
        err "Go installation failed"; exit 1
    fi
}

# ============================================================================
# Step 3: Node.js
# ============================================================================

install_nodejs() {
    print_step 3 "Installing Node.js $NODE_VERSION"

    if command -v node &>/dev/null; then
        NODE_MAJOR=$(node --version | sed 's/v//' | cut -d. -f1)
        if [[ "$NODE_MAJOR" -ge 18 ]]; then
            ok "Node.js $(node --version) already installed"
            return
        fi
        warn "Node.js too old, upgrading..."
    fi

    curl -fsSL "https://deb.nodesource.com/setup_${NODE_VERSION}.x" | bash - >/dev/null 2>&1
    apt-get install -y -qq nodejs >/dev/null 2>&1

    if command -v node &>/dev/null; then
        ok "Node.js $(node --version) installed"
        ok "npm $(npm --version) installed"
    else
        err "Node.js installation failed"; exit 1
    fi
}

# ============================================================================
# Step 4: SSH Key
# ============================================================================

setup_ssh_key() {
    print_step 4 "Setting up SSH key for Git"

    REAL_USER="${SUDO_USER:-$USER}"
    REAL_HOME=$(eval echo "~$REAL_USER")
    SSH_DIR="$REAL_HOME/.ssh"

    # If using HTTP clone, skip SSH entirely
    if [[ "$USE_HTTP" == "true" ]]; then
        info "Using HTTP clone — skipping SSH key setup"
        return
    fi

    if [[ -f "$SSH_DIR/id_ed25519" ]]; then
        ok "SSH key already exists at $SSH_DIR/id_ed25519"
    else
        info "Generating SSH key for $REAL_USER"
        mkdir -p "$SSH_DIR"
        ssh-keygen -t ed25519 -C "$REAL_USER@$(hostname)" -f "$SSH_DIR/id_ed25519" -N "" -q
        chown -R "$REAL_USER:$REAL_USER" "$SSH_DIR"
        chmod 700 "$SSH_DIR"
        chmod 600 "$SSH_DIR/id_ed25519"
        chmod 644 "$SSH_DIR/id_ed25519.pub"
        ok "SSH key generated"
    fi

    # Add server to known_hosts
    GIT_HOST=$(echo "$GIT_REPO" | sed 's/.*@\(.*\):.*/\1/')
    if ! grep -q "$GIT_HOST" "$SSH_DIR/known_hosts" 2>/dev/null; then
        ssh-keyscan -H "$GIT_HOST" >> "$SSH_DIR/known_hosts" 2>/dev/null || true
        chown "$REAL_USER:$REAL_USER" "$SSH_DIR/known_hosts" 2>/dev/null || true
        ok "Added $GIT_HOST to known_hosts"
    fi

    # Show key and ask user to add it
    echo ""
    echo -e "  ${YELLOW}┌─────────────────────────────────────────────────────────┐${NC}"
    echo -e "  ${YELLOW}│  Your SSH Public Key:                                   │${NC}"
    echo -e "  ${YELLOW}└─────────────────────────────────────────────────────────┘${NC}"
    echo ""
    echo -e "  ${BOLD}$(cat "$SSH_DIR/id_ed25519.pub")${NC}"
    echo ""
    echo -e "  ${YELLOW}Add this key to your Git server:${NC}"
    echo -e "    1. Open ${BOLD}http://$GIT_HOST${NC} in your browser"
    echo -e "    2. Login -> Settings -> SSH Keys"
    echo -e "    3. Paste the key above and Save"
    echo ""

    read -rp "  Press Enter after adding the key (or type 'skip' to use HTTP clone): " REPLY
    if [[ "$REPLY" == "skip" ]]; then
        USE_HTTP=true
        info "Switching to HTTP clone"
    fi
}

# ============================================================================
# Step 5: Clone Repository
# ============================================================================

clone_repository() {
    print_step 5 "Cloning EdgeFlow repository"

    REAL_USER="${SUDO_USER:-$USER}"
    REAL_HOME=$(eval echo "~$REAL_USER")
    CLONE_DIR="$REAL_HOME/edgeflow"

    # Pick clone URL
    if [[ "$USE_HTTP" == "true" ]]; then
        CLONE_URL="$GIT_HTTP_REPO"
    else
        CLONE_URL="$GIT_REPO"
    fi

    if [[ -d "$CLONE_DIR/.git" ]]; then
        info "Repository already exists, pulling latest..."
        cd "$CLONE_DIR"
        sudo -u "$REAL_USER" git pull --ff-only 2>/dev/null || sudo -u "$REAL_USER" git pull || true
        ok "Repository updated"
    else
        info "Cloning from $CLONE_URL"
        sudo -u "$REAL_USER" git clone "$CLONE_URL" "$CLONE_DIR"
        ok "Repository cloned to $CLONE_DIR"
    fi

    # Find project root (may be in a subfolder)
    if [[ -f "$CLONE_DIR/go.mod" ]]; then
        PROJECT_DIR="$CLONE_DIR"
    elif [[ -f "$CLONE_DIR/EdgeFlow Platform/go.mod" ]]; then
        PROJECT_DIR="$CLONE_DIR/EdgeFlow Platform"
    else
        PROJECT_DIR=$(find "$CLONE_DIR" -maxdepth 2 -name "go.mod" -exec dirname {} \; | head -1)
    fi

    if [[ -z "$PROJECT_DIR" || ! -f "$PROJECT_DIR/go.mod" ]]; then
        err "Could not find go.mod in repository"
        exit 1
    fi

    ok "Project directory: $PROJECT_DIR"
}

# ============================================================================
# Step 6: Build Backend
# ============================================================================

build_backend() {
    print_step 6 "Building EdgeFlow backend [$PROFILE profile]"

    cd "$PROJECT_DIR"
    export PATH=$PATH:/usr/local/go/bin

    info "Downloading Go modules..."
    sudo -u "$REAL_USER" env PATH="$PATH" HOME="$REAL_HOME" /usr/local/go/bin/go mod download
    ok "Dependencies downloaded"

    info "Compiling (profile=$PROFILE)..."
    sudo -u "$REAL_USER" env PATH="$PATH" HOME="$REAL_HOME" make build PROFILE="$PROFILE"

    if [[ -f "$PROJECT_DIR/bin/edgeflow" ]]; then
        BINARY_SIZE=$(du -h "$PROJECT_DIR/bin/edgeflow" | cut -f1)
        ok "Backend built: bin/edgeflow ($BINARY_SIZE)"
    else
        err "Build failed — binary not found"
        exit 1
    fi
}

# ============================================================================
# Step 7: Build Frontend
# ============================================================================

build_frontend() {
    print_step 7 "Building frontend"

    if [[ ! -f "$PROJECT_DIR/web/package.json" ]]; then
        warn "No web/package.json found, skipping frontend"
        return
    fi

    cd "$PROJECT_DIR/web"

    info "Installing npm packages..."
    sudo -u "$REAL_USER" npm install --loglevel=error 2>&1
    ok "npm packages installed"

    info "Building production bundle..."
    sudo -u "$REAL_USER" npm run build 2>&1

    if [[ -d "$PROJECT_DIR/web/dist" ]]; then
        ok "Frontend built successfully"
    else
        warn "Frontend dist/ not found (may be embedded in binary)"
    fi
}

# ============================================================================
# Step 8: Install Application
# ============================================================================

install_application() {
    print_step 8 "Installing to $INSTALL_DIR"

    # Create directory structure
    mkdir -p "$INSTALL_DIR"/{bin,data,logs,configs,web}

    # Copy binary
    cp "$PROJECT_DIR/bin/edgeflow" "$INSTALL_DIR/bin/edgeflow"
    chmod +x "$INSTALL_DIR/bin/edgeflow"
    ok "Binary installed"

    # Copy configs
    if [[ -d "$PROJECT_DIR/configs" ]]; then
        cp -r "$PROJECT_DIR/configs/"* "$INSTALL_DIR/configs/" 2>/dev/null || true
        ok "Config files copied"
    fi

    # Copy frontend
    if [[ -d "$PROJECT_DIR/web/dist" ]]; then
        cp -r "$PROJECT_DIR/web/dist/"* "$INSTALL_DIR/web/" 2>/dev/null || true
        ok "Frontend files copied"
    fi

    # Create service user
    if ! id "$SERVICE_USER" &>/dev/null; then
        useradd -r -s /bin/false -d "$INSTALL_DIR" -M "$SERVICE_USER"
        ok "User '$SERVICE_USER' created"
    fi

    # Hardware group access (GPIO, I2C, SPI)
    for grp in gpio i2c spi dialout; do
        if getent group "$grp" &>/dev/null; then
            usermod -aG "$grp" "$SERVICE_USER"
        fi
    done
    ok "Hardware group access configured"

    # Permissions
    chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"

    # Symlink to PATH
    ln -sf "$INSTALL_DIR/bin/edgeflow" /usr/local/bin/edgeflow
    ok "Symlink: /usr/local/bin/edgeflow"
}

# ============================================================================
# Step 9: Systemd Service
# ============================================================================

setup_systemd_service() {
    print_step 9 "Creating systemd service"

    cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<SVCEOF
[Unit]
Description=EdgeFlow - Edge & IoT Automation Platform
Documentation=https://edgx.cloud/docs
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/bin/edgeflow
Restart=on-failure
RestartSec=10s
StandardOutput=journal
StandardError=journal
SyslogIdentifier=edgeflow

# Environment
Environment=EDGEFLOW_SERVER_HOST=0.0.0.0
Environment=EDGEFLOW_SERVER_PORT=8080
Environment=EDGEFLOW_LOGGING_LEVEL=info

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$INSTALL_DIR/data $INSTALL_DIR/logs

[Install]
WantedBy=multi-user.target
SVCEOF

    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME" >/dev/null 2>&1
    systemctl start "$SERVICE_NAME"

    ok "Service '$SERVICE_NAME' enabled and started"
}

# ============================================================================
# Done
# ============================================================================

show_completion() {
    IP_ADDR=$(hostname -I | awk '{print $1}')
    STATUS=$(systemctl is-active "$SERVICE_NAME" 2>/dev/null || echo "unknown")

    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                               ║${NC}"
    echo -e "${GREEN}║   ${BOLD}EdgeFlow installed successfully!${NC}${GREEN}                            ║${NC}"
    echo -e "${GREEN}║                                                               ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${BOLD}Profile:${NC}      $PROFILE"
    echo -e "  ${BOLD}Install:${NC}      $INSTALL_DIR"
    echo -e "  ${BOLD}Source:${NC}       $PROJECT_DIR"
    echo -e "  ${BOLD}Service:${NC}      $STATUS"
    echo ""
    echo -e "  ${CYAN}┌─────────────────────────────────────────┐${NC}"
    echo -e "  ${CYAN}│${NC}  ${BOLD}Open in browser:${NC}                       ${CYAN}│${NC}"
    echo -e "  ${CYAN}│${NC}                                         ${CYAN}│${NC}"
    echo -e "  ${CYAN}│${NC}  http://${IP_ADDR}:8080               ${CYAN}│${NC}"
    echo -e "  ${CYAN}│${NC}  http://localhost:8080                  ${CYAN}│${NC}"
    echo -e "  ${CYAN}│${NC}                                         ${CYAN}│${NC}"
    echo -e "  ${CYAN}└─────────────────────────────────────────┘${NC}"
    echo ""
    echo -e "  ${BOLD}Commands:${NC}"
    echo -e "    sudo systemctl status $SERVICE_NAME     # status"
    echo -e "    sudo journalctl -u $SERVICE_NAME -f     # logs"
    echo -e "    sudo systemctl restart $SERVICE_NAME    # restart"
    echo -e "    sudo systemctl stop $SERVICE_NAME       # stop"
    echo ""
    echo -e "  ${BOLD}Update:${NC}"
    echo -e "    cd $PROJECT_DIR && git pull"
    echo -e "    make build PROFILE=$PROFILE"
    echo -e "    sudo cp bin/edgeflow $INSTALL_DIR/bin/"
    echo -e "    sudo systemctl restart $SERVICE_NAME"
    echo ""
}

# ============================================================================
# Main
# ============================================================================

main() {
    check_root
    print_banner
    detect_platform

    install_system_deps
    install_go
    install_nodejs

    if [[ "$SKIP_SSH" == "false" && "$USE_HTTP" == "false" ]]; then
        setup_ssh_key
    fi

    clone_repository
    build_backend

    if [[ "$SKIP_FRONTEND" == "false" ]]; then
        build_frontend
    fi

    install_application
    setup_systemd_service
    show_completion
}

main
