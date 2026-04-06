package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func newMockFileStorage(uploadError error) *mockFileStorage {
	return &mockFileStorage{
		uploadFunction:    func(ctx context.Context, key string, reader io.Reader, contentType string) error { return uploadError },
		openFunction:      func(ctx context.Context, key string) (io.ReadCloser, string, error) { return nil, "", nil },
		urlFunction:       func(key string) string { return "" },
		signedURLFunction: func(ctx context.Context, key string, expiry time.Duration) (string, error) { return "", nil },
		deleteFunction:    func(ctx context.Context, key string) error { return nil },
	}
}

func TestFileService_Upload_Extension(testingContext *testing.T) {
	testingContext.Parallel()

	tests := []struct {
		name             string
		originalFilename string
		contentType      string
		wantExtension    string
	}{
		{"known extension preserved", "photo.jpg", "image/jpeg", ".jpg"},
		{"uppercase extension lowered", "document.PDF", "application/pdf", ".pdf"},
		{"extension inferred from content type", "noext", "image/png", ".png"},
		{"extension from filename wins", "file.bin", "", ".bin"},
	}

	for _, testCase := range tests {
		testCase := testCase
		testingContext.Run(testCase.name, func(testingContext *testing.T) {
			testingContext.Parallel()

			service := NewFileService(newMockFileStorage(nil), 15*time.Minute)
			objectKey, uploadError := service.Upload(
				context.Background(), bytes.NewReader([]byte("data")),
				testCase.originalFilename, testCase.contentType,
			)
			if uploadError != nil {
				testingContext.Fatalf("unexpected error: %v", uploadError)
			}
			if !strings.HasSuffix(objectKey, testCase.wantExtension) {
				testingContext.Fatalf("expected key to end with %q, got %q", testCase.wantExtension, objectKey)
			}
		})
	}
}

func TestFileService_Upload_DefaultContentType(testingContext *testing.T) {
	testingContext.Parallel()

	var capturedContentType string
	storage := &mockFileStorage{
		uploadFunction: func(ctx context.Context, key string, reader io.Reader, contentType string) error {
			capturedContentType = contentType
			return nil
		},
		openFunction:      func(ctx context.Context, key string) (io.ReadCloser, string, error) { return nil, "", nil },
		urlFunction:       func(key string) string { return "" },
		signedURLFunction: func(ctx context.Context, key string, expiry time.Duration) (string, error) { return "", nil },
		deleteFunction:    func(ctx context.Context, key string) error { return nil },
	}

	service := NewFileService(storage, 15*time.Minute)
	_, uploadError := service.Upload(context.Background(), bytes.NewReader([]byte("data")), "file.bin", "")
	if uploadError != nil {
		testingContext.Fatalf("unexpected error: %v", uploadError)
	}
	if capturedContentType != "application/octet-stream" {
		testingContext.Fatalf("expected default content type, got %q", capturedContentType)
	}
}

func TestFileService_Upload_StorageError(testingContext *testing.T) {
	testingContext.Parallel()

	service := NewFileService(newMockFileStorage(errors.New("disk full")), 15*time.Minute)
	_, uploadError := service.Upload(context.Background(), bytes.NewReader([]byte("data")), "file.txt", "text/plain")
	if uploadError == nil {
		testingContext.Fatalf("expected error, got nil")
	}
}

func TestFileService_SignedDownloadURL(testingContext *testing.T) {
	testingContext.Parallel()

	testingContext.Run("returns signed URL from storage", func(testingContext *testing.T) {
		testingContext.Parallel()

		storage := &mockFileStorage{
			uploadFunction:    func(ctx context.Context, key string, reader io.Reader, contentType string) error { return nil },
			openFunction:      func(ctx context.Context, key string) (io.ReadCloser, string, error) { return nil, "", nil },
			urlFunction:       func(key string) string { return "" },
			signedURLFunction: func(ctx context.Context, key string, expiry time.Duration) (string, error) { return "https://signed/" + key, nil },
			deleteFunction:    func(ctx context.Context, key string) error { return nil },
		}
		service := NewFileService(storage, 15*time.Minute)

		signedURL, signError := service.SignedDownloadURL(context.Background(), "test-key.jpg")
		if signError != nil {
			testingContext.Fatalf("unexpected error: %v", signError)
		}
		if signedURL != "https://signed/test-key.jpg" {
			testingContext.Fatalf("expected signed URL, got %q", signedURL)
		}
	})

	testingContext.Run("returns error on storage failure", func(testingContext *testing.T) {
		testingContext.Parallel()

		storage := &mockFileStorage{
			uploadFunction:    func(ctx context.Context, key string, reader io.Reader, contentType string) error { return nil },
			openFunction:      func(ctx context.Context, key string) (io.ReadCloser, string, error) { return nil, "", nil },
			urlFunction:       func(key string) string { return "" },
			signedURLFunction: func(ctx context.Context, key string, expiry time.Duration) (string, error) { return "", errors.New("sign failed") },
			deleteFunction:    func(ctx context.Context, key string) error { return nil },
		}
		service := NewFileService(storage, 15*time.Minute)

		_, signError := service.SignedDownloadURL(context.Background(), "test-key.jpg")
		if signError == nil {
			testingContext.Fatalf("expected error, got nil")
		}
	})
}

func TestFileService_PublicURL(testingContext *testing.T) {
	testingContext.Parallel()

	storage := &mockFileStorage{
		uploadFunction:    func(ctx context.Context, key string, reader io.Reader, contentType string) error { return nil },
		openFunction:      func(ctx context.Context, key string) (io.ReadCloser, string, error) { return nil, "", nil },
		urlFunction:       func(key string) string { return "https://cdn.example.com/" + key },
		signedURLFunction: func(ctx context.Context, key string, expiry time.Duration) (string, error) { return "", nil },
		deleteFunction:    func(ctx context.Context, key string) error { return nil },
	}
	service := NewFileService(storage, 15*time.Minute)

	publicURL := service.PublicURL("my-file.png")
	if publicURL != "https://cdn.example.com/my-file.png" {
		testingContext.Fatalf("expected public URL, got %q", publicURL)
	}
}
