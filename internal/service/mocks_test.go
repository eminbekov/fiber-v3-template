package service

import (
	"context"
	"io"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/repository"
	"github.com/eminbekov/fiber-v3-template/internal/session"
	"github.com/eminbekov/fiber-v3-template/internal/storage"
	"github.com/gofrs/uuid/v5"
)

var (
	_ repository.UserRepository       = (*mockUserRepository)(nil)
	_ repository.RoleRepository       = (*mockRoleRepository)(nil)
	_ repository.PermissionRepository = (*mockPermissionRepository)(nil)
	_ session.Store                   = (*mockSessionStore)(nil)
	_ storage.FileStorage             = (*mockFileStorage)(nil)
)

type mockUserRepository struct {
	findByIDFunction       func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	findByUsernameFunction func(ctx context.Context, username string) (*domain.User, error)
	listFunction           func(ctx context.Context, page int, pageSize int) ([]domain.User, int64, error)
	createFunction         func(ctx context.Context, user *domain.User) error
	updateFunction         func(ctx context.Context, user *domain.User) error
	softDeleteFunction     func(ctx context.Context, id uuid.UUID) error
}

func (mockRepository *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return mockRepository.findByIDFunction(ctx, id)
}

func (mockRepository *mockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	return mockRepository.findByUsernameFunction(ctx, username)
}

func (mockRepository *mockUserRepository) List(ctx context.Context, page int, pageSize int) ([]domain.User, int64, error) {
	return mockRepository.listFunction(ctx, page, pageSize)
}

func (mockRepository *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return mockRepository.createFunction(ctx, user)
}

func (mockRepository *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return mockRepository.updateFunction(ctx, user)
}

func (mockRepository *mockUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return mockRepository.softDeleteFunction(ctx, id)
}

type mockRoleRepository struct {
	findByIDFunction         func(ctx context.Context, id int64) (*domain.Role, error)
	findByNameFunction       func(ctx context.Context, name string) (*domain.Role, error)
	findByUserIDFunction     func(ctx context.Context, userID uuid.UUID) ([]domain.Role, error)
	listFunction             func(ctx context.Context) ([]domain.Role, error)
	createFunction           func(ctx context.Context, role *domain.Role) error
	assignToUserFunction     func(ctx context.Context, userID uuid.UUID, roleID int64) error
	removeFromUserFunction   func(ctx context.Context, userID uuid.UUID, roleID int64) error
}

func (mockRepository *mockRoleRepository) FindByID(ctx context.Context, id int64) (*domain.Role, error) {
	return mockRepository.findByIDFunction(ctx, id)
}

func (mockRepository *mockRoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	return mockRepository.findByNameFunction(ctx, name)
}

func (mockRepository *mockRoleRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) {
	return mockRepository.findByUserIDFunction(ctx, userID)
}

func (mockRepository *mockRoleRepository) List(ctx context.Context) ([]domain.Role, error) {
	return mockRepository.listFunction(ctx)
}

func (mockRepository *mockRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	return mockRepository.createFunction(ctx, role)
}

func (mockRepository *mockRoleRepository) AssignToUser(ctx context.Context, userID uuid.UUID, roleID int64) error {
	return mockRepository.assignToUserFunction(ctx, userID, roleID)
}

func (mockRepository *mockRoleRepository) RemoveFromUser(ctx context.Context, userID uuid.UUID, roleID int64) error {
	return mockRepository.removeFromUserFunction(ctx, userID, roleID)
}

type mockPermissionRepository struct {
	findByUserIDFunction func(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error)
	findByRoleIDFunction func(ctx context.Context, roleID int64) ([]domain.Permission, error)
	listFunction         func(ctx context.Context) ([]domain.Permission, error)
}

func (mockRepository *mockPermissionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error) {
	return mockRepository.findByUserIDFunction(ctx, userID)
}

func (mockRepository *mockPermissionRepository) FindByRoleID(ctx context.Context, roleID int64) ([]domain.Permission, error) {
	return mockRepository.findByRoleIDFunction(ctx, roleID)
}

func (mockRepository *mockPermissionRepository) List(ctx context.Context) ([]domain.Permission, error) {
	return mockRepository.listFunction(ctx)
}

type mockCache struct {
	getFunction            func(ctx context.Context, key string, destination any) error
	setFunction            func(ctx context.Context, key string, value any, ttl time.Duration) error
	deleteFunction         func(ctx context.Context, keys ...string) error
	deleteByPrefixFunction func(ctx context.Context, prefix string) error
}

func (mockCacheClient *mockCache) Get(ctx context.Context, key string, destination any) error {
	return mockCacheClient.getFunction(ctx, key, destination)
}

func (mockCacheClient *mockCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return mockCacheClient.setFunction(ctx, key, value, ttl)
}

func (mockCacheClient *mockCache) Delete(ctx context.Context, keys ...string) error {
	return mockCacheClient.deleteFunction(ctx, keys...)
}

func (mockCacheClient *mockCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	return mockCacheClient.deleteByPrefixFunction(ctx, prefix)
}

type mockSessionStore struct {
	createFunction func(ctx context.Context, metadata session.Metadata) (string, error)
	getFunction    func(ctx context.Context, token string) (*session.Metadata, error)
	deleteFunction func(ctx context.Context, token string) error
	extendFunction func(ctx context.Context, token string) error
}

func (mockStore *mockSessionStore) Create(ctx context.Context, metadata session.Metadata) (string, error) {
	return mockStore.createFunction(ctx, metadata)
}

func (mockStore *mockSessionStore) Get(ctx context.Context, token string) (*session.Metadata, error) {
	return mockStore.getFunction(ctx, token)
}

func (mockStore *mockSessionStore) Delete(ctx context.Context, token string) error {
	return mockStore.deleteFunction(ctx, token)
}

func (mockStore *mockSessionStore) Extend(ctx context.Context, token string) error {
	return mockStore.extendFunction(ctx, token)
}

type mockPasswordHasher struct {
	hashFunction   func(password string) (string, error)
	verifyFunction func(password string, encodedHash string) (bool, error)
}

func (mockHasher *mockPasswordHasher) Hash(password string) (string, error) {
	return mockHasher.hashFunction(password)
}

func (mockHasher *mockPasswordHasher) Verify(password string, encodedHash string) (bool, error) {
	return mockHasher.verifyFunction(password, encodedHash)
}

type mockFileStorage struct {
	uploadFunction    func(ctx context.Context, key string, reader io.Reader, contentType string) error
	openFunction      func(ctx context.Context, key string) (io.ReadCloser, string, error)
	urlFunction       func(key string) string
	signedURLFunction func(ctx context.Context, key string, expiry time.Duration) (string, error)
	deleteFunction    func(ctx context.Context, key string) error
}

func (mockStorage *mockFileStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	return mockStorage.uploadFunction(ctx, key, reader, contentType)
}

func (mockStorage *mockFileStorage) Open(ctx context.Context, key string) (io.ReadCloser, string, error) {
	return mockStorage.openFunction(ctx, key)
}

func (mockStorage *mockFileStorage) URL(key string) string {
	return mockStorage.urlFunction(key)
}

func (mockStorage *mockFileStorage) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return mockStorage.signedURLFunction(ctx, key, expiry)
}

func (mockStorage *mockFileStorage) Delete(ctx context.Context, key string) error {
	return mockStorage.deleteFunction(ctx, key)
}
