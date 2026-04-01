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
	GRPCListenAddressVariableName = "GRPC_LISTEN_ADDRESS"
	ViewsPathVariableName         = "VIEWS_PATH"
	CORSAllowOriginsVariableName  = "CORS_ALLOW_ORIGINS"
	BodyLimitVariableName         = "BODY_LIMIT"
	OTELExporterEndpointVarName   = "OTEL_EXPORTER_ENDPOINT"
	DatabaseURLVariableName       = "DATABASE_URL"
	RedisURLVariableName          = "REDIS_URL"
	NATSURLVariableName           = "NATS_URL"
	SessionDurationVariableName   = "SESSION_DURATION"
	StorageTypeVariableName       = "STORAGE_TYPE"
	StorageLocalBasePathVarName   = "STORAGE_LOCAL_BASE_PATH"
	S3EndpointVariableName        = "S3_ENDPOINT"
	S3BucketVariableName          = "S3_BUCKET"
	S3AccessKeyVariableName       = "S3_ACCESS_KEY"
	S3SecretKeyVariableName       = "S3_SECRET_KEY"
	S3RegionVariableName          = "S3_REGION"
	CDNBaseURLVariableName        = "CDN_BASE_URL"
	FileSigningKeyVariableName    = "FILE_SIGNING_KEY"
	SignedURLTTLVariableName      = "SIGNED_URL_TTL"
)

const (
	DefaultEnvironment       = "development"
	DefaultLogLevel          = "debug"
	DefaultHTTPListenAddress = ":8080"
	DefaultGRPCListenAddress = ":9090"
	DefaultNATSURL           = "nats://localhost:4222"
	DefaultViewsPath         = "./views"
	DefaultBodyLimit         = 4 * 1024 * 1024
	DefaultSessionDuration   = "24h"
	DefaultStorageType       = "local"
	DefaultStorageLocalPath  = "./uploads"
	DefaultSignedURLTTL      = "15m"
)

// Config holds application settings loaded from the environment.
type Config struct {
	Environment          string
	LogLevel             string
	HTTPListenAddress    string
	GRPCListenAddress    string
	ViewsPath            string
	CORSAllowOrigins     string
	BodyLimit            int
	OTELExporterEndpoint string
	DatabaseURL          string
	RedisURL             string
	NATSURL              string
	SessionDuration      time.Duration
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
}

// Load reads configuration from environment variables, applies defaults, and validates.
func Load() (*Config, error) {
	loadedConfig := &Config{
		Environment:          strings.ToLower(strings.TrimSpace(getenvOrDefault(EnvironmentVariableName, DefaultEnvironment))),
		LogLevel:             strings.ToLower(strings.TrimSpace(getenvOrDefault(LogLevelVariableName, DefaultLogLevel))),
		HTTPListenAddress:    strings.TrimSpace(getenvOrDefault(HTTPListenAddressVariableName, DefaultHTTPListenAddress)),
		GRPCListenAddress:    strings.TrimSpace(getenvOrDefault(GRPCListenAddressVariableName, DefaultGRPCListenAddress)),
		ViewsPath:            strings.TrimSpace(getenvOrDefault(ViewsPathVariableName, DefaultViewsPath)),
		CORSAllowOrigins:     strings.TrimSpace(os.Getenv(CORSAllowOriginsVariableName)),
		BodyLimit:            DefaultBodyLimit,
		OTELExporterEndpoint: strings.TrimSpace(os.Getenv(OTELExporterEndpointVarName)),
		DatabaseURL:          strings.TrimSpace(os.Getenv(DatabaseURLVariableName)),
		RedisURL:             strings.TrimSpace(os.Getenv(RedisURLVariableName)),
		NATSURL:              strings.TrimSpace(getenvOrDefault(NATSURLVariableName, DefaultNATSURL)),
		SessionDuration:      24 * time.Hour,
		StorageType:          strings.ToLower(strings.TrimSpace(getenvOrDefault(StorageTypeVariableName, DefaultStorageType))),
		StorageLocalBasePath: strings.TrimSpace(getenvOrDefault(StorageLocalBasePathVarName, DefaultStorageLocalPath)),
		S3Endpoint:           strings.TrimSpace(os.Getenv(S3EndpointVariableName)),
		S3Bucket:             strings.TrimSpace(os.Getenv(S3BucketVariableName)),
		S3AccessKey:          strings.TrimSpace(os.Getenv(S3AccessKeyVariableName)),
		S3SecretKey:          strings.TrimSpace(os.Getenv(S3SecretKeyVariableName)),
		S3Region:             strings.TrimSpace(os.Getenv(S3RegionVariableName)),
		CDNBaseURL:           strings.TrimSpace(os.Getenv(CDNBaseURLVariableName)),
		FileSigningKey:       strings.TrimSpace(os.Getenv(FileSigningKeyVariableName)),
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

	signedURLTTLString := strings.TrimSpace(getenvOrDefault(SignedURLTTLVariableName, DefaultSignedURLTTL))
	parsedSignedURLTTL, signedTTLError := time.ParseDuration(signedURLTTLString)
	if signedTTLError != nil {
		return nil, fmt.Errorf("config: invalid %s: %w", SignedURLTTLVariableName, signedTTLError)
	}
	if parsedSignedURLTTL <= 0 {
		return nil, fmt.Errorf("config: %s must be greater than 0", SignedURLTTLVariableName)
	}
	loadedConfig.SignedURLTTL = parsedSignedURLTTL

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
	if loadedConfig.GRPCListenAddress == "" {
		return fmt.Errorf("config: %s cannot be empty", GRPCListenAddressVariableName)
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
	if loadedConfig.NATSURL == "" {
		return fmt.Errorf("config: %s cannot be empty", NATSURLVariableName)
	}

	switch loadedConfig.StorageType {
	case "s3", "local":
	default:
		return fmt.Errorf(
			"config: invalid %s %q (allowed: s3, local)",
			StorageTypeVariableName,
			loadedConfig.StorageType,
		)
	}

	if loadedConfig.FileSigningKey == "" {
		return fmt.Errorf("config: %s cannot be empty", FileSigningKeyVariableName)
	}

	if loadedConfig.StorageType == "s3" {
		if loadedConfig.S3Bucket == "" {
			return fmt.Errorf("config: %s cannot be empty when %s=s3", S3BucketVariableName, StorageTypeVariableName)
		}
		if loadedConfig.S3AccessKey == "" {
			return fmt.Errorf("config: %s cannot be empty when %s=s3", S3AccessKeyVariableName, StorageTypeVariableName)
		}
		if loadedConfig.S3SecretKey == "" {
			return fmt.Errorf("config: %s cannot be empty when %s=s3", S3SecretKeyVariableName, StorageTypeVariableName)
		}
		if loadedConfig.S3Region == "" {
			return fmt.Errorf("config: %s cannot be empty when %s=s3", S3RegionVariableName, StorageTypeVariableName)
		}
	}

	if loadedConfig.StorageType == "local" && loadedConfig.StorageLocalBasePath == "" {
		return fmt.Errorf("config: %s cannot be empty when %s=local", StorageLocalBasePathVarName, StorageTypeVariableName)
	}

	return nil
}
