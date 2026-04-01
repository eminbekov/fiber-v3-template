# fiber-v3-template

Starter layout for a Go HTTP API using [Fiber v3](https://github.com/gofiber/fiber). This repository will grow as the template is filled out.

## Requirements

- Go 1.26+

## Run

```bash
go run ./cmd/server
```

The server listens on `:8080` by default. Override with `HTTP_LISTEN_ADDRESS` (for example `HTTP_LISTEN_ADDRESS=:3000`).

## Endpoints

- `GET /` — JSON placeholder
- `GET /health/live`, `GET /health/ready` — liveness and readiness probes
