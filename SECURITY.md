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

#### Argon2id password hashing

Argon2id is memory-hard, meaning each hash attempt requires 64MB of RAM. This makes GPU-based brute force attacks significantly more expensive than bcrypt. OWASP (2025) recommends Argon2id as the primary choice for password hashing.

**Recommended parameters:**

| Parameter | Value |
|---|---|
| Time (iterations) | 3 |
| Memory | 65536 KB (64 MB) |
| Threads | 4 |
| Key length | 32 bytes |
| Salt length | 16 bytes (random, unique per password) |

**Migrating from bcrypt:** Do not force all users to reset passwords. Instead, migrate transparently on login -- check if the hash starts with `$2a$` (bcrypt), verify with bcrypt, then re-hash with Argon2id and save.

### Authorization

- Role-Based Access Control (RBAC) with database-backed roles and permissions.
- Permission checks enforced at the middleware layer before handlers execute.
- No client-side role enforcement; all authorization decisions are server-side.

#### RBAC database schema

Roles and permissions are NOT stored on the User struct. They live in separate tables with many-to-many relationships:

```sql
-- Roles table (internal-only: BIGSERIAL, never exposed in API URLs)
CREATE TABLE roles (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Permissions table (internal-only: BIGSERIAL)
CREATE TABLE permissions (
    id          BIGSERIAL PRIMARY KEY,
    resource    VARCHAR(100) NOT NULL,
    action      VARCHAR(50) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(resource, action)
);

-- Role-Permission mapping (many-to-many)
CREATE TABLE role_permissions (
    role_id       BIGINT REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- User-Role mapping (many-to-many: user_id is UUID because users are external)
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- Indexes for fast lookups
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
```

#### Permission checking in middleware

```go
func RequirePermission(resource string, action string) fiber.Handler {
    return func(ctx fiber.Ctx) error {
        userID := ctx.Locals("user_id").(uuid.UUID)
        hasPermission, err := authorizationService.HasPermission(ctx.Context(), userID, resource, action)
        if err != nil {
            return ctx.Status(fiber.StatusInternalServerError).JSON(response.Error{
                Message: "internal server error",
            })
        }
        if !hasPermission {
            return ctx.Status(fiber.StatusForbidden).JSON(response.Error{
                Message: "you do not have permission to perform this action",
            })
        }
        return ctx.Next()
    }
}

// Usage in router:
api.Get("/users", RequirePermission("users", "read"), userHandler.List)
api.Post("/users", RequirePermission("users", "create"), userHandler.Create)
```

#### Efficient permission lookup

Permissions are checked on every request -- use a single SQL query that JOINs through the tables, and cache the result in Redis (`permissions:user:{user_id}`, 10 min TTL). Invalidate on role change.

#### When to add ABAC

RBAC is sufficient for most applications. Consider ABAC rules when you need:
- **Ownership rules**: "Users can only edit their own profile"
- **Time-based rules**: "Reports can only be exported during business hours"
- **Resource-attribute rules**: "Managers can only see users in their own department"

### Input validation

- All request payloads are validated using `go-playground/validator/v10` with struct tags.
- Request body size is limited via Fiber's `BodyLimit` middleware (default 4 MB).
- SQL injection is structurally prevented through parameterized queries (`pgx` prepared statements).

```go
type CreateUser struct {
    Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Phone    string `json:"phone"    validate:"required,e164"`
    FullName string `json:"full_name" validate:"required,min=2,max=100"`
    Role     string `json:"role"     validate:"required,oneof=manager user"`
}
```

```go
// BAD: string concatenation -- SQL injection is trivial
query := "SELECT * FROM users WHERE username = '" + username + "'"

// GOOD: parameterized query -- database treats $1 as data, never as SQL code
query := "SELECT * FROM users WHERE username = $1"
row := pool.QueryRow(ctx, query, username)
```

### Fiber security middleware

| Middleware | What It Does | Why |
|---|---|---|
| **Helmet** | Adds security headers (HSTS, CSP, X-Frame-Options, X-Content-Type-Options) | Prevents clickjacking, MIME sniffing, forces HTTPS |
| **Rate Limiter** | Limits requests per time window | Prevents brute force, DDoS, and API abuse |
| **CSRF** | Validates a random token on every POST/PUT/DELETE | Prevents cross-site request forgery |
| **EncryptCookie** | AES-256-GCM encrypts cookie values | Prevents cookie tampering and information leakage |
| **CORS** | Controls which origins can make cross-origin requests | Deny-all by default; only open for trusted domains |
| **Body Limit** | Rejects request bodies larger than 4MB | Prevents memory exhaustion attacks |

#### Rate limiter layered strategy

```
Authenticated routes:   limit by User ID  -> each user gets own bucket
                        (fair on shared networks like offices, universities)

Unauthenticated routes: limit by IP       -> fallback when identity is unknown

Login endpoints:        limit by Username AND by IP
                        -> prevents brute force per account (5/min per username)
                        -> prevents credential stuffing per source (20/min per IP)
```

Why layered? If you only limit by IP, all users in the same office/university share one bucket. One active user blocks everyone else.

### Error sanitization

**Never expose internal details to clients.**

```go
// BAD: leaks database schema
return ctx.Status(500).JSON(fiber.Map{
    "error": "pq: duplicate key value violates unique constraint \"users_username_key\"",
})

// BAD: leaks file system paths
return ctx.Status(500).JSON(fiber.Map{
    "error": "open /app/uploads/photos/secret.jpg: no such file or directory",
})

// BAD: leaks raw error (could contain anything)
return ctx.Status(500).JSON(fiber.Map{"error": err.Error()})

// GOOD: generic error for the client, full details in server logs
slog.ErrorContext(ctx.Context(), "failed to create user",
    "error", err,
    "username", request.Username,
    "request_id", ctx.Get("X-Request-ID"),
)
return ctx.Status(500).JSON(response.Error{Message: "internal server error"})
```

### Uploaded file protection (signed URLs)

Do not serve uploaded files with `app.Static("/uploads", "./uploads")`. Use signed URLs with HMAC-SHA256 and expiry:

1. API response contains a signed URL instead of a raw file path
2. Middleware validates the HMAC token and expiry on each request
3. Token is valid and not expired -> serve the file; otherwise -> 403 Forbidden

```go
func SignURL(filename string, signingKey []byte, ttl time.Duration) string {
    expires := time.Now().Add(ttl).Unix()
    message := filename + strconv.FormatInt(expires, 10)
    mac := hmac.New(sha256.New, signingKey)
    mac.Write([]byte(message))
    token := hex.EncodeToString(mac.Sum(nil))
    return fmt.Sprintf("/api/files/%s?token=%s&expires=%d", filename, token, expires)
}

func ValidateSignedURL(filename string, token string, expires int64, signingKey []byte) bool {
    if time.Now().Unix() > expires {
        return false
    }
    message := filename + strconv.FormatInt(expires, 10)
    mac := hmac.New(sha256.New, signingKey)
    mac.Write([]byte(message))
    expectedToken := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(token), []byte(expectedToken))
}
```

### Transport

- Helmet middleware sets security headers: HSTS, X-Content-Type-Options, X-Frame-Options, CSP.
- CORS is configurable via `CORS_ALLOW_ORIGINS` with explicit allowlist (no wildcard in production).
- Kubernetes deployments include an EnvoyFilter WAF that blocks path traversal, SQL injection patterns, XSS, command injection, scanner bots, and SSRF attempts.

### Envoy WAF (Kubernetes deployments)

For Kubernetes deployments with Istio, an EnvoyFilter with Lua scripts acts as a Web Application Firewall (WAF).

**What the WAF blocks:**

| Attack Type | What It Is | How the WAF Blocks It |
|---|---|---|
| Path traversal | `../../etc/passwd` | Blocks `..` patterns in URLs |
| SQL injection | `' OR 1=1 --` | Pattern matching on URL and query params |
| XSS | `<script>alert(1)</script>` | Blocks script tags and event handlers |
| Command injection | `; rm -rf /` | Blocks shell command patterns |
| Scanner bots | nikto, sqlmap, nmap | Blocks known scanner user-agents |
| DDoS tools | LOIC, Slowloris | Blocks known DDoS tool user-agents |
| Sensitive files | `.git/config`, `.env` | Blocks access to config/secret files |
| SSRF | `http://169.254.169.254` | Blocks internal IP addresses |
| Oversized requests | 100MB POST body | 10MB limit at proxy level |

### Secrets management

- `.env` files are gitignored; `.env.example` documents variable names without values.
- Kubernetes secrets use the Secret resource (base64-encoded); for production, use a secrets manager (Vault, AWS Secrets Manager, etc.).
- The `setup.sh` installer generates `.env` interactively and never commits secret values.

### Dependencies

- `govulncheck` runs in CI to detect known vulnerabilities in dependencies.
- `gosec` performs static analysis for common security issues in Go code.
- Dependencies are pinned to specific versions in `go.mod`.

### Defense in depth

Security is not one layer. It is multiple layers, each catching what the previous one missed:

```
Internet
  |
  v
[Reverse Proxy / Envoy WAF]     Blocks known attack patterns (SQLi, XSS, bots, path traversal)
  |
  v
[Fiber Middleware]               Rate limiting, authentication, CSRF, Helmet headers, body limit
  |
  v
[Handler Layer]                  Request parsing, DTO validation (go-playground/validator)
  |
  v
[Service Layer]                  Business rule validation, role checks, authorization
  |
  v
[Repository Layer]               Parameterized queries (SQL injection structurally impossible)
  |
  v
[Database]                       Constraints, foreign keys, row-level security (future)
```

If an attacker bypasses the WAF (maybe it is a new attack pattern), the middleware catches them. If they bypass middleware (maybe they have valid credentials), the service layer checks their role. Every layer is an independent security check.

## Disclosure policy

We follow coordinated disclosure. Once a fix is released, we will:

1. Publish a security advisory on the repository
2. Credit the reporter (unless they prefer anonymity)
3. Tag a new release with the fix
