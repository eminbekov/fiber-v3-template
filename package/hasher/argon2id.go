package hasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2IDTimeIterations uint32 = 3
	argon2IDMemoryKiB      uint32 = 64 * 1024
	argon2IDThreads        uint8  = 4
	argon2IDKeyLength      uint32 = 32
	argon2IDSaltLength            = 16
)

// Hash hashes a password with Argon2id and returns a PHC-formatted string.
func Hash(password string) (string, error) {
	passwordBytes := []byte(password)
	saltBytes := make([]byte, argon2IDSaltLength)
	if _, readError := rand.Read(saltBytes); readError != nil {
		return "", fmt.Errorf("hasher.Hash salt: %w", readError)
	}

	hashBytes := argon2.IDKey(
		passwordBytes,
		saltBytes,
		argon2IDTimeIterations,
		argon2IDMemoryKiB,
		argon2IDThreads,
		argon2IDKeyLength,
	)

	saltBase64 := base64.RawStdEncoding.EncodeToString(saltBytes)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hashBytes)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2IDMemoryKiB,
		argon2IDTimeIterations,
		argon2IDThreads,
		saltBase64,
		hashBase64,
	), nil
}

// Verify verifies a plaintext password against a PHC-formatted Argon2id hash.
func Verify(password string, encodedHash string) (bool, error) {
	hashParts := strings.Split(encodedHash, "$")
	if len(hashParts) != 6 {
		return false, fmt.Errorf("hasher.Verify: invalid hash format")
	}
	if hashParts[1] != "argon2id" {
		return false, fmt.Errorf("hasher.Verify: unsupported algorithm")
	}

	var memoryKiB uint32
	var timeIterations uint32
	var threads uint8
	if _, scanError := fmt.Sscanf(hashParts[3], "m=%d,t=%d,p=%d", &memoryKiB, &timeIterations, &threads); scanError != nil {
		return false, fmt.Errorf("hasher.Verify params: %w", scanError)
	}

	saltBytes, saltError := base64.RawStdEncoding.DecodeString(hashParts[4])
	if saltError != nil {
		return false, fmt.Errorf("hasher.Verify salt: %w", saltError)
	}
	expectedHashBytes, hashError := base64.RawStdEncoding.DecodeString(hashParts[5])
	if hashError != nil {
		return false, fmt.Errorf("hasher.Verify hash: %w", hashError)
	}

	computedHashBytes := argon2.IDKey(
		[]byte(password),
		saltBytes,
		timeIterations,
		memoryKiB,
		threads,
		uint32(len(expectedHashBytes)),
	)

	return subtle.ConstantTimeCompare(computedHashBytes, expectedHashBytes) == 1, nil
}
