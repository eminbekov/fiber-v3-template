package generate

import (
	"strings"
	"unicode"
)

func ToSnakeCase(name string) string {
	if name == "" {
		return ""
	}

	var builder strings.Builder
	for index, runeValue := range name {
		if runeValue == '-' || runeValue == ' ' {
			builder.WriteRune('_')
			continue
		}
		if unicode.IsUpper(runeValue) {
			if index > 0 {
				builder.WriteRune('_')
			}
			builder.WriteRune(unicode.ToLower(runeValue))
			continue
		}
		builder.WriteRune(unicode.ToLower(runeValue))
	}
	return strings.Trim(builder.String(), "_")
}

func ToCamelCase(name string) string {
	normalizedName := ToSnakeCase(name)
	parts := strings.Split(strings.ReplaceAll(strings.ReplaceAll(normalizedName, "-", "_"), " ", "_"), "_")
	var builder strings.Builder
	for _, partValue := range parts {
		if partValue == "" {
			continue
		}
		loweredPart := strings.ToLower(partValue)
		builder.WriteString(strings.ToUpper(loweredPart[:1]))
		if len(loweredPart) > 1 {
			builder.WriteString(loweredPart[1:])
		}
	}
	return builder.String()
}

func ToLowerCamelCase(name string) string {
	camelCaseName := ToCamelCase(name)
	if camelCaseName == "" {
		return ""
	}
	return strings.ToLower(camelCaseName[:1]) + camelCaseName[1:]
}

func Pluralize(name string) string {
	if strings.HasSuffix(name, "s") {
		return name
	}
	if strings.HasSuffix(name, "y") && len(name) > 1 {
		return name[:len(name)-1] + "ies"
	}
	return name + "s"
}

func TableName(name string) string {
	return Pluralize(ToSnakeCase(name))
}
