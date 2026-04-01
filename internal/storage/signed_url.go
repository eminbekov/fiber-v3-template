package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ValidateSignedURL checks HMAC-SHA256 token and expiry for a file path segment (see GO_FIBER_PROJECT_GUIDE §11.1).
func ValidateSignedURL(filename string, token string, expiresUnix int64, signingKey []byte) bool {
	if len(signingKey) == 0 || filename == "" || token == "" {
		return false
	}
	if time.Now().Unix() > expiresUnix {
		return false
	}
	expected := filenameSignatureHex(filename, expiresUnix, signingKey)
	return hmac.Equal([]byte(token), []byte(expected))
}

func filenameSignatureHex(filename string, expiresUnix int64, signingKey []byte) string {
	message := filename + strconv.FormatInt(expiresUnix, 10)
	mac := hmac.New(sha256.New, signingKey)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// PublicSignedFileURL builds a signed URL path for the application file download route.
func PublicSignedFileURL(publicBaseURL string, servePrefix string, filename string, signingKey []byte, ttl time.Duration) (string, error) {
	if len(signingKey) == 0 {
		return "", fmt.Errorf("PublicSignedFileURL: signing key cannot be empty")
	}
	trimmedName := strings.TrimLeft(filename, "/")
	if trimmedName == "" || strings.Contains(trimmedName, "..") {
		return "", fmt.Errorf("PublicSignedFileURL: invalid filename")
	}
	prefix := strings.TrimRight(servePrefix, "/")
	expiresUnix := time.Now().Add(ttl).Unix()
	token := filenameSignatureHex(trimmedName, expiresUnix, signingKey)
	path := fmt.Sprintf("%s/%s?token=%s&expires=%d", prefix, trimmedName, token, expiresUnix)
	if publicBaseURL != "" {
		return strings.TrimRight(publicBaseURL, "/") + path, nil
	}
	return path, nil
}
