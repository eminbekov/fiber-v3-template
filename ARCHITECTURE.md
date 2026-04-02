# Architecture

Architecture guide for **fiber-v3-template**. Describes layering, dependency direction, and the reasoning behind the project layout.

## Domain-Driven Design layers

Code is organized around business concepts, not technical infrastructure. Each layer has a clear responsibility and strict boundaries.

```
┌─────────────────────────────────────────────────────────────────────┐
│  HTTP Layer (Handlers)                                              │
│  ONLY: parse request, call service, write response                  │
│  NEVER: SQL queries, business logic, direct DB access               │
├─────────────────────────────────────────────────────────────────────┤
│  Service Layer (Business Logic)                                     │
│  ONLY: orchestration, validation, caching, event publishing         │
│  NEVER: HTTP parsing, SQL queries, response formatting              │
├─────────────────────────────────────────────────────────────────────┤
│  Repository Layer (Data Access)                                     │
│  ONLY: SQL queries, row scanning, error mapping                     │
│  NEVER: business logic, HTTP concerns, caching decisions            │
├─────────────────────────────────────────────────────────────────────┤
│  Domain Layer (Pure Models)                                         │
│  ONLY: structs, enums, errors, simple business methods              │
│  NEVER: imports from other layers, framework dependencies           │
└─────────────────────────────────────────────────────────────────────┘
```

### Dependency direction

Dependencies flow downward only. A handler can call a service, but a service must never import a handler. A repository can use a domain struct, but the domain must never import from the repository.

```
Handler -> Service -> Repository -> Domain
```

If you change a handler, nothing below it breaks. If you change the domain, everything above may need to update — which is why the domain is kept stable and simple.

### What goes where

**Domain** (`internal/domain/`):
- Pure Go structs with no external-library tags
- Business methods like `IsActive()`, `IsDeleted()`
- Typed enums (`UserStatusActive`, `UserStatusDisabled`)
- Sentinel errors (`ErrNotFound`, `ErrUnauthorized`, `ErrConflict`)

**Repository** (`internal/repository/`):
- Interfaces defining data access contracts
- Implementations under `postgres/` with hand-written SQL and `pgx/v5`
- Each implementation receives a database pool via constructor injection

**Service** (`internal/service/`):
- Business logic, validation, cache-aside reads, event publishing
- Receives repositories and cache via constructor injection
- Never touches HTTP or SQL directly

**Handler** (`internal/handler/`):
- Thin HTTP layer — ideally under 30 lines per method
- Parse request, call service, write response
- Translate domain errors to HTTP status codes via a shared error handler

**DTO** (`internal/dto/`):
- `request/` — incoming payloads with `json` and `validate` struct tags
- `response/` — outgoing payloads with `json` tags, versioned under `v1/`
- Domain models remain clean and reusable across all layers

## Directory layout

```text
.
├── cmd/
│   ├── server/              # Main HTTP + app wiring
│   ├── migrate/             # Migration CLI
│   ├── seed/                # Development database seeder
│   ├── console/             # Optional console CLI commands
│   ├── generate/            # Optional code generator CLI
│   └── cron/                # Optional cron binary
├── deploy/
│   ├── docker/              # Dockerfile and compose manifests
│   ├── k8s/                 # Optional Kubernetes manifests and EnvoyFilter
│   └── monitoring/          # Optional observability stack configs
├── internal/
│   ├── config/              # Env config parsing/validation
│   ├── database/            # pgx pool, registry, query tracer
│   ├── domain/              # Pure business models and sentinel errors
│   ├── repository/          # Repository interfaces
│   │   └── postgres/        # PostgreSQL implementations
│   ├── service/             # Business services
│   ├── handler/             # HTTP handlers
│   │   ├── api/v1/          # Versioned REST API handlers
│   │   ├── admin/           # Optional admin HTML handlers
│   │   └── web/             # Optional public HTML handlers
│   ├── dto/
│   │   ├── request/         # Incoming request DTOs (validation tags)
│   │   └── response/        # Outgoing response DTOs (versioned)
│   ├── middleware/           # Request middleware stack
│   ├── router/              # Route registration
│   ├── cache/               # Cache interface and Redis implementation
│   ├── session/             # Session store (Redis backend)
│   ├── helpers/             # Shared validation/utility functions
│   ├── nats/                # Optional NATS module
│   ├── grpc/                # Optional gRPC module
│   ├── websocket/           # Optional websocket module
│   ├── storage/             # Optional file storage module
│   ├── i18n/                # Optional i18n module
│   ├── cron/                # Optional cron scheduler
│   ├── console/             # Optional console commands
│   └── generate/            # Optional code generator
├── package/                 # Reusable packages (could be extracted to separate repos)
│   ├── hasher/              # Argon2id password hashing
│   ├── health/              # Liveness and readiness checks
│   ├── logger/              # slog configuration
│   └── telemetry/           # OpenTelemetry setup
├── migrations/              # Sequential SQL migrations
├── views/                   # Optional HTML templates
├── proto/                   # Protobuf definitions
├── gen/                     # Generated protobuf Go code
├── docs/                    # Generated Swagger/OpenAPI files
├── .env.example
├── setup.sh
├── Makefile
├── AGENTS.md
├── CONVENTIONS.md
├── ARCHITECTURE.md
├── TESTING.md
├── SECURITY.md
└── README.md
```

### Why `internal/` vs `package/`

- `internal/` is a Go language feature: code inside `internal/` cannot be imported by other Go modules. This is private application code.
- `package/` contains reusable utilities (logger, telemetry, health) that could be shared across projects. Placing them outside `internal/` signals they have no application-specific dependencies.

### Why interfaces for repositories

1. **Testability**: unit tests use mock repositories without needing a real database.
2. **Swappability**: switching from PostgreSQL to another storage engine requires only a new implementation, no service changes.
3. **Multiple pools**: the same interface can work with different database pools (production, read replica) because the pool is injected via constructor.

### Why separate `dto/request/` and `dto/response/`

- **Request DTOs** have `json` and `validate` tags for parsing and validation.
- **Response DTOs** have `json` tags for serialization and may exclude sensitive fields.
- **Domain models** remain pure, reusable, and free of framework-specific tags.

## Entry point pattern

`cmd/server/main.go` follows the `run(ctx)` pattern:

1. Trap signals (`SIGINT`, `SIGTERM`)
2. Load and validate configuration
3. Set up structured logging
4. Initialize telemetry
5. Connect to databases, Redis, NATS
6. Wire repositories, services, handlers
7. Start HTTP, gRPC, consumers, and cron via `errgroup`
8. Graceful shutdown on context cancellation

All cleanup is handled via `defer` in reverse initialization order.

## Common mistakes

| Mistake | Fix |
|---|---|
| Business logic in handlers | Move logic to services; handlers only parse, delegate, respond |
| Global database variable | Inject pool via constructor |
| Ignoring errors with `_` | Handle or log every error; return critical errors |
| Leaking internal errors to clients | Log full error server-side; return generic message to client |
| N+1 query problem | Use `pgx.Batch` or JOINs to minimize round trips |
