package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/storage"
	"github.com/gofrs/uuid/v5"
)

type FileService struct {
	fileStorage  storage.FileStorage
	signedURLTTL time.Duration
}

func NewFileService(fileStorage storage.FileStorage, signedURLTTL time.Duration) *FileService {
	return &FileService{
		fileStorage:  fileStorage,
		signedURLTTL: signedURLTTL,
	}
}

func (service *FileService) Upload(ctx context.Context, reader io.Reader, originalFilename string, contentType string) (string, error) {
	newIdentifier, identifierError := uuid.NewV7()
	if identifierError != nil {
		return "", fmt.Errorf("fileService.Upload: new uuid v7: %w", identifierError)
	}
	extension := strings.ToLower(filepath.Ext(originalFilename))
	if extension == "" {
		extension = extensionFromContentType(contentType)
	}
	objectKey := newIdentifier.String() + extension
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	if uploadError := service.fileStorage.Upload(ctx, objectKey, reader, contentType); uploadError != nil {
		return "", fmt.Errorf("fileService.Upload: %w", uploadError)
	}
	return objectKey, nil
}

func (service *FileService) SignedDownloadURL(ctx context.Context, objectKey string) (string, error) {
	signed, err := service.fileStorage.SignedURL(ctx, objectKey, service.signedURLTTL)
	if err != nil {
		return "", fmt.Errorf("fileService.SignedDownloadURL: %w", err)
	}
	return signed, nil
}

func (service *FileService) PublicURL(objectKey string) string {
	return service.fileStorage.URL(objectKey)
}

func (service *FileService) Open(ctx context.Context, objectKey string) (io.ReadCloser, string, error) {
	reader, contentType, err := service.fileStorage.Open(ctx, objectKey)
	if err != nil {
		return nil, "", fmt.Errorf("fileService.Open: %w", err)
	}
	return reader, contentType, nil
}

func extensionFromContentType(contentType string) string {
	switch strings.TrimSpace(strings.ToLower(contentType)) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
	}
}
