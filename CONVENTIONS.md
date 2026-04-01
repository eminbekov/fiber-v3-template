# Coding conventions

Human-readable rules for **fiber-v3-template**. Automated checks may enforce subsets of these via `golangci-lint` and code review.

## Project rules (hard limits)

| Rule | Limit | Rationale |
|------|--------|-----------|
| Max function length | ~60 lines (linter often ~80) | Easier to test and reason about |
| Max nesting depth | 4 levels | Use guard clauses to stay flat |
| Error returns | No ignoring with `_` | Every error must be handled or logged |
| Global mutable state | Not allowed | Use constructors and explicit wiring |
| Empty interface | Prefer `any`; justify if needed | Clear, modern Go |
| HTTP responses | Typed structs | Avoid `fiber.Map` for fixed shapes |
| Dynamic maps | Only when keys are unknown at compile time | Otherwise use structs |

## Imports

1. Standard library  
2. Third-party  
3. This repository  

Separate groups with a blank line.

## Errors

- Last return value is always `error`.
- Wrap with `%w` when propagating: `fmt.Errorf("userRepository.FindByID: %w", err)`.
- Use sentinel errors in `internal/domain` for business outcomes (`ErrNotFound`, etc.).
- Return early on errors; avoid `else` after `return`.

## Functions

- `context.Context` is the first parameter when the work can be cancelled or timed out.
- Value receivers when the method only reads; pointer receivers when it mutates. If any method uses a pointer receiver, prefer pointer receivers for all methods on that type.

## Documentation

- Doc comments should explain behavior from the caller’s perspective and document error conditions (for example when `domain.ErrNotFound` is returned).

## Concurrency

- Do not start goroutines without cancellation (`context`) and panic recovery where appropriate.
- Set timeouts on external calls via `context.WithTimeout` when appropriate.

## Git

- Branch from `main`, open a pull request, keep changes reviewable.
- Conventional Commits for all messages.
