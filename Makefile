# EraLove Project Makefile

.PHONY: help frontend backend infra-up infra-down dev dev-stop all install-deps clean wire-gen swagger-gen

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Quick Start Commands
dev: infra-up ## Start development environment (infrastructure + backend + frontend)
	@echo "Starting backend and frontend..."
	@echo "Backend will run on http://localhost:8080"
	@echo "Frontend will run on http://localhost:5173"
	@echo "Nginx proxy will run on http://localhost:80"
	@echo ""
	@echo "Run 'make dev-stop' to stop all services"
	@$(MAKE) -j2 backend frontend

all: install-deps infra-up dev ## Install dependencies and start everything

dev-stop: ## Stop all development services
	@echo "Stopping infrastructure..."
	@$(MAKE) infra-down
	@echo "Note: Backend and frontend processes need to be stopped manually (Ctrl+C)"

# Infrastructure Commands (Docker)
infra-up: ## Start infrastructure services (MongoDB, Redis, MinIO, Nginx)
	@echo "Starting infrastructure services..."
	docker-compose up -d
	@echo "Waiting for services to be healthy..."
	@ping 127.0.0.1 -n 6 >nul 2>&1
	@echo "Infrastructure ready!"
	@echo "  - MongoDB: localhost:27017"
	@echo "  - Redis: localhost:6379"
	@echo "  - MinIO API: localhost:9000"
	@echo "  - MinIO Console: http://localhost:9001 (minioadmin/minioadmin123)"
	@echo "  - Nginx: localhost:80"

infra-down: ## Stop infrastructure services
	docker-compose down

infra-logs: ## Show infrastructure logs
	docker-compose logs -f

infra-restart: infra-down infra-up ## Restart infrastructure services

# Application Commands
frontend: ## Start frontend development server (Vite)
	@echo "Starting frontend on http://localhost:5173..."
	cd frontend && npm run dev

backend: ## Start backend development server (Go)
	@echo "Starting backend on http://localhost:8080..."
	cd backend && go run cmd/main.go

backend-build: ## Build backend binary
	cd backend && go build -o eralove-backend cmd/main.go

frontend-build: ## Build frontend for production
	cd frontend && npm run build

wire-gen: ## Generate Wire dependency injection code
	cd backend && wire gen ./internal/app

swagger-gen: ## Generate Swagger documentation
	cd backend && swag init -g cmd/main.go -o docs

install-deps: ## Install all dependencies
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Installing backend dependencies..."
	cd backend && go mod download
	@echo "Dependencies installed!"

# Database Commands
db-reset: ## Reset database (WARNING: This will delete all data)
	@echo "Resetting database..."
	docker-compose down -v
	docker volume rm eralove_mongodb_data eralove_redis_data 2>/dev/null || true
	@echo "Database reset complete!"

db-shell-mongo: ## Open MongoDB shell
	docker exec -it eralove-mongodb mongosh -u admin -p password123 --authenticationDatabase admin

db-shell-redis: ## Open Redis CLI
	docker exec -it eralove-redis redis-cli -a password123

# Testing Commands
test-backend: ## Run backend tests
	cd backend && go test ./...

test-frontend: ## Run frontend tests
	cd frontend && npm test

test-all: test-backend test-frontend ## Run all tests

# Linting Commands
lint-backend: ## Lint backend code
	cd backend && golangci-lint run

lint-frontend: ## Lint frontend code
	cd frontend && npm run lint

# Utility Commands
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	cd frontend && rm -rf dist node_modules/.cache
	cd backend && go clean
	@echo "Clean complete!"

clean-all: ## Clean everything including volumes (WARNING: This will delete all data)
	@echo "Cleaning everything..."
	cd frontend && rm -rf dist node_modules/.cache
	cd backend && go clean
	docker-compose down -v
	@echo "Clean complete!"

# Health Checks
health: ## Check health of all services
	@echo "Checking service health..."
	@echo "Backend API:"
	@curl -f http://localhost:8080/health || echo "  ❌ Backend: DOWN"
	@echo ""
	@echo "Frontend:"
	@curl -f http://localhost:5173 > /dev/null 2>&1 && echo "  ✅ Frontend: UP" || echo "  ❌ Frontend: DOWN"
	@echo ""
	@echo "MongoDB:"
	@docker exec eralove-mongodb mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1 && echo "  ✅ MongoDB: UP" || echo "  ❌ MongoDB: DOWN"
	@echo ""
	@echo "Redis:"
	@docker exec eralove-redis redis-cli -a password123 ping > /dev/null 2>&1 && echo "  ✅ Redis: UP" || echo "  ❌ Redis: DOWN"
	@echo ""
	@echo "MinIO:"
	@curl -f http://localhost:9000/minio/health/live > /dev/null 2>&1 && echo "  ✅ MinIO: UP" || echo "  ❌ MinIO: DOWN"

# Status Commands
status: ## Show status of all services
	@echo "=== Infrastructure Services ==="
	@docker-compose ps
	@echo ""
	@echo "=== Application Processes ==="
	@echo "Backend: Check terminal running 'make backend'"
	@echo "Frontend: Check terminal running 'make frontend'"

# Logs Commands
logs-mongo: ## Show MongoDB logs
	docker-compose logs -f mongodb

logs-redis: ## Show Redis logs
	docker-compose logs -f redis

logs-nginx: ## Show Nginx logs
	docker-compose logs -f nginx

logs-minio: ## Show MinIO logs
	docker-compose logs -f minio