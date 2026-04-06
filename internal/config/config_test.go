package config

import (
	"os"
	"testing"
)

func setMinimalRequiredEnvironment(testingContext *testing.T) {
	testingContext.Helper()
	testingContext.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	testingContext.Setenv("REDIS_URL", "redis://localhost:6379/0")
	testingContext.Setenv("FILE_SIGNING_KEY", "test-signing-key")
}

func TestLoad_DefaultValues(testingContext *testing.T) {
	setMinimalRequiredEnvironment(testingContext)

	loadedConfig, loadError := Load()
	if loadError != nil {
		testingContext.Fatalf("expected no error, got %v", loadError)
	}

	if loadedConfig.Environment != DefaultEnvironment {
		testingContext.Fatalf("expected environment=%q, got %q", DefaultEnvironment, loadedConfig.Environment)
	}
	if loadedConfig.LogLevel != DefaultLogLevel {
		testingContext.Fatalf("expected log level=%q, got %q", DefaultLogLevel, loadedConfig.LogLevel)
	}
	if loadedConfig.HTTPListenAddress != DefaultHTTPListenAddress {
		testingContext.Fatalf("expected listen address=%q, got %q", DefaultHTTPListenAddress, loadedConfig.HTTPListenAddress)
	}
	if loadedConfig.BodyLimit != DefaultBodyLimit {
		testingContext.Fatalf("expected body limit=%d, got %d", DefaultBodyLimit, loadedConfig.BodyLimit)
	}
}

func TestLoad_MissingDatabaseURL(testingContext *testing.T) {
	testingContext.Setenv("REDIS_URL", "redis://localhost:6379/0")
	testingContext.Setenv("FILE_SIGNING_KEY", "key")
	os.Unsetenv("DATABASE_URL") //nolint:errcheck // test helper

	_, loadError := Load()
	if loadError == nil {
		testingContext.Fatalf("expected error for missing DATABASE_URL")
	}
}

func TestLoad_MissingRedisURL(testingContext *testing.T) {
	testingContext.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	testingContext.Setenv("FILE_SIGNING_KEY", "key")
	os.Unsetenv("REDIS_URL") //nolint:errcheck // test helper

	_, loadError := Load()
	if loadError == nil {
		testingContext.Fatalf("expected error for missing REDIS_URL")
	}
}

func TestLoad_InvalidEnvironment(testingContext *testing.T) {
	setMinimalRequiredEnvironment(testingContext)
	testingContext.Setenv("ENVIRONMENT", "staging")

	_, loadError := Load()
	if loadError == nil {
		testingContext.Fatalf("expected error for invalid ENVIRONMENT")
	}
}

func TestLoad_InvalidLogLevel(testingContext *testing.T) {
	setMinimalRequiredEnvironment(testingContext)
	testingContext.Setenv("LOG_LEVEL", "verbose")

	_, loadError := Load()
	if loadError == nil {
		testingContext.Fatalf("expected error for invalid LOG_LEVEL")
	}
}

func TestLoad_InvalidBodyLimit(testingContext *testing.T) {
	setMinimalRequiredEnvironment(testingContext)
	testingContext.Setenv("BODY_LIMIT", "not-a-number")

	_, loadError := Load()
	if loadError == nil {
		testingContext.Fatalf("expected error for invalid BODY_LIMIT")
	}
}

func TestLoad_CustomSessionDuration(testingContext *testing.T) {
	setMinimalRequiredEnvironment(testingContext)
	testingContext.Setenv("SESSION_DURATION", "48h")

	loadedConfig, loadError := Load()
	if loadError != nil {
		testingContext.Fatalf("expected no error, got %v", loadError)
	}
	if loadedConfig.SessionDuration.Hours() != 48 {
		testingContext.Fatalf("expected 48h session duration, got %v", loadedConfig.SessionDuration)
	}
}

func TestLoad_ProductionEnvironment(testingContext *testing.T) {
	setMinimalRequiredEnvironment(testingContext)
	testingContext.Setenv("ENVIRONMENT", "production")

	loadedConfig, loadError := Load()
	if loadError != nil {
		testingContext.Fatalf("expected no error, got %v", loadError)
	}
	if loadedConfig.Environment != "production" {
		testingContext.Fatalf("expected environment=production, got %q", loadedConfig.Environment)
	}
}
