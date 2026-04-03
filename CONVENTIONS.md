# Coding conventions

Human-readable rules for **fiber-v3-template**. Automated checks may enforce subsets of these via `golangci-lint` and code review.

## Project rules (hard limits)

| Rule | Limit | Rationale |
|---|---|---|
| Max function length | ~60 lines (linter often ~80) | Easier to test and reason about |
| Max nesting depth | 4 levels | Use guard clauses to stay flat |
| Error returns | No ignoring with `_` | Every error must be handled or logged |
| Global mutable state | Not allowed | Use constructors and explicit wiring |
| Empty interface | Prefer `any`; justify if needed | Clear, modern Go |
| HTTP responses | Typed structs | Avoid `fiber.Map` for fixed shapes |
| Dynamic maps | Only when keys are unknown at compile time | Otherwise use structs |

## Naming

| Element | Convention | Good | Bad |
|---|---|---|---|
| Files | `snake_case.go` | `user_service.go` | `userService.go` |
| Packages | lowercase, no underscores | `package domain` | `package user_domain` |
| Exported identifiers | `CamelCase` | `UserService` | `User_service` |
| Unexported identifiers | `camelCase` | `buildQuery` | `build_query` |
| Constants | `CamelCase`/`camelCase` | `MaxRetries` | `MAX_RETRIES` |
| Acronyms | consistent all-caps | `UserID`, `ParseURL` | `UserId`, `ParseUrl` |

- Use **full words**, not abbreviations (`permission` not `perm`, `subscription` not `sub`).
- Fiber handlers: name the receiver parameter `ctx` (type `fiber.Ctx`), never `c`.
- Standard `context.Context` parameters: always named `ctx`.
- Avoid Java-style `Get` prefix on getters; use Go idioms (`Name()` not `GetName()`).
- Avoid package stuttering (`user.Service` not `user.UserService`).

## Imports

1. Standard library
2. Third-party
3. This repository

Separate groups with a blank line.

```go
import (
    "context"
    "fmt"

    "github.com/gofiber/fiber/v3"

    "github.com/eminbekov/fiber-v3-template/internal/domain"
)
```

## Errors

- Last return value is always `error`.
- Wrap with `%w` when propagating: `fmt.Errorf("userRepository.FindByID: %w", err)`.
- Use sentinel errors in `internal/domain` for business outcomes (`ErrNotFound`, `ErrUnauthorized`, `ErrConflict`, `ErrValidation`, `ErrForbidden`).
- Return early on errors; avoid `else` after `return`.

```go
// Preferred: guard clause style
if err != nil {
    return fmt.Errorf("operationName: %w", err)
}
```

## Functions

- `context.Context` is the first parameter when the work can be canceled or timed out.
- Value receivers when the method only reads; pointer receivers when it mutates. If any method uses a pointer receiver, prefer pointer receivers for all methods on that type.

```go
// Value receiver: does not change the User
func (user User) IsActive() bool {
    return user.Status == "active"
}

// Pointer receiver: changes the User
func (user *User) Deactivate() {
    user.Status = "inactive"
}
```

## Documentation

- Doc comments should explain behavior from the caller's perspective and document error conditions (for example when `domain.ErrNotFound` is returned).
- Do not restate the obvious. `// FindByID finds a user by ID` adds nothing. `// FindByID returns the active user with the given ID. Returns domain.ErrNotFound if the user does not exist or has been soft-deleted.` is useful.

## Concurrency

- Do not start goroutines without cancellation (`context`) and panic recovery where appropriate.
- Set timeouts on external calls via `context.WithTimeout` when appropriate.
- Use `errgroup` for managed goroutine groups with error collection.

```go
group, groupContext := errgroup.WithContext(ctx)
group.Go(func() error {
    defer func() {
        if recovered := recover(); recovered != nil {
            slog.Error("panic in worker", "panic", recovered)
        }
    }()
    return processWork(groupContext)
})
```

## REST API best practices

### HTTP methods and status codes

| Method | Purpose | Success Code | Example |
|---|---|---|---|
| `GET` | Retrieve a resource or list | `200 OK` | `GET /api/users/123` |
| `POST` | Create a new resource | `201 Created` | `POST /api/users` |
| `PUT` | Full update of a resource | `200 OK` | `PUT /api/users/123` |
| `PATCH` | Partial update of a resource | `200 OK` | `PATCH /api/users/123` |
| `DELETE` | Delete a resource | `204 No Content` | `DELETE /api/users/123` |

Common error codes:

| Code | When to Use |
|---|---|
| `400 Bad Request` | Invalid JSON, validation failed, missing required fields |
| `401 Unauthorized` | Missing or invalid authentication token |
| `403 Forbidden` | Authenticated but not authorized (wrong role/permission) |
| `404 Not Found` | Resource does not exist |
| `409 Conflict` | Duplicate (e.g., username already taken) |
| `413 Payload Too Large` | Request body exceeds body limit |
| `422 Unprocessable Entity` | Valid JSON but semantically wrong (e.g., end date before start date) |
| `429 Too Many Requests` | Rate limit exceeded |
| `500 Internal Server Error` | Unexpected server error (never expose details) |

### URL naming conventions

```
GET    /api/users              # List users (plural noun)
GET    /api/users/:id          # Get single user
POST   /api/users              # Create user
PUT    /api/users/:id          # Update user
DELETE /api/users/:id          # Delete user
GET    /api/users/:id/orders   # List orders for a user (nested resource)
POST   /api/users/:id/orders   # Create order for a user
```

**Rules:**
- Always use **plural nouns** for resources: `/users` not `/user`
- Use **kebab-case** for multi-word resources: `/privacy-policies` not `/privacyPolicies`
- Use **path parameters** for resource identity: `/users/:id`
- Use **query parameters** for filtering, sorting, pagination: `/users?role=manager&sort=created_at&page=2`
- Never use verbs in URLs: `/users` with `POST` method, not `/create-user`
- Version your API with a prefix: `/api/v1/users` (add when needed, not from day one)

### Pagination response headers

Return pagination metadata in HTTP response headers, not in the JSON body.

**OFFSET pagination response:**

```go
ctx.Set("X-Total-Count", strconv.FormatInt(totalCount, 10))
ctx.Set("X-Page", strconv.Itoa(page))
ctx.Set("X-Page-Size", strconv.Itoa(pageSize))
ctx.Set("X-Total-Pages", strconv.Itoa(totalPages))
```

**Cursor pagination response (for log tables):**

```go
ctx.Set("X-Next-Cursor", nextCursor)
ctx.Set("X-Has-More", strconv.FormatBool(hasMore))
```

### Rate limit response headers

Follow the IETF standard (RFC draft `ratelimit-headers`):

| Header | Meaning | Example |
|---|---|---|
| `RateLimit-Limit` | Max requests allowed per window | `100` |
| `RateLimit-Remaining` | Requests remaining in current window | `87` |
| `RateLimit-Reset` | Seconds until the window resets | `42` |
| `Retry-After` | Seconds to wait (only on `429` responses) | `30` |

### Standard response envelope

**Success (single resource):**
```json
{
  "data": {
    "id": "01961a2c-3e4f-7abc-8def-1234567890ab",
    "username": "john",
    "full_name": "John Doe"
  }
}
```

**Success (list -- pagination info is in response headers):**
```json
{
  "data": [
    {"id": "...", "username": "john"},
    {"id": "...", "username": "jane"}
  ]
}
```

**Error:**
```json
{
  "error": {
    "message": "validation failed",
    "details": [
      {"field": "phone", "message": "must be a valid E.164 phone number"},
      {"field": "username", "message": "must be at least 3 characters"}
    ]
  }
}
```

### API versioning

| Strategy | Format | Example |
|---|---|---|
| **URL path** (recommended) | `/api/v1/users` | `GET /api/v1/users`, `GET /api/v2/users` |
| **Header** | `Accept-Version: v1` | Custom header per request |

Each API version gets its own package under `handler/api/`. Response DTOs are versioned under `dto/response/v1/`, `dto/response/v2/`.

## Performance

- Use typed response structs (4-9x faster than `fiber.Map`).
- Pre-allocate slices when capacity is known: `make([]T, 0, capacity)`.
- Prefer string concatenation over `fmt.Sprintf` in hot paths.
- Every column in a `WHERE` clause should have an index.
- Use `pgx.Batch` or JOINs to avoid N+1 query problems.
- Use `sync.Pool` for reusable buffers in hot paths.

```go
// BAD: fmt.Sprintf allocates a new string every call
key := fmt.Sprintf("user:%s", id.String())

// GOOD: string concatenation is faster for simple cases
key := "user:" + id.String()

// BAD: append grows the slice, causing re-allocations
var users []domain.User
for rows.Next() {
    users = append(users, scanUser(rows))
}

// GOOD: pre-allocate with known capacity
users := make([]domain.User, 0, pageSize)
for rows.Next() {
    users = append(users, scanUser(rows))
}
```

## Recommended dependencies

### Use these

| Package | Purpose | Why This One |
|---|---|---|
| `github.com/gofiber/fiber/v3` | HTTP framework | Fastest Go framework, Express-like API |
| `github.com/jackc/pgx/v5` | PostgreSQL driver | Fastest PG driver, native pgxpool, OTEL tracer hook |
| `github.com/redis/go-redis/v9` | Redis client | Official Redis client, full feature coverage |
| `github.com/nats-io/nats.go` | Message queue | JetStream persistence, lightweight |
| `go.opentelemetry.io/otel` | Tracing/metrics | Industry standard, vendor-neutral |
| `github.com/prometheus/client_golang` | Metrics | De facto standard for metrics in Go |
| `github.com/bytedance/sonic` | Fast JSON | 2-3x faster than encoding/json, drop-in replacement |
| `github.com/stretchr/testify` | Testing | Assertions, mocking, test suites |
| `golang.org/x/sync` | errgroup | Managed goroutine groups with error collection |
| `golang.org/x/crypto` | Argon2id | OWASP-recommended password hashing |
| `github.com/go-playground/validator/v10` | Validation | Struct tag-based validation, extensive rule set |
| `github.com/golang-migrate/migrate/v4` | Migrations | SQL migration files, up/down, CLI and library modes |
| `github.com/gofrs/uuid/v5` | UUID v7 | Time-sortable UUIDs for external-facing primary keys |
| `github.com/swaggo/swag` | Swagger code gen | Generates OpenAPI spec from Go annotations |
| `google.golang.org/grpc` | gRPC server/client | Official Go gRPC implementation |
| `google.golang.org/protobuf` | Protobuf runtime | Official protobuf Go runtime |

### Prefer standard library

| Use stdlib | Instead of | Why |
|---|---|---|
| `log/slog` | Zap, Logrus, zerolog | Zero dependency, built into Go 1.21+, JSON output for Grafana/Loki |
| `crypto/hmac` + `crypto/sha256` | Third-party HMAC libs | Stdlib is audited, maintained by Go team |
| `testing` + `testing.F` (fuzz) | Custom test frameworks | Native support, `go test` just works |
| `net/http` (for HTTP clients) | resty, gentleman | Stdlib is simple and sufficient for most cases |

### Avoid these

| Pattern | Why | Use Instead |
|---|---|---|
| GORM / any ORM | Hides SQL complexity, slower, opaque errors | pgx + hand-written SQL |
| `fiber.Map` / `map[string]any` | 4-9x slower than structs, GC pressure, no type safety | Typed response structs |
| `interface{}` syntax | Legacy Go syntax | `any` (Go 1.18+ alias) |
| Global mutable state | Hidden dependencies, untestable, race conditions | Dependency injection via constructors |
| `init()` with side effects | Runs before main(), hard to test, unpredictable order | Explicit initialization in `main.go` |
| Fire-and-forget goroutines | Unrecoverable panics, leaked goroutines, no shutdown | `errgroup` + `context.Context` |
| `fmt.Sprintf` in hot paths | Allocates a new string every call | `strconv` or string concatenation |
| `SELECT *` | Fetches unnecessary data, breaks when columns change | Explicit column lists |

## Git

- Branch from `main`, open a pull request, keep changes reviewable.
- Conventional Commits for all messages: `<type>(<scope>): <description>`.
- Keep PRs under 400 changed lines when possible.
- Run `make verify` before every push to catch lint, vet, and test failures early.
- Never commit directly to `main`; all changes go through pull requests.

## Makefile reference

All common development tasks are available as `make` targets. Run `make help` for the full list.

### Build and run

| Target | Description |
|---|---|
| `make build` | Build the application binary (`CGO_ENABLED=0`, stripped) |
| `make run` | Run the application (development) |
| `make clean` | Remove build artifacts |
| `make dev` | Run with hot reload (requires `air`) |
| `make fmt` | Format all Go files |
| `make tidy` | Clean up `go.mod` and `go.sum` |

### Database

| Target | Description |
|---|---|
| `make migrate-up` | Run all pending migrations |
| `make migrate-down` | Rollback the last migration |
| `make migrate-create NAME=...` | Create a new migration |
| `make seed` | Seed the development database |

### Code quality

| Target | Description |
|---|---|
| `make lint` | Run golangci-lint |
| `make vet` | Run `go vet` |
| `make security` | Run `govulncheck` + `gosec` |
| `make verify` | Full pre-push check (fmt, tidy, build, vet, lint, test) |

### Testing

| Target | Description |
|---|---|
| `make test` | Run all unit tests with race detector |
| `make test-verbose` | Run tests with verbose output |
| `make test-cover` | Generate coverage report (`coverage.out` + `coverage.html`) |
| `make test-integration` | Run integration tests (requires Docker) |
| `make bench` | Run benchmark tests |
| `make fuzz` | Run fuzz tests (30 second budget) |

### Docker

| Target | Description |
|---|---|
| `make docker-build` | Build Docker image |
| `make docker-up` | Start full stack with Docker Compose |
| `make docker-down` | Stop all Docker Compose services |
| `make docker-dev` | Start development services (DB, Redis, NATS only) |
| `make docker-logs` | Tail Docker Compose logs |

### Console commands

| Target | Description |
|---|---|
| `make create-admin USERNAME=...` | Create an admin user |
| `make assign-role USER_ID=... ROLE=...` | Assign a role to a user |
| `make cache-clear` | Flush all Redis cache |
| `make export-users` | Export all users to CSV |

## Common mistakes to avoid

| Mistake | Fix |
|---|---|
| Business logic in handlers | Move logic to services; handlers only parse, delegate, respond |
| Global database variable | Inject pool via constructor |
| Ignoring errors with `_` | Handle or log every error; return critical errors |
| Leaking internal errors to clients | Log full error server-side; return generic message to client |
| N+1 query problem | Use `pgx.Batch` or JOINs to minimize round trips |
| Using `fiber.Map` for responses | Use typed response structs (4-9x faster) |
| `fmt.Sprintf` in hot paths | Use string concatenation or `strconv` |
| `SELECT *` in queries | Use explicit column lists |
| Fire-and-forget goroutines | Use `errgroup` + `context.Context` |
| Using ORMs (GORM, etc.) | Use `pgx` + hand-written SQL |
