package v1

import (
	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	"github.com/gofiber/fiber/v3"
)

type Handler struct{}

type PingResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewHandler() *Handler {
	return &Handler{}
}

func (handler *Handler) Ping(ctx fiber.Ctx) error {
	return ctx.JSON(response.Response{
		Data: PingResponse{
			Name:    "fiber-v3-template",
			Version: "v1",
		},
	})
}
