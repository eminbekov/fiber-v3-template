# Testing

Testing strategy and patterns for **fiber-v3-template**.

## Test types

| Type | Speed | What it catches | Database needed? |
|---|---|---|---|
| Unit | Milliseconds | Logic bugs in services and domain | No (mocked) |
| Integration | Seconds | SQL bugs, constraint violations, migration errors | Yes (testcontainers) |
| HTTP handler | Milliseconds | Request parsing, auth, response format | No (mocked services) |
| Fuzz | 30+ seconds | Crashes, panics from unexpected input | No |
| Benchmark | Seconds | Performance regressions between commits | Depends |
| Security | Seconds | Injection, auth bypass, data leakage | Yes |

## Running tests

```bash
# Unit tests (fast, no external deps)
go test ./...

# All tests with race detector
make test

# Verbose output
make test-verbose

# Integration tests (requires Docker for testcontainers)
make test-integration

# Fuzz tests with 30-second budget
make fuzz

# Benchmarks
make bench

# Coverage report (generates coverage.out and coverage.html)
make test-cover

# Security scanners
make security
```

CI runs them in order: lint -> test (with Postgres, Redis, NATS services) -> security (govulncheck) -> swagger verification -> build and push image (on push to main only).

### Test organization commands

```bash
go test ./...                                          # Unit tests only (fast)
go test -count=1 -run Integration ./...                # Integration tests (needs Docker)
go test -fuzz FuzzCreateUserInput -fuzztime 30s ./...  # Fuzz tests (30 second budget)
go test -bench=. -benchmem ./...                       # Benchmark tests
go test -race ./...                                    # Race condition detection
```

## Unit tests

Test business logic in isolation by mocking dependencies. Use table-driven tests for comprehensive coverage.

```go
func TestUserService_FindByID(t *testing.T) {
    tests := []struct {
        name      string
        id        uuid.UUID
        mockSetup func(repository *mocks.UserRepository, cache *mocks.Cache)
        wantUser  *domain.User
        wantError error
    }{
        {
            name: "returns user from cache when cached",
            id:   uuid.Must(uuid.NewV7()),
            mockSetup: func(repository *mocks.UserRepository, cache *mocks.Cache) {
                cache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil)
            },
            wantUser:  &domain.User{Username: "cached_user"},
            wantError: nil,
        },
        {
            name: "returns user from DB on cache miss",
            id:   uuid.Must(uuid.NewV7()),
            mockSetup: func(repository *mocks.UserRepository, cache *mocks.Cache) {
                cache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("miss"))
                repository.On("FindByID", mock.Anything, mock.Anything).Return(&domain.User{Username: "db_user"}, nil)
                cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
            },
            wantUser:  &domain.User{Username: "db_user"},
            wantError: nil,
        },
        {
            name: "returns ErrNotFound when user does not exist",
            id:   uuid.Must(uuid.NewV7()),
            mockSetup: func(repository *mocks.UserRepository, cache *mocks.Cache) {
                cache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("miss"))
                repository.On("FindByID", mock.Anything, mock.Anything).Return(nil, domain.ErrNotFound)
            },
            wantUser:  nil,
            wantError: domain.ErrNotFound,
        },
        {
            name: "returns ErrValidation for zero UUID",
            id:   uuid.Nil,
            mockSetup: func(repository *mocks.UserRepository, cache *mocks.Cache) {
                // nothing should be called -- validation fails first
            },
            wantUser:  nil,
            wantError: domain.ErrValidation,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepository := new(mocks.UserRepository)
            mockCache := new(mocks.Cache)
            tt.mockSetup(mockRepository, mockCache)

            service := NewUserService(mockRepository, mockCache)
            user, err := service.FindByID(context.Background(), tt.id)

            if tt.wantError != nil {
                require.Error(t, err)
                assert.True(t, errors.Is(err, tt.wantError))
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.wantUser.Username, user.Username)
            }

            mockRepository.AssertExpectations(t)
            mockCache.AssertExpectations(t)
        })
    }
}
```

## Integration tests

Test repository implementations against a real PostgreSQL database running in Docker via testcontainers:

```go
func TestUserRepository_Create_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    ctx := context.Background()

    container, err := postgres.Run(ctx, "postgres:18.3-alpine",
        postgres.WithDatabase("test_db"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        postgres.WithInitScripts("../../migrations/000001_create_users.up.sql"),
    )
    require.NoError(t, err)
    defer container.Terminate(ctx)

    connectionString, _ := container.ConnectionString(ctx)
    pool, _ := pgxpool.New(ctx, connectionString)
    defer pool.Close()

    repository := postgresrepo.NewUserRepository(pool)

    user := &domain.User{
        ID:       uuid.Must(uuid.NewV7()),
        Username: "testuser",
        FullName: "Test User",
        Phone:    "+998901234567",
        Status:   "active",
    }

    err = repository.Create(ctx, user)
    require.NoError(t, err)

    found, err := repository.FindByID(ctx, user.ID)
    require.NoError(t, err)
    assert.Equal(t, user.Username, found.Username)
    assert.Equal(t, user.FullName, found.FullName)
}
```

## Fuzz tests

Go's built-in fuzz testing sends random/mutated input to find crashes:

```go
func FuzzCreateUserInput(f *testing.F) {
    // Seed corpus: valid examples that the fuzzer will mutate
    f.Add(`{"username":"john","phone":"+998901234567","full_name":"John Doe"}`)
    f.Add(`{"username":"","phone":"","full_name":""}`)
    f.Add(`{"username":"a","phone":"123","full_name":"x"}`)

    f.Fuzz(func(t *testing.T, input string) {
        var dto request.CreateUser
        if err := json.Unmarshal([]byte(input), &dto); err != nil {
            return // Invalid JSON is expected, just skip
        }

        // The validator must NEVER panic, regardless of input
        require.NotPanics(t, func() {
            _ = validator.ValidateStruct(&dto)
        })
    })
}
```

Run: `go test -fuzz FuzzCreateUserInput -fuzztime 30s ./internal/handler/api/`

## Security tests

Verify that parameterized queries prevent SQL injection, that role escalation is blocked, and that internal errors are never leaked:

### SQL injection prevention

```go
func TestSQLInjection(t *testing.T) {
    payloads := []string{
        "'; DROP TABLE users; --",
        "' OR '1'='1",
        "admin'--",
        "1; SELECT * FROM information_schema.tables",
        "' UNION SELECT username, password FROM users --",
    }

    for _, payload := range payloads {
        t.Run(payload, func(t *testing.T) {
            // pgx parameterized queries treat this as a literal string, not SQL
            user, err := repository.FindByUsername(ctx, payload)
            assert.Nil(t, user)
            assert.True(t, errors.Is(err, domain.ErrNotFound))
        })
    }
}
```

### Role escalation prevention

```go
func TestRoleEscalation(t *testing.T) {
    // A regular user tries to set their role to superadmin
    body := `{"role": "superadmin"}`
    request := httptest.NewRequest("PUT", "/api/users/me", strings.NewReader(body))
    request.Header.Set("Authorization", "Bearer "+regularUserToken)

    response, _ := app.Test(request)
    assert.Equal(t, 403, response.StatusCode)
}
```

### Error sanitization

```go
func TestErrorSanitization(t *testing.T) {
    // Trigger an internal error and verify no internals are leaked
    response, _ := app.Test(httptest.NewRequest("GET", "/api/users/invalid-uuid", nil))
    body, _ := io.ReadAll(response.Body)

    bodyString := string(body)
    assert.NotContains(t, bodyString, "pq:")          // no Postgres errors
    assert.NotContains(t, bodyString, "/app/")         // no file paths
    assert.NotContains(t, bodyString, "goroutine")     // no stack traces
    assert.NotContains(t, bodyString, "password")      // no sensitive fields
}
```

## Test best practices

| Practice | Why |
|---|---|
| Table-driven tests | Cover many cases in one function; easy to add new scenarios |
| Mock at interface boundaries | Services mock repositories; handlers mock services |
| Use `testcontainers` for integration | Real database catches SQL bugs that mocks miss |
| Run race detector (`-race`) | Catches concurrent access bugs before production |
| Set timeouts with `context.WithTimeout` | Prevents tests from hanging indefinitely |
| Separate unit from integration tests | `make test` stays fast; `make test-integration` for thorough checks |

## NASA P10 safety rules (Go adaptation)

NASA's Power of Ten rules were designed for mission-critical C code. Here is how each rule translates to Go:

| # | Original Rule | Go Adaptation | Enforcement | Why It Matters |
|---|---|---|---|---|
| 1 | Simple control flow | No `goto`, max nesting 4, no recursion | `nestif`, `cyclop` linters | Deep nesting and complex flow are the #1 source of bugs |
| 2 | Fixed upper bounds on loops | `context.WithTimeout` on all external calls | Context propagation | Unbounded operations can hang forever |
| 3 | No dynamic memory after init | `sync.Pool`, pre-allocated slices | `prealloc` linter | Uncontrolled allocation causes GC pauses |
| 4 | Short functions | Max ~60 lines | `funlen: 80` linter | Long functions are hard to understand and test |
| 5 | Minimal assertions | Guard clauses, input validation, error checks | `validator`, `errcheck` | Silent failures cause data corruption |
| 6 | Declare at smallest scope | Variables declared where first used | `varnamelen` linter | Wide scope = more places for bugs |
| 7 | Check all return values | Never use `_` for error returns | `errcheck` linter | Ignored errors hide bugs that surface later |
| 8 | Limit preprocessor use | No build tags / conditional compilation | Policy | Conditional code doubles the paths to test |
| 9 | Restrict pointer use | Value receivers where possible | Code review | Pointers create aliasing bugs |
| 10 | Compile with all warnings | golangci-lint with 20+ linters | CI pipeline gate | Catch bugs before they reach production |

### Recommended `.golangci.yml`

```yaml
linters:
  enable:
    - errcheck       # All errors must be checked
    - govet          # Go vet checks (shadow, printf, etc.)
    - staticcheck    # Advanced static analysis
    - unused         # Find unused code
    - gosimple       # Simplify code
    - ineffassign    # Detect useless assignments
    - typecheck      # Type checking
    - gocritic       # Opinionated code quality
    - cyclop         # Cyclomatic complexity
    - nestif         # Deeply nested ifs
    - funlen         # Function length
    - gocognit       # Cognitive complexity
    - prealloc       # Suggest pre-allocation
    - errorlint      # Error wrapping correctness
    - wrapcheck      # Errors from external packages must be wrapped
    - noctx          # HTTP requests must use context
    - gosec          # Security issues
    - bodyclose      # HTTP response body must be closed

linters-settings:
  funlen:
    lines: 80
  cyclop:
    max-complexity: 15
  nestif:
    min-complexity: 5
```
