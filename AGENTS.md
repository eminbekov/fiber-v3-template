# AI agent instructions

Rules for automated tools and contributors working on **fiber-v3-template**. Aligns with `CONVENTIONS.md` and project linting.

## Git workflow

- Branch from `main` for every change. Never commit directly to `main`.
- Branch names: lowercase, hyphen-separated, with a prefix: `feature/`, `fix/`, `chore/`, `docs/`, `refactor/`, `test/`, `hotfix/` as appropriate.
- Use [Conventional Commits](https://www.conventionalcommits.org/): `<type>(<scope>): <description>`.
  - Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`.
  - Example: `feat(config): add environment validation`.

## Imports

Group imports in three blocks separated by blank lines:

1. Standard library
2. Third-party modules
3. This module (`github.com/eminbekov/fiber-v3-template/...`)

## Naming

- Files: `snake_case.go`
- Exported identifiers: `CamelCase`; unexported: `camelCase`
- Use **full words**, not cryptic abbreviations (for example `permission` not `perm`).
- Fiber handlers: name the receiver parameter `ctx` (type `fiber.Ctx`), never `c`.
- Standard `context.Context` parameters: always named `ctx`.
- Prefer `any` over `interface{}` when an empty interface is required.
- Avoid Java-style `Get` prefixes on getters; use Go idioms (`Name()` not `GetName()`).

## Errors

- Error is always the last return value.
- Wrap errors with context using `fmt.Errorf("operationName: %w", err)` so callers can use `errors.Is` / `errors.As`.
- Use domain sentinel errors for known business conditions (`domain.ErrNotFound`, etc.).
- Prefer guard clauses and early returns; avoid deep nesting and `else` after `return`.

## Architecture and style

- No global mutable state; inject dependencies via constructors.
- Do not use `fiber.Map` or `map[string]any` for known response shapes; use typed structs.
- `context.Context` must be the first parameter on functions that perform I/O or respect cancellation.
- Keep handlers thin: parse request, call service, write response.

## Secrets

- Never read, commit, or paste contents of `.env`. Use `.env.example` for documented variable names only.
