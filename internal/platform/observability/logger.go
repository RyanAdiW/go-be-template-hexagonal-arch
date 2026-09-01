package observability

import (
	"log/slog"
	"os"
	"strings"
)

func NewLogger(environment, level string) *slog.Logger {
	var parsed slog.Level
	if err := parsed.UnmarshalText([]byte(level)); err != nil {
		parsed = slog.LevelInfo
	}
	options := &slog.HandlerOptions{Level: parsed}
	if strings.EqualFold(environment, "development") {
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, options))
}
