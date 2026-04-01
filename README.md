# fiber-v3-template

Starter layout for a Go HTTP API using [Fiber v3](https://github.com/gofiber/fiber). This repository will grow as the template is filled out.

## Requirements

- Go 1.26+

## Project layout

```text
.
├── cmd/server/          # HTTP server entrypoint (config, logger, graceful shutdown)
├── internal/
│   ├── config/          # Typed configuration from environment variables
│   ├── domain/          # Sentinel business errors and core domain types
│   ├── dto/response/    # Standard API success/error envelopes
│   ├── handler/         # Centralized error handler + API v1 handlers
│   ├── middleware/      # Recovery, metrics, request ID, logging, CORS, Helmet, body limit
│   └── router/          # Fiber app + middleware and route registration
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

## Run

```bash
go run ./cmd/server
```

With overrides:

```bash
ENVIRONMENT=production \
LOG_LEVEL=info \
HTTP_LISTEN_ADDRESS=:3000 \
CORS_ALLOW_ORIGINS=https://example.com,https://admin.example.com \
BODY_LIMIT=4194304 \
OTEL_EXPORTER_ENDPOINT=localhost:4317 \
go run ./cmd/server
```

## Endpoints

- `GET /` — service info (`data` envelope with typed payload)
- `GET /health/live`, `GET /health/ready` — liveness and readiness probes
- `GET /metrics` — Prometheus metrics endpoint
- `GET /api/v1/ping` — versioned API scaffold endpoint

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
