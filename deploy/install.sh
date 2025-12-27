#!/bin/bash
set -e

# BabyTrack Installation Script for Ubuntu 24.04
# Run as root or with sudo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
echo_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
echo_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo_error "This script must be run as root (use sudo)"
   exit 1
fi

# Configuration
APP_USER="babytrack"
APP_DIR="/opt/babytrack"
DB_NAME="babytrack"
DB_USER="babytrack"
DOMAIN="${1:-}"

echo_info "Starting BabyTrack installation..."

# Update system
echo_info "Updating system packages..."
apt-get update
apt-get upgrade -y

# Install PostgreSQL
echo_info "Installing PostgreSQL..."
apt-get install -y postgresql postgresql-contrib

# Start and enable PostgreSQL
systemctl start postgresql
systemctl enable postgresql

# Generate random database password
DB_PASSWORD=$(openssl rand -base64 24 | tr -dc 'a-zA-Z0-9' | head -c 24)

# Create database and user
echo_info "Setting up PostgreSQL database..."
sudo -u postgres psql <<EOF
CREATE USER ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';
CREATE DATABASE ${DB_NAME} OWNER ${DB_USER};
GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME} TO ${DB_USER};
\c ${DB_NAME}
GRANT ALL ON SCHEMA public TO ${DB_USER};
EOF

# Create application user
echo_info "Creating application user..."
if ! id "$APP_USER" &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin "$APP_USER"
fi

# Create application directory
echo_info "Setting up application directory..."
mkdir -p "$APP_DIR"

# Check if binary exists in current directory
if [[ -f "./babytrack" ]]; then
    cp ./babytrack "$APP_DIR/"
    chmod +x "$APP_DIR/babytrack"
    echo_info "Binary copied to $APP_DIR"
else
    echo_warn "Binary 'babytrack' not found in current directory"
    echo_warn "You'll need to copy it manually to $APP_DIR/babytrack"
fi

# Generate JWT secret
JWT_SECRET=$(openssl rand -base64 32)

# Create config file
echo_info "Creating configuration file..."
cat > "$APP_DIR/config.yaml" <<EOF
server:
  port: 8080
  base_url: https://${DOMAIN:-your-domain.com}

database:
  dsn: postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable

auth:
  google_client_id: YOUR_GOOGLE_CLIENT_ID
  google_client_secret: YOUR_GOOGLE_CLIENT_SECRET
  jwt_secret: ${JWT_SECRET}

notifications:
  enabled: true
EOF

chmod 600 "$APP_DIR/config.yaml"

# Set ownership
chown -R "$APP_USER:$APP_USER" "$APP_DIR"

# Install systemd service
echo_info "Installing systemd service..."
cp ./babytrack.service /etc/systemd/system/
systemctl daemon-reload

# Create Caddy log directory
mkdir -p /var/log/caddy
chown caddy:caddy /var/log/caddy 2>/dev/null || true

echo ""
echo_info "=========================================="
echo_info "Installation complete!"
echo_info "=========================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Edit the configuration file:"
echo "   sudo nano $APP_DIR/config.yaml"
echo ""
echo "2. Add your Google OAuth credentials:"
echo "   - Go to https://console.cloud.google.com/apis/credentials"
echo "   - Create OAuth 2.0 Client ID"
echo "   - Add authorized redirect URI: https://${DOMAIN:-your-domain.com}/auth/google/callback"
echo ""
echo "3. Update Caddyfile with your domain:"
echo "   sudo nano /etc/caddy/Caddyfile"
echo "   # Add the contents from deploy/Caddyfile"
echo ""
echo "4. Start the services:"
echo "   sudo systemctl enable --now babytrack"
echo "   sudo systemctl reload caddy"
echo ""
echo "Database credentials (save these!):"
echo "  Database: ${DB_NAME}"
echo "  User: ${DB_USER}"
echo "  Password: ${DB_PASSWORD}"
echo ""
echo_warn "Make sure to save the database password above!"
echo ""
