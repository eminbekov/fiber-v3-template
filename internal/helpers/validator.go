package helpers

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	slugPattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	e164Pattern     = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
)

// IsValidUsername reports whether the value contains only alphanumeric characters and underscores.
func IsValidUsername(value string) bool {
	return len(value) >= 3 && len(value) <= 50 && usernamePattern.MatchString(value)
}

// IsValidSlug reports whether the value is a valid URL slug (lowercase alphanumeric with hyphens).
func IsValidSlug(value string) bool {
	return len(value) >= 1 && len(value) <= 200 && slugPattern.MatchString(value)
}

// IsValidE164Phone reports whether the value matches E.164 international phone format.
func IsValidE164Phone(value string) bool {
	return e164Pattern.MatchString(value)
}

// SanitizeString trims whitespace and collapses internal runs of whitespace to a single space.
func SanitizeString(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

// TruncateString shortens a string to at most maxRunes UTF-8 runes, appending a suffix when truncated.
func TruncateString(value string, maxRunes int, suffix string) string {
	if utf8.RuneCountInString(value) <= maxRunes {
		return value
	}

	runes := []rune(value)
	suffixLength := utf8.RuneCountInString(suffix)
	if maxRunes <= suffixLength {
		return string(runes[:maxRunes])
	}

	return string(runes[:maxRunes-suffixLength]) + suffix
}

// IsNotBlank reports whether the string contains at least one non-whitespace character.
func IsNotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}
