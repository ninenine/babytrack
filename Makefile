.PHONY: help install dev dev-web db-up db-down db-reset migrate build build-web build-server run clean lint format pre-commit test test-web test-all

# Default target
help:
	@echo "BabyTrack - Available commands:"
	@echo ""
	@echo "  Development:"
	@echo "    make install    - Install all dependencies"
	@echo "    make dev        - Start database and run server (development mode)"
	@echo "    make dev-web    - Start Vite dev server with hot reload"
	@echo ""
	@echo "  Database:"
	@echo "    make db-up      - Start PostgreSQL container"
	@echo "    make db-down    - Stop PostgreSQL container"
	@echo "    make db-reset   - Reset database (drop and recreate)"
	@echo "    make migrate    - Run database migrations"
	@echo ""
	@echo "  Build:"
	@echo "    make build      - Build web UI and server binary"
	@echo "    make build-web  - Build only the web UI"
	@echo "    make build-server - Build only the server binary"
	@echo ""
	@echo "  Code Quality:"
	@echo "    make lint       - Run linters"
	@echo "    make format     - Format all code"
	@echo "    make pre-commit - Run pre-commit hooks on all files"
	@echo ""
	@echo "  Testing:"
	@echo "    make test       - Run Go tests"
	@echo "    make test-web   - Run web tests"
	@echo "    make test-all   - Run all tests"
	@echo ""
	@echo "  Other:"
	@echo "    make run        - Run the built binary"
	@echo "    make clean      - Clean build artifacts"

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing web dependencies..."
	cd web && pnpm install
	@echo "Setting up pre-commit hooks..."
	pre-commit install
	pre-commit install --hook-type commit-msg

# Database commands
db-up:
	docker compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 2
	@echo "Database is up"

db-down:
	docker compose down

db-reset:
	docker compose down -v
	docker compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 3
	@echo "Database reset complete"

migrate:
	go run ./cmd/server -migrate -config ./configs/config.local.yaml

# Development
dev: db-up
	go run ./cmd/server -config ./configs/config.local.yaml

dev-web:
	cd web && pnpm dev

# Build commands
build-web:
	cd web && pnpm run build

build-server: build-web
	@echo "Building Go binary..."
	CGO_ENABLED=0 go build -o babytrack ./cmd/server
	@echo "Build complete: ./babytrack"

build: build-server

# Run built binary
run:
	./babytrack -config ./configs/config.yaml

# Clean build artifacts
clean:
	rm -f babytrack server
	rm -rf internal/app/web_dist
	rm -rf web/node_modules/.vite

# Linting
lint:
	@echo "Linting Go..."
	go vet ./...
	@echo "Linting web..."
	cd web && pnpm lint

# Formatting
format:
	@echo "Formatting Go..."
	gofmt -w .
	@echo "Formatting web..."
	cd web && pnpm prettier --write "src/**/*.{ts,tsx,css}"

# Pre-commit hooks
pre-commit:
	pre-commit run --all-files

# Tests
test:
	@echo "Running Go tests..."
	go test ./...

test-web:
	@echo "Running web tests..."
	cd web && pnpm test

test-all: test test-web
