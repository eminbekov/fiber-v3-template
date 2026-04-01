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

// Ping returns service identity and API version.
//
// @Summary      Ping API
// @Description  Returns a simple payload to verify API availability.
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       /v1/ping [get]
func (handler *Handler) Ping(ctx fiber.Ctx) error {
	return ctx.JSON(response.Response{
		Data: PingResponse{
			Name:    "fiber-v3-template",
			Version: "v1",
		},
	})
}
