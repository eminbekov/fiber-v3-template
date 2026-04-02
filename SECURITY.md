# Security policy

## Reporting vulnerabilities

If you discover a security vulnerability, **do not open a public issue**. Instead, email **security@example.com** with:

1. Description of the vulnerability
2. Steps to reproduce
3. Potential impact
4. Suggested fix (if any)

We acknowledge reports within 48 hours and aim to provide a fix or mitigation within 7 days for critical issues.

## Supported versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |
| < latest | Best effort |

## Security practices

### Authentication

- Passwords are hashed with **Argon2id** (OWASP-recommended parameters).
- API authentication uses bearer tokens stored in Redis with configurable expiration.
- Admin panel uses HttpOnly, SameSite=Lax, Secure (in production) session cookies.
- Session tokens are cryptographically random (32 bytes, hex-encoded).

### Authorization

- Role-Based Access Control (RBAC) with database-backed roles and permissions.
- Permission checks enforced at the middleware layer before handlers execute.
- No client-side role enforcement; all authorization decisions are server-side.

### Input validation

- All request payloads are validated using `go-playground/validator/v10` with struct tags.
- Request body size is limited via Fiber's `BodyLimit` middleware (default 4 MB).
- SQL injection is structurally prevented through parameterized queries (`pgx` prepared statements).

### Transport

- Helmet middleware sets security headers: HSTS, X-Content-Type-Options, X-Frame-Options, CSP.
- CORS is configurable via `CORS_ALLOW_ORIGINS` with explicit allowlist (no wildcard in production).
- Kubernetes deployments include an EnvoyFilter WAF that blocks path traversal, SQL injection patterns, XSS, command injection, scanner bots, and SSRF attempts.

### Secrets management

- `.env` files are gitignored; `.env.example` documents variable names without values.
- Kubernetes secrets use the Secret resource (base64-encoded); for production, use a secrets manager (Vault, AWS Secrets Manager, etc.).
- The `setup.sh` installer generates `.env` interactively and never commits secret values.

### Dependencies

- `govulncheck` runs in CI to detect known vulnerabilities in dependencies.
- `gosec` performs static analysis for common security issues in Go code.
- Dependencies are pinned to specific versions in `go.mod`.

### Defense in depth

```
Internet
  |
  v
[Envoy WAF]           Network-level attack blocking (K8s deployments)
  |
  v
[Fiber Middleware]     Rate limiting, authentication, CSRF, security headers
  |
  v
[Service Layer]        Business rule enforcement, authorization checks
  |
  v
[Repository Layer]     Parameterized queries (SQL injection structurally impossible)
```

## Disclosure policy

We follow coordinated disclosure. Once a fix is released, we will:

1. Publish a security advisory on the repository
2. Credit the reporter (unless they prefer anonymity)
3. Tag a new release with the fix
