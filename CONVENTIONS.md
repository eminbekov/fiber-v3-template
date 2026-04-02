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

## Performance

- Use typed response structs (4-9x faster than `fiber.Map`).
- Pre-allocate slices when capacity is known: `make([]T, 0, capacity)`.
- Prefer string concatenation over `fmt.Sprintf` in hot paths.
- Every column in a `WHERE` clause should have an index.
- Use `pgx.Batch` or JOINs to avoid N+1 query problems.

## Git

- Branch from `main`, open a pull request, keep changes reviewable.
- Conventional Commits for all messages: `<type>(<scope>): <description>`.
- Keep PRs under 400 changed lines when possible.
