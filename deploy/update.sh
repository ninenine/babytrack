#!/bin/bash
set -e

# BabyTrack Update Script
# Usage: ./update.sh [path-to-new-binary]

APP_DIR="/opt/babytrack"
APP_USER="babytrack"
BINARY_PATH="${1:-./babytrack}"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
echo_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo_error "This script must be run as root (use sudo)"
   exit 1
fi

# Check if binary exists
if [[ ! -f "$BINARY_PATH" ]]; then
    echo_error "Binary not found: $BINARY_PATH"
    echo "Usage: $0 [path-to-binary]"
    exit 1
fi

echo_info "Creating backup before update..."
if [[ -f "$APP_DIR/backup.sh" ]]; then
    "$APP_DIR/backup.sh" || true
fi

echo_info "Stopping service..."
systemctl stop babytrack

echo_info "Backing up current binary..."
if [[ -f "$APP_DIR/babytrack" ]]; then
    cp "$APP_DIR/babytrack" "$APP_DIR/babytrack.bak"
fi

echo_info "Installing new binary..."
cp "$BINARY_PATH" "$APP_DIR/babytrack"
chown "$APP_USER:$APP_USER" "$APP_DIR/babytrack"
chmod +x "$APP_DIR/babytrack"

echo_info "Running migrations..."
sudo -u "$APP_USER" "$APP_DIR/babytrack" -config "$APP_DIR/config.yaml" -migrate || {
    echo_error "Migration failed, rolling back..."
    mv "$APP_DIR/babytrack.bak" "$APP_DIR/babytrack"
    systemctl start babytrack
    exit 1
}

echo_info "Starting service..."
systemctl start babytrack

echo_info "Checking service status..."
sleep 2
if systemctl is-active --quiet babytrack; then
    echo_info "Update complete! Service is running."
    rm -f "$APP_DIR/babytrack.bak"
else
    echo_error "Service failed to start, rolling back..."
    mv "$APP_DIR/babytrack.bak" "$APP_DIR/babytrack"
    systemctl start babytrack
    exit 1
fi
