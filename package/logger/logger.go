package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Setup configures the process-wide default slog logger.
// Production uses JSON; non-production uses text. Log level comes from the configuration string.
func Setup(logLevel string, environment string) {
	parsedLevel := parseLogLevel(logLevel)
	handlerOptions := &slog.HandlerOptions{
		Level: parsedLevel,
	}

	var handler slog.Handler
	if strings.EqualFold(strings.TrimSpace(environment), "production") {
		handler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	} else {
		handler = slog.NewTextHandler(os.Stdout, handlerOptions)
	}

	slog.SetDefault(slog.New(handler))
}

func parseLogLevel(logLevel string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(logLevel)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
