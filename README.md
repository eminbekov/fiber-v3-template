# fiber-v3-template

Production-ready Go backend template built with [Fiber v3](https://github.com/gofiber/fiber), inspired by Yii2 advanced-template style modularity: start fast, keep only what you need, remove what you do not.

## Quick Start

```bash
git clone https://github.com/eminbekov/fiber-v3-template.git my-project
cd my-project
./setup.sh
```

`setup.sh` is the one-command terminal installer. It:

1. checks prerequisites
2. asks for your Go module path
3. asks which optional modules to keep/remove
4. asks for environment values and generates `.env`
5. runs `go mod tidy` and `gofmt`

## Requirements

- Go `1.26+` (repo uses `1.26.1`)
- Git
- Optional: Docker and Docker Compose

## One-Command Installer

The installer script is at `./setup.sh`.

### What It Changes

- rewrites module path from `github.com/eminbekov/fiber-v3-template` to your project path
- strips optional modules you disable
- keeps selected modules and removes marker comments
- creates `.env` interactively from `.env.example`
- finalizes with `go mod tidy` and `gofmt -s -w .`

### Marker-Based Modularity

Optional blocks are wrapped with markers:

```go
// [module:<key>:start]
// optional code
// [module:<key>:end]
```

Installer behavior:

- module disabled -> delete marker block
- module enabled -> keep code, remove marker comments

## Project Layout

```text
.
├── cmd/
│   ├── server/          # main HTTP server entrypoint
│   ├── migrate/         # migration CLI entrypoint
│   └── cron/            # optional separate cron binary
├── deploy/docker/       # Dockerfile + compose manifests
├── internal/
│   ├── config/          # env config parsing/validation
│   ├── database/        # pgx pools and DB wiring
│   ├── domain/          # domain models + sentinel errors
│   ├── repository/      # repository interfaces and postgres implementations
│   ├── service/         # business services
│   ├── handler/         # API, admin HTML, public HTML handlers
│   ├── middleware/      # request middleware chain
│   ├── nats/            # optional NATS module
│   ├── grpc/            # optional gRPC module
│   ├── websocket/       # optional websocket module
│   └── storage/         # optional local/S3 storage module
├── migrations/          # sequential SQL migrations
├── monitoring/          # optional Prometheus/Loki/Tempo/Grafana configs
├── views/               # optional server-rendered templates
├── .env.example
├── setup.sh
└── Makefile
```

## Optional Module Catalog

| Module | Purpose | Main Paths | Key Environment Variables |
|---|---|---|---|
| `nats` | async events and consumers | `internal/nats`, `internal/nats/consumers` | `NATS_URL` |
| `grpc` | gRPC server and protobuf contracts | `internal/grpc`, `proto`, `gen` | `GRPC_LISTEN_ADDRESS` |
| `websocket` | real-time websocket endpoint | `internal/websocket` | none |
| `admin` | admin HTML login and dashboard | `internal/handler/admin`, `views/admin` | session-related values |
| `web` | public landing HTML | `internal/handler/web`, `views/public` | none |
| `i18n` | translation loader and locale middleware | `internal/i18n`, `internal/middleware/language.go` | none |
| `storage` | upload/download and signed file URLs | `internal/storage`, `uploads`, `internal/middleware/signed_url.go` | `STORAGE_*`, `S3_*`, `FILE_SIGNING_KEY`, `SIGNED_URL_TTL`, `CDN_BASE_URL` |
| `cron` | in-process scheduler + optional cron binary | `internal/cron`, `cmd/cron` | none |
| `monitoring` | local observability stack | `monitoring` | `OTEL_EXPORTER_ENDPOINT` |
| `swagger` | generated OpenAPI docs and route | `docs` | none |

### Manual Removal (Without Installer)

If needed, remove modules manually:

1. delete module-owned files/directories
2. remove module env vars from `.env` and `.env.example`
3. remove module marker blocks in:
   - `cmd/server/main.go`
   - `internal/router/router.go`
   - `internal/config/config.go`
   - `Makefile`
   - `deploy/docker/Dockerfile`
   - `deploy/docker/docker-compose.yml`
   - `deploy/docker/docker-compose.dev.yml`
4. run:

```bash
go mod tidy
gofmt -s -w .
go build ./...
go vet ./...
make lint
```

## Configuration

Copy `.env.example` to `.env` or let `setup.sh` generate it.

| Variable | Required | Default | Description |
|---|---|---|---|
| `ENVIRONMENT` | No | `development` | `development` or `production` |
| `LOG_LEVEL` | No | `debug` | `debug`, `info`, `warn`, `error` |
| `HTTP_LISTEN_ADDRESS` | No | `:8080` | HTTP bind address |
| `VIEWS_PATH` | if HTML modules kept | `./views` | template root path |
| `CORS_ALLOW_ORIGINS` | No | empty | comma-separated browser origins |
| `BODY_LIMIT` | No | `4194304` | max request body bytes |
| `OTEL_EXPORTER_ENDPOINT` | No | empty | OTEL endpoint `host:port` |
| `DATABASE_URL` | Yes | none | PostgreSQL DSN |
| `REDIS_URL` | Yes | none | Redis DSN |
| `NATS_URL` | if `nats` kept | `nats://localhost:4222` | NATS URL |
| `GRPC_LISTEN_ADDRESS` | if `grpc` kept | `:9090` | gRPC bind address |
| `SESSION_DURATION` | No | `24h` | session lifetime |
| `STORAGE_TYPE` | if `storage` kept | `local` | `local` or `s3` |
| `STORAGE_LOCAL_BASE_PATH` | if local storage | `./uploads` | local upload path |
| `S3_ENDPOINT` | if s3 storage | empty | custom endpoint (MinIO/AWS compatible) |
| `S3_BUCKET` | if s3 storage | none | S3 bucket |
| `S3_ACCESS_KEY` | if s3 storage | none | access key |
| `S3_SECRET_KEY` | if s3 storage | none | secret key |
| `S3_REGION` | if s3 storage | none | region |
| `CDN_BASE_URL` | optional | empty | public CDN prefix |
| `FILE_SIGNING_KEY` | if storage kept | none | HMAC key for file URLs |
| `SIGNED_URL_TTL` | if storage kept | `15m` | signed URL lifetime |

## Development

### Local Run

```bash
go run ./cmd/server
```

### Common Make Targets

```bash
make build
make run
make lint
make migrate-up
make migrate-down
make help
```

### Docker Workflows

Dependencies only:

```bash
make docker-dev
```

Full stack:

```bash
make up
```

Stop stack:

```bash
make down
```

Logs:

```bash
make logs
```

## Migrations

```bash
# apply pending migrations
make migrate-up

# rollback last migration (or set N=2, N=3, ...)
make migrate-down

# create next migration
make migrate-create NAME=create_orders
```

Direct CLI:

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down 1
go run ./cmd/migrate version
go run ./cmd/migrate force 1
```

## Architecture

```mermaid
flowchart LR
    Client --> FiberHTTP
    FiberHTTP --> MiddlewareStack
    MiddlewareStack --> Router
    Router --> Handlers
    Handlers --> Services
    Services --> Repositories
    Repositories --> PostgreSQL
    Services --> Redis
    Services --> NATS
    GRPCClient --> GRPCServer
    GRPCServer --> Services
```

## Endpoints

- `GET /health/live`
- `GET /health/ready`
- `GET /metrics`
- `GET /api/v1/ping`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/:id`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id`
- `POST /api/v1/files` (storage module)
- `GET /api/files/:filename` (storage module)
- `GET /ws` (websocket module)
- `GET /admin/login`, `POST /admin/login`, `POST /admin/logout`, `GET /admin/dashboard` (admin module)
- `GET /swagger/*` (swagger module, non-production)

## Observability

- health: `GET /health/live`, `GET /health/ready`
- metrics: `GET /metrics` for Prometheus
- traces: set `OTEL_EXPORTER_ENDPOINT` for OTLP export

Data flow:

- logs -> Promtail -> Loki -> Grafana
- metrics -> Prometheus scrape -> Grafana
- traces -> OTEL collector -> Tempo -> Grafana

Config files live under `monitoring/`.

## HTML Views and Sessions

Server-rendered templates are under `views/`.

| Area | Layout | Handlers | Notes |
|---|---|---|---|
| Public | `layouts/public.html`, `views/public/` | `internal/handler/web` | landing page `/` |
| Admin | `layouts/base.html`, `layouts/auth.html`, `views/admin/` | `internal/handler/admin` | login `/admin/login`, dashboard `/admin/dashboard` |

Admin session behavior:

- login sets HttpOnly `session_token` cookie (`Secure` in production)
- admin routes use `middleware.NewAdminAuthenticate`
- API auth continues with `Authorization: Bearer <token>`

## Middleware Stack Order

1. Recovery
2. Metrics
3. Request ID
4. Request logging
5. Helmet headers
6. CORS
7. Body limit

## Repository and Storage Notes

- `internal/repository/user_repository.go` defines repository contract
- `internal/repository/postgres/user.go` implements PostgreSQL using `pgx/v5`
- wiring is assembled in `cmd/server/main.go`
- `/health/ready` includes PostgreSQL readiness check

Storage:

- `internal/storage/` supports local and S3-compatible backends
- signed URL verification middleware: `internal/middleware/signed_url.go`

## Cron and Scheduled Jobs

- in-process scheduler is wired in `cmd/server/main.go`
- separate binary mode is available in `cmd/cron/main.go` for multi-instance deployments

Commands:

```bash
make build-cron
make run-cron
```

## CI/CD and Deployment

Workflows:

- `.github/workflows/ci.yml`: lint, tests, swagger verification, image build/push
- `.github/workflows/deploy.yml`: manual deploy via `workflow_dispatch`

Image tags:

- push to `main` -> `main-<short-sha>`
- tag `vX.Y.Z` -> `X.Y.Z` and `latest`

Deploy secrets:

- `SERVER_HOST`
- `SERVER_USER`
- `SSH_PRIVATE_KEY`
- `APP_DIR`

Server prerequisites:

- Docker and Docker Compose
- repository deployed with `deploy/docker/docker-compose.yml`

## Git Best Practices (Guide §22)

Follow these rules for every change:

### Branch Rules

- branch from `main` only
- never commit directly to `main`
- use lowercase, hyphen-separated names with prefix:
  - `feature/`
  - `fix/`
  - `hotfix/`
  - `refactor/`
  - `docs/`
  - `chore/`
  - `test/`
  - `release/`

Examples:

- `feature/setup-installer`
- `docs/readme-polish`
- `fix/migration-command`

### Commit Rules (Conventional Commits)

Format:

```text
<type>(<scope>): <description>
```

Types:

- `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`

Examples:

- `feat(scripts): add interactive setup installer`
- `docs(readme): improve module removal guide`

### Remote Branch + PR Flow (No Local-Only Branches)

```bash
git checkout main && git pull origin main
git checkout -b <prefix>/<name>

# implement changes

gofmt -s -w .
go mod tidy
go build ./...
go vet ./...
make lint
go test -race -count=1 ./...   # when Go code changes

git add -A
git commit -m "type(scope): message"
git push -u origin <prefix>/<name>

gh pr create --title "type(scope): message" --body "..."
gh pr merge --squash --delete-branch

git checkout main && git pull origin main
```

Before every push, run local checks and fix issues first.

## FAQ

### Can I remove modules after setup?

Yes. Remove module-owned files/blocks and run:

```bash
go mod tidy
go build ./...
go vet ./...
make lint
```

### Do I need Docker?

No. You can run with local PostgreSQL/Redis/NATS and:

```bash
go run ./cmd/server
```

### Can I use this as a minimal API template?

Yes. Disable HTML, gRPC, websocket, monitoring, storage, and NATS during `./setup.sh`.
# fiber-v3-template

Production-ready Go API template built with [Fiber v3](https://github.com/gofiber/fiber), designed to be installed with a single interactive command and then trimmed to your real project scope.

## Quick Start

```bash
git clone https://github.com/eminbekov/fiber-v3-template.git my-project
cd my-project
./setup.sh
```

`setup.sh` does all initial setup in terminal:

1. Checks prerequisites (`go`, `git`, optional `docker`, `make`)
2. Asks your new module path (`github.com/you/project`)
3. Lets you keep/remove optional modules
4. Builds your `.env` from prompts
5. Runs `go mod tidy` and `gofmt`

## Requirements

- Go `1.26+` (repo currently uses `1.26.1`)
- Git
- Optional for containerized workflows: Docker + Docker Compose

## One-Command Installer

`setup.sh` is the primary entry point for new users.

### What it changes

- Rewrites module import path from `github.com/eminbekov/fiber-v3-template` to your chosen path
- Removes optional modules you disable (files + marker blocks)
- Regenerates `.env` interactively from `.env.example`
- Cleans dependencies/formatting

### Marker-based removal

Optional code sections are wrapped with:

```go
// [module:<key>:start]
// optional code
// [module:<key>:end]
```

The installer removes blocks for disabled modules and removes marker comments for enabled modules.

## Optional Module Catalog

| Module | Purpose | Main Paths | Key Env Vars |
|---|---|---|---|
| `nats` | Async events, JetStream consumers | `internal/nats`, `internal/nats/consumers` | `NATS_URL` |
| `grpc` | gRPC server and protobuf contracts | `internal/grpc`, `proto`, `gen` | `GRPC_LISTEN_ADDRESS` |
| `websocket` | Realtime websocket endpoint | `internal/websocket` | - |
| `admin` | Admin HTML login/dashboard area | `internal/handler/admin`, `views/admin` | uses session/cookie settings |
| `web` | Public HTML welcome area | `internal/handler/web`, `views/public` | - |
| `i18n` | Locale files + language middleware | `internal/i18n`, `internal/middleware/language.go` | - |
| `storage` | File upload/download and signed URLs | `internal/storage`, `uploads`, `internal/middleware/signed_url.go` | `STORAGE_TYPE`, `S3_*`, `FILE_SIGNING_KEY`, `SIGNED_URL_TTL` |
| `cron` | Separate cron binary + scheduler wiring | `cmd/cron`, `internal/cron` | - |
| `monitoring` | Local observability stack configs | `monitoring` | `OTEL_EXPORTER_ENDPOINT` (when using collector) |
| `swagger` | Generated OpenAPI docs and route | `docs` | - |

### Manual removal (without installer)

1. Delete module-owned directories/files
2. Remove module-specific env vars from `.env` / `.env.example`
3. Remove marker blocks for that module from:
   - `cmd/server/main.go`
   - `internal/router/router.go`
   - `internal/config/config.go`
   - `Makefile`
   - `deploy/docker/Dockerfile`
   - `deploy/docker/docker-compose.yml`
   - `deploy/docker/docker-compose.dev.yml`
4. Run:

```bash
go mod tidy
gofmt -s -w .
go build ./...
make lint
```

## Project Layout

```text
.
├── cmd/
│   ├── server/              # Main HTTP + app wiring
│   ├── migrate/             # Migration CLI
│   └── cron/                # Optional cron binary
├── deploy/docker/           # Dockerfile and compose manifests
├── internal/
│   ├── config/              # Env config parsing/validation
│   ├── database/            # pgx pool and DB helpers
│   ├── repository/          # Repository interfaces + postgres impl
│   ├── service/             # Business services
│   ├── handler/             # API, admin, and web handlers
│   ├── middleware/          # Request middleware stack
│   ├── nats/                # Optional NATS module
│   ├── grpc/                # Optional gRPC module
│   ├── websocket/           # Optional websocket module
│   └── storage/             # Optional storage module
├── migrations/              # SQL migrations
├── monitoring/              # Optional observability stack configs
├── views/                   # Optional HTML templates
├── .env.example
├── setup.sh
└── Makefile
```

## Configuration

Copy `.env.example` to `.env` (or run `./setup.sh` which generates it interactively).

| Variable | Required | Default | Description |
|---|---|---|---|
| `ENVIRONMENT` | No | `development` | `development` or `production` |
| `LOG_LEVEL` | No | `debug` | `debug`, `info`, `warn`, `error` |
| `HTTP_LISTEN_ADDRESS` | No | `:8080` | HTTP listen address |
| `VIEWS_PATH` | If HTML modules enabled | `./views` | Template root path |
| `CORS_ALLOW_ORIGINS` | No | empty | Comma-separated browser origins |
| `BODY_LIMIT` | No | `4194304` | Max request body bytes |
| `OTEL_EXPORTER_ENDPOINT` | No | empty | OTEL collector endpoint (`host:port`) |
| `DATABASE_URL` | Yes | none | PostgreSQL DSN |
| `REDIS_URL` | Yes | none | Redis URL |
| `NATS_URL` | If `nats` enabled | `nats://localhost:4222` | NATS server URL |
| `GRPC_LISTEN_ADDRESS` | If `grpc` enabled | `:9090` | gRPC bind address |
| `SESSION_DURATION` | No | `24h` | Session lifetime |
| `STORAGE_TYPE` | If `storage` enabled | `local` | `local` or `s3` |
| `STORAGE_LOCAL_BASE_PATH` | If local storage | `./uploads` | Local storage root |
| `S3_ENDPOINT` | If S3 storage | empty | MinIO/custom endpoint |
| `S3_BUCKET` | If S3 storage | none | Bucket name |
| `S3_ACCESS_KEY` | If S3 storage | none | Access key |
| `S3_SECRET_KEY` | If S3 storage | none | Secret key |
| `S3_REGION` | If S3 storage | none | Region |
| `CDN_BASE_URL` | Optional | empty | Public URL prefix |
| `FILE_SIGNING_KEY` | If storage enabled | none | HMAC key for file links |
| `SIGNED_URL_TTL` | If storage enabled | `15m` | Signed URL duration |

## Development Workflow

### Local run

```bash
go run ./cmd/server
```

### Common make targets

```bash
make build
make run
make lint
make migrate-up
make migrate-down
make help
```

### Docker workflows

Start only dependencies:

```bash
make docker-dev
```

Start full stack:

```bash
make up
```

Stop stack:

```bash
make down
```

Tail logs:

```bash
make logs
```

## Architecture Overview

```mermaid
flowchart LR
    client[Client] --> httpApp[FiberHTTP]
    httpApp --> middleware[MiddlewareStack]
    middleware --> router[Router]
    router --> handlers[Handlers]
    handlers --> services[Services]
    services --> repositories[Repositories]
    repositories --> postgres[(PostgreSQL)]
    services --> redis[(Redis)]
    services --> nats[NATS]
    grpcClient[gRPCClient] --> grpcServer[gRPCServer]
    grpcServer --> services
```

## Endpoints

- `GET /health/live`
- `GET /health/ready`
- `GET /metrics`
- `GET /api/v1/ping`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/:id`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id`
- `POST /api/v1/files` (storage module)
- `GET /api/files/:filename` (storage module)
- `GET /ws` (websocket module)
- `GET /admin/login`, `POST /admin/login`, `POST /admin/logout`, `GET /admin/dashboard` (admin module)
- `GET /swagger/*` (swagger module, non-production)

## CI/CD

GitHub Actions:

- `.github/workflows/ci.yml`:
  - lint
  - tests
  - swagger generation check
  - image build/push on push events
- `.github/workflows/deploy.yml`:
  - manual deploy workflow (`workflow_dispatch`)

## Coding and Git Rules

This template follows:

- `GO_FIBER_PROJECT_GUIDE.md` (architecture, coding, testing, git practices)
- `AGENTS.md` (project-specific implementation rules)
- `CONVENTIONS.md` (coding conventions)

Minimum pre-push local checks:

```bash
gofmt -s -w .
go mod tidy
go build ./...
go vet ./...
make lint
go test -race -count=1 ./...
```

## FAQ

### Can I remove modules after setup?

Yes. Re-run from a fresh branch and remove by marker key/path ownership, then run `go mod tidy`, `go build`, and `make lint`.

### How do I add a migration?

```bash
make migrate-create NAME=create_orders
make migrate-up
```

### Do I need Docker?

No. You can run locally with native PostgreSQL/Redis/NATS and `go run ./cmd/server`.

### Can I use this as a minimal API template?

Yes. Disable HTML, gRPC, websocket, monitoring, storage, and NATS during `./setup.sh` to keep only API-focused components.

## Additional Detailed Reference

The sections below were restored to keep the previous detailed operational guidance available.

### Database setup

Install PostgreSQL locally (or run it in Docker), then provide `DATABASE_URL`.

Example:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/fiber_template?sslmode=disable"
```

The HTTP server validates `DATABASE_URL` on startup and fails fast if it is missing or invalid.

### Cache and sessions

- `internal/cache/cache.go` defines the cache contract used by services.
- `internal/cache/redis.go` provides the Redis implementation with `Get`, `Set`, `Delete`, and prefix invalidation.
- `internal/cache/keys.go` centralizes typed cache key builders to avoid string typos.
- `internal/service/user_service.go` uses cache-aside reads and invalidates stale keys after writes.
- `internal/session/` backs login sessions for both JSON API and admin HTML flow (Redis). Admin UI uses an HttpOnly `session_token` cookie referencing the same session store as API Bearer tokens.

### Migrations

The project includes `cmd/migrate` and root `Makefile` targets for schema lifecycle.

```bash
# apply pending migrations
make migrate-up

# rollback last migration (or set N=2, N=3, ...)
make migrate-down

# create the next sequential migration files
make migrate-create NAME=create_orders
```

You can also run the CLI directly:

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down 1
go run ./cmd/migrate version
go run ./cmd/migrate force 1
```

### Deployment setup (template reuse)

`deploy.yml` is a reusable template. Configure these repository secrets before manual deploy:

- `SERVER_HOST` - target server hostname or IP
- `SERVER_USER` - SSH username
- `SSH_PRIVATE_KEY` - private key for SSH auth
- `APP_DIR` - absolute path to the app directory on the server

Server prerequisites:

- Docker and Docker Compose installed
- repository deployed on the server with `deploy/docker/docker-compose.yml` available

Manual deploy flow:

1. Open GitHub Actions and select `Deploy`.
2. Click `Run workflow`.
3. Enter `image_tag` (for example `main-a1b2c3d` or `1.2.3`).
4. Run and monitor deployment logs.

### Image tagging behavior

- Pushes to `main` publish `main-<short-sha>` tags.
- Version tags like `v1.2.3` publish `1.2.3` and `latest`.

### Cron and scheduled jobs

- In-process mode is wired in `cmd/server/main.go` and runs jobs under the same `errgroup` cancellation context as HTTP, gRPC, and consumers.
- Separate mode is available in `cmd/cron/main.go` for production where cron should run only once across multiple app instances.
- Jobs are registered through `internal/cron/scheduler.go` with structured start/completion/failure logging and graceful stop via `context.Context`.

Useful commands:

```bash
make build-cron
make run-cron
```

### Repository layer

- `internal/repository/user_repository.go` defines the data access contract.
- `internal/repository/postgres/user.go` provides PostgreSQL implementation with `pgx/v5`.
- Server wiring happens in `cmd/server/main.go` via `router.Dependencies`.
- Readiness endpoint `/health/ready` includes a PostgreSQL ping checker.

### HTML views (public and admin)

Server-rendered pages use `html/template` under `views/`.

| Area | Layout | Handlers | Notes |
|---|---|---|---|
| Public (end-user) | `layouts/public.html`, `views/public/` | `internal/handler/web` | Landing page at `/`. |
| Admin | `layouts/base.html`, `layouts/auth.html`, `views/admin/` | `internal/handler/admin` | Sign-in at `/admin/login`; dashboard at `/admin/dashboard`. |

Admin browser sessions:

- After successful `POST /admin/login`, server sets HttpOnly `session_token` cookie (`SameSite=Lax`, `Secure` in production).
- Protected admin routes use `middleware.NewAdminAuthenticate`.
- JSON API under `/api/v1` continues to use `Authorization: Bearer <token>` via `middleware.NewAuthenticate`.

### Observability details

Health and metrics:

- `GET /health/live` and `GET /health/ready` return typed JSON health responses.
- `GET /metrics` exposes Prometheus metrics including request totals, request durations, and in-flight requests.
- Set `OTEL_EXPORTER_ENDPOINT` to enable OTLP gRPC export for OpenTelemetry providers.

Data flow:

- Logs: app stdout -> Promtail -> Loki -> Grafana
- Metrics: Prometheus scrapes `http://app:8080/metrics` -> Grafana
- Traces: app exports OTLP gRPC to `otel-collector:4317` -> Tempo -> Grafana

Monitoring configuration files live under `monitoring/`.

### Middleware stack order

Registered in this order:

1. Recovery middleware (panic protection with stack-trace logging)
2. Prometheus metrics middleware
3. Request ID middleware (`X-Request-ID`)
4. Structured request logging middleware (`slog`)
5. Helmet security headers middleware
6. CORS middleware (configurable allowlist)
7. Body limit enforcement middleware
