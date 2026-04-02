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
	// [module:grpc:start]
	GRPCListenAddressVariableName = "GRPC_LISTEN_ADDRESS"
	// [module:grpc:end]
	// [module:views:start]
	ViewsPathVariableName = "VIEWS_PATH"
	// [module:views:end]
	CORSAllowOriginsVariableName = "CORS_ALLOW_ORIGINS"
	BodyLimitVariableName        = "BODY_LIMIT"
	OTELExporterEndpointVarName  = "OTEL_EXPORTER_ENDPOINT"
	DatabaseURLVariableName      = "DATABASE_URL"
	RedisURLVariableName         = "REDIS_URL"
	// [module:nats:start]
	NATSURLVariableName = "NATS_URL"
	// [module:nats:end]
	SessionDurationVariableName = "SESSION_DURATION"
	// [module:storage:start]
	StorageTypeVariableName     = "STORAGE_TYPE"
	StorageLocalBasePathVarName = "STORAGE_LOCAL_BASE_PATH"
	S3EndpointVariableName      = "S3_ENDPOINT"
	S3BucketVariableName        = "S3_BUCKET"
	S3AccessKeyVariableName     = "S3_ACCESS_KEY"
	S3SecretKeyVariableName     = "S3_SECRET_KEY" //nolint:gosec // env var name, not a credential
	S3RegionVariableName        = "S3_REGION"
	CDNBaseURLVariableName      = "CDN_BASE_URL"
	FileSigningKeyVariableName  = "FILE_SIGNING_KEY"
	SignedURLTTLVariableName    = "SIGNED_URL_TTL"
	// [module:storage:end]
)

const (
	DefaultEnvironment       = "development"
	DefaultLogLevel          = "debug"
	DefaultHTTPListenAddress = ":8080"
	// [module:grpc:start]
	DefaultGRPCListenAddress = ":9090"
	// [module:grpc:end]
	// [module:nats:start]
	DefaultNATSURL = "nats://localhost:4222"
	// [module:nats:end]
	// [module:views:start]
	DefaultViewsPath = "./views"
	// [module:views:end]
	DefaultBodyLimit       = 4 * 1024 * 1024
	DefaultSessionDuration = "24h"
	// [module:storage:start]
	DefaultStorageType      = "local"
	DefaultStorageLocalPath = "./uploads"
	DefaultSignedURLTTL     = "15m"
	// [module:storage:end]
)

// Config holds application settings loaded from the environment.
type Config struct {
	Environment       string
	LogLevel          string
	HTTPListenAddress string
	// [module:grpc:start]
	GRPCListenAddress string
	// [module:grpc:end]
	// [module:views:start]
	ViewsPath string
	// [module:views:end]
	CORSAllowOrigins     string
	BodyLimit            int
	OTELExporterEndpoint string
	DatabaseURL          string
	RedisURL             string
	// [module:nats:start]
	NATSURL string
	// [module:nats:end]
	SessionDuration time.Duration
	// [module:storage:start]
	StorageType          string
	StorageLocalBasePath string
	S3Endpoint           string
	S3Bucket             string
	S3AccessKey          string
	S3SecretKey          string
	S3Region             string
	CDNBaseURL           string
	FileSigningKey       string
	SignedURLTTL         time.Duration
	// [module:storage:end]
}

// Load reads configuration from environment variables, applies defaults, and validates.
func Load() (*Config, error) {
	loadedConfig := &Config{
		Environment:       strings.ToLower(strings.TrimSpace(getenvOrDefault(EnvironmentVariableName, DefaultEnvironment))),
		LogLevel:          strings.ToLower(strings.TrimSpace(getenvOrDefault(LogLevelVariableName, DefaultLogLevel))),
		HTTPListenAddress: strings.TrimSpace(getenvOrDefault(HTTPListenAddressVariableName, DefaultHTTPListenAddress)),
		// [module:grpc:start]
		GRPCListenAddress: strings.TrimSpace(getenvOrDefault(GRPCListenAddressVariableName, DefaultGRPCListenAddress)),
		// [module:grpc:end]
		// [module:views:start]
		ViewsPath: strings.TrimSpace(getenvOrDefault(ViewsPathVariableName, DefaultViewsPath)),
		// [module:views:end]
		CORSAllowOrigins:     strings.TrimSpace(os.Getenv(CORSAllowOriginsVariableName)),
		BodyLimit:            DefaultBodyLimit,
		OTELExporterEndpoint: strings.TrimSpace(os.Getenv(OTELExporterEndpointVarName)),
		DatabaseURL:          strings.TrimSpace(os.Getenv(DatabaseURLVariableName)),
		RedisURL:             strings.TrimSpace(os.Getenv(RedisURLVariableName)),
		// [module:nats:start]
		NATSURL: strings.TrimSpace(getenvOrDefault(NATSURLVariableName, DefaultNATSURL)),
		// [module:nats:end]
		SessionDuration: 24 * time.Hour,
		// [module:storage:start]
		StorageType:          strings.ToLower(strings.TrimSpace(getenvOrDefault(StorageTypeVariableName, DefaultStorageType))),
		StorageLocalBasePath: strings.TrimSpace(getenvOrDefault(StorageLocalBasePathVarName, DefaultStorageLocalPath)),
		S3Endpoint:           strings.TrimSpace(os.Getenv(S3EndpointVariableName)),
		S3Bucket:             strings.TrimSpace(os.Getenv(S3BucketVariableName)),
		S3AccessKey:          strings.TrimSpace(os.Getenv(S3AccessKeyVariableName)),
		S3SecretKey:          strings.TrimSpace(os.Getenv(S3SecretKeyVariableName)),
		S3Region:             strings.TrimSpace(os.Getenv(S3RegionVariableName)),
		CDNBaseURL:           strings.TrimSpace(os.Getenv(CDNBaseURLVariableName)),
		FileSigningKey:       strings.TrimSpace(os.Getenv(FileSigningKeyVariableName)),
		// [module:storage:end]
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

	// [module:storage:start]
	signedURLTTLString := strings.TrimSpace(getenvOrDefault(SignedURLTTLVariableName, DefaultSignedURLTTL))
	parsedSignedURLTTL, signedTTLError := time.ParseDuration(signedURLTTLString)
	if signedTTLError != nil {
		return nil, fmt.Errorf("config: invalid %s: %w", SignedURLTTLVariableName, signedTTLError)
	}
	if parsedSignedURLTTL <= 0 {
		return nil, fmt.Errorf("config: %s must be greater than 0", SignedURLTTLVariableName)
	}
	loadedConfig.SignedURLTTL = parsedSignedURLTTL
	// [module:storage:end]

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

	if enumError := validateEnum(loadedConfig.Environment, EnvironmentVariableName, []string{"development", "production"}); enumError != nil {
		return enumError
	}
	if enumError := validateEnum(loadedConfig.LogLevel, LogLevelVariableName, []string{"debug", "info", "warn", "error"}); enumError != nil {
		return enumError
	}
	if fieldError := firstError(
		validateRequiredField(loadedConfig.HTTPListenAddress, HTTPListenAddressVariableName),
		// [module:grpc:start]
		validateRequiredField(loadedConfig.GRPCListenAddress, GRPCListenAddressVariableName),
		// [module:grpc:end]
		// [module:views:start]
		validateRequiredField(loadedConfig.ViewsPath, ViewsPathVariableName),
		// [module:views:end]
		validatePositiveInteger(loadedConfig.BodyLimit, BodyLimitVariableName),
		validateRequiredField(loadedConfig.DatabaseURL, DatabaseURLVariableName),
		validateRequiredField(loadedConfig.RedisURL, RedisURLVariableName),
		// [module:nats:start]
		validateRequiredField(loadedConfig.NATSURL, NATSURLVariableName),
		// [module:nats:end]
	); fieldError != nil {
		return fieldError
	}
	// [module:storage:start]
	if enumError := validateEnum(loadedConfig.StorageType, StorageTypeVariableName, []string{"s3", "local"}); enumError != nil {
		return enumError
	}
	if fieldError := validateRequiredField(loadedConfig.FileSigningKey, FileSigningKeyVariableName); fieldError != nil {
		return fieldError
	}
	if loadedConfig.StorageType == "s3" {
		if s3Error := loadedConfig.validateS3Config(); s3Error != nil {
			return s3Error
		}
	}
	if loadedConfig.StorageType == "local" && loadedConfig.StorageLocalBasePath == "" {
		return fmt.Errorf("config: %s cannot be empty when %s=local", StorageLocalBasePathVarName, StorageTypeVariableName)
	}
	// [module:storage:end]

	return nil
}

// [module:storage:start]
func (loadedConfig *Config) validateS3Config() error {
	return firstError(
		validateRequiredField(loadedConfig.S3Bucket, S3BucketVariableName),
		validateRequiredField(loadedConfig.S3AccessKey, S3AccessKeyVariableName),
		validateRequiredField(loadedConfig.S3SecretKey, S3SecretKeyVariableName),
		validateRequiredField(loadedConfig.S3Region, S3RegionVariableName),
	)
}

// [module:storage:end]

func validateRequiredField(value string, fieldName string) error {
	if value == "" {
		return fmt.Errorf("config: %s cannot be empty", fieldName)
	}
	return nil
}

func validatePositiveInteger(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("config: %s must be greater than 0", fieldName)
	}
	return nil
}

func validateEnum(value string, fieldName string, allowed []string) error {
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return nil
		}
	}
	return fmt.Errorf("config: invalid %s %q (allowed: %s)", fieldName, value, strings.Join(allowed, ", "))
}

func firstError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
