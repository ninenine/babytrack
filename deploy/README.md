# BabyTrack Deployment Guide

Deploy BabyTrack on Ubuntu 24.04 with PostgreSQL and Caddy.

## Prerequisites

- Ubuntu 24.04 LTS server
- Domain name pointing to your server
- Caddy installed (`apt install caddy`)
- Google OAuth credentials

## Quick Start

### 1. Build the binary

On your development machine:

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o babytrack ./cmd/server

# Copy to server
scp babytrack user@your-server:/tmp/
scp -r deploy/* user@your-server:/tmp/deploy/
```

### 2. Run the installer

On the server:

```bash
cd /tmp/deploy
sudo ./install.sh your-domain.com
```

This will:
- Install PostgreSQL
- Create database and user
- Create system user `babytrack`
- Set up the application directory `/opt/babytrack`
- Generate secure passwords and JWT secret
- Install the systemd service

### 3. Configure Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create a new OAuth 2.0 Client ID
3. Set authorized JavaScript origins: `https://your-domain.com`
4. Set authorized redirect URI: `https://your-domain.com/auth/google/callback`
5. Copy the Client ID and Secret

Update the config:

```bash
sudo nano /opt/babytrack/config.yaml
```

Replace `YOUR_GOOGLE_CLIENT_ID` and `YOUR_GOOGLE_CLIENT_SECRET` with your values.

### 4. Configure Caddy

Add to `/etc/caddy/Caddyfile`:

```bash
sudo nano /etc/caddy/Caddyfile
```

Copy the contents from `Caddyfile` in this directory, replacing `your-domain.com` with your actual domain.

### 5. Start services

```bash
# Start BabyTrack
sudo systemctl enable --now babytrack

# Reload Caddy
sudo systemctl reload caddy
```

## File Structure

```
/opt/babytrack/
├── babytrack     # Binary
├── config.yaml        # Configuration
└── backups/           # Database backups
```

## Commands

### Service Management

```bash
# Start/stop/restart
sudo systemctl start babytrack
sudo systemctl stop babytrack
sudo systemctl restart babytrack

# View status
sudo systemctl status babytrack

# View logs
sudo journalctl -u babytrack -f
```

### Database

```bash
# Connect to database
sudo -u postgres psql babytrack

# Run migrations manually
sudo -u babytrack /opt/babytrack/babytrack -config /opt/babytrack/config.yaml -migrate
```

### Backup

```bash
# Manual backup
sudo /opt/babytrack/backup.sh

# Set up automated daily backups (2 AM)
sudo crontab -e
# Add: 0 2 * * * /opt/babytrack/backup.sh
```

### Restore from Backup

```bash
# Stop the service
sudo systemctl stop babytrack

# Restore
gunzip -c /opt/babytrack/backups/babytrack_TIMESTAMP.sql.gz | sudo -u postgres psql babytrack

# Start the service
sudo systemctl start babytrack
```

## Updating

```bash
# Build new binary
GOOS=linux GOARCH=amd64 go build -o babytrack ./cmd/server

# Copy to server
scp babytrack user@your-server:/tmp/

# On server: update binary
sudo systemctl stop babytrack
sudo cp /tmp/babytrack /opt/babytrack/
sudo chown babytrack:babytrack /opt/babytrack/babytrack
sudo chmod +x /opt/babytrack/babytrack
sudo systemctl start babytrack
```

## Troubleshooting

### Check service status

```bash
sudo systemctl status babytrack
sudo journalctl -u babytrack --since "10 minutes ago"
```

### Check Caddy

```bash
sudo systemctl status caddy
sudo journalctl -u caddy --since "10 minutes ago"
```

### Test database connection

```bash
sudo -u postgres psql -c "SELECT 1" babytrack
```

### Check ports

```bash
# App should be on 8080
ss -tlnp | grep 8080

# Caddy on 80/443
ss -tlnp | grep -E ':(80|443)'
```

## Security Notes

- Config file has restricted permissions (600)
- App runs as unprivileged `babytrack` user
- Systemd service has security hardening enabled
- Database password is randomly generated
- JWT secret is randomly generated
