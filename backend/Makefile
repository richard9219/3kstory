# Makefile for 3kstory Backend

.PHONY: help build run test clean docker-build docker-up docker-down migrate

help:
@echo "Available commands:"
@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build backend
go build -o server ./cmd/server

run: ## Run backend
go run ./cmd/server/main.go

test: ## Run tests
go test -v ./...

clean: ## Clean build files
rm -f server

deps: ## Install dependencies
go mod download

tidy: ## Tidy dependencies
go mod tidy

docker-build: ## Build Docker image
docker build -t 3kstory-backend:latest .

docker-up: ## Start all services
docker-compose up -d

docker-down: ## Stop all services
docker-compose down

docker-logs: ## View logs
docker-compose logs -f backend

migrate: ## Run database migration
psql -h localhost -U postgres -d 3kvedio -f ../docs/database_schema.sql

dev: ## Development mode
go run ./cmd/server/main.go
