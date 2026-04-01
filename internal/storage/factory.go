package storage

import (
	"context"
	"fmt"

	"github.com/eminbekov/fiber-v3-template/internal/config"
)

// NewFromApplicationConfig builds the FileStorage implementation selected by configuration.
func NewFromApplicationConfig(ctx context.Context, applicationConfiguration *config.Config) (FileStorage, error) {
	switch applicationConfiguration.StorageType {
	case TypeLocal:
		localBackend, localError := NewLocalFileStorage(LocalFileStorageOptions{
			BasePath:            applicationConfiguration.StorageLocalBasePath,
			CDNBaseURL:          applicationConfiguration.CDNBaseURL,
			FileSigningKey:      []byte(applicationConfiguration.FileSigningKey),
			FileServePathPrefix: "/api/files",
		})
		if localError != nil {
			return nil, fmt.Errorf("NewFromApplicationConfig: %w", localError)
		}
		return localBackend, nil
	case TypeS3:
		usePathStyle := applicationConfiguration.S3Endpoint != ""
		s3Backend, s3Error := NewS3FileStorage(ctx, S3FileStorageOptions{
			Endpoint:     applicationConfiguration.S3Endpoint,
			Region:       applicationConfiguration.S3Region,
			Bucket:       applicationConfiguration.S3Bucket,
			AccessKey:    applicationConfiguration.S3AccessKey,
			SecretKey:    applicationConfiguration.S3SecretKey,
			CDNBaseURL:   applicationConfiguration.CDNBaseURL,
			UsePathStyle: usePathStyle,
		})
		if s3Error != nil {
			return nil, fmt.Errorf("NewFromApplicationConfig: %w", s3Error)
		}
		return s3Backend, nil
	default:
		return nil, fmt.Errorf("NewFromApplicationConfig: unknown %s %q", config.StorageTypeVariableName, applicationConfiguration.StorageType)
	}
}
