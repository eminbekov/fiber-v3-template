# fiber-v3-template

Starter layout for a Go HTTP API using [Fiber v3](https://github.com/gofiber/fiber). This repository will grow as the template is filled out.

## Requirements

- Go 1.26+

## Project layout

```text
.
├── cmd/server/          # HTTP server entrypoint (config, logger, graceful shutdown)
├── cmd/cron/            # Scheduled job runner entrypoint (separate deployment option)
├── cmd/migrate/         # Database migration CLI entrypoint
├── deploy/docker/       # Dockerfile and compose manifests
├── .github/workflows/   # CI and deploy automation
├── internal/
│   ├── cron/            # Ticker-based scheduler for periodic jobs
│   ├── config/          # Typed configuration from environment variables
│   ├── cache/           # Redis client, cache interface, key builders, cache implementation
│   ├── database/        # PostgreSQL pool setup and pool registry
│   ├── domain/          # Sentinel business errors and core domain types
│   ├── repository/      # Repository interfaces and PostgreSQL implementations
│   ├── session/         # Optional Redis-backed session store (removable)
│   ├── dto/response/    # Standard API success/error envelopes
│   ├── handler/         # Error handler; API v1 (JSON); admin HTML; public HTML (`web/`)
│   ├── middleware/      # Recovery, metrics, request ID, logging, CORS, Helmet, body limit
│   ├── storage/         # File storage abstraction (local filesystem, S3-compatible)
│   └── router/          # Fiber app + middleware and route registration
├── views/               # HTML templates (public + admin layouts and pages)
├── migrations/          # Sequential SQL migrations (up/down)
├── monitoring/          # Prometheus, Loki, Grafana, Tempo, OTEL Collector configs (Docker Compose)
├── package/
│   ├── health/          # Reusable liveness and readiness handlers
│   ├── logger/          # slog setup (JSON in production, text in development)
│   └── telemetry/       # OpenTelemetry tracer and meter setup
├── .env.example         # Documented environment variables (copy to .env locally)
├── AGENTS.md            # Rules for agents and tooling
├── CONVENTIONS.md       # Contributor coding conventions
└── README.md
```

## Configuration

Copy [`.env.example`](.env.example) to `.env` for local development and adjust values. The application reads the same variables from the process environment (export them or use a tool that loads `.env`).

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ENVIRONMENT` | No | `development` | `development` (text logs) or `production` (JSON logs). |
| `LOG_LEVEL` | No | `debug` | `debug`, `info`, `warn`, or `error`. |
| `HTTP_LISTEN_ADDRESS` | No | `:8080` | Listen address (for example `:3000` or `127.0.0.1:8080`). |
| `CORS_ALLOW_ORIGINS` | No | (empty) | Comma-separated list of allowed CORS origins (empty means deny cross-origin browser access). |
| `BODY_LIMIT` | No | `4194304` | Maximum request body size in bytes (4 MB default). |
| `OTEL_EXPORTER_ENDPOINT` | No | (empty) | OpenTelemetry collector endpoint (`host:port`). Empty disables telemetry export. |
| `DATABASE_URL` | Yes | (none) | PostgreSQL connection URL used by the server and migration CLI. |
| `REDIS_URL` | Yes | (none) | Redis connection URL used for cache and optional session storage. |
| `NATS_URL` | No | `nats://localhost:4222` | NATS server URL. |
| `GRPC_LISTEN_ADDRESS` | No | `:9090` | gRPC listen address. |
| `SESSION_DURATION` | No | `24h` | Session lifetime (Go duration). |
| `STORAGE_TYPE` | No | `local` | `local` or `s3`. |
| `STORAGE_LOCAL_BASE_PATH` | No | `./uploads` | Root directory for uploads when `STORAGE_TYPE=local`. |
| `S3_ENDPOINT` | When `STORAGE_TYPE=s3` | (empty) | Custom endpoint for MinIO; leave empty for AWS S3. |
| `S3_BUCKET` | When `STORAGE_TYPE=s3` | (none) | Bucket name. |
| `S3_ACCESS_KEY` | When `STORAGE_TYPE=s3` | (none) | Access key. |
| `S3_SECRET_KEY` | When `STORAGE_TYPE=s3` | (none) | Secret key. |
| `S3_REGION` | When `STORAGE_TYPE=s3` | (none) | Region (for example `us-east-1`). |
| `CDN_BASE_URL` | No | (empty) | Optional public URL prefix for CDN or reverse-proxy (no trailing slash). |
| `FILE_SIGNING_KEY` | Yes | (none) | Secret for HMAC-signed file URLs (use a long random value in production). |
| `SIGNED_URL_TTL` | No | `15m` | How long presigned / HMAC download links remain valid (Go duration). |

## Run

```bash
go run ./cmd/server
```

## Docker

Build the app image:

```bash
make docker-build
```

**Full stack (recommended):** builds the app image, starts PostgreSQL, Redis, NATS, the HTTP/gRPC app (with **automatic DB migrations** on container start), and the observability stack (Prometheus, Loki, Promtail, Grafana Tempo, OpenTelemetry Collector, Grafana).

```bash
make up
```

Equivalent:

```bash
docker compose -f deploy/docker/docker-compose.yml up --build -d
```

Stop everything:

```bash
make down
```

Tail all service logs:

```bash
make logs
```

Start **only** the observability services (when the app and databases already run elsewhere):

```bash
make monitoring-up
```

Stop those services:

```bash
make monitoring-down
```

Restart stack without rebuilding images:

```bash
make docker-up
```

Start development dependencies only (PostgreSQL, Redis, NATS — **no** app or monitoring):

```bash
make docker-dev
```

Stop containers:

```bash
make docker-down
make docker-dev-down
```

Run migrations manually inside a running app container (usually not needed — the image entrypoint runs `./migrate up` before `./server`):

```bash
make docker-migrate
```

### Observability (Compose)

After `make up`, services are available as follows:

| Service | URL | Notes |
|--------|-----|--------|
| App HTTP | [http://localhost:8080](http://localhost:8080) | Health: `/health/live`, `/health/ready`, metrics: `/metrics` |
| Grafana | [http://localhost:3000](http://localhost:3000) | Default login `admin` / `admin` (change in production). |
| Prometheus | [http://localhost:9090](http://localhost:9090) | UI and `/targets` — scrape target `fiber-v3-template-app`. |
| Loki | [http://localhost:3100](http://localhost:3100) | Log store; query via Grafana Explore. |
| Tempo | [http://localhost:3200](http://localhost:3200) | Trace backend; explore traces in Grafana. |

Data flow (see `GO_FIBER_PROJECT_GUIDE.md` §19.3):

- **Logs:** app stdout (JSON in production) → Promtail → Loki → Grafana.
- **Metrics:** Prometheus scrapes `http://app:8080/metrics` → Grafana.
- **Traces:** app exports OTLP gRPC to `otel-collector:4317` (set `OTEL_EXPORTER_ENDPOINT` in Compose) → Tempo → Grafana.

Configuration files live under [`monitoring/`](monitoring/).

Cron worker (separate process):

```bash
go run ./cmd/cron
```

With overrides:

```bash
ENVIRONMENT=production \
LOG_LEVEL=info \
HTTP_LISTEN_ADDRESS=:3000 \
CORS_ALLOW_ORIGINS=https://example.com,https://admin.example.com \
BODY_LIMIT=4194304 \
OTEL_EXPORTER_ENDPOINT=localhost:4317 \
REDIS_URL=redis://localhost:6379/0 \
go run ./cmd/server
```

## Database setup

Install PostgreSQL locally (or run it in Docker), then provide `DATABASE_URL`.

Example:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/fiber_template?sslmode=disable"
```

The HTTP server validates `DATABASE_URL` on startup and fails fast if it is missing or invalid.

## Cache and sessions

- `internal/cache/cache.go` defines the cache contract used by services.
- `internal/cache/redis.go` provides the Redis implementation with `Get`, `Set`, `Delete`, and prefix invalidation.
- `internal/cache/keys.go` centralizes typed cache key builders to avoid string typos.
- `internal/service/user_service.go` now uses cache-aside reads and invalidates stale keys after writes.
- `internal/session/` backs login sessions for both the JSON API and the admin HTML flow (Redis); the admin UI uses a `session_token` cookie that references the same session store as API Bearer tokens.

## Migrations

The project includes `cmd/migrate` and root `Makefile` targets for database schema lifecycle.

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

## CI/CD

The repository includes GitHub Actions workflows:

- `.github/workflows/ci.yml` runs lint, tests, swagger generation checks, and Docker build/push on push events.
- `.github/workflows/deploy.yml` provides manual production deployment via `workflow_dispatch`.

Image tagging behavior:

- Pushes to `main` publish `main-<short-sha>` tags.
- Version tags like `v1.2.3` publish `1.2.3` and `latest`.

## Deployment Setup (Template Reuse)

`deploy.yml` is a reusable template. Configure these repository secrets before using manual deploy:

- `SERVER_HOST` - target server hostname or IP
- `SERVER_USER` - SSH username
- `SSH_PRIVATE_KEY` - private key for SSH auth
- `APP_DIR` - absolute path to the app directory on the server

Server prerequisites:

- Docker and Docker Compose installed
- repository deployed on the server with `deploy/docker/docker-compose.yml` available

Manual deploy flow:

1. Open GitHub Actions and select the `Deploy` workflow.
2. Click `Run workflow`.
3. Enter `image_tag` (for example `main-a1b2c3d` or `1.2.3`).
4. Run and monitor deployment logs.

## Cron / scheduled jobs

- In-process mode is wired in `cmd/server/main.go` and runs jobs under the same `errgroup` cancellation context as HTTP, gRPC, and consumers.
- Separate mode is available in `cmd/cron/main.go` for production deployments where cron should run only once across multiple app instances.
- Jobs are registered through `internal/cron/scheduler.go` with structured start/completion/failure logging and graceful stop via `context.Context`.

Useful commands:

```bash
make build-cron
make run-cron
```

## Repository layer

- `internal/repository/user_repository.go` defines the data access contract.
- `internal/repository/postgres/user.go` provides the PostgreSQL implementation with `pgx/v5`.
- The server wires repositories in `cmd/server/main.go` and injects them via `router.Dependencies`.
- Readiness (`/health/ready`) now includes a PostgreSQL ping checker.

## HTML views (public site and admin)

Server-rendered pages use `html/template` under `views/`. Two layout families:

| Area | Layout | Handlers | Notes |
|------|--------|----------|--------|
| **Public (end-user)** | `layouts/public.html`, `views/public/` | `internal/handler/web` | Landing page at `/`. |
| **Admin** | `layouts/base.html`, `layouts/auth.html`, `views/admin/` | `internal/handler/admin` | Sign-in at `/admin/login` (form); dashboard at `/admin/dashboard`. |

**Admin browser sessions:** after a successful `POST /admin/login`, the server sets an HttpOnly cookie `session_token` (SameSite=Lax; `Secure` in production). Protected admin routes read this cookie via `middleware.NewAdminAuthenticate`. The JSON API under `/api/v1` continues to use `Authorization: Bearer <token>` from `POST /api/v1/auth/login` (`middleware.NewAuthenticate`).

## Endpoints

- `GET /` — public welcome page (HTML)
- `GET /admin/login`, `POST /admin/login` — admin sign-in form (HTML); sets session cookie on success
- `POST /admin/logout` — ends admin session (requires admin cookie)
- `GET /admin/dashboard` — admin dashboard (HTML; requires admin cookie)
- `GET /health/live`, `GET /health/ready` — liveness and readiness probes
- `GET /metrics` — Prometheus metrics endpoint
- `GET /api/v1/ping` — versioned API scaffold endpoint
- `POST /api/v1/files` — multipart upload (authenticated; requires `files:create` for manager role after migration `000005`)
- `GET /api/files/:filename` — download when `token` and `expires` HMAC query parameters are valid

## Observability

- **Health checks:** `GET /health/live` and `GET /health/ready` return typed JSON health responses.
- **Metrics:** `GET /metrics` exposes Prometheus metrics including request totals, request durations, and in-flight requests.
- **Tracing/metrics export:** set `OTEL_EXPORTER_ENDPOINT` to enable OTLP gRPC export for OpenTelemetry providers. Keep it empty to run with no-op export.

## Middleware stack

Registered in this order:
1. Recovery middleware (panic protection with stack-trace logging)
2. Prometheus metrics middleware
3. Request ID middleware (`X-Request-ID`)
4. Structured request logging middleware (`slog`)
5. Helmet security headers middleware
6. CORS middleware (configurable allowlist)
7. Body limit enforcement middleware
