# EraLove Project Makefile

.PHONY: help frontend backend docker-up docker-down docker-dev-up docker-dev-down docker-build docker-logs clean wire-gen swagger-gen install-deps

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development Commands
frontend: ## Start frontend development server
	cd frontend && npm run dev

backend: ## Start backend development server
	cd backend && go run cmd/main.go

wire-gen: ## Generate Wire dependency injection code
	cd backend && wire gen ./internal/app

swagger-gen: ## Generate Swagger documentation
	cd backend && swag init -g cmd/main.go -o docs

install-deps: ## Install all dependencies
	cd frontend && npm install
	cd backend && go mod download

# Docker Commands - Production
docker-up: ## Start all services with Docker Compose (Production)
	docker-compose up -d

docker-down: ## Stop all services (Production)
	docker-compose down

docker-build: ## Build all Docker images
	docker-compose build

docker-rebuild: ## Rebuild all Docker images without cache
	docker-compose build --no-cache

docker-logs: ## Show logs from all services
	docker-compose logs -f

docker-logs-backend: ## Show backend logs
	docker-compose logs -f backend

docker-logs-frontend: ## Show frontend logs
	docker-compose logs -f frontend

# Docker Commands - Development
docker-dev-up: ## Start development environment with hot reload
	docker-compose -f docker-compose.dev.yml up -d

docker-dev-down: ## Stop development environment
	docker-compose -f docker-compose.dev.yml down

docker-dev-logs: ## Show development logs
	docker-compose -f docker-compose.dev.yml logs -f

docker-dev-rebuild: ## Rebuild development images
	docker-compose -f docker-compose.dev.yml build --no-cache

# Database Commands
db-up: ## Start only database services
	docker-compose up -d mongodb redis

db-down: ## Stop database services
	docker-compose stop mongodb redis

db-reset: ## Reset database (WARNING: This will delete all data)
	docker-compose down -v
	docker volume rm eralove_mongodb_data eralove_redis_data 2>/dev/null || true

# Utility Commands
clean: ## Clean build artifacts and Docker resources
	cd frontend && rm -rf dist node_modules/.cache
	cd backend && go clean
	docker system prune -f

clean-all: ## Clean everything including volumes (WARNING: This will delete all data)
	cd frontend && rm -rf dist node_modules/.cache node_modules
	cd backend && go clean
	docker-compose down -v
	docker system prune -af
	docker volume prune -f

# Health Checks
health: ## Check health of all services
	@echo "Checking service health..."
	@curl -f http://localhost:8080/health || echo "Backend: DOWN"
	@curl -f http://localhost:3000/health || echo "Frontend: DOWN"

# Quick Start Commands
dev: docker-dev-up ## Quick start development environment (backend services only)

dev-full: ## Start full development environment (backend services + frontend)
	@echo "Starting backend services..."
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Backend services started. Now starting frontend..."
	@echo "Open a new terminal and run: make frontend"
	@echo "Or run: cd frontend && npm run dev"

prod: docker-up ## Quick start production environment

stop: docker-down docker-dev-down ## Stop all environments

restart: stop docker-up ## Restart production environment

restart-dev: docker-dev-down docker-dev-up ## Restart development environment