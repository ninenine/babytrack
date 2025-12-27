#!/bin/bash
set -e

# Family Tracker Database Backup Script
# Add to crontab for automated backups:
# 0 2 * * * /opt/family-tracker/backup.sh

# Configuration
DB_NAME="family_tracker"
DB_USER="family_tracker"
BACKUP_DIR="/opt/family-tracker/backups"
RETENTION_DAYS=30

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Generate backup filename with timestamp
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/${DB_NAME}_${TIMESTAMP}.sql.gz"

# Create backup
echo "Creating backup: $BACKUP_FILE"
sudo -u postgres pg_dump "$DB_NAME" | gzip > "$BACKUP_FILE"

# Set permissions
chmod 600 "$BACKUP_FILE"

# Remove old backups
echo "Removing backups older than $RETENTION_DAYS days..."
find "$BACKUP_DIR" -name "*.sql.gz" -type f -mtime +$RETENTION_DAYS -delete

# List current backups
echo "Current backups:"
ls -lh "$BACKUP_DIR"

echo "Backup complete: $BACKUP_FILE"
