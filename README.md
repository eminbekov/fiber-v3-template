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

### Database setup

Install PostgreSQL locally, or run it with Docker, then set `DATABASE_URL`.

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/fiber_template?sslmode=disable"
```

The HTTP server validates `DATABASE_URL` during startup and exits early when the value is missing or malformed.

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

### Migrations

The project includes `cmd/migrate` and matching root `Makefile` targets for the schema lifecycle.

```bash
# Apply pending migrations.
make migrate-up

# Roll back the latest migration (or set N=2, N=3, ...).
make migrate-down

# Create the next sequential migration files.
make migrate-create NAME=create_orders
```

You can also run the migration CLI directly:

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down 1
go run ./cmd/migrate version
go run ./cmd/migrate force 1
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

### Repository layer

- `internal/repository/user_repository.go` defines the data-access contract.
- `internal/repository/postgres/user.go` provides the PostgreSQL implementation using `pgx/v5`.
- Wiring is done in `cmd/server/main.go` through `router.Dependencies`.
- Readiness endpoint `/health/ready` includes a PostgreSQL ping check.

### Cache and sessions

- `internal/cache/cache.go` defines the cache contract consumed by services.
- `internal/cache/redis.go` implements Redis cache operations (`Get`, `Set`, `Delete`, prefix invalidation).
- `internal/cache/keys.go` centralizes typed key builders to avoid string mistakes.
- `internal/service/user_service.go` uses cache-aside reads and invalidates stale keys after writes.
- `internal/session/` stores login sessions for both JSON API and admin HTML flows (Redis backend).
- Admin UI uses an HttpOnly `session_token` cookie pointing to the same session store used by API bearer-token sessions.

### Cron and scheduled jobs

- In-process mode is wired in `cmd/server/main.go` and runs jobs under the same `errgroup` cancellation context as HTTP, gRPC, and consumers.
- Separate mode is available in `cmd/cron/main.go` when cron should run once across multiple app instances.
- Jobs are registered in `internal/cron/scheduler.go` with structured start/completion/failure logs and graceful stop through `context.Context`.

Useful commands:

```bash
make build-cron
make run-cron
```

### HTML views (public and admin)

Server-rendered pages use `html/template` under `views/`.

| Area | Layout | Handlers | Notes |
|---|---|---|---|
| Public (end user) | `layouts/public.html`, `views/public/` | `internal/handler/web` | Landing page at `/`. |
| Admin | `layouts/base.html`, `layouts/auth.html`, `views/admin/` | `internal/handler/admin` | Sign-in at `/admin/login`; dashboard at `/admin/dashboard`. |

Admin browser sessions:

- After successful `POST /admin/login`, the server sets an HttpOnly `session_token` cookie (`SameSite=Lax`, `Secure` in production).
- Protected admin routes use `middleware.NewAdminAuthenticate`.
- JSON API routes under `/api/v1` continue to use `Authorization: Bearer <token>` via `middleware.NewAuthenticate`.

### Middleware stack order

Registered in this order:

1. Recovery middleware (panic protection with stack-trace logging)
2. Prometheus metrics middleware
3. Request ID middleware (`X-Request-ID`)
4. Structured request logging middleware (`slog`)
5. Helmet security headers middleware
6. CORS middleware (configurable allowlist)
7. Body-limit enforcement middleware

### Observability details

Health and metrics:

- `GET /health/live` and `GET /health/ready` return typed JSON health responses.
- `GET /metrics` exposes Prometheus metrics, including request totals, durations, and in-flight requests.
- Set `OTEL_EXPORTER_ENDPOINT` to enable OTLP gRPC export for OpenTelemetry providers.

Telemetry flow:

- Logs: app stdout -> Promtail -> Loki -> Grafana
- Metrics: Prometheus scrapes `http://app:8080/metrics` -> Grafana
- Traces: app exports OTLP gRPC to `otel-collector:4317` -> Tempo -> Grafana

Monitoring configuration is stored under `monitoring/`.

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
- `GET /` (public welcome page, web module)
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

### Deployment setup (template reuse)

`deploy.yml` is designed as a reusable workflow template. Configure these repository secrets before running a manual deploy:

- `SERVER_HOST`: target server hostname or IP
- `SERVER_USER`: SSH username
- `SSH_PRIVATE_KEY`: private key used for SSH authentication
- `APP_DIR`: absolute path to the app directory on the server

Server prerequisites:

- Docker and Docker Compose installed
- Repository deployed on the server with `deploy/docker/docker-compose.yml` available

Manual deploy flow:

1. Open GitHub Actions and select `Deploy`.
2. Click `Run workflow`.
3. Set `image_tag` (for example `main-a1b2c3d` or `1.2.3`).
4. Run and monitor deployment logs.

### Image tagging behavior

- Pushes to `main` publish `main-<short-sha>` tags.
- Version tags like `v1.2.3` publish `1.2.3` and `latest`.

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

### Branch and commit rules by change type

Use GitHub Flow from `GO_FIBER_PROJECT_GUIDE.md`:

1. Sync `main`.
2. Create a new branch from `main`.
3. Commit with Conventional Commits (`<type>(<scope>): <description>`).
4. Push the branch to origin.
5. Open a PR, request review, merge, and delete the branch.

Rules for common template change areas:

| Change area | Branch prefix example | Commit message format example |
|---|---|---|
| Database wiring or DSN validation | `feature/database-startup-validation` | `feat(config): validate database url at startup` |
| Cache implementation or key strategy | `feature/cache-invalidation` | `feat(cache): add prefix invalidation for user keys` |
| Session behavior or auth flow | `feature/admin-session-flow` | `feat(auth): unify admin and api session storage` |
| Migration tooling or migration files | `chore/migration-tooling` | `chore(migrate): add migrate make targets and cli docs` |
| Deployment workflow or image publishing | `chore/deploy-workflow` | `chore(ci): document deploy workflow secrets and runbook` |
| Cron scheduler behavior | `feature/cron-runner-mode` | `feat(cron): add dedicated cron process mode` |
| Repository layer refactor/fix | `refactor/repository-postgres` | `refactor(repository): align postgres user repository contract` |
| HTML view/admin UX changes | `feature/admin-views` | `feat(admin): add dashboard view flow documentation` |
| Observability/metrics/tracing setup | `feature/observability-pipeline` | `feat(observability): document metrics logs and traces flow` |
| Middleware ordering/security changes | `fix/middleware-order` | `fix(middleware): enforce stable request middleware order` |
| Documentation-only updates | `docs/readme-runtime-sections` | `docs(readme): add runtime setup and operations sections` |

Prefer small PRs (roughly under 400 changed lines when possible), include a clear description, and link the related issue/ticket when available.

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

