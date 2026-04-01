package health

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
)

// Checker reports readiness state for one dependency.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// Response is the health response payload.
type Response struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details,omitempty"`
}

// Handler serves liveness and readiness endpoints.
type Handler struct {
	checkers []Checker
}

// NewHandler creates a health handler with optional readiness checkers.
func NewHandler(checkers ...Checker) *Handler {
	return &Handler{
		checkers: checkers,
	}
}

// Liveness returns process liveness.
func (healthHandler *Handler) Liveness(ctx fiber.Ctx) error {
	return ctx.JSON(Response{
		Status: "ok",
	})
}

// Readiness returns dependency readiness.
func (healthHandler *Handler) Readiness(ctx fiber.Ctx) error {
	for _, checker := range healthHandler.checkers {
		if checkerError := checker.Check(ctx.Context()); checkerError != nil {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(Response{
				Status: "unhealthy",
				Details: map[string]string{
					checker.Name(): checkerError.Error(),
				},
			})
		}
	}

	return ctx.JSON(Response{
		Status: "ok",
	})
}

// StaticErrorChecker keeps readiness extensible for future phases.
type StaticErrorChecker struct {
	name  string
	error error
}

// NewStaticErrorChecker returns a checker with a preconfigured error.
func NewStaticErrorChecker(name string, checkerError error) Checker {
	return &StaticErrorChecker{
		name:  name,
		error: checkerError,
	}
}

// Name returns checker name.
func (checker *StaticErrorChecker) Name() string {
	return checker.name
}

// Check returns the configured checker error.
func (checker *StaticErrorChecker) Check(context.Context) error {
	return checker.error
}

// ErrNotReady is a helper sentinel for readiness checks.
var ErrNotReady = errors.New("not ready")

// DatabaseChecker runs a ping function against a database dependency.
type DatabaseChecker struct {
	name         string
	pingFunction func(context.Context) error
}

// NewDatabaseChecker creates a readiness checker backed by a ping function.
func NewDatabaseChecker(name string, pingFunction func(context.Context) error) Checker {
	return &DatabaseChecker{
		name:         name,
		pingFunction: pingFunction,
	}
}

// Name returns checker name.
func (checker *DatabaseChecker) Name() string {
	return checker.name
}

// Check calls the configured ping function.
func (checker *DatabaseChecker) Check(ctx context.Context) error {
	if checker.pingFunction == nil {
		return ErrNotReady
	}

	return checker.pingFunction(ctx)
}

// RedisChecker runs a ping function against a Redis dependency.
type RedisChecker struct {
	name         string
	pingFunction func(context.Context) error
}

// NewRedisChecker creates a readiness checker backed by a ping function.
func NewRedisChecker(name string, pingFunction func(context.Context) error) Checker {
	return &RedisChecker{
		name:         name,
		pingFunction: pingFunction,
	}
}

// Name returns checker name.
func (checker *RedisChecker) Name() string {
	return checker.name
}

// Check calls the configured ping function.
func (checker *RedisChecker) Check(ctx context.Context) error {
	if checker.pingFunction == nil {
		return ErrNotReady
	}

	return checker.pingFunction(ctx)
}
