.PHONY: help build run test test-verbose test-coverage clean fmt vet lint install-deps docker-build docker-up docker-down migrate-up migrate-down

# Variables
APP_NAME=payforwardnow
MAIN_PATH=./cmd/server
BINARY_NAME=server
DOCKER_COMPOSE=docker-compose

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m

help: ## Display this help message
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'

install-deps: ## Install Go dependencies
	@echo "$(COLOR_BOLD)Installing dependencies...$(COLOR_RESET)"
	go mod download
	go mod tidy
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

build: ## Build the application
	@echo "$(COLOR_BOLD)Building $(APP_NAME)...$(COLOR_RESET)"
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(COLOR_GREEN)Build complete: bin/$(BINARY_NAME)$(COLOR_RESET)"

run: ## Run the application
	@echo "$(COLOR_BOLD)Running $(APP_NAME)...$(COLOR_RESET)"
	go run $(MAIN_PATH)/main.go

test: ## Run tests
	@echo "$(COLOR_BOLD)Running tests...$(COLOR_RESET)"
	go test -v ./...

test-verbose: ## Run tests with verbose output
	@echo "$(COLOR_BOLD)Running tests (verbose)...$(COLOR_RESET)"
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "$(COLOR_BOLD)Running tests with coverage...$(COLOR_RESET)"
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)Coverage report generated: coverage.html$(COLOR_RESET)"

test-short: ## Run short tests only
	@echo "$(COLOR_BOLD)Running short tests...$(COLOR_RESET)"
	go test -short ./...

benchmark: ## Run benchmarks
	@echo "$(COLOR_BOLD)Running benchmarks...$(COLOR_RESET)"
	go test -bench=. -benchmem ./...

clean: ## Clean build artifacts
	@echo "$(COLOR_BOLD)Cleaning...$(COLOR_RESET)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean
	@echo "$(COLOR_GREEN)Clean complete$(COLOR_RESET)"

fmt: ## Format code
	@echo "$(COLOR_BOLD)Formatting code...$(COLOR_RESET)"
	go fmt ./...

vet: ## Run go vet
	@echo "$(COLOR_BOLD)Running go vet...$(COLOR_RESET)"
	go vet ./...

lint: ## Run linter
	@echo "$(COLOR_BOLD)Running golangci-lint...$(COLOR_RESET)"
	golangci-lint run ./...

check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)
	@echo "$(COLOR_GREEN)All checks passed!$(COLOR_RESET)"

docker-build: ## Build Docker image
	@echo "$(COLOR_BOLD)Building Docker image...$(COLOR_RESET)"
	docker build -t $(APP_NAME):latest .

docker-up: ## Start Docker Compose services
	@echo "$(COLOR_BOLD)Starting services...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop Docker Compose services
	@echo "$(COLOR_BOLD)Stopping services...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) down

docker-logs: ## Show Docker Compose logs
	$(DOCKER_COMPOSE) logs -f

dev: ## Run in development mode with hot reload (requires air)
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

.DEFAULT_GOAL := help
