# AI agent instructions

Rules for automated tools and contributors working on **fiber-v3-template**. Aligns with `CONVENTIONS.md`, `ARCHITECTURE.md`, and project linting.

## Git workflow

### Branch naming

| Prefix | When to Use | Example |
|---|---|---|
| `feature/` or `feat/` | New functionality | `feature/user-export` |
| `fix/` or `bugfix/` | Bug fixes | `fix/login-timeout` |
| `hotfix/` | Urgent production fixes | `hotfix/payment-crash` |
| `refactor/` | Code restructuring | `refactor/user-service` |
| `chore/` | Non-code tasks (CI, deps) | `chore/update-dependencies` |
| `docs/` | Documentation | `docs/api-guide` |
| `test/` | Adding or fixing tests | `test/user-service-unit` |
| `release/` | Release preparation | `release/v1.3.0` |

**Rules:**
- All lowercase, words separated by hyphens: `feature/user-export`
- Include ticket number when applicable: `fix/CM-234-login-timeout`
- Short but descriptive: `feature/rbac` not `feature/implement-full-role-based-access-control`
- Never commit directly to `main`

### Branching strategy (GitHub Flow)

```
main (always deployable)
  |
  +-- feature/user-export -----> PR -> Review -> Merge to main
  |
  +-- fix/login-timeout -------> PR -> Review -> Merge to main
  |
  +-- hotfix/payment-crash -----> PR -> Review -> Merge to main (urgent)
```

1. `main` is always deployable
2. Branch from `main` for every change
3. Commit frequently with clear messages
4. Open a pull request when ready
5. Get at least one approval
6. Merge (squash for features, regular for large refactors)
7. Delete the branch after merge
8. Deploy from `main` with tagged releases

### Commit messages (Conventional Commits)

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**

| Type | When | Semver Impact |
|---|---|---|
| `feat` | New feature | MINOR (1.x.0) |
| `fix` | Bug fix | PATCH (1.0.x) |
| `docs` | Documentation only | None |
| `style` | Formatting (no logic change) | None |
| `refactor` | Restructuring (no behavior change) | None |
| `perf` | Performance improvement | PATCH |
| `test` | Adding or fixing tests | None |
| `build` | Build system or dependencies | None |
| `ci` | CI/CD configuration | None |
| `chore` | Maintenance | None |

**Breaking changes** add `!`: `feat(api)!: change pagination format`

**Good examples:**

```bash
git commit -m "feat(auth): add Argon2id password hashing"
git commit -m "fix(api): return 404 instead of 500 for missing user"
git commit -m "refactor(repository): extract row scanning helpers"
git commit -m "perf(database): add composite index on users(role, status)"
git commit -m "test(service): add table-driven tests for UserService.FindByID"
```

**Bad examples:**

```bash
git commit -m "fix stuff"       # vague, no type, no context
git commit -m "update"          # meaningless
git commit -m "WIP"             # do not commit work-in-progress to main
git commit -m "changes"         # says nothing
```

### Pull request best practices

| Practice | Why |
|---|---|
| Keep PRs under 400 lines changed | Large PRs get rubber-stamped without real review |
| Write a clear description | Reviewers need context |
| Link to the issue/ticket | Traceability |
| Run CI before requesting review | Do not waste reviewer time on broken code |
| Squash merge for features | Clean `main` history |
| Delete branch after merge | Keep the repo tidy |

### Git commands reference

```bash
# Start new work
git checkout main && git pull origin main
git checkout -b feature/user-export

# Commit
git add .
git commit -m "feat(export): add Excel export for user list"

# Push and create PR
git push -u origin feature/user-export

# Keep branch up to date
git checkout main && git pull origin main
git checkout feature/user-export && git rebase main
git push --force-with-lease

# After merge, clean up
git checkout main && git pull origin main
git branch -d feature/user-export

# Tag a release
git tag -a v1.3.0 -m "Release v1.3.0: add RBAC and audit logging"
git push origin v1.3.0
```

### Pre-push verification

Run `make verify` (fmt, tidy, build, vet, lint, test) before every push. Push the branch to origin and open a pull request; do not merge locally.

### Branch and commit rules by change area

| Change area | Branch prefix example | Commit message example |
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
| Setup script or installer changes | `fix/setup-git-cleanup` | `fix(setup): remove template git origin on clone` |

## Imports

Group imports in three blocks separated by blank lines:

1. Standard library
2. Third-party modules
3. This module (`github.com/eminbekov/fiber-v3-template/...`)

```go
import (
    "context"
    "fmt"

    "github.com/gofiber/fiber/v3"

    "github.com/eminbekov/fiber-v3-template/internal/domain"
)
```

## Naming

- Files: `snake_case.go`
- Packages: lowercase, no underscores (`package domain` not `package user_domain`)
- Exported identifiers: `CamelCase`; unexported: `camelCase`
- Constants: `CamelCase`/`camelCase` (`MaxRetries` not `MAX_RETRIES`)
- Acronyms: consistent all-caps (`UserID`, `ParseURL` not `UserId`, `ParseUrl`)
- Use **full words**, not cryptic abbreviations (`permission` not `perm`, `subscription` not `sub`).
- Fiber handlers: name the receiver parameter `ctx` (type `fiber.Ctx`), never `c`.
- Standard `context.Context` parameters: always named `ctx`.
- Prefer `any` over `interface{}` when an empty interface is required.
- Avoid Java-style `Get` prefixes on getters; use Go idioms (`Name()` not `GetName()`).
- Avoid package stuttering (`user.Service` not `user.UserService`).

## Errors

- Error is always the last return value.
- Wrap errors with context using `fmt.Errorf("operationName: %w", err)` so callers can use `errors.Is` / `errors.As`.
- Use domain sentinel errors for known business conditions (`domain.ErrNotFound`, `domain.ErrUnauthorized`, `domain.ErrConflict`, `domain.ErrValidation`, `domain.ErrForbidden`).
- Prefer guard clauses and early returns; avoid deep nesting and `else` after `return`.

```go
// Preferred: guard clause style
if err != nil {
    return fmt.Errorf("operationName: %w", err)
}
```

## Architecture and style

- No global mutable state; inject dependencies via constructors.
- Do not use `fiber.Map` or `map[string]any` for known response shapes; use typed structs.
- `context.Context` must be the first parameter on functions that perform I/O or respect cancellation.
- Keep handlers thin: parse request, call service, write response (ideally under 30 lines per method).
- Max function length: ~60 lines (linter at 80).
- Max nesting depth: 4 levels.

## Project rules (hard limits)

| Rule | Limit | Rationale |
|---|---|---|
| Max function length | ~60 lines (linter at 80) | Easier to test and reason about |
| Max nesting depth | 4 levels | Use guard clauses to stay flat |
| Error returns | No ignoring with `_` | Every error must be handled or logged |
| Global mutable state | Not allowed | Use constructors and explicit wiring |
| Empty interface | Prefer `any`; justify if needed | Clear, modern Go |
| HTTP responses | Typed structs | Avoid `fiber.Map` for fixed shapes |
| Dynamic maps | Only when keys are unknown at compile time | Otherwise use structs |

## Receivers

- Value receivers when the method only reads; pointer receivers when it mutates.
- If any method on a type uses a pointer receiver, prefer pointer receivers for all methods on that type.

## Secrets

- Never read, commit, or paste contents of `.env`. Use `.env.example` for documented variable names only.

## Optional module markers

- Optional feature blocks are wrapped with marker comments: `// [module:<key>:start]` and `// [module:<key>:end]`.
- Marker keys currently used: `nats`, `grpc`, `websocket`, `admin`, `web`, `i18n`, `storage`, `cron`, `console`, `generate`, `k8s`, `monitoring`, `swagger`, `views`.
- The installer script `setup.sh` uses these markers to remove disabled modules safely.
- When adding a new removable feature, mark imports, config fields, wiring blocks, routes, and infra snippets consistently.
- Keep marker blocks statement-scoped; do not wrap partial expressions.

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
