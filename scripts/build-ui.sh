#!/bin/bash
set -e

echo "Building UI..."

cd "$(dirname "$0")/../web"

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "Installing pnpm dependencies..."
    pnpm install
fi

# Build the UI
pnpm run build

echo "UI build complete. Output in web_dist/"
