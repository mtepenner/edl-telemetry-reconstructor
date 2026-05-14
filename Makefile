.PHONY: help build up down logs clean test

help:
	@echo "EDL Telemetry Reconstructor - Available targets:"
	@echo "  make build       - Build all Docker images"
	@echo "  make up          - Start all services with Docker Compose"
	@echo "  make down        - Stop all services"
	@echo "  make logs        - View service logs"
	@echo "  make test        - Run tests for all components"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make simulator   - Run flight simulator locally"
	@echo "  make fusion      - Run fusion engine locally"
	@echo "  make ui          - Run React UI locally"

build:
	@echo "Building Docker images..."
	docker-compose build

up:
	@echo "Starting all services..."
	docker-compose up -d
	@echo "Services started!"
	@echo "- Flight Simulator: UDP on localhost:5005"
	@echo "- Fusion Engine: HTTP on localhost:8080"
	@echo "- Mission UI: HTTP on localhost:3000"
	@echo "- Kafka: localhost:9092"
	@echo "- TimescaleDB: localhost:5432"

down:
	@echo "Stopping all services..."
	docker-compose down

logs:
	docker-compose logs -f

clean:
	@echo "Cleaning build artifacts..."
	rm -rf flight_simulator/__pycache__
	rm -rf flight_simulator/app/__pycache__
	rm -rf flight_simulator/app/dynamics/__pycache__
	rm -rf flight_simulator/app/sensors/__pycache__
	rm -rf telemetry_pipeline/fusion_server
	rm -rf mission_playback_ui/node_modules
	rm -rf mission_playback_ui/build
	docker-compose down -v

test:
	@echo "Running tests..."
	cd telemetry_pipeline && go test -v ./internal/estimation/... && cd ..
	@echo "All tests passed!"

simulator:
	@echo "Running flight simulator (UDP mode)..."
	cd flight_simulator && python app/main.py --udp-host 127.0.0.1 --udp-port 5005 --verbose

fusion:
	@echo "Running fusion engine locally..."
	cd telemetry_pipeline && go run ./cmd/fusion_server/main.go --http-port 8080 --verbose

ui:
	@echo "Starting React UI..."
	cd mission_playback_ui && npm start

status:
	@echo "Service status:"
	docker-compose ps

shell-python:
	docker-compose exec flight-simulator bash

shell-go:
	docker-compose exec fusion-engine sh

shell-node:
	docker-compose exec mission-ui sh

shell-db:
	docker-compose exec timescaledb psql -U postgres -d telemetry
