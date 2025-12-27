#!/bin/bash
set -e

# Family Tracker Update Script
# Usage: ./update.sh [path-to-new-binary]

APP_DIR="/opt/family-tracker"
APP_USER="family-tracker"
BINARY_PATH="${1:-./family-tracker}"

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
systemctl stop family-tracker

echo_info "Backing up current binary..."
if [[ -f "$APP_DIR/family-tracker" ]]; then
    cp "$APP_DIR/family-tracker" "$APP_DIR/family-tracker.bak"
fi

echo_info "Installing new binary..."
cp "$BINARY_PATH" "$APP_DIR/family-tracker"
chown "$APP_USER:$APP_USER" "$APP_DIR/family-tracker"
chmod +x "$APP_DIR/family-tracker"

echo_info "Running migrations..."
sudo -u "$APP_USER" "$APP_DIR/family-tracker" -config "$APP_DIR/config.yaml" -migrate || {
    echo_error "Migration failed, rolling back..."
    mv "$APP_DIR/family-tracker.bak" "$APP_DIR/family-tracker"
    systemctl start family-tracker
    exit 1
}

echo_info "Starting service..."
systemctl start family-tracker

echo_info "Checking service status..."
sleep 2
if systemctl is-active --quiet family-tracker; then
    echo_info "Update complete! Service is running."
    rm -f "$APP_DIR/family-tracker.bak"
else
    echo_error "Service failed to start, rolling back..."
    mv "$APP_DIR/family-tracker.bak" "$APP_DIR/family-tracker"
    systemctl start family-tracker
    exit 1
fi
