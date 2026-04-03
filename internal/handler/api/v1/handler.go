package v1

import (
	"fmt"

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
// @Success      200  {object}  response.Response[PingResponse]
// @Router       /v1/ping [get]
func (handler *Handler) Ping(ctx fiber.Ctx) error {
	if jsonError := ctx.JSON(response.Response[PingResponse]{
		Data: PingResponse{
			Name:    "fiber-v3-template",
			Version: "v1",
		},
	}); jsonError != nil {
		return fmt.Errorf("handler.Ping: %w", jsonError)
	}
	return nil
}
