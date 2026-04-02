package web

import (
	"fmt"

	"github.com/eminbekov/fiber-v3-template/internal/i18n"
	"github.com/gofiber/fiber/v3"
)

// WelcomeHandler serves public HTML pages (end-user site).
type WelcomeHandler struct {
	translator *i18n.Translator
}

// NewWelcomeHandler constructs a [WelcomeHandler].
func NewWelcomeHandler(translator *i18n.Translator) *WelcomeHandler {
	return &WelcomeHandler{
		translator: translator,
	}
}

type welcomePageViewData struct {
	Title       string
	Heading     string
	Description string
	T           func(string) string
}

// Index renders the public landing page.
func (handler *WelcomeHandler) Index(ctx fiber.Ctx) error {
	language, _ := ctx.Locals("language").(string)
	translate := func(key string) string {
		return handler.translator.Translate(language, key)
	}

	if renderError := ctx.Render("public/welcome", welcomePageViewData{
		Title:       translate("welcome.title"),
		Heading:     translate("welcome.heading"),
		Description: translate("welcome.description"),
		T:           translate,
	}, "layouts/public"); renderError != nil {
		return fmt.Errorf("welcomeHandler.Index: %w", renderError)
	}

	return nil
}
