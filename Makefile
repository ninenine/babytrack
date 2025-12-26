.PHONY: help dev db-up db-down db-reset migrate build run clean

help:
	@echo "Available commands:"
	@echo "  make dev       - Start database and run server in development mode"
	@echo "  make db-up     - Start PostgreSQL container"
	@echo "  make db-down   - Stop PostgreSQL container"
	@echo "  make db-reset  - Reset database (drop and recreate)"
	@echo "  make migrate   - Run database migrations"
	@echo "  make build     - Build the application"
	@echo "  make run       - Run the application"
	@echo "  make clean     - Clean build artifacts"

db-up:
	docker compose up -d

db-down:
	docker compose down

db-reset:
	docker compose down -v
	docker compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 3

migrate:
	go run ./cmd/server -migrate -config ./configs/config.yaml

build:
	./scripts/build-server.sh

run:
	go run ./cmd/server -config ./configs/config.yaml

dev: db-up
	@echo "Waiting for database to be ready..."
	@sleep 2
	go run ./cmd/server -config ./configs/config.yaml

clean:
	rm -f family-tracker
	rm -rf web_dist
	rm -rf internal/app/web_dist
