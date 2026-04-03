# Architecture

Architecture guide for **fiber-v3-template**. Describes layering, dependency direction, database patterns, caching, observability, deployment, and the reasoning behind the project layout.

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

If you change a handler, nothing below it breaks. If you change the domain, everything above may need to update -- which is why the domain is kept stable and simple.

### What goes where

**Domain** (`internal/domain/`):
- Pure Go structs with no external-library tags
- Business methods like `IsActive()`, `IsDeleted()`
- Typed enums (`UserStatusActive`, `UserStatusDisabled`)
- Sentinel errors (`ErrNotFound`, `ErrUnauthorized`, `ErrConflict`, `ErrValidation`, `ErrForbidden`)

**Repository** (`internal/repository/`):
- Interfaces defining data access contracts
- Implementations under `postgres/` with hand-written SQL and `pgx/v5`
- Each implementation receives a database pool via constructor injection

**Service** (`internal/service/`):
- Business logic, validation, cache-aside reads, event publishing
- Receives repositories and cache via constructor injection
- Never touches HTTP or SQL directly

**Handler** (`internal/handler/`):
- Thin HTTP layer -- ideally under 30 lines per method
- Parse request, call service, write response
- Translate domain errors to HTTP status codes via a shared error handler

**DTO** (`internal/dto/`):
- `request/` -- incoming payloads with `json` and `validate` struct tags
- `response/` -- outgoing payloads with `json` tags, versioned under `v1/`
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

## Database patterns

### Multi-database pool registry

In production, you may need multiple database pools: production, development, read replica, or a separate logs database. The registry pattern manages all of these.

```go
type Registry struct {
    pools map[string]*pgxpool.Pool
}

func NewRegistry(ctx context.Context, configs map[string]PostgresConfig) (*Registry, error) {
    registry := &Registry{pools: make(map[string]*pgxpool.Pool)}
    for name, config := range configs {
        pool, err := createPool(ctx, config)
        if err != nil {
            registry.Close()
            return nil, fmt.Errorf("registry: pool %s: %w", name, err)
        }
        registry.pools[name] = pool
    }
    return registry, nil
}

func (registry *Registry) Pool(name string) *pgxpool.Pool {
    return registry.pools[name]
}
```

Pool selection happens in `main.go`, not in business code:

```go
productionPool := registry.Pool("production")
userRepository := postgres.NewUserRepository(productionPool)
userService := service.NewUserService(userRepository, cache)
```

### Automatic struct scanning with pgx v5

Use `db` struct tags to map column names to struct fields. pgx v5 has built-in generic scanning:

| Function | Use Case |
|---|---|
| `pgx.RowToStructByName[T]` | Maps columns to struct fields by `db` tag name (recommended) |
| `pgx.RowToStructByNameLax[T]` | Same but ignores missing columns (for partial SELECTs or LEFT JOINs) |
| `pgx.RowToStructByPos[T]` | Maps columns to struct fields by position order (fragile, avoid) |
| `pgx.CollectRows(rows, fn)` | Collects all rows into a `[]T` |
| `pgx.CollectOneRow(row, fn)` | Collects a single row into `T` |
| `pgx.CollectExactlyOneRow(rows, fn)` | Returns error if not exactly one row |

Use `ByName` (strict) for standard queries. Use `ByNameLax` for LEFT JOINs or optional columns.

### SQL query performance

**Rule: every column in a `WHERE` clause must have an index.**

Index types and when to use them:

| Index Type | Use When | Example |
|---|---|---|
| B-Tree (default) | Equality (`=`) and range (`>`, `<`, `BETWEEN`) | `CREATE INDEX idx_users_phone ON users(phone)` |
| Unique B-Tree | Column values must be unique | `CREATE UNIQUE INDEX idx_users_username ON users(username)` |
| Composite | Multiple columns filtered together | `CREATE INDEX idx_users_role_status ON users(role, status)` |
| Covering (`INCLUDE`) | Avoid reading the table entirely | `CREATE INDEX idx_users_status ON users(status) INCLUDE (full_name, phone)` |
| GIN | JSONB fields, full-text search, arrays | `CREATE INDEX idx_users_metadata ON users USING GIN(metadata)` |
| BRIN | Time-series data (logs, events) | `CREATE INDEX idx_logs_created ON audit_logs USING BRIN(created_at)` |
| Partial | Subset of rows | `CREATE INDEX idx_active ON users(id) WHERE status = 'active'` |

**Composite index column order matters.** Put equality columns first, then range columns:

```sql
-- GOOD: equality (role) first, then range (created_at)
CREATE INDEX idx_users_role_created ON users(role, created_at);

-- BAD: range first -- PostgreSQL cannot use the index efficiently for role filtering
CREATE INDEX idx_users_created_role ON users(created_at, role);
```

**Validate every query with EXPLAIN during development:**

```sql
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM users WHERE role = 'manager' AND status = 'active';
-- GOOD output: "Index Scan using idx_users_role_status"
-- BAD output: "Seq Scan on users" -- means no index is being used
```

**Pagination strategy:**
- OFFSET for small tables (thousands of rows, numbered page buttons)
- Cursor/keyset for large tables (millions of rows, "load more" UI)

**Minimize database round trips** with `pgx.Batch`:

```go
batch := &pgx.Batch{}
batch.Queue("SELECT ... FROM users WHERE id = $1", id)
batch.Queue("SELECT ... FROM subscriptions WHERE user_id = $1", id)
batch.Queue("SELECT ... FROM user_stats WHERE user_id = $1", id)
results := pool.SendBatch(ctx, batch)
defer results.Close()
```

## Caching with Redis

### Cache interface

```go
type Cache interface {
    Get(ctx context.Context, key string, destination any) error
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Delete(ctx context.Context, keys ...string) error
    DeleteByPrefix(ctx context.Context, prefix string) error
}
```

### Type-safe key builders

Never build cache keys with raw string concatenation. Use centralized key builders:

```go
package keys

func UserByID(id uuid.UUID) string         { return "user:" + id.String() }
func UserList(filterHash string) string     { return "user:list:" + filterHash }
func DashboardStats() string               { return "dashboard:stats" }
func TokenToUser(token string) string       { return "token:" + token }
```

### Caching strategy

| Endpoint | Cache Key | TTL | Invalidate When |
|---|---|---|---|
| GET /api/users/:id | `user:{id}` | 5 min | User updated or deleted |
| GET /api/users | `user:list:{filter_hash}` | 2 min | Any user mutated |
| GET /admin/dashboard | `dashboard:stats` | 1 min | Periodic refresh |
| GET /api/settings | `settings:{key}` | 10 min | Setting updated |
| Token validation | `token:{token_string}` | Matches token expiry | Logout |

### Cache-aside pattern

```go
func (service *UserService) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    var user domain.User
    cacheKey := keys.UserByID(id)
    if err := service.cache.Get(ctx, cacheKey, &user); err == nil {
        return &user, nil // Cache hit
    }

    result, err := service.userRepository.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("userService.FindByID: %w", err)
    }

    _ = service.cache.Set(ctx, cacheKey, result, 5*time.Minute)
    return result, nil
}
```

**Rule: always delete a cache entry rather than trying to update it.** Deleting is simple and safe. Updating risks cache/DB inconsistency.

## Observability

### Three pillars

- **Logs**: What happened? (structured text records of events via `slog`)
- **Metrics**: How is it performing? (numbers: request count, latency, error rate via Prometheus)
- **Traces**: Where is the time spent? (request flow across services via OpenTelemetry)

### Monitoring stack

```
Your Go App
  |
  |--(logs to stdout)-----> Promtail -----> Loki -----> Grafana (log search)
  |
  |--(/metrics endpoint)--> Prometheus ---> Grafana (dashboards, alerts)
  |
  |--(OTLP gRPC)---------> OTEL Collector -> Prometheus -> Grafana (traces)
```

| Component | Role |
|---|---|
| **Promtail** | Reads container logs, ships to Loki |
| **Loki** | Stores and indexes logs |
| **Prometheus** | Scrapes /metrics, stores time-series data |
| **Grafana** | Visualizes logs and metrics in dashboards |
| **OTEL Collector** | Receives traces/metrics, exports to backends |

### Structured logging with slog

```go
// Production: JSON output (parsed by Promtail)
slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})

// Development: human-readable text output
slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})

// Usage
slog.Info("user created", "user_id", user.ID, "username", user.Username)
slog.ErrorContext(ctx, "failed to process payment", "error", err, "order_id", order.ID)
```

### Database query time logging

pgx's tracer hook logs every SQL query's execution time:
- **Every query at DEBUG level** (visible in development)
- **Slow queries (>100ms) at WARN level** (visible in production)
- **Failed queries at ERROR level** (always visible)

### Health checks

- **`/health/live`** (liveness): "Is the process running?" Kubernetes uses this to decide whether to restart the container.
- **`/health/ready`** (readiness): "Can the app serve traffic?" Returns 200 only if all dependencies (database, Redis, NATS) are healthy.

## Docker and deployment

### Multi-stage Dockerfile

Build image is ~1.5GB (full Go toolchain). Runtime image is ~15MB (just Alpine + binary).

```dockerfile
# Stage 1: Build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Stage 2: Runtime
FROM alpine:3.23
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
USER appuser
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:8080/health/live || exit 1
EXPOSE 8080
CMD ["./server"]
```

### Docker Compose

- **Production** (`docker-compose.yml`): full stack with app, Redis, NATS, Prometheus, Loki, Grafana
- **Development** (`docker-compose.dev.yml`): minimal with Postgres, Redis, NATS only

### Image tagging strategy

| Trigger | Image Tag | Latest? |
|---|---|---|
| Git tag `v1.2.3` on main | `1.2.3` | Yes |
| Commit on main (no tag) | `main-{short-sha}` | No |
| Branch (not main) | Not pushed | -- |

## CI/CD pipeline

### CI (automatic on every push)

```
Push to Git
  |
  v
[lint] golangci-lint run          Catches code quality issues
  |
  v
[security] govulncheck ./...      Checks for known Go CVEs
           trivy image scan       Docker image CVEs
  |
  v
[test] unit -> fuzz -> integration -> benchmark
  |
  v
[build] docker build + push       Build and push to registry
```

### CD (manual deployment)

CD is never automatic for production. A human clicks "Deploy" or approves, specifying which image tag. Rollback is instant: deploy the previous tag.

Supported platforms: GitLab CI/CD, GitHub Actions, ArgoCD, Jenkins, Drone CI, Flux, Terraform.

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

### Graceful shutdown

When you press Ctrl+C or Kubernetes sends SIGTERM:

1. **Stop accepting new connections** (close HTTP and gRPC listeners)
2. **Finish all in-flight requests** (up to 10 seconds)
3. **Stop gRPC server gracefully** (finish active RPCs, reject new ones)
4. **Drain NATS messages** (finish processing, acknowledge)
5. **Close connections** (Redis, database pools)
6. **Flush telemetry** (send remaining traces/metrics)
7. **Exit cleanly** (exit code 0)

Without graceful shutdown, in-flight requests get killed mid-execution, database transactions are left open, and users see random errors during deployments.

## Common mistakes

| Mistake | Fix |
|---|---|
| Business logic in handlers | Move logic to services; handlers only parse, delegate, respond |
| Global database variable | Inject pool via constructor |
| Ignoring errors with `_` | Handle or log every error; return critical errors |
| Leaking internal errors to clients | Log full error server-side; return generic message to client |
| N+1 query problem | Use `pgx.Batch` or JOINs to minimize round trips |
