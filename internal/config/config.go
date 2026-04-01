package config

import (
	"fmt"
	"os"
	"strings"
)

// Environment variable names (stable contract for operators and .env.example).
const (
	EnvironmentVariableName       = "ENVIRONMENT"
	LogLevelVariableName          = "LOG_LEVEL"
	HTTPListenAddressVariableName = "HTTP_LISTEN_ADDRESS"
)

const (
	DefaultEnvironment       = "development"
	DefaultLogLevel          = "debug"
	DefaultHTTPListenAddress = ":8080"
)

// Config holds application settings loaded from the environment.
type Config struct {
	Environment       string
	LogLevel          string
	HTTPListenAddress string
}

// Load reads configuration from environment variables, applies defaults, and validates.
func Load() (*Config, error) {
	loadedConfig := &Config{
		Environment:       strings.ToLower(strings.TrimSpace(getenvOrDefault(EnvironmentVariableName, DefaultEnvironment))),
		LogLevel:          strings.ToLower(strings.TrimSpace(getenvOrDefault(LogLevelVariableName, DefaultLogLevel))),
		HTTPListenAddress: strings.TrimSpace(getenvOrDefault(HTTPListenAddressVariableName, DefaultHTTPListenAddress)),
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

	return nil
}
