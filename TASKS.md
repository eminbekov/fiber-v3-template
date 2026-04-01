# fiber-v3-template — phased task list

Work through phases in order unless noted. Each checkbox is a concrete deliverable. Cross-references point to sections in `GO_FIBER_PROJECT_GUIDE.md` (local file; not committed).

**Codegen (Gii-style):** a dedicated phase builds a `cmd/gen` (or `console generate …`) tool that scaffolds files from minimal input—mirroring Yii2 Gii workflows for CRUD, migrations, and boilerplate.

---

## Phase 0 — Baseline and conventions

- Confirm module path, Go version, and `README` basics stay aligned with the repo.
- Add `AGENTS.md` / `CONVENTIONS.md` only when you want contributor-facing rules (guide §5, §22).
- Enforce import grouping, error wrapping, and naming per guide §3–§4.
- Optional: strict `golangci-lint` config as in guide §21 / examples in §25.

---

## Phase 1 — Application shell (entrypoint, config, logging)

- Typed configuration from environment with validation (guide §2 `internal/config`, §5).
- Structured logging with `slog` (JSON in prod, text in dev) via `package/logger` (guide §19.1, §25).
- `run(ctx)` pattern: load config → logger → defer cleanups (guide §27).
- Wire HTTP listen address from config/env (already partially in `cmd/server`); unify with future config package.
- Document required env vars in `README` (minimal table; expand later).

---

## Phase 2 — HTTP routing, middleware, and API shape

- Central router registration in `internal/router` (guide §2); keep handlers thin.
- Middleware stack: recovery, request ID, structured request logging (guide §2, §19).
- Security-related middleware where applicable: CORS, Helmet-analog, body limits (guide §11.3).
- Standard JSON success/error envelope and stable error codes (guide §6.5, §11.4).
- API versioning: `/api/v1/...` groups (guide §6.6).
- Pagination and rate-limit response headers when endpoints need them (guide §6.3–§6.4).
- Replace ad-hoc `fiber.Map` responses with typed DTO structs for hot paths (guide §25 “Avoid”).

---

## Phase 3 — Health, observability hooks

- Reusable health package: liveness/readiness (guide §19.4, `package/health` in §2).
- Readiness checks optional dependencies (DB, Redis) once those layers exist.
- OpenTelemetry: tracer/meter setup and graceful shutdown (guide §19, §25 `otel`).
- Prometheus HTTP metrics middleware (guide §19.3, §25).
- Optional: DB query logging helpers (guide §19.2).

---

## Phase 4 — Database layer and migrations

- PostgreSQL with `pgx/v5` and connection pool tuning (guide §13, §25).
- Multi-pool or named registry if you need more than one DB (guide §13.1).
- `cmd/migrate` using `golang-migrate` (up/down, version) (guide §2, §25, §28).
- `migrations/` SQL files with sequential naming; conventions from §10.3.
- Repository interfaces in `internal/repository`; implementations under `internal/repository/postgres` (guide §1–§2, §13.2).
- Domain models in `internal/domain` (no framework imports) and sentinel errors (guide §1).

---

## Phase 5 — Vertical slice (prove DDD wiring)

- One full flow: domain → repository → service → handler → request/response DTOs (guide §1–§2, §6).
- Validation on input DTOs (`validator/v10` per guide §11.5, §25).
- Integration test pattern for repository or HTTP (guide §23.2).

---

## Phase 6 — Cache and sessions (when needed)

- `Cache` interface + Redis implementation (guide §14, §25).
- Type-safe cache keys (guide §14.2).
- Cache-aside in services; invalidation strategy (guide §14.4–§14.5).
- If using admin HTML sessions: Redis session store (guide §15.3).

---

## Phase 7 — Auth, passwords, RBAC

- Password hashing with Argon2id (guide §11.2, §25).
- Authentication mechanism (Bearer JWT or session) per product choice (guide §6, §12).
- RBAC schema, domain models, permission checks in middleware (guide §12).
- Seed data for roles/permissions (guide §12.6).
- Optional: ABAC extension points (guide §12.5).

---

## Phase 8 — API documentation (Swagger)

- `swag` annotations on `main` and handlers (guide §7).
- Generated `docs/` package; route `/swagger/`* via `gofiber/swagger` (guide §7.6, §25).
- Makefile `swagger` target; CI check that spec generation succeeds (guide §7.7, §7.10).
- Disable or protect Swagger in production (guide §7.11).

---

## Phase 9 — HTML templates and i18n (optional product features)

- Fiber HTML engine setup, layouts, embed (guide §8).
- Admin handler area under `internal/handler/admin` if needed (guide §2).
- i18n loader, middleware for `Accept-Language`, API and template usage (guide §9).

---

## Phase 10 — Async and inter-service (optional, scale-out)

- [x] NATS connection and JetStream patterns (guide §16, §25).
- [x] Background consumers under `internal/nats/consumers` (guide §2).
- [x] `proto/` definitions and generated `gen/` code; gRPC server alongside Fiber (guide §17, §25).
- [x] gRPC interceptors (recovery, OTEL) (guide §17.4–§17.5).

---

## Phase 11 — Storage, CDN, multi-instance concerns

- [x] File storage abstraction (S3/MinIO/local) (guide §2 `internal/storage`, §11.1).
- [x] Signed URLs for uploads/downloads where required (guide §11.1, §15.1).
- [x] WebSocket or broadcast notes if horizontal scale (guide §15.4).

---

## Phase 12 — Cron and scheduled jobs

- In-process scheduler *or* separate `cmd/cron` binary (guide §30).
- Register jobs with shared services via DI; graceful stop on signal (guide §30).

---

## Phase 13 — Docker, deploy, CI/CD

- Multi-stage `Dockerfile` and `.dockerignore` (guide §20.1).
- `docker-compose` for app + Postgres + Redis + NATS as needed (guide §20.2).
- GitHub Actions (or chosen CI): test, lint, `swagger` gen, build image (guide §21).
- Image tagging and manual deploy gates for production (guide §21.2–§21.3).

---

## Phase 14 — Testing and hardening

- Unit tests for services and pure domain (guide §23.1).
- Integration tests with real Postgres/Redis in CI (guide §23.2).
- Fuzz tests for parsers/validators where valuable (guide §23.3).
- `govulncheck` / `gosec` in CI (guide §28 `security` target, §25).
- Review NASA P10 rules for critical paths (guide §24).

---

## Phase 15 — Makefile and developer ergonomics

- `Makefile`: build, run, fmt, tidy, test, lint, migrate up/down/create, swagger, proto, docker, help (guide §28).
- `make help` as default goal with discoverable targets.

---

## Phase 16 — Console CLI (`cmd/console`)

- Entry point routing subcommands with `context.Context` (guide §29).
- Shared wiring: config, DB pool, Redis—reuse services, not duplicate logic (guide §29.5).
- Example commands: `create-admin`, `assign-role`, `cache-clear`, `export-users` (guide §29.3–§29.4).
- Document commands in `README` and mirror important flows in `Makefile` (guide §29.4).

---

## Phase 17 — Code generator (Yii2 Gii–style)

- **CLI design:** e.g. `go run ./cmd/gen …` or `./bin/console generate …` with subcommands.
- **Migration:** `generate migration <name>` — create timestamped up/down SQL stubs under `migrations/` (wraps `migrate create` or equivalent).
- **CRUD / module:** `generate resource <name>` — optional flags for `--with-repo`, `--with-handler`, API version.
  - Emit: `internal/domain/<entity>.go`, repository interface, `postgres` stub, service stub, handler stubs, DTOs under `dto/request` and `dto/response/v1`, router registration snippet or file.
- **Templates:** use `text/template` or `embed` for codegen templates; keep templates versioned in repo.
- **Idempotence / safety:** dry-run flag; never overwrite without `--force`.
- **Tests:** generator golden-file tests for emitted code shape.
- **Docs:** short “Codegen” section in `README` listing commands and examples.

---

## Phase 18 — Kubernetes / advanced edge (when you deploy to K8s)

- Manifests under `deploy/k8s` (guide §2).
- Optional Envoy WAF / filters if using service mesh (guide §26).

---

## Phase 19 — Microservices split (only if needed)

- Document service boundaries and data ownership (guide §18).
- Extract services only after monolith proves the domain (guide §18.1).

---

### How to use this file

1. Pick the next unchecked item in the earliest incomplete phase.
2. Implement in a branch, open a PR, merge.
3. Adjust tasks if the product does not need optional phases (9–11, 16–19).

---

### Reference index (guide sections → topics)


| Sections | Topics                                     |
| -------- | ------------------------------------------ |
| §1–§2    | DDD layers, directory layout               |
| §3–§5    | Naming, code style, project rules          |
| §6       | REST API, envelopes, versioning            |
| §7       | Swagger                                    |
| §8–§9    | Templates, i18n                            |
| §10–§11  | Performance, security                      |
| §12      | RBAC/ABAC                                  |
| §13–§14  | DB, Redis cache                            |
| §15–§18  | CDN, NATS, gRPC, microservices             |
| §19–§23  | Observability, Docker, CI/CD, Git, testing |
| §24–§26  | NASA rules, dependencies, Envoy WAF        |
| §27–§30  | Entrypoint, Makefile, console CLI, cron    |


