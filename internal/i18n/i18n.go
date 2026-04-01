package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

//go:embed locales/*.json
var localeFiles embed.FS

type Translator struct {
	translations map[string]map[string]string
	fallback     string
}

func NewTranslator(fallbackLanguage string) (*Translator, error) {
	translator := &Translator{
		translations: make(map[string]map[string]string),
		fallback:     strings.TrimSpace(fallbackLanguage),
	}

	entries, readDirectoryError := localeFiles.ReadDir("locales")
	if readDirectoryError != nil {
		return nil, fmt.Errorf("i18n: read locales dir: %w", readDirectoryError)
	}

	for _, entry := range entries {
		language := strings.TrimSuffix(entry.Name(), ".json")
		data, readFileError := localeFiles.ReadFile("locales/" + entry.Name())
		if readFileError != nil {
			return nil, fmt.Errorf("i18n: read %s: %w", entry.Name(), readFileError)
		}

		var messages map[string]string
		if unmarshalError := json.Unmarshal(data, &messages); unmarshalError != nil {
			return nil, fmt.Errorf("i18n: parse %s: %w", entry.Name(), unmarshalError)
		}
		translator.translations[language] = messages
	}

	if translator.fallback == "" {
		translator.fallback = "en"
	}

	return translator, nil
}

func (translator *Translator) Translate(language string, key string, params ...map[string]any) string {
	message := translator.resolve(language, key)
	if len(params) == 0 || params[0] == nil {
		return message
	}

	parsedTemplate, parseError := template.New("translation").Parse(message)
	if parseError != nil {
		return message
	}

	var builder strings.Builder
	if executeError := parsedTemplate.Execute(&builder, params[0]); executeError != nil {
		return message
	}

	return builder.String()
}

func (translator *Translator) resolve(language string, key string) string {
	normalizedLanguage := strings.TrimSpace(language)
	if messages, exists := translator.translations[normalizedLanguage]; exists {
		if message, messageExists := messages[key]; messageExists {
			return message
		}
	}

	if messages, exists := translator.translations[translator.fallback]; exists {
		if message, messageExists := messages[key]; messageExists {
			return message
		}
	}

	return key
}
