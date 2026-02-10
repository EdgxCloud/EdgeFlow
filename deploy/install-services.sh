#!/bin/bash
# EdgeFlow Service Installer for Raspberry Pi
# Run with: sudo bash install-services.sh

set -e

EDGEFLOW_DIR="/home/admin/edgeflow"
USER="admin"

echo "========================================="
echo "  EdgeFlow Service Installer"
echo "========================================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root: sudo bash install-services.sh"
  exit 1
fi

# Check if edgeflow directory exists
if [ ! -d "$EDGEFLOW_DIR" ]; then
  echo "Error: $EDGEFLOW_DIR not found"
  exit 1
fi

# Build backend if binary doesn't exist
if [ ! -f "$EDGEFLOW_DIR/bin/edgeflow" ]; then
  echo "Building EdgeFlow backend..."
  cd "$EDGEFLOW_DIR"
  sudo -u $USER GOPROXY=direct go build -o bin/edgeflow cmd/edgeflow/main.go
fi

# Create data directory
mkdir -p "$EDGEFLOW_DIR/data"
chown $USER:$USER "$EDGEFLOW_DIR/data"

# Copy service files
echo "Installing systemd services..."
cp "$EDGEFLOW_DIR/deploy/edgeflow-backend.service" /etc/systemd/system/
cp "$EDGEFLOW_DIR/deploy/edgeflow-frontend.service" /etc/systemd/system/

# Reload systemd
systemctl daemon-reload

# Enable services (start on boot)
systemctl enable edgeflow-backend.service
systemctl enable edgeflow-frontend.service

# Start services
echo "Starting EdgeFlow backend..."
systemctl start edgeflow-backend.service
sleep 2

echo "Starting EdgeFlow frontend..."
systemctl start edgeflow-frontend.service
sleep 2

# Show status
echo ""
echo "========================================="
echo "  Service Status"
echo "========================================="
systemctl status edgeflow-backend.service --no-pager -l || true
echo ""
systemctl status edgeflow-frontend.service --no-pager -l || true

echo ""
echo "========================================="
echo "  EdgeFlow is running!"
echo "========================================="
echo "Backend:  http://0.0.0.0:8080"
echo "Frontend: http://0.0.0.0:3000"
echo ""
echo "Useful commands:"
echo "  sudo systemctl status edgeflow-backend"
echo "  sudo systemctl status edgeflow-frontend"
echo "  sudo systemctl restart edgeflow-backend"
echo "  sudo systemctl restart edgeflow-frontend"
echo "  sudo journalctl -u edgeflow-backend -f"
echo "  sudo journalctl -u edgeflow-frontend -f"
echo "  sudo systemctl stop edgeflow-backend edgeflow-frontend"
echo "========================================="
