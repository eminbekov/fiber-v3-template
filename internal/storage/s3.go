package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3FileStorageOptions configures S3-compatible storage (AWS S3, MinIO, etc.).
type S3FileStorageOptions struct {
	Endpoint   string
	Region     string
	Bucket     string
	AccessKey  string
	SecretKey  string
	CDNBaseURL string
	// UsePathStyle enables path-style addressing (typical for MinIO).
	UsePathStyle bool
}

type s3FileStorage struct {
	client     *s3.Client
	presign    *s3.PresignClient
	bucket     string
	cdnBaseURL string
}

// NewS3FileStorage builds an S3-backed FileStorage.
func NewS3FileStorage(ctx context.Context, options S3FileStorageOptions) (*s3FileStorage, error) {
	if options.Bucket == "" {
		return nil, fmt.Errorf("NewS3FileStorage: bucket cannot be empty")
	}
	if options.Region == "" {
		return nil, fmt.Errorf("NewS3FileStorage: region cannot be empty")
	}

	configuration, loadError := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(options.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(options.AccessKey, options.SecretKey, ""),
		),
	)
	if loadError != nil {
		return nil, fmt.Errorf("NewS3FileStorage: load aws config: %w", loadError)
	}

	client := s3.NewFromConfig(configuration, func(clientOptions *s3.Options) {
		clientOptions.UsePathStyle = options.UsePathStyle
		if options.Endpoint != "" {
			clientOptions.BaseEndpoint = aws.String(strings.TrimRight(options.Endpoint, "/"))
		}
	})

	return &s3FileStorage{
		client:     client,
		presign:    s3.NewPresignClient(client),
		bucket:     options.Bucket,
		cdnBaseURL: strings.TrimRight(options.CDNBaseURL, "/"),
	}, nil
}

func (backend *s3FileStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	_, err := backend.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(backend.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("s3FileStorage.Upload: %w", err)
	}
	return nil
}

func (backend *s3FileStorage) URL(key string) string {
	if backend.cdnBaseURL != "" {
		return backend.cdnBaseURL + "/" + strings.TrimLeft(key, "/")
	}
	return "/" + strings.TrimLeft(key, "/")
}

func (backend *s3FileStorage) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	request, err := backend.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(backend.bucket),
		Key:    aws.String(key),
	}, func(presignOptions *s3.PresignOptions) {
		presignOptions.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("s3FileStorage.SignedURL: %w", err)
	}
	return request.URL, nil
}

func (backend *s3FileStorage) Delete(ctx context.Context, key string) error {
	_, err := backend.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(backend.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3FileStorage.Delete: %w", err)
	}
	return nil
}
