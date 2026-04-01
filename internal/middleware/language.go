package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

func LanguageDetector(supportedLanguages []string, defaultLanguage string) fiber.Handler {
	supportedLanguagesMap := make(map[string]bool, len(supportedLanguages))
	for _, language := range supportedLanguages {
		supportedLanguagesMap[strings.TrimSpace(language)] = true
	}

	return func(ctx fiber.Ctx) error {
		detectedLanguage := parseAcceptLanguage(
			ctx.Get("Accept-Language"),
			supportedLanguagesMap,
			defaultLanguage,
		)
		ctx.Locals("language", detectedLanguage)

		return ctx.Next()
	}
}

func parseAcceptLanguage(header string, supportedLanguagesMap map[string]bool, defaultLanguage string) string {
	if strings.TrimSpace(header) == "" {
		return defaultLanguage
	}

	for _, part := range strings.Split(header, ",") {
		tag := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if tag == "" {
			continue
		}
		if supportedLanguagesMap[tag] {
			return tag
		}

		baseLanguage := strings.SplitN(tag, "-", 2)[0]
		if supportedLanguagesMap[baseLanguage] {
			return baseLanguage
		}
	}

	return defaultLanguage
}
