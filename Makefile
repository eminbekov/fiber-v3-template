APP_NAME := server
BUILD_DIR := bin
SERVER_MAIN_PATH := cmd/server/main.go
# [module:cron:start]
CRON_MAIN_PATH := cmd/cron/main.go
# [module:cron:end]
# [module:console:start]
CONSOLE_MAIN_PATH := cmd/console/main.go
# [module:console:end]
# [module:generate:start]
GENERATE_MAIN_PATH := cmd/generate/main.go
# [module:generate:end]
MIGRATE_MAIN_PATH := cmd/migrate/main.go
DOCKERFILE_PATH := deploy/docker/Dockerfile
DOCKER_COMPOSE_PATH := deploy/docker/docker-compose.yml
DOCKER_COMPOSE_DEV_PATH := deploy/docker/docker-compose.dev.yml

# --- Build ---
.PHONY: build
build: ## Build the HTTP server binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(SERVER_MAIN_PATH)

.PHONY: run
run: ## Run the HTTP server
	go run $(SERVER_MAIN_PATH)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)

# [module:cron:start]
.PHONY: build-cron
build-cron: ## Build the cron binary
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/cron $(CRON_MAIN_PATH)

.PHONY: run-cron
run-cron: ## Run the cron binary
	go run $(CRON_MAIN_PATH)
# [module:cron:end]

# [module:console:start]
.PHONY: build-console
build-console: ## Build the console CLI binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/console $(CONSOLE_MAIN_PATH)

.PHONY: create-admin
create-admin: ## Create admin user (make create-admin USERNAME=x PASSWORD=x PHONE=x)
	go run $(CONSOLE_MAIN_PATH) create-admin --username=$(USERNAME) --password=$(PASSWORD) --phone=$(PHONE)

.PHONY: assign-role
assign-role: ## Assign role to user (make assign-role USER_ID=... ROLE=admin)
	go run $(CONSOLE_MAIN_PATH) assign-role --user-id=$(USER_ID) --role=$(ROLE)

.PHONY: cache-clear
cache-clear: ## Clear Redis cache (optional: make cache-clear PREFIX=user:)
	go run $(CONSOLE_MAIN_PATH) cache-clear $(if $(PREFIX),--prefix=$(PREFIX))

.PHONY: export-users
export-users: ## Export users to CSV (optional: make export-users OUTPUT=users.csv)
	go run $(CONSOLE_MAIN_PATH) export-users $(if $(OUTPUT),--output=$(OUTPUT))
# [module:console:end]

# [module:generate:start]
.PHONY: build-generate
build-generate: ## Build the code generator binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/generate $(GENERATE_MAIN_PATH)

.PHONY: generate-migration
generate-migration: ## Create migration stubs (make generate-migration NAME=create_orders)
	go run $(GENERATE_MAIN_PATH) migration $(NAME)

.PHONY: generate-resource
generate-resource: ## Scaffold CRUD resource (make generate-resource NAME=order)
	go run $(GENERATE_MAIN_PATH) resource $(NAME)
# [module:generate:end]

# --- Development ---
.PHONY: tidy
tidy: ## Tidy go module files
	go mod tidy

.PHONY: fmt
fmt: ## Format all Go source files
	gofmt -s -w .

.PHONY: dev
dev: ## Run with hot reload (requires: go install github.com/air-verse/air@latest)
	air

# --- Code Quality ---
.PHONY: lint
lint: ## Run linter
	golangci-lint run ./...

.PHONY: vet
vet: ## Run go vet checks
	go vet ./...

# --- Testing ---
.PHONY: test
test: ## Run all tests with race detector
	go test -race -count=1 ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	go test -race -count=1 -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-integration
test-integration: ## Run integration tests
	go test -race -count=1 -run Integration ./...

.PHONY: bench
bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

.PHONY: fuzz
fuzz: ## Run fuzz tests for 30 seconds
	go test -fuzz=. -fuzztime=30s ./...

.PHONY: security
security: ## Run security scanners
	govulncheck ./...
	gosec ./...

.PHONY: verify
verify: ## Run full pre-push verification sequence (fmt, tidy, build, vet, lint, test)
	$(MAKE) fmt
	$(MAKE) tidy
	go build ./...
	$(MAKE) vet
	$(MAKE) lint
	$(MAKE) test
	@echo "All pre-push checks passed"

# --- Database ---
.PHONY: migrate-up
migrate-up: ## Run pending migrations
	go run $(MIGRATE_MAIN_PATH) up

.PHONY: migrate-down
migrate-down: ## Roll back migration steps (default 1)
	go run $(MIGRATE_MAIN_PATH) down $(or $(N),1)

.PHONY: migrate-create
migrate-create: ## Create a new migration (usage: make migrate-create NAME=create_orders)
	migrate create -ext sql -dir migrations -seq $(NAME)

# --- Swagger ---
# [module:swagger:start]
.PHONY: swagger
swagger: ## Generate Swagger docs from handler annotations
	swag init -g cmd/server/main.go -o docs --parseInternal --parseDependency

.PHONY: swagger-fmt
swagger-fmt: ## Format Swagger annotations
	swag fmt
# [module:swagger:end]

# --- Protobuf ---
# [module:grpc:start]
.PHONY: proto
proto: ## Generate Go code from protobuf definitions
	protoc --go_out=gen --go_opt=paths=source_relative \
	       --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
	       proto/**/**/*.proto
# [module:grpc:end]

# --- Docker ---
.PHONY: docker-build
docker-build: ## Build application Docker image
	docker build -t $(APP_NAME):latest -f $(DOCKERFILE_PATH) .

.PHONY: up
up: ## Build and start full stack (app, Postgres, Redis, NATS, observability)
	docker compose -f $(DOCKER_COMPOSE_PATH) up --build -d

.PHONY: down
down: ## Stop full Docker Compose stack
	docker compose -f $(DOCKER_COMPOSE_PATH) down

.PHONY: logs
logs: ## Tail logs for all Compose services
	docker compose -f $(DOCKER_COMPOSE_PATH) logs -f

# [module:monitoring:start]
.PHONY: monitoring-up
monitoring-up: ## Start observability services only (Prometheus, Loki, Promtail, Tempo, OTEL Collector, Grafana)
	docker compose -f $(DOCKER_COMPOSE_PATH) up -d prometheus loki promtail tempo otel-collector grafana

.PHONY: monitoring-down
monitoring-down: ## Stop observability services only
	docker compose -f $(DOCKER_COMPOSE_PATH) stop prometheus loki promtail tempo otel-collector grafana
# [module:monitoring:end]

.PHONY: docker-up
docker-up: ## Start full application stack with Docker Compose (no rebuild)
	docker compose -f $(DOCKER_COMPOSE_PATH) up -d

.PHONY: docker-down
docker-down: ## Stop full application stack with Docker Compose
	docker compose -f $(DOCKER_COMPOSE_PATH) down

.PHONY: docker-dev
docker-dev: ## Start development dependencies only (Postgres, Redis, NATS)
	docker compose -f $(DOCKER_COMPOSE_DEV_PATH) up -d

.PHONY: docker-dev-down
docker-dev-down: ## Stop development dependencies
	docker compose -f $(DOCKER_COMPOSE_DEV_PATH) down

.PHONY: docker-logs
docker-logs: ## Tail full stack container logs
	docker compose -f $(DOCKER_COMPOSE_PATH) logs -f

.PHONY: docker-ps
docker-ps: ## Show full stack container status
	docker compose -f $(DOCKER_COMPOSE_PATH) ps

.PHONY: docker-migrate
docker-migrate: ## Run database migrations inside app container
	docker compose -f $(DOCKER_COMPOSE_PATH) exec app ./migrate up

.PHONY: docker-health
docker-health: ## Check server health endpoints
	curl -fsS http://localhost:8080/health/live
	curl -fsS http://localhost:8080/health/ready
	curl -fsS http://localhost:8080/metrics >/dev/null

# --- Help ---
.PHONY: help
help: ## Show available make targets
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z_-]+:.*## / {printf "%-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
