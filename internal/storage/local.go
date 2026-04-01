package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// LocalFileStorageOptions configures filesystem-backed storage (development or single-node).
type LocalFileStorageOptions struct {
	BasePath            string
	CDNBaseURL          string
	FileSigningKey      []byte
	FileServePathPrefix string // e.g. "/api/files" (no trailing slash)
}

type localFileStorage struct {
	basePath        string
	publicURLPrefix string
	signingKey      []byte
	servePrefix     string
}

// NewLocalFileStorage stores objects on the local filesystem under basePath.
func NewLocalFileStorage(options LocalFileStorageOptions) (*localFileStorage, error) {
	if options.BasePath == "" {
		return nil, fmt.Errorf("NewLocalFileStorage: base path cannot be empty")
	}
	if len(options.FileSigningKey) == 0 {
		return nil, fmt.Errorf("NewLocalFileStorage: file signing key cannot be empty")
	}
	prefix := strings.TrimSpace(options.FileServePathPrefix)
	if prefix == "" {
		prefix = "/api/files"
	}
	prefix = strings.TrimRight(prefix, "/")

	if mkdirError := os.MkdirAll(options.BasePath, 0o750); mkdirError != nil {
		return nil, fmt.Errorf("NewLocalFileStorage: mkdir: %w", mkdirError)
	}

	return &localFileStorage{
		basePath:        filepath.Clean(options.BasePath),
		publicURLPrefix: strings.TrimRight(options.CDNBaseURL, "/"),
		signingKey:      options.FileSigningKey,
		servePrefix:     prefix,
	}, nil
}

func (backend *localFileStorage) resolvePath(key string) (string, error) {
	trimmedKey := strings.Trim(key, "/")
	if trimmedKey == "" || strings.Contains(trimmedKey, "..") {
		return "", fmt.Errorf("localFileStorage.resolvePath: invalid key")
	}
	fullPath := filepath.Join(backend.basePath, filepath.FromSlash(trimmedKey))
	baseAbsolute, baseError := filepath.Abs(backend.basePath)
	if baseError != nil {
		return "", fmt.Errorf("localFileStorage.resolvePath: %w", baseError)
	}
	pathAbsolute, pathError := filepath.Abs(fullPath)
	if pathError != nil {
		return "", fmt.Errorf("localFileStorage.resolvePath: %w", pathError)
	}
	relative, relativeError := filepath.Rel(baseAbsolute, pathAbsolute)
	if relativeError != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("localFileStorage.resolvePath: path escapes base")
	}
	return pathAbsolute, nil
}

func (backend *localFileStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	fullPath, resolveError := backend.resolvePath(key)
	if resolveError != nil {
		return fmt.Errorf("localFileStorage.Upload: %w", resolveError)
	}
	if mkdirError := os.MkdirAll(filepath.Dir(fullPath), 0o750); mkdirError != nil {
		return fmt.Errorf("localFileStorage.Upload: mkdir: %w", mkdirError)
	}
	file, createError := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if createError != nil {
		return fmt.Errorf("localFileStorage.Upload: %w", createError)
	}
	defer func() { _ = file.Close() }()
	if _, copyError := io.Copy(file, reader); copyError != nil {
		return fmt.Errorf("localFileStorage.Upload: write: %w", copyError)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (backend *localFileStorage) URL(key string) string {
	path := fmt.Sprintf("%s/%s", backend.servePrefix, strings.TrimLeft(key, "/"))
	if backend.publicURLPrefix != "" {
		return backend.publicURLPrefix + path
	}
	return path
}

func (backend *localFileStorage) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	expiresUnix := time.Now().Add(expiry).Unix()
	trimmedKey := strings.TrimLeft(key, "/")
	token := signFilenameToken(trimmedKey, expiresUnix, backend.signingKey)
	path := fmt.Sprintf("%s/%s?token=%s&expires=%d", backend.servePrefix, trimmedKey, token, expiresUnix)
	if backend.publicURLPrefix != "" {
		return backend.publicURLPrefix + path, nil
	}
	return path, nil
}

func signFilenameToken(filename string, expiresUnix int64, signingKey []byte) string {
	message := filename + strconv.FormatInt(expiresUnix, 10)
	mac := hmac.New(sha256.New, signingKey)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func (backend *localFileStorage) Delete(ctx context.Context, key string) error {
	fullPath, resolveError := backend.resolvePath(key)
	if resolveError != nil {
		return fmt.Errorf("localFileStorage.Delete: %w", resolveError)
	}
	if removeError := os.Remove(fullPath); removeError != nil && !os.IsNotExist(removeError) {
		return fmt.Errorf("localFileStorage.Delete: %w", removeError)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
