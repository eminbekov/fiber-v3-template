package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Environment variable names (stable contract for operators and .env.example).
const (
	EnvironmentVariableName       = "ENVIRONMENT"
	LogLevelVariableName          = "LOG_LEVEL"
	HTTPListenAddressVariableName = "HTTP_LISTEN_ADDRESS"
	CORSAllowOriginsVariableName  = "CORS_ALLOW_ORIGINS"
	BodyLimitVariableName         = "BODY_LIMIT"
	OTELExporterEndpointVarName   = "OTEL_EXPORTER_ENDPOINT"
)

const (
	DefaultEnvironment       = "development"
	DefaultLogLevel          = "debug"
	DefaultHTTPListenAddress = ":8080"
	DefaultBodyLimit         = 4 * 1024 * 1024
)

// Config holds application settings loaded from the environment.
type Config struct {
	Environment          string
	LogLevel             string
	HTTPListenAddress    string
	CORSAllowOrigins     string
	BodyLimit            int
	OTELExporterEndpoint string
}

// Load reads configuration from environment variables, applies defaults, and validates.
func Load() (*Config, error) {
	loadedConfig := &Config{
		Environment:          strings.ToLower(strings.TrimSpace(getenvOrDefault(EnvironmentVariableName, DefaultEnvironment))),
		LogLevel:             strings.ToLower(strings.TrimSpace(getenvOrDefault(LogLevelVariableName, DefaultLogLevel))),
		HTTPListenAddress:    strings.TrimSpace(getenvOrDefault(HTTPListenAddressVariableName, DefaultHTTPListenAddress)),
		CORSAllowOrigins:     strings.TrimSpace(os.Getenv(CORSAllowOriginsVariableName)),
		BodyLimit:            DefaultBodyLimit,
		OTELExporterEndpoint: strings.TrimSpace(os.Getenv(OTELExporterEndpointVarName)),
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
	if loadedConfig.BodyLimit <= 0 {
		return fmt.Errorf("config: %s must be greater than 0", BodyLimitVariableName)
	}

	return nil
}
