package v1

import (
	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	"github.com/eminbekov/fiber-v3-template/internal/repository"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	userRepository repository.UserRepository
}

type PingResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewHandler(userRepository repository.UserRepository) *Handler {
	return &Handler{
		userRepository: userRepository,
	}
}

func (handler *Handler) Ping(ctx fiber.Ctx) error {
	return ctx.JSON(response.Response{
		Data: PingResponse{
			Name:    "fiber-v3-template",
			Version: "v1",
		},
	})
}
