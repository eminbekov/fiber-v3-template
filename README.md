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
│   └── router/          # Fiber app and route registration
├── package/
│   └── logger/          # slog setup (JSON in production, text in development)
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

## Run

```bash
go run ./cmd/server
```

With overrides:

```bash
ENVIRONMENT=production LOG_LEVEL=info HTTP_LISTEN_ADDRESS=:3000 go run ./cmd/server
```

## Endpoints

- `GET /` — JSON placeholder
- `GET /health/live`, `GET /health/ready` — liveness and readiness probes
