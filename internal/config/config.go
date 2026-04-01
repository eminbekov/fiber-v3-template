package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment variable names (stable contract for operators and .env.example).
const (
	EnvironmentVariableName       = "ENVIRONMENT"
	LogLevelVariableName          = "LOG_LEVEL"
	HTTPListenAddressVariableName = "HTTP_LISTEN_ADDRESS"
	ViewsPathVariableName         = "VIEWS_PATH"
	CORSAllowOriginsVariableName  = "CORS_ALLOW_ORIGINS"
	BodyLimitVariableName         = "BODY_LIMIT"
	OTELExporterEndpointVarName   = "OTEL_EXPORTER_ENDPOINT"
	DatabaseURLVariableName       = "DATABASE_URL"
	RedisURLVariableName          = "REDIS_URL"
	SessionDurationVariableName   = "SESSION_DURATION"
)

const (
	DefaultEnvironment       = "development"
	DefaultLogLevel          = "debug"
	DefaultHTTPListenAddress = ":8080"
	DefaultViewsPath         = "./views"
	DefaultBodyLimit         = 4 * 1024 * 1024
	DefaultSessionDuration   = "24h"
)

// Config holds application settings loaded from the environment.
type Config struct {
	Environment          string
	LogLevel             string
	HTTPListenAddress    string
	ViewsPath            string
	CORSAllowOrigins     string
	BodyLimit            int
	OTELExporterEndpoint string
	DatabaseURL          string
	RedisURL             string
	SessionDuration      time.Duration
}

// Load reads configuration from environment variables, applies defaults, and validates.
func Load() (*Config, error) {
	loadedConfig := &Config{
		Environment:          strings.ToLower(strings.TrimSpace(getenvOrDefault(EnvironmentVariableName, DefaultEnvironment))),
		LogLevel:             strings.ToLower(strings.TrimSpace(getenvOrDefault(LogLevelVariableName, DefaultLogLevel))),
		HTTPListenAddress:    strings.TrimSpace(getenvOrDefault(HTTPListenAddressVariableName, DefaultHTTPListenAddress)),
		ViewsPath:            strings.TrimSpace(getenvOrDefault(ViewsPathVariableName, DefaultViewsPath)),
		CORSAllowOrigins:     strings.TrimSpace(os.Getenv(CORSAllowOriginsVariableName)),
		BodyLimit:            DefaultBodyLimit,
		OTELExporterEndpoint: strings.TrimSpace(os.Getenv(OTELExporterEndpointVarName)),
		DatabaseURL:          strings.TrimSpace(os.Getenv(DatabaseURLVariableName)),
		RedisURL:             strings.TrimSpace(os.Getenv(RedisURLVariableName)),
		SessionDuration:      24 * time.Hour,
	}

	if configuredBodyLimit := strings.TrimSpace(os.Getenv(BodyLimitVariableName)); configuredBodyLimit != "" {
		parsedBodyLimit, parseError := strconv.Atoi(configuredBodyLimit)
		if parseError != nil {
			return nil, fmt.Errorf("config: invalid %s: %w", BodyLimitVariableName, parseError)
		}
		loadedConfig.BodyLimit = parsedBodyLimit
	}

	if validationError := loadedConfig.validate(); validationError != nil {
		return nil, validationError
	}

	return loadedConfig, nil
}

func getenvOrDefault(key string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	return value
}

func (loadedConfig *Config) validate() error {
	configuredSessionDuration := strings.TrimSpace(getenvOrDefault(SessionDurationVariableName, DefaultSessionDuration))
	parsedSessionDuration, parseError := time.ParseDuration(configuredSessionDuration)
	if parseError != nil {
		return fmt.Errorf("config: invalid %s: %w", SessionDurationVariableName, parseError)
	}
	loadedConfig.SessionDuration = parsedSessionDuration

	switch loadedConfig.Environment {
	case "development", "production":
	default:
		return fmt.Errorf(
			"config: invalid %s %q (allowed: development, production)",
			EnvironmentVariableName,
			loadedConfig.Environment,
		)
	}

	switch loadedConfig.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf(
			"config: invalid %s %q (allowed: debug, info, warn, error)",
			LogLevelVariableName,
			loadedConfig.LogLevel,
		)
	}

	if loadedConfig.HTTPListenAddress == "" {
		return fmt.Errorf("config: %s cannot be empty", HTTPListenAddressVariableName)
	}
	if loadedConfig.ViewsPath == "" {
		return fmt.Errorf("config: %s cannot be empty", ViewsPathVariableName)
	}
	if loadedConfig.BodyLimit <= 0 {
		return fmt.Errorf("config: %s must be greater than 0", BodyLimitVariableName)
	}
	if loadedConfig.DatabaseURL == "" {
		return fmt.Errorf("config: %s cannot be empty", DatabaseURLVariableName)
	}
	if loadedConfig.RedisURL == "" {
		return fmt.Errorf("config: %s cannot be empty", RedisURLVariableName)
	}

	return nil
}
