package commands

import (
	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/repository"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	"github.com/redis/go-redis/v9"
)

// Dependencies aggregates shared services for console commands.
type Dependencies struct {
	UserService    *service.UserService
	RoleRepository repository.RoleRepository
	Cache          cache.Cache
	RedisClient    *redis.Client
}
