.PHONY: help up down restart backend-up backend-down logs clean backend frontend test build docker-build install

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)PayForward - Available Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

up: ## Start all services
	@echo "$(BLUE)Starting Keycloak services...$(NC)"
	docker-compose -f keycloak.yml up -d
	@echo "$(BLUE)Starting application services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Services started$(NC)"
	@echo ""
	@echo "$(YELLOW)Access points:$(NC)"
	@echo "  Frontend:  http://localhost:3000"
	@echo "  Backend:   http://localhost:8080"
	@echo "  Keycloak:  http://localhost:8180 (admin/admin)"
	@echo "  Neo4j:     http://localhost:7474 (neo4j/password123)"

down: ## Stop all services
	@echo "$(BLUE)Stopping all services...$(NC)"
	docker compose down
	docker compose -f keycloak.yml down
	@echo "$(GREEN)✓ Services stopped$(NC)"

restart: down up ## Restart all services

backend-up: ## Start backend service (with dependencies)
	@echo "$(BLUE)Starting Keycloak services...$(NC)"
	docker compose -f keycloak.yml up -d
	@echo "$(BLUE)Starting Neo4j...$(NC)"
	docker compose up -d neo4j
	@echo "$(BLUE)Starting backend...$(NC)"
	docker compose up -d backend
	@echo "$(GREEN)✓ Backend services started$(NC)"
	@echo ""
	@echo "$(YELLOW)Access points:$(NC)"
	@echo "  Backend:   http://localhost:8080"
	@echo "  Keycloak:  http://localhost:8180 (admin/admin)"
	@echo "  Neo4j:     http://localhost:7474 (neo4j/password123)"

backend-down: ## Stop backend service
	@echo "$(BLUE)Stopping backend...$(NC)"
	docker compose stop backend
	@echo "$(GREEN)✓ Backend stopped$(NC)"

logs: ## Show logs from all services
	docker compose -f keycloak.yml -f docker-compose.yml logs -f

logs-backend: ## Show backend logs
	docker compose logs -f backend

logs-frontend: ## Show frontend logs
	docker compose logs -f frontend

logs-keycloak: ## Show Keycloak logs
	docker compose -f keycloak.yml logs -f keycloak

logs-neo4j: ## Show Neo4j logs
	docker compose logs -f neo4j

clean: ## Stop services and remove volumes
	@echo "$(RED)Removing all services and volumes...$(NC)"
	docker compose down -v
	docker compose -f keycloak.yml down -v
	@echo "$(GREEN)✓ Cleanup complete$(NC)"

install: ## Install dependencies
	@echo "$(BLUE)Installing backend dependencies...$(NC)"
	cd backend && go mod download && go mod tidy
	@echo "$(BLUE)Installing frontend dependencies...$(NC)"
	cd frontend && npm install
	@echo "$(GREEN)✓ Dependencies installed$(NC)"

backend-dev: ## Run backend in development mode
	@echo "$(BLUE)Starting backend server...$(NC)"
	cd backend && go run cmd/server/main.go

frontend-dev: ## Run frontend in development mode
	@echo "$(BLUE)Starting frontend dev server...$(NC)"
	cd frontend && npm run dev

test: ## Run all tests
	@echo "$(BLUE)Running backend tests...$(NC)"
	cd backend && go test -v ./...
	@echo "$(BLUE)Running frontend tests...$(NC)"
	cd frontend && npm run test || true

test-backend: ## Run backend tests
	cd backend && go test -v -race -coverprofile=coverage.out ./...

test-frontend: ## Run frontend tests
	cd frontend && npm run test

lint: ## Run linters
	@echo "$(BLUE)Linting backend...$(NC)"
	cd backend && go vet ./... && gofmt -s -l .
	@echo "$(BLUE)Linting frontend...$(NC)"
	cd frontend && npm run lint

build: ## Build both backend and frontend
	@echo "$(BLUE)Building backend...$(NC)"
	cd backend && CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o server ./cmd/server
	@echo "$(BLUE)Building frontend...$(NC)"
	cd frontend && npm run build
	@echo "$(GREEN)✓ Build complete$(NC)"

docker-build: ## Build Docker images
	@echo "$(BLUE)Building Docker images...$(NC)"
	docker compose -f keycloak.yml build
	docker compose build
	@echo "$(GREEN)✓ Docker images built$(NC)"

ps: ## Show running services
	@echo "$(BLUE)Keycloak services:$(NC)"
	@docker compose -f keycloak.yml ps
	@echo ""
	@echo "$(BLUE)Application services:$(NC)"
	@docker compose ps

health: ## Check health of all services
	@echo "$(BLUE)Checking service health...$(NC)"
	@curl -s http://localhost:8080/api/health | jq '.' || echo "$(RED)Backend not responding$(NC)"
	@curl -s http://localhost:8180/health | jq '.' || echo "$(RED)Keycloak not responding$(NC)"

keycloak-setup: ## Open Keycloak admin console
	@echo "$(YELLOW)Opening Keycloak admin console...$(NC)"
	@echo "URL: http://localhost:8180"
	@echo "Username: admin"
	@echo "Password: admin"
	@open http://localhost:8180 2>/dev/null || xdg-open http://localhost:8180 2>/dev/null || echo "Please open http://localhost:8180 manually"

neo4j-browser: ## Open Neo4j browser
	@echo "$(YELLOW)Opening Neo4j browser...$(NC)"
	@echo "URL: http://localhost:7474"
	@echo "Username: neo4j"
	@echo "Password: password123"
	@open http://localhost:7474 2>/dev/null || xdg-open http://localhost:7474 2>/dev/null || echo "Please open http://localhost:7474 manually"
