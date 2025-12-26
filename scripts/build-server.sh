#!/bin/bash
set -e

echo "Building server..."

cd "$(dirname "$0")/.."

# Build the UI first
./scripts/build-ui.sh

# Copy web_dist into internal/app for embedding
echo "Preparing embedded files..."
rm -rf internal/app/web_dist
cp -r web_dist internal/app/

# Build the Go binary
echo "Building Go binary..."
go build -o family-tracker ./cmd/server

echo "Build complete. Binary: ./family-tracker"
echo ""
echo "Run with: ./family-tracker --config ./configs/config.yaml"
