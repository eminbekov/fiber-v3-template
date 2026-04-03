# Contributing to fiber-v3-template

Thank you for investing time in this project. This document explains **how we work in Git**, how to set up a **local development environment**, what we expect in **pull requests**, and where to read deeper rules (**architecture**, **conventions**, **testing**, **security**).

If anything here conflicts with repository-specific automation (for example CI), prefer the **checked-in scripts and workflows** and open an issue or pull request to align the docs.

---

## Table of contents

1. [Code of Conduct](#code-of-conduct)
2. [What “contributing” means here](#what-contributing-means-here)
3. [Security disclosures](#security-disclosures)
4. [Development environment](#development-environment)
5. [Git workflow (branches, commits, pull requests)](#git-workflow-branches-commits-pull-requests)
6. [Pull request checklist](#pull-request-checklist)
7. [Project structure and coding standards](#project-structure-and-coding-standards)
8. [Testing and verification](#testing-and-verification)
9. [Optional modules and installer markers](#optional-modules-and-installer-markers)
10. [Documentation you may need](#documentation-you-may-need)
11. [Licensing](#licensing)

---

## Code of Conduct

All contributors are expected to follow our **[Code of Conduct](./CODE_OF_CONDUCT.md)**. It applies to issues, pull requests, discussions, and any other project space. If you see behavior that violates it, use the reporting paths described there.

---

## What “contributing” means here

Reasonable contributions include:

- **Bug fixes** with tests or clear reproduction notes when tests are not feasible.
- **Features** that fit the template’s scope (production-oriented Go API with Fiber v3, optional modules, clear wiring).
- **Documentation** improvements (README, guides, comments where they remove ambiguity—avoid noisy comments).
- **Tooling** improvements (Makefile, CI, Docker) that help developers without breaking downstream clones.
- **Tests** that increase confidence without flaking in CI.

If you plan a **large change**, open an issue or discussion first so maintainers can agree on direction and avoid rework.

---

## Security disclosures

**Do not** open a public issue for security vulnerabilities. Follow **[SECURITY.md](./SECURITY.md)** instead.

---

## Development environment

### Prerequisites

- **Go:** version compatible with `go` directive in `go.mod` (see README for the current minimum).
- **Git:** for version control.
- **PostgreSQL and Redis:** required for a running application; integration tests may use containers (see `TESTING.md`).
- **Optional:** Docker and Docker Compose for container workflows; `make` for standard tasks.

### First-time setup (typical)

1. **Clone** the repository (or your fork).

   ```bash
   git clone https://github.com/eminbekov/fiber-v3-template.git
   cd fiber-v3-template
   ```

2. **Run the installer** (recommended for new projects derived from the template):

   ```bash
   ./setup.sh
   ```

   The script can rewrite the module path, trim optional modules, and help produce a valid `.env`. It is the primary entry point documented in the README.

3. **Configure environment:** copy from `.env.example` to `.env` if you are not using the interactive installer. **Never commit secrets.** Do not paste real credentials into issues or pull requests.

4. **Build** to confirm the tree compiles:

   ```bash
   go build ./...
   ```

5. Before opening a pull request, run the full verification sequence (see [Testing and verification](#testing-and-verification)):

   ```bash
   make verify
   ```

---

## Git workflow (branches, commits, pull requests)

This project follows **GitHub Flow**: `main` is expected to stay **deployable**; work happens on **short-lived branches** merged via **pull requests**.

The rules below align with **[AGENTS.md](./AGENTS.md)** and are expanded here so humans and automation share the same vocabulary.

### Rules that always apply

1. **Do not commit directly to `main`.** Every change should land through a **pull request**, except for maintainers performing exceptional operational tasks—and even then, prefer a PR for auditability.
2. **Branch from an up-to-date `main`:**

   ```bash
   git checkout main
   git pull origin main
   git checkout -b <prefix>/<short-description>
   ```

3. **Push your branch to the remote** (`origin`) and open a **pull request** early if the change is non-trivial. This matches a **remote-first** collaboration style: the branch exists on the server, CI can run, and others can see work in progress.

   ```bash
   git push -u origin <prefix>/<short-description>
   ```

4. **Keep pull requests reviewable.** Aim for roughly **under 400 lines** of diff where practical; split large work into stacked or sequential PRs with clear dependencies.

5. **Rebase when necessary** to incorporate upstream changes (coordinate if others depend on your branch):

   ```bash
   git checkout main && git pull origin main
   git checkout <your-branch>
   git rebase main
   git push --force-with-lease
   ```

6. **Delete the remote branch after merge** when GitHub offers it, or use the CLI; keep the repository tidy.

### Branch naming

Use **lowercase**, words separated by **hyphens**, and a **prefix** that reflects intent:

| Prefix | When to use | Example |
|--------|-------------|---------|
| `feature/` or `feat/` | New functionality | `feature/user-export` |
| `fix/` or `bugfix/` | Bug fixes | `fix/login-timeout` |
| `hotfix/` | Urgent production fixes | `hotfix/payment-crash` |
| `refactor/` | Restructuring without behavior change | `refactor/user-service` |
| `chore/` | Maintenance: CI, dependencies, bot updates, GitHub templates | `chore/update-dependencies` |
| `docs/` | Documentation only | `docs/api-guide` |
| `test/` | Adding or fixing tests | `test/user-service-unit` |
| `release/` | Release preparation | `release/v1.3.0` |

**Additional rules:**

- Prefer **short but descriptive** names: `feature/rbac` rather than `feature/implement-full-role-based-access-control`.
- If your organization uses tickets, include the id: `fix/CM-234-login-timeout`.
- **Map your change type to a prefix** so reviewers immediately know the risk profile:
  - Documentation-only updates to guides or README → `docs/...`
  - GitHub issue templates, PR templates, Dependabot config → `chore/...` (or `docs/...` if you only touch explanatory markdown under `.github` without workflow logic—either is acceptable if you stay consistent within the PR)
  - New tests only → `test/...`
  - User-visible API or behavior change → `feature/...` or `fix/...`

### Commit messages (Conventional Commits)

Use **[Conventional Commits](https://www.conventionalcommits.org/)**:

```text
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Common types:**

| Type | When | Typical semver impact |
|------|------|------------------------|
| `feat` | New feature | Minor |
| `fix` | Bug fix | Patch |
| `docs` | Documentation only | None |
| `style` | Formatting, no logic change | None |
| `refactor` | Refactor without behavior change | None |
| `perf` | Performance improvement | Patch |
| `test` | Tests | None |
| `build` | Build system or dependencies | None |
| `ci` | CI configuration | None |
| `chore` | Other maintenance | None |

**Breaking changes** must be indicated with `!` after the type, for example: `feat(api)!: change pagination format`.

**Good examples:**

```bash
git commit -m "feat(auth): add Argon2id password hashing"
git commit -m "fix(api): return 404 instead of 500 for missing user"
git commit -m "docs(readme): clarify setup.sh module selection"
git commit -m "chore(ci): cache Go modules in workflow"
```

**Poor examples (avoid):**

```text
fix stuff
update
WIP
changes
```

Each commit should be **understandable on its own**; squash locally before pushing if your branch contains experimental commits.

### Opening and merging pull requests

1. Push your branch to `origin`.
2. Open a PR against `main`.
3. Fill in the **description**: problem, approach, risk, and how you tested (see [Pull request checklist](#pull-request-checklist)).
4. Wait for **CI** to pass.
5. Address review feedback with additional commits or a rebase, as maintainers prefer.
6. Merge strategy: maintainers may use **squash merge** for feature branches to keep `main` linear and readable; large refactors may merge with a merge commit when history preservation matters.

---

## Pull request checklist

Before you request review:

- [ ] **Scope:** One cohesive concern per PR when possible.
- [ ] **Tests:** Added or updated where behavior changed; existing tests pass.
- [ ] **Verification:** `make verify` passes locally (format, tidy, build, vet, lint, test).
- [ ] **Documentation:** README or other docs updated if user-facing behavior or setup changed.
- [ ] **Security:** No secrets committed; security-sensitive issues use SECURITY.md, not the public PR description.
- [ ] **Style:** Matches `CONVENTIONS.md` and project layout in `ARCHITECTURE.md`.

---

## Project structure and coding standards

High-level layout is described in the **README** and in depth in **[ARCHITECTURE.md](./ARCHITECTURE.md)**.

**Layering:** keep **handlers** thin, put orchestration in **services**, put SQL in **repositories**, and keep **domain** models free of framework imports.

**Naming, imports, errors, and limits** are defined in **[CONVENTIONS.md](./CONVENTIONS.md)**. Automated linting may enforce parts of these rules.

---

## Testing and verification

- **Day-to-day:** `make test` or targeted `go test ./...` paths you touch.
- **Before push / PR:** `make verify` runs formatting, `go mod tidy`, `go build ./...`, vet, lint, and tests—this is the **canonical pre-push gate** referenced in `AGENTS.md`.

See **[TESTING.md](./TESTING.md)** for philosophy (unit vs integration), database usage, and troubleshooting.

---

## Optional modules and installer markers

Optional features are guarded by **marker comments** such as `// [module:nats:start]` … `// [module:nats:end]`. The **`setup.sh`** installer removes disabled modules. If you change optional wiring, keep markers consistent and update **README** / **`.env.example`** when new variables appear.

---

## Documentation you may need

| Document | Purpose |
|----------|---------|
| [README.md](./README.md) | Quick start, layout, configuration tables |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Layers, dependencies, routing |
| [CONVENTIONS.md](./CONVENTIONS.md) | Naming, errors, style limits |
| [TESTING.md](./TESTING.md) | How and what to test |
| [SECURITY.md](./SECURITY.md) | Reporting vulnerabilities |
| [AGENTS.md](./AGENTS.md) | Guidance for automation and contributors (Git, imports, markers) |

---

## Licensing

By contributing, you agree that your contributions are licensed under the same terms as the project. See **[LICENSE](./LICENSE)**.

---

## Questions

Use [GitHub Discussions](https://github.com/eminbekov/fiber-v3-template/discussions) for general questions, or open an issue when you have a concrete bug report or feature proposal. Choose the appropriate form under [`.github/ISSUE_TEMPLATE/`](./.github/ISSUE_TEMPLATE/) so maintainers get structured fields.

Thank you again for helping improve this template.
