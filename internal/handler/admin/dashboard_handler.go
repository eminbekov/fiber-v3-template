package admin

import (
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/i18n"
	"github.com/gofiber/fiber/v3"
)

type DashboardHandler struct {
	translator *i18n.Translator
}

func NewDashboardHandler(translator *i18n.Translator) *DashboardHandler {
	return &DashboardHandler{
		translator: translator,
	}
}

func (handler *DashboardHandler) Index(ctx fiber.Ctx) error {
	type DashboardViewData struct {
		UsersCount   int
		SessionsOpen int
		LastUpdated  time.Time
	}

	language, _ := ctx.Locals("language").(string)
	translate := func(key string) string {
		return handler.translator.Translate(language, key)
	}

	return ctx.Render("admin/dashboard", fiber.Map{
		"Title": translate("dashboard.title"),
		"T":     translate,
		"Stats": DashboardViewData{
			UsersCount:   0,
			SessionsOpen: 0,
			LastUpdated:  time.Now().UTC(),
		},
	}, "layouts/base")
}
