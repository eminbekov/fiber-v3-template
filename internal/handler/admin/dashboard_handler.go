package admin

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (handler *DashboardHandler) Index(ctx fiber.Ctx) error {
	type DashboardViewData struct {
		UsersCount   int
		SessionsOpen int
		LastUpdated  time.Time
	}

	return ctx.Render("admin/dashboard", fiber.Map{
		"Title": "Dashboard",
		"Stats": DashboardViewData{
			UsersCount:   0,
			SessionsOpen: 0,
			LastUpdated:  time.Now().UTC(),
		},
	}, "layouts/base")
}
