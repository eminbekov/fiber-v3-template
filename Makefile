APP_NAME := server
BUILD_DIR := bin
SERVER_MAIN_PATH := cmd/server/main.go
MIGRATE_MAIN_PATH := cmd/migrate/main.go

.PHONY: build
build: ## Build the HTTP server binary
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(APP_NAME) $(SERVER_MAIN_PATH)

.PHONY: run
run: ## Run the HTTP server
	go run $(SERVER_MAIN_PATH)

.PHONY: tidy
tidy: ## Tidy go module files
	go mod tidy

.PHONY: lint
lint: ## Run linter
	golangci-lint run ./...

.PHONY: migrate-up
migrate-up: ## Run pending migrations
	go run $(MIGRATE_MAIN_PATH) up

.PHONY: migrate-down
migrate-down: ## Roll back migration steps (default 1)
	go run $(MIGRATE_MAIN_PATH) down $(or $(N),1)

.PHONY: migrate-create
migrate-create: ## Create a new migration (usage: make migrate-create NAME=create_orders)
	migrate create -ext sql -dir migrations -seq $(NAME)

.PHONY: swagger
swagger: ## Generate Swagger docs from handler annotations
	swag init -g cmd/server/main.go -o docs --parseInternal --parseDependency

.PHONY: swagger-fmt
swagger-fmt: ## Format Swagger annotations
	swag fmt

.PHONY: proto
proto: ## Generate Go code from protobuf definitions
	protoc --go_out=gen --go_opt=paths=source_relative \
	       --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
	       proto/**/**/*.proto

.PHONY: help
help: ## Show available make targets
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z_-]+:.*## / {printf "%-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
