#!/bin/bash

# EdgeFlow Installation Script
# Supports: Raspberry Pi OS, Ubuntu, Debian

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
EDGEFLOW_VERSION="${EDGEFLOW_VERSION:-latest}"
INSTALL_DIR="/opt/edgeflow"
BINARY_NAME="edgeflow"
SERVICE_NAME="edgeflow"
USER="edgeflow"

# Functions
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

detect_platform() {
    OS=$(uname -s)
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l|armv6l)
            ARCH="arm"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    case "$OS" in
        Linux)
            OS="linux"
            ;;
        Darwin)
            OS="darwin"
            ;;
        *)
            print_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac

    print_info "Detected platform: $OS-$ARCH"
}

download_binary() {
    print_info "Downloading EdgeFlow $EDGEFLOW_VERSION..."

    DOWNLOAD_URL="https://github.com/EdgxCloud/EdgeFlow/releases/download/${EDGEFLOW_VERSION}/edgeflow-${OS}-${ARCH}"

    if ! curl -fsSL "$DOWNLOAD_URL" -o "/tmp/$BINARY_NAME"; then
        print_error "Failed to download EdgeFlow"
        print_info "Falling back to building from source..."
        build_from_source
        return
    fi

    chmod +x "/tmp/$BINARY_NAME"
    print_success "Downloaded EdgeFlow"
}

build_from_source() {
    print_info "Building EdgeFlow from source..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Installing Go..."
        install_go
    fi

    # Clone repository
    cd /tmp
    rm -rf edgeflow
    git clone https://github.com/EdgxCloud/EdgeFlow.git
    cd edgeflow

    # Build
    make build

    cp "bin/$BINARY_NAME" "/tmp/$BINARY_NAME"
    print_success "Built EdgeFlow from source"
}

install_go() {
    GO_VERSION="1.21.5"
    print_info "Installing Go $GO_VERSION..."

    wget "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-${ARCH}.tar.gz"
    export PATH=$PATH:/usr/local/go/bin

    print_success "Go installed"
}

install_binary() {
    print_info "Installing EdgeFlow to $INSTALL_DIR..."

    # Create installation directory
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$INSTALL_DIR/data"
    mkdir -p "$INSTALL_DIR/logs"
    mkdir -p "$INSTALL_DIR/configs"

    # Copy binary
    cp "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    # Create default config
    cat > "$INSTALL_DIR/configs/default.yaml" <<EOF
server:
  host: 0.0.0.0
  port: 8080

database:
  type: sqlite
  path: $INSTALL_DIR/data/edgeflow.db

logger:
  level: info
  format: json
EOF

    print_success "Binary installed"
}

create_user() {
    if ! id "$USER" &>/dev/null; then
        print_info "Creating user $USER..."
        useradd -r -s /bin/false -d "$INSTALL_DIR" "$USER"
        print_success "User created"
    fi

    chown -R "$USER:$USER" "$INSTALL_DIR"
}

create_systemd_service() {
    print_info "Creating systemd service..."

    cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=EdgeFlow - Edge & IoT Automation Platform
After=network.target

[Service]
Type=simple
User=$USER
Group=$USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME
Restart=on-failure
RestartSec=10s
StandardOutput=journal
StandardError=journal
SyslogIdentifier=edgeflow

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$INSTALL_DIR/data $INSTALL_DIR/logs

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    print_success "Systemd service created"
}

enable_service() {
    print_info "Enabling EdgeFlow service..."
    systemctl enable "$SERVICE_NAME"
    systemctl start "$SERVICE_NAME"
    print_success "EdgeFlow service enabled and started"
}

show_status() {
    echo ""
    echo "═══════════════════════════════════════════════════════"
    echo "  EdgeFlow Installation Complete!"
    echo "═══════════════════════════════════════════════════════"
    echo ""
    echo "  Version:      $EDGEFLOW_VERSION"
    echo "  Install dir:  $INSTALL_DIR"
    echo "  Web UI:       http://$(hostname -I | awk '{print $1}'):8080"
    echo "  Local:        http://localhost:8080"
    echo ""
    echo "Commands:"
    echo "  Status:       sudo systemctl status $SERVICE_NAME"
    echo "  Logs:         sudo journalctl -u $SERVICE_NAME -f"
    echo "  Stop:         sudo systemctl stop $SERVICE_NAME"
    echo "  Start:        sudo systemctl start $SERVICE_NAME"
    echo "  Restart:      sudo systemctl restart $SERVICE_NAME"
    echo ""
    echo "═══════════════════════════════════════════════════════"
}

# Main installation flow
main() {
    echo ""
    echo "═══════════════════════════════════════════════════════"
    echo "  EdgeFlow Installer"
    echo "═══════════════════════════════════════════════════════"
    echo ""

    check_root
    detect_platform
    download_binary
    install_binary
    create_user
    create_systemd_service
    enable_service
    show_status
}

# Run main
main
