.PHONY: all build clean test generate help dev-db db-migrate-up db-migrate-down

# Variables
BUILD_DIR := app/bin
APP_NAME := snippets
MAKEFILE_LIST := $(firstword $(MAKEFILE_LIST))

# Environment variables
export DB_USER=go-snippets
export DB_PASSWORD=password
export DB_NAME=snippets
export DB_HOST=localhost
export DB_PORT=5432
export DB_SSL_MODE=disable
export APP_HOST=localhost
export APP_PORT=8080


SYSTEM_KEY := "U2FsdGVkX1/K0w2X9/0jJeJk+nGGchmRtIpC/FP4YI0="
DB_PRIMARY_DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)
# Default is to show help
.DEFAULT_GOAL := help

# Main targets
all: generate build ## Generate code and build the application

generate: ## Generate code 
	@echo "Generating API code..."
	go generate ./...
	@echo "Generating DB code..."
	sqlc generate -f app/internal/db/sqlc.yaml

build: generate ## Build the application
	@echo "Building... "
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./app/cmd/api/main.go

run: build ## Run the application
	@echo "Running $(APP_NAME)..."
	DB_PRIMARY_DSN=$(DB_PRIMARY_DSN) SERVER_HOST=$(APP_HOST) SERVER_PORT=$(APP_PORT) ENCRYPTION_KEY=$(SYSTEM_KEY) ./$(BUILD_DIR)/$(APP_NAME)

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning build directory..."
	rm -rf $(BUILD_DIR)

dev-db: ## Start development database
	@echo "Starting development database..."
	docker-compose -f docker-compose.dev.yml up -d

db-stop: ## Stop development database
	@echo "Stopping development database..."
	docker-compose -f docker-compose.dev.yml down

db-shell: ## Connect to the database shell
	@echo "Connecting to database shell..."
	docker exec -it postgres psql -U $(DB_USER) -d $(DB_NAME)

db-migrate-up: ## Run database migrations
	@echo "Running all migrations..."
	migrate -source file://app/internal/db/migrations -database "$(DB_PRIMARY_DSN)" up

db-migrate-down: ## Rollback all database migrations
	@echo "Rolling back all migrations..."
	migrate -source file://db/migrations -database "$(DB_PRIMARY_DSN)" down

help: ## Show help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
