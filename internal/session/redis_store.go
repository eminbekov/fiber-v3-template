package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/redis/go-redis/v9"
)

const sessionTokenLengthBytes = 32

type RedisStore struct {
	redisClient       *redis.Client
	sessionExpiration time.Duration
}

func NewRedisStore(redisClient *redis.Client, sessionExpiration time.Duration) *RedisStore {
	return &RedisStore{
		redisClient:       redisClient,
		sessionExpiration: sessionExpiration,
	}
}

func (store *RedisStore) Create(ctx context.Context, metadata Metadata) (string, error) {
	sessionToken, tokenError := generateSecureToken()
	if tokenError != nil {
		return "", fmt.Errorf("redisStore.Create token: %w", tokenError)
	}

	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = time.Now().UTC()
	}

	sessionData, marshalError := sonic.Marshal(metadata)
	if marshalError != nil {
		return "", fmt.Errorf("redisStore.Create marshal: %w", marshalError)
	}

	sessionKey := sessionKeyByToken(sessionToken)
	if setError := store.redisClient.Set(ctx, sessionKey, sessionData, store.sessionExpiration).Err(); setError != nil {
		return "", fmt.Errorf("redisStore.Create set: %w", setError)
	}

	return sessionToken, nil
}

func (store *RedisStore) Get(ctx context.Context, token string) (*Metadata, error) {
	sessionData, getError := store.redisClient.Get(ctx, sessionKeyByToken(token)).Bytes()
	if getError != nil {
		if errors.Is(getError, redis.Nil) {
			return nil, domain.ErrUnauthorized
		}

		return nil, fmt.Errorf("redisStore.Get: %w", getError)
	}

	var metadata Metadata
	if unmarshalError := sonic.Unmarshal(sessionData, &metadata); unmarshalError != nil {
		return nil, fmt.Errorf("redisStore.Get unmarshal: %w", unmarshalError)
	}

	return &metadata, nil
}

func (store *RedisStore) Delete(ctx context.Context, token string) error {
	if deleteError := store.redisClient.Del(ctx, sessionKeyByToken(token)).Err(); deleteError != nil {
		return fmt.Errorf("redisStore.Delete: %w", deleteError)
	}

	return nil
}

func (store *RedisStore) Extend(ctx context.Context, token string) error {
	if expireError := store.redisClient.Expire(ctx, sessionKeyByToken(token), store.sessionExpiration).Err(); expireError != nil {
		return fmt.Errorf("redisStore.Extend: %w", expireError)
	}

	return nil
}

func sessionKeyByToken(token string) string {
	return "session:" + token
}

func generateSecureToken() (string, error) {
	randomBytes := make([]byte, sessionTokenLengthBytes)
	if _, readError := rand.Read(randomBytes); readError != nil {
		return "", fmt.Errorf("generateSecureToken read: %w", readError)
	}

	return hex.EncodeToString(randomBytes), nil
}
